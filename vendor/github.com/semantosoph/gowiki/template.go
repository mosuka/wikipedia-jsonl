/*
Copyright (C) IBM Corporation 2015, Michele Franceschini <franceschini@us.ibm.com>
Copyright (C) 2021 Sven Windisch <semantosoph@posteo.de>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gowiki

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

type Template struct {
	Typ        string            `json:"type"` //magic,normal,ext,param
	Name       string            `json:"name"`
	Attr       string            `json:"attr"` //text after the ':' in magic templates
	Parameters map[string]string `json:"parameters"`
}

func (a *Article) parseTemplateEtc(l string) []Template {
	return nil
}

type streak struct {
	opening bool
	length  int
	b       int
	e       int
}

type template struct {
	b        int
	e        int
	isparam  bool
	children []*template
	rt       string
	rendered bool
}

type byStart []*template

func (a byStart) Len() int           { return len(a) }
func (a byStart) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byStart) Less(i, j int) bool { return a[i].b < a[j].b }

var templateStreaksRe = regexp.MustCompile(`(?:\{\{+)|(?:\}\}+)`)

func findCurlyStreaks(mw string) [][]int {
	out := [][]int{}
	found := '.'
	beg := 0
	for i, r := range mw {
		switch r {
		case found:
		default:
			if i-beg > 1 && (found == '{' || found == '}') {
				out = append(out, []int{beg, i})
			}
			beg = i
			found = r
		}
	}
	if beg < len(mw)-1 && (found == '{' || found == '}') {
		out = append(out, []int{beg, len(mw)})
	}
	return out
}

func findTemplates(mw string) []*template {
	tsl := findCurlyStreaks(mw)
	streaks := make([]streak, 0, len(tsl))
	for _, pair := range tsl {
		streaks = append(streaks, streak{
			opening: (mw[pair[0]] == '{'),
			length:  pair[1] - pair[0],
			b:       pair[0],
			e:       pair[1],
		})
	}
	tl := make([]*template, 0, 8)
	i := 0
	for i < len(streaks) {
		if !streaks[i].opening && streaks[i].length > 1 { // found a closing set: search for the opening
			found := false
			for j := i - 1; j >= 0; j-- {
				if streaks[j].opening && streaks[j].length > 1 {
					found = true
					n := 2
					isparam := false
					if streaks[i].length > 2 && streaks[j].length > 2 {
						n = 3
						isparam = true
					}
					tl = append(tl, &template{
						isparam: isparam,
						b:       streaks[j].e - n,
						e:       streaks[i].b + n,
					})
					streaks[i].length -= n
					streaks[i].b += n
					streaks[j].length -= n
					streaks[j].e -= n
					break
				}
			}
			if found {
				continue
			}
		}
		i++
	}
	sort.Sort(byStart(tl))
	out := make([]*template, 0, 4)
	cur_end := 0
	for i := range tl {
		tl[i].children = []*template{}
		if tl[i].b >= cur_end {
			cur_end = tl[i].e
			out = append(out, tl[i])
		} else {
			for j := i - 1; j >= 0; j-- {
				if tl[j].e > tl[i].e {
					tl[j].children = append(tl[j].children, tl[i])
					break
				}
			}
		}
	}
	return out
}

func findTemplateParamPos(mw string, t *template) [][]int { //first is position of pipe, second is position of first equal
	out := make([][]int, 0, 1)
	inChildTemplate := false
	inlink := false
	lastopen := false
	lastclosed := false
	for i, rv := range mw[t.b:t.e] {
		inChildTemplate = false
		open := false
		closed := false
		for _, ct := range t.children {
			if i+t.b >= ct.b && i+t.b < ct.e {
				inChildTemplate = true
				break
			}
		}
		if !inChildTemplate {
			switch {
			case rv == '[':
				if lastopen {
					inlink = true
				}
				open = true
			case rv == ']':
				if lastclosed {
					inlink = false
				}
				closed = true
			case rv == '|' && !inlink:
				out = append(out, []int{i + t.b})
			case rv == '=' && len(out) > 0 && len(out[len(out)-1]) == 1 && !inlink:
				out[len(out)-1] = append(out[len(out)-1], i+t.b)
			}
		}
		lastopen = open
		lastclosed = closed
	}
	return out
}

func (a *Article) processTemplates(mws string, tokens map[string]*Token, g PageGetter) (string, map[string]*Token) {

	mlt := findTemplates(mws)

	last := 0
	out := make([]byte, 0, len(mws))
	for i, t := range mlt {
		sb := fmt.Sprintf("\x07tb%05d", i)
		se := fmt.Sprintf("\x07te%05d", i)
		tn, pm := a.renderInnerTemplates(mws, t, nil, g, 0)
		a.addTemplate(tn, pm)
		out = append(out, []byte(mws[last:t.b])...)
		out = append(out, []byte(sb+t.rt+se)...)
		last = t.e
		tokens[sb] = &Token{
			TText: fmt.Sprintf("%d", i),
			TType: "tb",
		}
		tokens[se] = &Token{
			TText: fmt.Sprintf("%d", i),
			TType: "te",
		}
	}
	out = append(out, []byte(mws[last:])...)

	//unstrip here

	return string(out), tokens
}

func (a *Article) addTemplate(tn string, pm map[string]string) {
	outT := Template{Parameters: pm}
	base, attr, typ, _ := detectTemplateType(tn)
	outT.Typ = typ
	outT.Name = base
	outT.Attr = attr
	a.Templates = append(a.Templates, &outT)
	return
}

func (a *Article) renderTemplate(mw string, t *template) string {
	pp := findTemplateParamPos(mw, t)
	n := 2
	if t.isparam {
		n = 3
	}
	var tn string
	if len(pp) > 0 {
		tn = fmt.Sprint(strings.TrimSpace(mw[t.b+n : pp[0][0]]))
	} else {
		tn = fmt.Sprint(strings.TrimSpace(mw[t.b+n : t.e-n]))
	}
	pm := make(map[string]string, len(pp))
	pp = append(pp, []int{t.e - n})
	for i := 0; i < len(pp)-1; i++ {
		var name string
		var param string
		if len(pp[i]) > 1 { //named param
			name = fmt.Sprint(strings.TrimSpace(mw[pp[i][0]+1 : pp[i][1]]))
			param = fmt.Sprint(strings.TrimSpace(mw[pp[i][1]+1 : pp[i+1][0]]))
		} else {
			name = fmt.Sprint(i + 1)
			param = fmt.Sprint(strings.TrimSpace(mw[pp[i][0]+1 : pp[i+1][0]]))
		}
		pm[name] = param
	}

	outT := Template{Parameters: pm}
	base, attr, typ, text := detectTemplateType(tn)
	switch {
	case t.isparam:
		outT.Typ = "param"
		outT.Name = tn
		text = ""
	default:
		outT.Typ = typ
		outT.Name = base
		outT.Attr = attr
	}
	a.Templates = append(a.Templates, &outT)
	return text
}

func detectTemplateType(tn string) (string, string, string, string) {
	index := strings.Index(tn, ":")
	var base string
	var attr string
	if index > 0 {
		base = strings.TrimSpace(tn[:index])
		attr = strings.TrimSpace(tn[index+1:])
	} else {
		base = tn
	}
	_, ok := MagicMap[base]
	if ok {
		return base, attr, "magic", ""
	}

	return tn, "", "normal", ""
}

type TemplateRenderer func(name string, params map[string]string) string

var MagicMap map[string]TemplateRenderer = map[string]TemplateRenderer{
	"DISPLAYTITLE": nil,
}

var NoHashFunctionsMap map[string]bool = map[string]bool{
	"displaytitle":     true,
	"formatdate":       true,
	"int":              true,
	"namespace":        true,
	"pagesinnamespace": true,
	"speciale":         true,
	"special":          true,
	"tag":              true,
	"anchorencode":     true, "basepagenamee": true, "basepagename": true, "canonicalurle": true,
	"canonicalurl": true, "cascadingsources": true, "defaultsort": true, "filepath": true,
	"formatnum": true, "fullpagenamee": true, "fullpagename": true, "fullurle": true,
	"fullurl": true, "gender": true, "grammar": true, "language": true,
	"lcfirst": true, "lc": true, "localurle": true, "localurl": true,
	"namespacee": true, "namespacenumber": true, "nse": true, "ns": true,
	"numberingroup": true, "numberofactiveusers": true, "numberofadmins": true, "numberofarticles": true,
	"numberofedits": true, "numberoffiles": true, "numberofpages": true, "numberofusers": true,
	"numberofviews": true, "padleft": true, "padright": true, "pageid": true,
	"pagenamee": true, "pagename": true, "pagesincategory": true, "pagesize": true,
	"plural": true, "protectionlevel": true, "revisionday2": true, "revisionday": true,
	"revisionid": true, "revisionmonth1": true, "revisionmonth": true, "revisiontimestamp": true,
	"revisionuser": true, "revisionyear": true, "rootpagenamee": true, "rootpagename": true,
	"subjectpagenamee": true, "subjectpagename": true, "subjectspacee": true, "subjectspace": true,
	"subpagenamee": true, "subpagename": true, "talkpagenamee": true, "talkpagename": true,
	"talkspacee": true, "talkspace": true, "ucfirst": true, "uc": true,
	"urlencode": true,
}
var VariablesMap map[string]bool = map[string]bool{
	"articlepath":         true,
	"basepagenamee":       true,
	"basepagename":        true,
	"cascadingsources":    true,
	"contentlanguage":     true,
	"currentday2":         true,
	"currentdayname":      true,
	"currentday":          true,
	"currentdow":          true,
	"currenthour":         true,
	"currentmonth1":       true,
	"currentmonthabbrev":  true,
	"currentmonthnamegen": true,
	"currentmonthname":    true,
	"currentmonth":        true,
	"currenttimestamp":    true,
	"currenttime":         true,
	"currentversion":      true,
	"currentweek":         true,
	"currentyear":         true,
	"directionmark":       true,
	"fullpagenamee":       true,
	"fullpagename":        true,
	"localday2":           true,
	"localdayname":        true,
	"localday":            true,
	"localdow":            true,
	"localhour":           true,
	"localmonth1":         true,
	"localmonthabbrev":    true,
	"localmonthnamegen":   true,
	"localmonthname":      true,
	"localmonth":          true,
	"localtimestamp":      true,
	"localtime":           true,
	"localweek":           true,
	"localyear":           true,
	"namespacee":          true,
	"namespacenumber":     true,
	"namespace":           true,
	"numberofactiveusers": true,
	"numberofadmins":      true,
	"numberofarticles":    true,
	"numberofedits":       true,
	"numberoffiles":       true,
	"numberofpages":       true,
	"numberofusers":       true,
	"numberofviews":       true,
	"pageid":              true,
	"pagenamee":           true,
	"pagename":            true,
	"revisionday2":        true,
	"revisionday":         true,
	"revisionid":          true,
	"revisionmonth1":      true,
	"revisionmonth":       true,
	"revisionsize":        true,
	"revisiontimestamp":   true,
	"revisionuser":        true,
	"revisionyear":        true,
	"rootpagenamee":       true,
	"rootpagename":        true,
	"scriptpath":          true,
	"servername":          true,
	"server":              true,
	"sitename":            true,
	"stylepath":           true,
	"subjectpagenamee":    true,
	"subjectpagename":     true,
	"subjectspacee":       true,
	"subjectspace":        true,
	"subpagenamee":        true,
	"subpagename":         true,
	"talkpagenamee":       true,
	"talkpagename":        true,
	"talkspacee":          true,
	"talkspace":           true,
}

func (a *Article) renderTemplateMagic(name string, params map[string]string) string {
	renderer, ok := MagicMap[name]
	text := ""
	if ok && renderer != nil {
		text = renderer(name, params)
	}

	return text
}

func (a *Article) renderTemplateExt(name string, params map[string]string) string {
	return ""
}

func (a *Article) renderTemplateRecursive(name string, params map[string]string, g PageGetter, depth int) string {
	if depth > 4 {
		return ""
	}
	//name and parameters have already been substituted so they are guaranteed not to contain any template

	//establish the type of template
	switch templateType(name) {
	case "magic":
		return a.renderTemplateMagic(name, params)
	case "ext":
		return a.renderTemplateExt(name, params)
	}
	//case "normal"
	//based on the type of template
	//for the name and each parameter, find templates and substite them in the proper order
	mw, err := g.Get(WikiCanonicalFormNamespace(name, "Template"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Title:", a.Title, " Error retrieving:", name, " ->", err)
		return ""
	}
	return a.TranscludeTemplatesRecursive(mw, params, g, depth)
}

func (a *Article) TranscludeTemplatesRecursive(mw string, params map[string]string, g PageGetter, depth int) string {
	var mws string
	followed := 0
	for {
		if followed > 4 {
			return ""
		}
		//strip nowiki noinclude etc here
		mws := a.stripComments(mw)
		isRedirect, redirect := a.checkRedirect(mws)
		if !isRedirect {
			break
		}
		var err error
		mw, err = g.Get(*redirect)
		if err != nil {
			return ""
		}
		followed++
	}
	mws = a.stripNoinclude(mws)
	mlt := findTemplates(mws)

	last := 0
	out := make([]byte, 0, len(mws))
	for _, t := range mlt {
		a.renderInnerTemplates(mws, t, params, g, depth)
		out = append(out, []byte(mws[last:t.b])...)
		out = append(out, []byte(t.rt)...)
		last = t.e
	}
	out = append(out, []byte(mws[last:])...)

	//unstrip here

	return string(out)
}

var ds []string = []string{"   ", "      ", "         ", "            ", "               ", "                  "}

func (a *Article) renderInnerTemplates(mws string, t *template, params map[string]string, g PageGetter, depth int) (string, map[string]string) {
	// render inner templates first
	for _, it := range t.children {
		if !it.rendered {
			a.renderInnerTemplates(mws, it, params, g, depth)
		}
	}
	pp := findTemplateParamPos(mws, t) //position of the pipes for this template

	n := 2
	if t.isparam {
		n = 3
	}
	pp = append(pp, []int{t.e - n})

	var mw string
	var tb int
	if len(t.children) == 0 {
		mw = mws
		tb = t.b
		//		te = t.e
	} else {
		//substitute the strings and update pp
		tci := 0
		ioff := t.children[tci].b
		tb = 0
		mw = mws[t.b:ioff]
		//		fmt.Println(*t)
		ooff := -t.b
		ppi0 := 0
		ppi1 := 0
		for ppi0 < len(pp) {
			if pp[ppi0][ppi1] <= ioff {
				pp[ppi0][ppi1] += ooff
				ppi1++
				if ppi1 >= len(pp[ppi0]) {
					ppi0++
					ppi1 = 0
				}
			} else {
				mw += t.children[tci].rt
				ooff += len(t.children[tci].rt) - (t.children[tci].e - t.children[tci].b)
				teoff := t.children[tci].e
				tci++
				if tci >= len(t.children) {
					ioff = t.e
				} else {
					ioff = t.children[tci].b
				}
				mw += mws[teoff:ioff]
			}
		}
	}

	var tn string
	if len(pp) > 1 {
		tn = fmt.Sprint(strings.TrimSpace(mw[tb+n : pp[0][0]]))
	} else {
		tn = fmt.Sprint(strings.TrimSpace(mw[tb+n : pp[len(pp)-1][0]]))
	}

	t.rendered = true
	if t.isparam { //it's a parameter substitution
		text, ok := params[tn]
		if ok {
			t.rt = text
			return "", nil
		}
		if len(pp) == 1 { //no default
			t.rt = "{{{" + tn + "}}}"
			return "", nil
		}
		t.rt = mw[pp[0][0]+1 : pp[len(pp)-1][0]]
		return "", nil
	}
	pm := make(map[string]string, len(pp))
	for i := 0; i < len(pp)-1; i++ {
		var name string
		var param string
		if len(pp[i]) > 1 { //named param
			name = fmt.Sprint(strings.TrimSpace(mw[pp[i][0]+1 : pp[i][1]]))
			param = fmt.Sprint(strings.TrimSpace(mw[pp[i][1]+1 : pp[i+1][0]]))
		} else {
			name = fmt.Sprint(i + 1)
			param = fmt.Sprint(strings.TrimSpace(mw[pp[i][0]+1 : pp[i+1][0]]))
		}
		pm[name] = param
	}
	t.rt = a.renderTemplateRecursive(tn, pm, g, depth+1)
	return tn, pm
}

func templateType(tn string) string {
	index := strings.Index(tn, ":")
	tns := strings.TrimSpace(tn)
	var base string
	if index > 0 {
		base = strings.TrimSpace(tn[:index])
	} else {
		base = tns
	}
	base = strings.ToLower(base)
	_, ok1 := NoHashFunctionsMap[base]
	_, ok2 := VariablesMap[base]
	if ok1 || ok2 {
		return "magic"
	}
	if strings.HasPrefix(tns, "#") {
		return "ext"
	}
	return "normal"
}

var noincludeRe = regexp.MustCompile(`(?isU)<noinclude>.*(?:</noinclude>|\z)`)
var includeonlyRe = regexp.MustCompile(`(?isU)<includeonly>(.*)(?:</includeonly>|\z)`)

func (a *Article) stripNoinclude(mw string) string {
	mwni := noincludeRe.ReplaceAllLiteralString(mw, "")
	ssl := includeonlyRe.FindAllStringSubmatch(mwni, -1)
	if len(ssl) == 0 {
		return mwni
	}
	sl := make([]string, 0, len(ssl))
	for _, s := range ssl {
		sl = append(sl, s[1])
	}
	return strings.Join(sl, "")
}

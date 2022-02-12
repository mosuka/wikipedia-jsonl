/*
Copyright (C) IBM Corporation 2015, Michele Franceschini <franceschini@us.ibm.com>

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
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Token struct {
	TText   string   `json:"tText,omitempty"`
	TType   string   `json:"tType,omitempty"`
	TAttr   string   `json:"tAttr,omitempty"`
	TLink   WikiLink `json:"tLink,omitempty"`
	TClosed bool     `json:"tClosed,omitempty"`
	TPipes  []string `json:"tPipes,omitempty"`
}

type Prefixes []string

var ExtLinkPrefixes Prefixes = []string{"https://", "http://", "ftp://", "//"}
var FileLinkPrefixes Prefixes = []string{"[[image:", "[[media:", "[[file:"}

func (a *Article) parseRedirectLine(l string) ([]*Token, error) {
	nt := make([]*Token, 0, 2)
	nt = append(nt, &Token{TType: "redirect"})
	nnt, err := a.parseInlineText(l, 9, len(l))
	if err != nil {
		return nil, err
	}
	nt = append(nt, nnt...)
	return nt, nil
}

func (a *Article) parseWikiPreLine(l string) ([]*Token, error) {
	nt := make([]*Token, 0, 2)
	nt = append(nt, &Token{TType: "wikipre"})
	nnt, err := a.parseInlineText(l, 1, len(l))
	if err != nil {
		return nil, err
	}
	nt = append(nt, nnt...)
	return nt, nil
}

func (a *Article) parseHRuler(l string) ([]*Token, error) {
	pos := 0
	for i, rv := range l {
		if rv != '-' {
			pos = i
			break
		}
	}
	nt := make([]*Token, 0, 2)
	nt = append(nt, &Token{TType: "hrule"})
	if pos != 0 {
		nnt, err := a.parseInlineText(l, pos, len(l))
		if err != nil {
			return nil, err
		}
		nt = append(nt, nnt...)
	}
	return nt, nil
}

func (a *Article) parseHeadingLine(l string) ([]*Token, error) {
	pf := 0
	pl := 0
	for i, rv := range l {
		if rv == '=' {
			pl = i
		}
	}
	for {
		pf++
		if pf == pl || l[pf] != '=' {
			pf--
			break
		}
		pl--
		if pf == pl || l[pl] != '=' {
			pl++
			pf--
			break
		}
	}
	pf++
	if pf > 6 {
		diff := pf - 6
		pf -= diff
		pl += diff
	}
	nt := make([]*Token, 0, 2)
	nt = append(nt, &Token{TType: fmt.Sprintf("h%d", pf)})
	nnt, err := a.parseInlineText(l, pf, pl)
	if err != nil {
		return nil, err
	}
	nt = append(nt, nnt...)
	return nt, nil
}

func (a *Article) parseListLine(l string) ([]*Token, error) {
	nt := make([]*Token, 0, 2)
	pos := 0
	for ; pos < len(l); pos++ {
		switch l[pos] {
		case ';', ':', '*', '#':
			nt = append(nt, &Token{TType: l[pos : pos+1]})
			continue
		}
		break
	}
	if pos < len(l) {
		nnt, err := a.parseInlineText(l, pos, len(l))
		if err != nil {
			return nil, err
		}
		nt = append(nt, nnt...)
	}
	return nt, nil
}

func (a *Article) parseTableLine(l string) ([]*Token, error) {
	nt := make([]*Token, 0, 0)
	return nt, nil
}

func isValidHTMLtag(tag string) bool {
	return true
}

func (a *Article) decodeHTMLtag(l string) (int, string, string, bool, bool) {
	matchingpos := 0
	inquote := false
	lastbackslash := false
	quote := '#'
	closefound := false
	tagend := 0
	tagstart := 0
dhtLoop:
	for idx, rv := range l {
		//		fmt.Println(string(rv), inquote, string(quote), idx, matchingpos)
		switch rv {
		case '>':
			if !inquote {
				matchingpos = idx
				break dhtLoop
			}
		case '\'', '"':
			switch {
			case inquote && quote == rv && !lastbackslash:
				inquote = false
			case !inquote:
				inquote = true
				quote = rv
			}
		case ' ', '\t', '\r':
		case '/':
			closefound = true
		}
		lastbackslash = (rv == '\\')
		if !unicode.IsSpace(rv) && tagstart == 0 {
			tagstart = idx
		}
		if rv != '/' && !unicode.IsSpace(rv) {
			closefound = false
		}
		if unicode.IsSpace(rv) && tagstart != 0 && tagend == 0 {
			tagend = idx
		}
	}
	if matchingpos == 0 || tagstart == 0 {
		return 0, "", "", false, false
	}
	var tag string
	var attr string

	if tagend == 0 {
		tag = l[tagstart:matchingpos]
		attr = ""
	} else {
		tag = l[tagstart:tagend]
		attr = l[tagend:matchingpos]
	}
	return matchingpos + 1, tag, attr, closefound, true
}

func matchPrefixes(s string, prefixes []string) bool {
	for i := range prefixes {
		if len(s) >= len(prefixes[i]) && strings.EqualFold(s[:len(prefixes[i])], prefixes[i]) {
			return true
		}
	}
	return false
}

func isExtLink(l string) bool {
	return matchPrefixes(l, ExtLinkPrefixes)
}

func possibleFileLink(l string) bool {
	return matchPrefixes(l, FileLinkPrefixes)
}

func (a *Article) parseLink(l string) (int, []*Token, bool) {
	if len(l) < 5 {
		return 0, nil, false
	}
	if l[1] == '[' {
		if possibleFileLink(l) {
			return a.parseFileLink(l)
		}
		return a.parseInternalLink(l)
	}
	return a.parseExternalLink(l)
}

func (a *Article) parseInternalLink(l string) (int, []*Token, bool) {

	// possible internal link
	pipepos := 0
	closed := false
	matchingpos := 0
	linktrail := 0
	//plLoop:
	for idx, rv := range l {
		if idx < 2 {
			continue
		}
		if matchingpos == 0 {
			switch rv {
			case '\x07': //prevent special tags in internal link
				if pipepos == 0 { //only in the link portion
					return 0, nil, false
				}
			case '[':
				if idx == 2 || len(l) > idx+1 && l[idx+1] == '[' {
					return 0, nil, false
				}

			case ']':
				if len(l) > idx+1 && l[idx+1] == ']' {
					matchingpos = idx
				}
			case '|':
				if pipepos == 0 {
					pipepos = idx
				}
			default:
			}
			continue
		}
		if !closed {
			closed = true
			continue
		}
		if unicode.IsLetter(rv) {
			linktrail = idx
			continue
		}
		break
	}
	if !closed {
		return 0, nil, false
	}
	var link WikiLink
	var nt []*Token = nil
	var err error = nil
	if pipepos == 0 {
		innerstring := l[2:matchingpos]
		if linktrail != 0 {
			innerstring += l[matchingpos+2 : linktrail+1]
		}
		link = WikiCanonicalForm(l[2:matchingpos])
		nt = []*Token{&Token{TText: innerstring, TType: "text"}}

	} else {
		innerstring := l[pipepos+1 : matchingpos]
		if linktrail != 0 {
			innerstring += l[matchingpos+2 : linktrail+1]
		}
		link = WikiCanonicalForm(l[2:pipepos])
		if pipepos+1 < matchingpos {
			nt, err = a.parseInlineText(innerstring, 0, len(innerstring))
			if err != nil {
				return 0, nil, false
			}
		}
	}
	tokens := make([]*Token, 0, 2)
	tokens = append(tokens, &Token{TLink: link, TType: "link"})
	if nt != nil {
		tokens = append(tokens, nt...)
	}
	tokens = append(tokens, &Token{TType: "closelink"})
	if linktrail != 0 {
		return linktrail + 1, tokens, true
	}
	return matchingpos + 2, tokens, true
}

func (a *Article) parseExternalLink(l string) (int, []*Token, bool) {
	// possible external link
	spacepos := 0
	matchingpos := 0
	endpos := 0
	intLinkOpen := false
	skipNext := false
plLoop2:
	for idx, rv := range l {
		if idx < 1 {
			continue
		}
		if skipNext {
			skipNext = false
			continue
		}
		switch rv {
		case '\x07':
			if spacepos == 0 {
				return 0, nil, false
			}
		case '[':
			if len(l) > idx+1 && l[idx+1] == '[' {
				intLinkOpen = true
			}
		case ' ':
			if spacepos == 0 {
				spacepos = idx
			}
		case '<':
			if spacepos > 0 {
				_, tag, _, _, ok := a.decodeHTMLtag(l[idx:len(l)])
				if ok && tag == "/ref" {
					matchingpos = idx
					endpos = idx
					break plLoop2
				}
			}
		case ']':
			if intLinkOpen && len(l) > idx+1 && l[idx+1] == ']' {
				intLinkOpen = false
				skipNext = true
				continue
			}
			matchingpos = idx
			endpos = idx + 1
			break plLoop2
		}
	}
	if matchingpos == 0 {
		return 0, nil, false
	}
	var link string
	var nt []*Token = nil
	var err error = nil
	if spacepos == 0 {
		link = l[1:matchingpos]
		if !isExtLink(link) {
			return 0, nil, false
		}
	} else {
		link = l[1:spacepos]
		if !isExtLink(link) {
			return 0, nil, false
		}
		if spacepos+1 < matchingpos {
			nt, err = a.parseInlineText(l, spacepos+1, matchingpos)
			if err != nil {
				return 0, nil, false
			}
		}
	}
	tokens := make([]*Token, 0, 2)
	tokens = append(tokens, &Token{TText: link, TType: "extlink"})
	if nt != nil {
		tokens = append(tokens, nt...)
	}
	tokens = append(tokens, &Token{TType: "closeextlink"})
	return endpos, tokens, true
}

func (a *Article) parseFileLink(l string) (int, []*Token, bool) {
	// possible internal link
	pipepos := make([]int, 0, 0)
	closed := false
	matchingpos := 0
	intLinkOpen := false
	skipNext := false
plLoop:
	for idx, rv := range l {
		if idx < 2 {
			continue
		}
		if skipNext {
			skipNext = false
			continue
		}
		switch rv {
		case '\x07': //prevent special tags in internal link
			if len(pipepos) == 0 { //only in the link portion
				return 0, nil, false
			}
		case '[':
			if len(l) > idx+1 && l[idx+1] == '[' {
				intLinkOpen = true
				skipNext = true
				continue
			}

		case ']':
			if len(l) > idx+1 && l[idx+1] == ']' {
				if intLinkOpen {
					intLinkOpen = false
					skipNext = true
					continue
				}
				matchingpos = idx
				closed = true
				break plLoop
			}
		case '|':
			if !intLinkOpen {
				pipepos = append(pipepos, idx)
			}
		default:
		}
	}
	if !closed {
		return 0, nil, false
	}
	var link WikiLink
	var pipes = make([]string, 0, 0)
	var nt []*Token = nil
	var err error = nil
	if len(pipepos) == 0 {
		link = WikiCanonicalForm(l[2:matchingpos])
		nt = []*Token{&Token{TText: l[2:matchingpos], TType: "text"}}
	} else {
		link = WikiCanonicalForm(l[2:pipepos[0]])
		for i := 0; i < len(pipepos)-1; i++ {
			pipes = append(pipes, l[pipepos[i]+1:pipepos[i+1]])
		}
		if pipepos[len(pipepos)-1]+1 < matchingpos {
			nt, err = a.parseInlineText(l, pipepos[len(pipepos)-1]+1, matchingpos)
			if err != nil {
				return 0, nil, false
			}
		}
	}
	tokens := make([]*Token, 0, 2)
	tokens = append(tokens, &Token{TLink: link, TType: "filelink", TPipes: pipes})
	if nt != nil {
		tokens = append(tokens, nt...)
	}
	tokens = append(tokens, &Token{TType: "closefilelink"})
	return matchingpos + 2, tokens, true
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

var behavswitchre = regexp.MustCompile(`^__[A-Z]+__`)

func (a *Article) decodeBehavSwitch(l string) (int, bool) {
	match := behavswitchre.FindString(l)
	if len(match) == 0 {
		return 0, false
	} else {
		return len(match), true
	}
}

func (a *Article) parseInlineText(l string, start, end int) ([]*Token, error) {
	nt := make([]*Token, 0)

	tStart, tEnd := start, start

	for pos := start; pos < end; {
		rv, rune_len := utf8.DecodeRuneInString(l[pos:end])
		switch rv {
		case '<':
			e, tag, attr, closed, ok := a.decodeHTMLtag(l[pos:end])
			if ok {
				pos += e
				if isValidHTMLtag(tag) {
					if tEnd > tStart {
						nt = append(nt, &Token{TText: l[tStart:tEnd], TType: "text"})
					}
					nt = append(nt, &Token{TType: "html", TText: tag, TAttr: attr, TClosed: closed})
					tStart = pos
				}
				tEnd = pos
				continue
			}
		case '[':
			e, lt, ok := a.parseLink(l[pos:end])
			if ok {
				if tEnd > tStart {
					nt = append(nt, &Token{TText: l[tStart:tEnd], TType: "text"})
				}
				nt = append(nt, lt...)
				pos += e
				tStart, tEnd = pos, pos
				continue
			}
		case '_':
			e, ok := a.decodeBehavSwitch(l[pos:end])
			if ok {
				if tEnd > tStart {
					nt = append(nt, &Token{TText: l[tStart:tEnd], TType: "text"})
				}
				nt = append(nt, &Token{TType: "magic", TAttr: l[pos : pos+e]})
				pos += e
				tStart, tEnd = pos, pos
				continue
			}
		case ' ', '\t', '\r':
			if tEnd > tStart {
				nt = append(nt, &Token{TText: l[tStart:tEnd], TType: "text"})
			}
			nt = append(nt, &Token{TType: "space"})
			tStart = pos + rune_len
		case '\'':
			if tEnd > tStart {
				nt = append(nt, &Token{TText: l[tStart:tEnd], TType: "text"})
			}
			nt = append(nt, &Token{TType: "quote"})
			tStart = pos + rune_len
		case ':':
			if tEnd > tStart {
				nt = append(nt, &Token{TText: l[tStart:tEnd], TType: "text"})
			}
			nt = append(nt, &Token{TType: "colon"})
			tStart = pos + rune_len
		case '\x07':
			//		case '@':
			if tEnd > tStart {
				nt = append(nt, &Token{TText: l[tStart:tEnd], TType: "text"})
			}
			nt = append(nt, &Token{TType: "special", TText: l[pos : pos+8]})
			pos += 8
			tStart, tEnd = pos, pos
			continue
		}
		pos += rune_len
		tEnd = pos
	}
	if tEnd > tStart {
		nt = append(nt, &Token{TText: l[tStart:tEnd], TType: "text"})
	}
	return nt, nil
}

func (a *Article) isHeading(l string) bool {
	if l[0] != '=' {
		return false
	}
	done := 0
	lastEqual := false
	for _, rv := range l {
		done++
		if done > 2 {
			if unicode.IsSpace(rv) {
				continue
			}
			if rv == '=' {
				lastEqual = true
				continue
			}
			lastEqual = false
		}

	}
	return lastEqual
}

func (a *Article) isTable(l string) bool {
	return (len(l) > 1 && (l[0:2] == "{|" || l[0:2] == "|}" || l[0:2] == "|+" || l[0:2] == "|-")) || (len(l) > 0 && (l[0:1] == "|" || l[0:1] == "!"))
}

func (a *Article) lineType(l string) string {
	switch {
	case len(l) == 0:
		return "blank"
	case len(l) > 8 && strings.ToLower(l[0:9]) == "#redirect":
		return "redirect"
	case len(l) > 3 && l[0:4] == "----":
		return "hr"
	case a.isHeading(l):
		return "heading"
	case l[0] == ';' || l[0] == ':' || l[0] == '*' || l[0] == '#':
		return "list"
	case a.isTable(l):
		return "table"
	case l[0] == ' ':
		return "wikipre"
	}
	return "normal"
}

func (a *Article) Tokenize(mw string, g PageGetter) ([]*Token, error) {
	mwnc := a.stripComments(mw)
	mwnr := a.stripRefs(mwnc)
	mw_stripped, nowikipremathmap := a.stripNowikiPreMath(mwnr)
	mw_tmpl, templatemap := a.processTemplates(mw_stripped, nowikipremathmap, g)
	mw_links := a.preprocessLinks(mw_tmpl)

	lines := strings.Split(mw_links, "\n")
	tokens := make([]*Token, 0, 16)
	for _, l := range lines {
		var nt []*Token
		var err error = nil
		lt := a.lineType(l)
		switch lt {
		case "normal":
			nt, err = a.parseInlineText(l, 0, len(l))
		case "redirect":
			nt, err = a.parseRedirectLine(l)
		case "hr":
			nt, err = a.parseHRuler(l)
		case "heading":
			nt, err = a.parseHeadingLine(l)
		case "list":
			nt, err = a.parseListLine(l)
		case "table":
			nt, err = a.parseTableLine(l)
		case "wikipre":
			nt, err = a.parseWikiPreLine(l)
		case "blank":
			nt = []*Token{&Token{TType: "blank"}}
		}
		if err != nil {
			return nil, err
		}
		nt = append(nt, &Token{TType: "newline"})
		tokens = append(tokens, nt...)
	}
	specialcount := 0
	for i := range tokens {
		if tokens[i].TType == "special" {
			specialcount++
			t, ok := templatemap[tokens[i].TText]
			if !ok {
				return nil, errors.New("special not in map")
			}
			tokens[i] = t
		}
	}

	if specialcount != len(templatemap) {
		if DebugLevel > 0 {
			fmt.Println("[Tokenize] Warning: number of specials in map differs from number found")
		}
	}
	return tokens, nil
}

var commentsRe = regexp.MustCompile(`(?isU)<!--.*(?:-->|\z)`)

func (a *Article) stripComments(mw string) string {
	return commentsRe.ReplaceAllLiteralString(mw, "")
}

var refRe = regexp.MustCompile(`(?msU)<ref.*</ref>`)

func (a *Article) stripRefs(mw string) string {
	return refRe.ReplaceAllLiteralString(mw, "")
}

var nowikiOpenRe = regexp.MustCompile(`(?i)<\s*(nowiki)\s*[^>/]*>`)
var nowikiCloseRe = regexp.MustCompile(`(?i)<(/nowiki)\s*[^>/]*>`)
var preOpenRe = regexp.MustCompile(`(?i)<\s*(pre)\s*[^>]*>`)
var preCloseRe = regexp.MustCompile(`(?i)<(/pre)\s*[^>]*>`)
var mathOpenRe = regexp.MustCompile(`(?i)<\s*(math)\s*[^>]*>`)
var mathCloseRe = regexp.MustCompile(`(?i)<(/math)\s*[^>]*>`)

type ssInt [][]int

func (a ssInt) Len() int           { return len(a) }
func (a ssInt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ssInt) Less(i, j int) bool { return a[i][0] < a[j][0] }

func (a *Article) stripNowikiPreMath(mw string) (string, map[string]*Token) {
	nwoc := nowikiOpenRe.FindAllStringSubmatchIndex(mw, -1)
	nwcc := nowikiCloseRe.FindAllStringSubmatchIndex(mw, -1)
	poc := preOpenRe.FindAllStringSubmatchIndex(mw, -1)
	pcc := preCloseRe.FindAllStringSubmatchIndex(mw, -1)
	moc := mathOpenRe.FindAllStringSubmatchIndex(mw, -1)
	mcc := mathCloseRe.FindAllStringSubmatchIndex(mw, -1)

	for i := range nwoc {
		nwoc[i] = append(nwoc[i], 0)
	}
	for i := range nwcc {
		nwcc[i] = append(nwcc[i], 1)
	}
	for i := range poc {
		poc[i] = append(poc[i], 2)
	}
	for i := range pcc {
		pcc[i] = append(pcc[i], 3)
	}
	for i := range moc {
		moc[i] = append(moc[i], 4)
	}
	for i := range mcc {
		mcc[i] = append(mcc[i], 5)
	}
	am := make([][]int, 0, len(nwoc)+len(nwcc)+len(poc)+len(pcc)+len(moc)+len(mcc))
	am = append(am, nwoc...)
	am = append(am, nwcc...)
	am = append(am, poc...)
	am = append(am, pcc...)
	am = append(am, moc...)
	am = append(am, mcc...)
	sort.Sort(ssInt(am))

	tokens := make(map[string]*Token, len(am))
	if len(am) == 0 {
		return mw, tokens
	}

	ctype := -1
	out := ""
	lastclose := 0
	openidx := 0
	count := 0
	for i := range am {
		if (ctype != -1) && (am[i][4] == ctype+1) && (am[openidx][1] <= am[i][0]) {
			// closing an open one
			special := fmt.Sprintf("\x07%07d", count)

			tokens[special] = &Token{
				TText: mw[am[openidx][1]:am[i][0]],
				TType: strings.ToLower(mw[am[openidx][2]:am[openidx][3]]),
				TAttr: mw[am[openidx][3] : am[openidx][1]-1],
			}
			out += special
			ctype = -1
			lastclose = am[i][1]
			count++
		} else if (ctype == -1) && (am[i][4]&1 == 0) && (lastclose <= am[i][0]) {
			// open a new one
			out += mw[lastclose:am[i][0]]
			ctype = am[i][4]
			openidx = i
		}
	}
	if ctype != -1 {
		//it's open: close it
		special := fmt.Sprintf("\x07%07d", count)

		tokens[special] = &Token{
			TText: mw[am[openidx][1]:len(mw)],
			TType: strings.ToLower(mw[am[openidx][2]:am[openidx][3]]),
			TAttr: mw[am[openidx][3] : am[openidx][1]-1],
		}
		out += special
		ctype = -1
		count++
	} else {
		out += mw[lastclose:]
	}
	return out, tokens
}

var multiLineLinksRe = regexp.MustCompile(`(?sm)\[\[[^\n|]*\|.*?\]\]`)

/* TODO: add preprocessing as in Parser.php:pstPass2() to enable pipe tricks
 */
func (a *Article) preprocessLinks(s string) string {
	mw := []byte(s)
	mll := multiLineLinksRe.FindAllSubmatchIndex(mw, -1)
	for _, pair := range mll {
		for i := pair[0]; i < pair[1]; {
			// we have to walk this string carefully, by rune, not by i
			rv, rlen := utf8.DecodeRune(mw[i:])
			if rv == '\n' {
				mw[i] = ' '
			}
			i += rlen
		}
	}
	return string(mw)
}

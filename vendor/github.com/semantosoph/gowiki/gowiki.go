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
	"bytes"
	"html"
	"regexp"
	"strings"
)

var DebugLevel int = 0

type Article struct {
	MediaWiki    string
	Title        string
	Links        []WikiLink
	ExtLinks     []string
	Type         string
	AbstractText string
	Media        []WikiLink
	Tokens       []*Token
	Root         *ParseNode
	Parsed       bool
	Text         string
	TextLinks    []FullWikiLink
	Templates    []*Template

	// unexported fields
	gt                   bool
	ga                   bool
	text                 *bytes.Buffer
	nchar                int
	innerParseErrorCount int
}
type WikiLink struct {
	Namespace string
	PageName  string
	Anchor    string
}
type FullWikiLink struct {
	Link  WikiLink
	Text  string
	Start int // rune offset of beginning
	End   int // rune offset of end (index of the char after the last)
}

type PageGetter interface {
	Get(page WikiLink) (string, error)
}

func NewArticle(title, text string) (*Article, error) {
	a := new(Article)
	a.Title = title
	a.MediaWiki = text
	a.Links = make([]WikiLink, 0, 16)
	a.Media = make([]WikiLink, 0, 16)
	a.TextLinks = make([]FullWikiLink, 0, 16)
	a.ExtLinks = make([]string, 0, 16)
	return a, nil
}

func (a *Article) GetText() string {
	if !a.gt {
		a.genText()
	}
	return a.Text
}

func (a *Article) GetAbstract() string {
	if !a.ga {
		a.genAbstract()
	}
	return a.AbstractText
}

func (a *Article) GetLinks() []WikiLink {
	return a.Links
}

func (a *Article) GetExternalLinks() []string {
	return a.ExtLinks
}

func (a *Article) GetMedia() []WikiLink {
	return a.Media
}

func (a *Article) GetTextLinks() []FullWikiLink {
	if !a.gt {
		a.genText()
	}
	return a.TextLinks
}

var canoReSpaces = regexp.MustCompile(`[ _]+`)

func WikiCanonicalFormEsc(l string, unescape bool) WikiLink {
	return StandardNamespaces.WikiCanonicalFormNamespaceEsc(l, "", unescape)
}

func WikiCanonicalForm(l string) WikiLink {
	return StandardNamespaces.WikiCanonicalFormNamespaceEsc(l, "", true)
}

func WikiCanonicalFormNamespace(l string, defaultNamespace string) WikiLink {
	return StandardNamespaces.WikiCanonicalFormNamespaceEsc(l, defaultNamespace, true)
}

func (namespaces Namespaces) WikiCanonicalFormNamespaceEsc(l string, defaultNamespace string, unescape bool) WikiLink {
	hpos := strings.IndexRune(l, '#')
	anchor := ""
	if hpos >= 0 {
		anchor = l[hpos+1:]
		l = l[0:hpos]
	}
	i := strings.Index(l, ":")
	namespace := defaultNamespace
	if i >= 0 {
		cns := strings.TrimSpace(canoReSpaces.ReplaceAllString(l[:i], " "))
		if unescape {
			cns = html.UnescapeString(cns)
		}
		ns, ok := namespaces[strings.ToLower(cns)]
		switch {
		case ok && len(cns) > 0:
			namespace = ns
		case ok:
			namespace = ""
		default:
			i = -1
		}
	}
	article := strings.TrimSpace(canoReSpaces.ReplaceAllString(l[i+1:], " "))
	anchor = canoReSpaces.ReplaceAllString(anchor, " ")
	if unescape {
		article = html.UnescapeString(article)
		anchor = html.UnescapeString(anchor)
	}
	if len(article) > 0 {
		article = strings.ToUpper(article[0:1]) + article[1:]
	}
	return WikiLink{Namespace: namespace, PageName: article, Anchor: anchor}
}

func (wl *WikiLink) FullPagename() string {
	if len(wl.Namespace) == 0 {
		return wl.PageName
	}
	return wl.Namespace + ":" + wl.PageName
}

func (wl *WikiLink) FullPagenameAnchor() string {
	ns := ""
	if len(wl.Namespace) != 0 {
		ns = wl.Namespace + ":"
	}
	an := ""
	if len(wl.Anchor) != 0 {
		an = "#" + wl.Anchor
	}
	return ns + wl.PageName + an
}

func (wl *WikiLink) IsImplicitSelfLink() bool {
	return len(wl.PageName) == 0
}

func (wl *WikiLink) HasAnchor() bool {
	return len(wl.Anchor) != 0
}

func (wl *WikiLink) GetAnchor() string {
	return wl.Anchor
}

type Namespaces map[string]string

var StandardNamespaces Namespaces = map[string]string{
	"media":                  "Media",
	"special":                "Special",
	"talk":                   "Talk",
	"user":                   "User",
	"user talk":              "User talk",
	"wikipedia":              "Wikipedia",
	"wikipedia talk":         "Wikipedia talk",
	"file":                   "File",
	"file talk":              "File talk",
	"mediawiki":              "MediaWiki",
	"mediawiki talk":         "MediaWiki talk",
	"template":               "Template",
	"template talk":          "Template talk",
	"help":                   "Help",
	"help talk":              "Help talk",
	"category":               "Category",
	"category talk":          "Category talk",
	"portal":                 "Portal",
	"portal talk":            "Portal talk",
	"book":                   "Book",
	"book talk":              "Book talk",
	"draft":                  "Draft",
	"draft talk":             "Draft talk",
	"education program":      "Education Program",
	"education program talk": "Education Program talk",
	"timedtext":              "TimedText",
	"timedtext talk":         "TimedText talk",
	"module":                 "Module",
	"module talk":            "Module talk",
	"topic":                  "Topic",
}

type DummyPageGetter struct{}

func (g *DummyPageGetter) Get(wl WikiLink) (string, error) {
	return "", nil
}

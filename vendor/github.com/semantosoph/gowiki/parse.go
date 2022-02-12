/*
Copyright (C) IBM Corporation 2015, Michele Franceschini <franceschini@us.ibm.com>
Copyright (C) 2021, Sven Windisch <semantosoph@posteo.de>

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
	"html"
	"log"
	"strconv"
	"strings"
)

const maxInnerParseErrorCount = 100

type ParseNode struct {
	NType    string
	NSubType string
	Link     WikiLink
	Contents string
	Flags    int
	Nodes    []*ParseNode
}

func (a *Article) PrintParseTree() {
	a.printParseTree(a.Root, 0)
}

func (a *Article) printParseTree(root *ParseNode, depth int) {
	if depth > 20 {
		return
	}
	spaces := "......................................"
	min := len(spaces)
	if depth < len(spaces) {
		min = depth
	}
	if depth < 0 {
		min = 0
	}
	prefix := spaces[0:min]
	for _, n := range root.Nodes {
		fmt.Printf("%s NType: %10s  NSubType: %10s  Contents: %16s  Flags: %d\n", prefix, n.NType, n.NSubType, n.Contents, n.Flags)
		if len(n.Nodes) > 0 {
			a.printParseTree(n, depth+1)
		}
	}
}

const (
	TClosed int = 1 << iota
)

const (
	QS_none int = iota
	QS_i
	QS_b
	QS_ib
	QS_bi
)

func ParseArticle(title, text string, g PageGetter) (*Article, error) {
	a, err := NewArticle(title, text)
	if err != nil {
		return nil, err
	}
	a.Tokens, err = a.Tokenize(a.MediaWiki, g)
	if err != nil {
		return a, err
	}
	err = a.parse()
	if err != nil {
		return a, err
	}
	a.gt = false
	return a, nil
}

func (a *Article) doQuotes() {
	log.SetFlags(log.Lshortfile) // | log.Ldate | log.Ltime)
	state := QS_none
	save := QS_none
	l := 0
	ni := 0
	tn := make([]*Token, 0, len(a.Tokens))
	t := a.Tokens
	for ; ni < len(t); ni++ {
		if t[ni].TType == "quote" {
			l++
		}
		if t[ni].TType != "quote" || ni == len(t)-1 {
			switch {
			case l == 0:
			case l == 1:
				tn = append(tn, &Token{TText: "'", TType: "text"})
			case l == 2:
				switch state {
				case QS_b:
					tn = append(tn, &Token{TType: "html", TText: "i"})
					state = QS_bi
				case QS_i:
					tn = append(tn, &Token{TType: "html", TText: "/i"})
					state = QS_none
				case QS_bi:
					tn = append(tn, &Token{TType: "html", TText: "/i"})
					state = QS_b
				case QS_ib:
					tn = append(tn, &Token{TType: "html", TText: "/b"})
					tn = append(tn, &Token{TType: "html", TText: "/i"})
					tn = append(tn, &Token{TType: "html", TText: "b"})
					state = QS_b
				case QS_none:
					tn = append(tn, &Token{TType: "html", TText: "i"})
					state = QS_i
				}
			case l == 3, l == 4:
				if l == 4 {
					tn = append(tn, &Token{TText: "'", TType: "text"})
				}
				switch state {
				case QS_b:
					tn = append(tn, &Token{TType: "html", TText: "/b"})
					state = QS_none
				case QS_i:
					tn = append(tn, &Token{TType: "html", TText: "b"})
					state = QS_ib
				case QS_ib:
					tn = append(tn, &Token{TType: "html", TText: "/b"})
					state = QS_i
				case QS_bi:
					tn = append(tn, &Token{TType: "html", TText: "/i"})
					tn = append(tn, &Token{TType: "html", TText: "/b"})
					tn = append(tn, &Token{TType: "html", TText: "i"})
					state = QS_i
				case QS_none:
					tn = append(tn, &Token{TType: "html", TText: "b"})
					state = QS_b
				}
			case l >= 5:
				s := ""
				for i := 5; i < l; i++ {
					s += "'"
				}
				if len(s) > 0 {
					tn = append(tn, &Token{TText: s, TType: "text"})
				}
				switch state {
				case QS_b:
					tn = append(tn, &Token{TType: "html", TText: "/b"})
					tn = append(tn, &Token{TType: "html", TText: "i"})
					state = QS_i
				case QS_i:
					tn = append(tn, &Token{TType: "html", TText: "/i"})
					tn = append(tn, &Token{TType: "html", TText: "b"})
					state = QS_b
				case QS_ib:
					tn = append(tn, &Token{TType: "html", TText: "/b"})
					tn = append(tn, &Token{TType: "html", TText: "/i"})
					state = QS_none
				case QS_bi:
					tn = append(tn, &Token{TType: "html", TText: "/i"})
					tn = append(tn, &Token{TType: "html", TText: "/b"})
					state = QS_none
				case QS_none:
					tn = append(tn, &Token{TType: "html", TText: "b"})
					tn = append(tn, &Token{TType: "html", TText: "i"})
					state = QS_bi
				}
			}
			l = 0
		}

		if t[ni].TType == "link" || t[ni].TType == "extlink" || t[ni].TType == "filelink" {
			save = state
			switch state {
			case QS_b:
				tn = append(tn, &Token{TType: "html", TText: "/b"})
			case QS_i:
				tn = append(tn, &Token{TType: "html", TText: "/i"})
			case QS_ib:
				tn = append(tn, &Token{TType: "html", TText: "/b"})
				tn = append(tn, &Token{TType: "html", TText: "/i"})
			case QS_bi:
				tn = append(tn, &Token{TType: "html", TText: "/i"})
				tn = append(tn, &Token{TType: "html", TText: "/b"})
			}
			state = QS_none
			l = 0
		}
		if t[ni].TType == "closelink" || t[ni].TType == "closeextlink" || t[ni].TType == "closefilelink" {
			switch state {
			case QS_b:
				tn = append(tn, &Token{TType: "html", TText: "/b"})
			case QS_i:
				tn = append(tn, &Token{TType: "html", TText: "/i"})
			case QS_ib:
				tn = append(tn, &Token{TType: "html", TText: "/b"})
				tn = append(tn, &Token{TType: "html", TText: "/i"})
			case QS_bi:
				tn = append(tn, &Token{TType: "html", TText: "/i"})
				tn = append(tn, &Token{TType: "html", TText: "/b"})
			}
			state = save
			save = QS_none
			l = 0
		}

		if t[ni].TType != "quote" && t[ni].TType != "newline" {
			tn = append(tn, t[ni])
		}
		if t[ni].TType == "newline" || ni == len(t)-1 {
			switch state {
			case QS_b:
				tn = append(tn, &Token{TType: "html", TText: "/b"})
			case QS_i:
				tn = append(tn, &Token{TType: "html", TText: "/i"})
			case QS_ib:
				tn = append(tn, &Token{TType: "html", TText: "/b"})
				tn = append(tn, &Token{TType: "html", TText: "/i"})
			case QS_bi:
				tn = append(tn, &Token{TType: "html", TText: "/i"})
				tn = append(tn, &Token{TType: "html", TText: "/b"})
			}
			state = QS_none
			l = 0
			save = QS_none
		}
		if t[ni].TType == "newline" {
			tn = append(tn, t[ni])
		}

	}
	a.Tokens = tn
}

//nowiki, wikipre, pre, math, quote, colon, magic, h?, *, #, ;, :, html,
func (a *Article) parse() error {
	a.doQuotes()
	nodes, err := a.internalParse(a.Tokens)
	if err != nil {
		return err
	}
	root := &ParseNode{NType: "root", Nodes: nodes}
	a.Root = root
	a.Parsed = true
	return nil
}
func isImage(t *Token) bool {
	return strings.ToLower(t.TLink.Namespace) == "file"
}

func (a *Article) internalParse(t []*Token) ([]*ParseNode, error) {
	ti := 0
	nl := make([]*ParseNode, 0, 0)
	lastti := -1
	for ti < len(t) {
		if ti == lastti {
			return nil, errors.New("parsing issue")
		}
		lastti = ti
		switch t[ti].TType {
		case "nowiki":
			n := &ParseNode{NType: "text", NSubType: "nowiki", Contents: html.UnescapeString(t[ti].TText)}
			nl = append(nl, n)
			ti++
			/*		case "curlyblock":
					n := &ParseNode{NType: "curly", Contents: t[ti].TText}
					nl = append(nl, n)
					ti++ */
		case "text":
			n := &ParseNode{NType: "text", Contents: html.UnescapeString(t[ti].TText)}
			nl = append(nl, n)
			ti++
		case "math":
			n := &ParseNode{NType: "math", Contents: t[ti].TText}
			nl = append(nl, n)
			ti++
		case "pre":
			n2 := &ParseNode{NType: "text", NSubType: "pre", Contents: html.UnescapeString(t[ti].TText)}
			n1 := &ParseNode{NType: "html", NSubType: "pre", Contents: t[ti].TAttr, Nodes: []*ParseNode{n2}}
			nl = append(nl, n1)
			ti++
		case "nop":
			ti++
		case "wikipre":
			closebefore := len(t)
			ni := ti + 1
			for ; ni < len(t)-1; ni++ {
				if t[ni].TType == "newline" {
					if t[ni+1].TType == "wikipre" {
						t[ni+1].TType = "nop"
					} else {
						closebefore = ni
						break
					}
				}
			}
			if closebefore <= ni+1 {
				n := &ParseNode{NType: "html", NSubType: "pre"}
				nl = append(nl, n)
				ti++
			} else {
				nodes, err := a.internalParse(t[ti+1 : closebefore])
				if err != nil {
					return nil, err
				}
				n := &ParseNode{NType: "html", NSubType: "pre", Nodes: nodes}
				nl = append(nl, n)
				ti = closebefore
			}
		case "extlink":
			ni := ti + 1
			for ; ni < len(t); ni++ {
				if t[ni].TType == "closeextlink" {
					break
				}
			}
			if ni == len(t) {
				return nil, errors.New("Unmatched external link token for link: " + t[ti].TText)
			}
			n := &ParseNode{NType: "extlink", NSubType: "", Contents: t[ti].TText}
			a.ExtLinks = append(a.ExtLinks, t[ti].TText)
			if ni > ti+1 {
				nodes, err := a.internalParse(t[ti+1 : ni])
				if err != nil {
					return nil, err
				}
				n.Nodes = nodes
			}
			nl = append(nl, n)
			ti = ni + 1

		case "closeextlink":
			return nil, errors.New("Unmatched close external link token")
		case "hrule":
			n := &ParseNode{NType: "html", NSubType: "hr"}
			nl = append(nl, n)
			ti++
		case "magic":
			n := &ParseNode{NType: "magic", Contents: t[ti].TText}
			nl = append(nl, n)
			ti++
		case "colon":
			n := &ParseNode{NType: "text", Contents: ":"}
			nl = append(nl, n)
			ti++
		case "space":
			n := &ParseNode{NType: "space", Contents: " "}
			nl = append(nl, n)
			ti++
		case "blank":
			n := &ParseNode{NType: "break"}
			nl = append(nl, n)
			ti++
		case "redirect":
			ni := ti + 1
			for ; ni < len(t); ni++ {
				if t[ni].TType == "newline" {
					break
				}
				if t[ni].TType == "link" {
					break
				}
			}
			if ni == len(t) || t[ni].TType == "newline" {
				n := &ParseNode{NType: "text", Contents: html.UnescapeString(t[ti].TText)}
				nl = append(nl, n)
				ti++
			} else {
				n := &ParseNode{NType: "redirect", Link: t[ni].TLink, NSubType: t[ni].TAttr}
				nl = append(nl, n)
				ti++
			}
		case "link":
			ni := ti + 1
			nopen := 1
			for ; ni < len(t); ni++ {
				switch t[ni].TType {
				case "link":
					nopen++
				case "closelink":
					nopen--
				}
				if nopen == 0 {
					break
				}
			}
			if ni == len(t) {
				return nil, errors.New("Unmatched link token for link: " + t[ti].TLink.PageName + " namespace: " + t[ti].TLink.Namespace)
			}
			var n *ParseNode
			n = &ParseNode{NType: "link", Link: t[ti].TLink}
			a.Links = append(a.Links, t[ti].TLink)
			if ni > ti+1 {
				nodes, err := a.internalParse(t[ti+1 : ni])
				if err != nil {
					return nil, err
				}
				n.Nodes = nodes
			}
			nl = append(nl, n)
			ti = ni + 1
		case "filelink":
			ni := ti + 1
			nopen := 1
			for ; ni < len(t); ni++ {
				switch t[ni].TType {
				case "filelink":
					nopen++
				case "closefilelink":
					nopen--
				}
				if nopen == 0 {
					break
				}
			}
			if ni == len(t) {
				return nil, errors.New("Unmatched filelink token for filelink: " + t[ti].TLink.PageName + " namespace: " + t[ti].TLink.Namespace)
			}
			var n *ParseNode
			n = &ParseNode{NType: "image", Link: t[ti].TLink}
			a.Media = append(a.Media, t[ti].TLink)
			if ni > ti+1 {
				nodes, err := a.internalParse(t[ti+1 : ni])
				if err != nil {
					return nil, err
				}
				n.Nodes = nodes
			}
			nl = append(nl, n)
			ti = ni + 1

		case "closelink":
			return nil, errors.New("Unmatched close link token")
		case "closefilelink":
			return nil, errors.New("Unmatched close file link token")
		case "html":
			tag := strings.ToLower(t[ti].TText)
			if tag[0] == '/' {
				ti++
				continue
			}
			n := &ParseNode{NType: "html", NSubType: tag, Contents: t[ti].TAttr}
			if t[ti].TClosed == true {
				flags := TClosed
				n.Flags = flags
				nl = append(nl, n)
				ti++
				continue
			}
			ni := ti + 1
			nopen := 1
			for ; ni < len(t); ni++ {
				if t[ni].TType == "html" {
					ntag := strings.ToLower(t[ni].TText)
					switch ntag {
					case tag:
						nopen++
					case "/" + tag:
						nopen--
					}
					if nopen == 0 {
						break
					}
				}
			}
			if ni > ti+1 {
				nodes, err := a.internalParse(t[ti+1 : ni])
				if err != nil {
					a.innerParseErrorCount++
					if a.innerParseErrorCount >= maxInnerParseErrorCount {
						return nil, err
					}
					ti++
					continue
				}
				n.Nodes = nodes
			}
			nl = append(nl, n)
			ti = ni + 1
			if ti > len(t) {
				ti = len(t)
			}
		case "*", "#", ";", ":":
			ti += 1
		case "newline":
			n := &ParseNode{NType: "text", Contents: "\n"}
			nl = append(nl, n)
			ti++
		case "h1", "h2", "h3", "h4", "h5", "h6":
			ni := ti + 1
			for ; ni < len(t); ni++ {
				if t[ni].TType == "newline" {
					break
				}
			}
			if ni == len(t) {
				return nil, errors.New("No newline after heading")
			}
			n := &ParseNode{NType: "html", NSubType: t[ti].TType}
			if ni > ti+1 {
				nodes, err := a.internalParse(t[ti+1 : ni])
				if err != nil {
					return nil, err
				}
				n.Nodes = nodes
			}
			nl = append(nl, n)
			ti = ni + 1
		case "tb", "te":
			templateIndex, err := strconv.Atoi(t[ti].TText)
			if err != nil {
				return nil, errors.New("Malformed tb token")
			}
			if templateIndex >= len(a.Templates) {
				return nil, errors.New("Template index out of range")
			} else {
				n := &ParseNode{NType: t[ti].TType, Contents: a.Templates[templateIndex].Name}
				nl = append(nl, n)
			}
			ti++

		default:
			return nil, errors.New("Unrecognized token type: " + t[ti].TType)
		}
	}
	return nl, nil
}

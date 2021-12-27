package util

import (
	"strings"

	"github.com/semantosoph/gowiki"
)

func ParseArticle(title string, text string) string {
	article, err := gowiki.ParseArticle(title, text, &gowiki.DummyPageGetter{})
	if err != nil {
		return ""
	}

	articleText := article.GetText()
	return strings.TrimSpace(articleText)
}

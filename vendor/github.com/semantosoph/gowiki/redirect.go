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

import "strings"

func (a *Article) checkRedirect(mw string) (bool, *WikiLink) {
	if len(mw) < 9 || strings.ToLower(mw[0:9]) != "#redirect" {
		return false, nil
	}
	idx := strings.Index(mw, "\n")
	if idx < 0 {
		idx = len(mw)
	}
	nnt, err := a.parseInlineText(mw, 9, idx)
	if err != nil {
		return false, nil
	}
	for _, t := range nnt {
		if t.TType == "link" {
			return true, &t.TLink
		}
	}
	return false, nil
}

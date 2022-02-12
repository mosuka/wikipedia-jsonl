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
	"strings"
)

func (a *Article) CheckRedirect() (bool, *WikiLink) {

	rf := false
	for i, t := range a.Tokens {
		if i > 10 {
			break
		}
		switch t.TType {
		case "redirect":
			rf = true
		case "link":
			if rf {
				return true, &t.TLink
			}
		}
	}
	return false, nil
}

func (a *Article) CheckDisambiguation() bool {
	for _, t := range a.Templates {
		if t.Typ != "normal" {
			continue
		}
		ln := strings.ToLower(t.Name)
		if strings.Contains(ln, "disambig") ||
			ln == "dab" ||
			ln == "geodis" ||
			ln == "hndis" ||
			ln == "hndis-cleanup" ||
			ln == "numberdis" {
			return true
		}
	}
	return false
}

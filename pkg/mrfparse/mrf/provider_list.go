/*
Copyright © 2023 Daniel Chalef

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
package mrf

import (
	mapset "github.com/deckarep/golang-set/v2"
)

var providersFilter = NewProviderList()

type ProviderList struct {
	Providers StringSet
}

// NewProviderList returns a new ProviderList containing a StringSet
func NewProviderList() *ProviderList {
	return &ProviderList{
		Providers: mapset.NewSet[string](),
	}
}

// Add adds a string (or many strings) to the ProviderList
// Returns true if all values were added, false otherwise
func (p *ProviderList) Add(vals ...string) bool {
	var added = true

	for _, val := range vals {
		r := p.Providers.Add(val)
		added = added && r
	}

	return added
}

// Slice returns a slice of strings from the ProviderList
func (p *ProviderList) Slice() []string {
	return p.Providers.ToSlice()
}

// Len returns the number of elements in the ProviderList
func (p *ProviderList) Len() int {
	return p.Providers.Cardinality()
}

// Contains returns true if the ProviderList contains a string
func (p *ProviderList) Contains(s string) bool {
	return p.Providers.Contains(s)
}

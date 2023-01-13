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
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

func TestProviderList(t *testing.T) {
	pl := NewProviderList()

	assert.Equal(t, 0, pl.Len())
	assert.Equal(t, false, pl.Contains("a"))

	pl.Add("a", "b", "c")

	assert.Equal(t, 3, pl.Len())
	assert.Equal(t, true, pl.Contains("a"))
	assert.Equal(t, true, pl.Contains("b"))
	assert.Equal(t, true, pl.Contains("c"))
	assert.Equal(t, false, pl.Contains("d"))

	pl.Add("a", "b", "c")

	assert.Equal(t, 3, pl.Len())
	assert.Equal(t, true, pl.Contains("a"))
	assert.Equal(t, true, pl.Contains("b"))
	assert.Equal(t, true, pl.Contains("c"))
	assert.Equal(t, false, pl.Contains("d"))

	pl.Add("d")

	assert.Equal(t, 4, pl.Len())
	assert.Equal(t, true, pl.Contains("a"))
	assert.Equal(t, true, pl.Contains("b"))
	assert.Equal(t, true, pl.Contains("c"))
	assert.Equal(t, true, pl.Contains("d"))

}

func TestProviderListReturn(t *testing.T) {
	pl := NewProviderList()

	pl.Add("a", "b", "c")

	r := pl.Add("a", "b", "c", "d")
	assert.Equal(t, false, r)

	r = pl.Add("e")
	assert.Equal(t, true, r)
}

func TestProviderListSlice(t *testing.T) {
	pl := NewProviderList()

	pl.Add("a", "b", "c")

	s := pl.Slice()

	assert.Equal(t, 3, len(s))
	assert.Equal(t, true, slices.Contains(s, "a"))
	assert.Equal(t, true, slices.Contains(s, "b"))
	assert.Equal(t, true, slices.Contains(s, "c"))
	assert.Equal(t, false, slices.Contains(s, "d"))
}

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
package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test Filter function

func TestFilterInt(t *testing.T) {
	var slice = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	var even = Filter(slice, func(i int) bool {
		return i%2 == 0
	})

	assert.Equal(t, []int{2, 4, 6, 8, 10}, even)
}

func TestFilterString(t *testing.T) {
	var slice = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}

	var eplus = Filter(slice, func(s string) bool {
		return s > "d"
	})

	assert.Equal(t, []string{"e", "f", "g", "h", "i", "j"}, eplus)
}

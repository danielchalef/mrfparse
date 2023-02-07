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

// test Sha256Sum
func TestSha256Sum(t *testing.T) {
	s := "filename_test.gz"
	hash_expected := "cc13984a42a92b86c46c861655e91bda947325361fe6427a611be61053366877"

	hash := Sha256Sum(s)
	assert.Equal(t, hash_expected, hash)
}

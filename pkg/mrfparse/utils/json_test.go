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

func TestGetElementValue(t *testing.T) {
	js := []byte(`{"string_val": "12345", "int_val": 12345, "float_val": 0.12345}`)

	jp, err := ParseJSON(&js, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	// test string
	var retString string
	retString, err = GetElementValue[string]("string_val", &iter)
	assert.NoError(t, err)
	assert.Equal(t, "12345", retString)

	// test int64
	var retInt int64
	retInt, err = GetElementValue[int64]("int_val", &iter)
	assert.NoError(t, err)
	assert.Equal(t, int64(12345), retInt)

	// test float64
	var retFloat float64
	retFloat, err = GetElementValue[float64]("float_val", &iter)
	assert.NoError(t, err)
	assert.Equal(t, 0.12345, retFloat)

	// Test string with float value (we see provider_group_id as floats in MRFs, but they should be strings)
	retString, err = GetElementValue[string]("float_val", &iter)
	assert.NoError(t, err)
	assert.Equal(t, "0.12345", retString)
}

// test GetArrayElementAsSlice
func TestGetArrayElementAsSlice(t *testing.T) {
	js := []byte(`{"string_vals": ["12345", "54321"], "int_vals": [12345, 54321], "float_vals": [0.12345, 0.54321]}`)
	jp, err := ParseJSON(&js, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	// test string
	var retString []string
	retString, err = GetArrayElementAsSlice[string]("string_vals", &iter)
	assert.NoError(t, err)
	assert.Equal(t, []string{"12345", "54321"}, retString)

	// test int64
	var retInt []int64
	retInt, err = GetArrayElementAsSlice[int64]("int_vals", &iter)
	assert.NoError(t, err)
	assert.Equal(t, []int64{12345, 54321}, retInt)

	// test float64
	var retFloat []float64
	retFloat, err = GetArrayElementAsSlice[float64]("float_vals", &iter)
	assert.NoError(t, err)
	assert.Equal(t, []float64{0.12345, 0.54321}, retFloat)
}

func TestTestElementNotPresent(t *testing.T) {
	js := []byte(`{"string_val": "12345", "int_val": 12345, "float_vals": [0.12345, 0.54321]}`)

	jp, err := ParseJSON(&js, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	path := "float_vals"
	_, err = iter.FindElement(nil, path)
	r := TestElementNotPresent(err, path)
	assert.False(t, r)

	path = "int_val"
	_, err = iter.FindElement(nil, path)
	r = TestElementNotPresent(err, path)
	assert.False(t, r)

	path = "missing_val"
	_, err = iter.FindElement(nil, path)
	r = TestElementNotPresent(err, path)
	assert.True(t, r)
}

// test GetArrayIterForElement
func TestGetArrayIterForElement(t *testing.T) {
	js := []byte(`{"string_vals": ["abc", "def"], "int_vals": [12345, 54321], "float_vals": [0.12345, 0.54321]}`)
	path := "string_vals"
	jp, err := ParseJSON(&js, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	// test string
	a, err := GetArrayForElement(path, &iter)
	assert.NoError(t, err)

	ret, err := a.AsString()
	assert.NoError(t, err)

	expectedRet := []string{"abc", "def"}

	assert.Equal(t, expectedRet, ret)

}

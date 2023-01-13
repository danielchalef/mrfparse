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
	"errors"
	"fmt"

	"github.com/minio/simdjson-go"
)

// GetElementValue extracts the value at the given path in the Iter. Supports string, int64, float64 types.
func GetElementValue[T string | int64 | float64](path string, iter *simdjson.Iter) (T, error) {
	var ret T

	e, err := iter.FindElement(nil, path)
	if err != nil {
		return ret, err
	}

	switch p := any(&ret).(type) {
	case *string:
		*p, err = e.Iter.StringCvt()

	case *int64:
		*p, err = e.Iter.Int()

	case *float64:
		*p, err = e.Iter.Float()
	}

	return ret, err
}

// GetArrayElementAsSlice extracts the array at the given path in the Iter, returning the array as a slice.
// Supports string, int64, float64 slices.
func GetArrayElementAsSlice[T string | int64 | float64](path string, iter *simdjson.Iter) ([]T, error) {
	var ret []T

	a, err := GetArrayForElement(path, iter)
	if err != nil {
		return ret, err
	}

	switch p := any(&ret).(type) {
	case *[]string:
		*p, err = a.AsStringCvt()

	case *[]int64:
		*p, err = a.AsInteger()

	case *[]float64:
		*p, err = a.AsFloat()
	}

	return ret, err
}

// GetArrayForElement extracts the array at the given path in the Iter, returning the simdjson.Array
func GetArrayForElement(path string, iter *simdjson.Iter) (*simdjson.Array, error) {
	e, err := iter.FindElement(nil, path)
	if err != nil {
		return nil, err
	}

	a, err := e.Iter.Array(nil)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// TestElementNotPresent evaluates simdjson error for ErrPathNotFound and exits if the error is something else.
// Otherwise, returns true if the element is missing.
func TestElementNotPresent(err error, path string) bool {
	if err != nil {
		if errors.Is(err, simdjson.ErrPathNotFound) {
			log.Tracef("Element not found: %s", path)
			return true
		}

		ExitOnError(err)
	}

	return false
}

// CheckCPU checks that we're running on a CPU that supports the required SIMD instructions
func CheckCPU() {
	if !simdjson.SupportedCPU() {
		ExitOnError(fmt.Errorf("unsupported cpu"))
	}
}

// ParseJSON parses []byte as Json document, while string is assumed to be NDJson
func ParseJSON[T *[]byte | *string](s T, r *simdjson.ParsedJson) (*simdjson.ParsedJson, error) {
	CheckCPU()

	switch p := any(s).(type) {
	case *[]byte:
		return simdjson.Parse(*p, r)
	case *string:
		return simdjson.ParseND([]byte(*p), r)
	default:
		return nil, fmt.Errorf("invalid type")
	}
}

//go:build !amd64 || !(linux || darwin)

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
	"github.com/kiwicom/fakesimdjson"
	"github.com/minio/simdjson-go"
)

func simdParse(b []byte, _ *simdjson.ParsedJson) (*simdjson.ParsedJson, error) {
	return fakesimdjson.Parse(b)
}

func simdParseND(b []byte, _ *simdjson.ParsedJson) (*simdjson.ParsedJson, error) {
	return fakesimdjson.ParseND(b)
}

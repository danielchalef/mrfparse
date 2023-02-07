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
	"crypto/sha256"
	"encoding/hex"

	"github.com/rs/xid"
)

// GetUniqueID generates an xid, a fast, sortable globally unique id that is only 20 characters long.
func GetUniqueID() string {
	guid := xid.New()

	return guid.String()
}

// Generate sha256sum for a string. Not intended to be cryptographically secure.
func Sha256Sum(s string) string {
	h := sha256.New()
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}

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
package cloud

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsCloudURITrue(t *testing.T) {
	var uri = "gs://bucket/path/to/file.json"
	var expected = true
	var actual = IsCloudURI(uri)

	require.Equal(t, expected, actual)
}

func TestIsCloudURIFalse(t *testing.T) {
	var uri = "/path/to/file.json"
	var expected = false
	var actual = IsCloudURI(uri)

	require.Equal(t, expected, actual)
}

func TestParseBlobUriCloud(t *testing.T) {
	var uri = "gs://bucket/path/to/file.json"
	var exoectedScheme = "gs"
	var expectedBucket = "bucket"
	var expectedKey = "path/to/file.json"
	var actualScheme, actualBucket, actualKey, err = ParseBlobURI(uri)

	require.Nil(t, err)
	require.Equal(t, exoectedScheme, actualScheme)
	require.Equal(t, expectedBucket, actualBucket)
	require.Equal(t, expectedKey, actualKey)
}

func TestParseBlobUriFS(t *testing.T) {
	var uri = "/path/to/file.json"
	var exoectedScheme = ""
	var expectedBucket = ""
	var expectedKey = "path/to/file.json"
	var actualScheme, actualBucket, actualKey, err = ParseBlobURI(uri)

	require.Nil(t, err)
	require.Equal(t, exoectedScheme, actualScheme)
	require.Equal(t, expectedBucket, actualBucket)
	require.Equal(t, expectedKey, actualKey)
}

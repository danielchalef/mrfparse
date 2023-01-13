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
)

// test findRootFile function
func TestFindRootFile(t *testing.T) {
	var filesList = []string{"s3://some_bucket/some/path/in_network_001.json", "s3://some_bucket/some/path/root.json",
		"s3://some_bucket/some/path/provider_references001.json", "s3://some_bucket/some/path/"}

	filename, err := findRootFile(filesList)
	assert.NoError(t, err)
	assert.Equal(t, "s3://some_bucket/some/path/root.json", filename)
}

// test parseMrfRoot function
func TestParseMrfRoot(t *testing.T) {
	doc := []byte(`{"reporting_entity_name":"Aetna Health Insurance Company",
        "reporting_entity_type":"Health Insurance Issuer",
        "last_updated_on":"2022-11-05",
		"plan_market_type":"group",
		"plan_id":"1234-5678-9012-3456",
		"plan_id_type":"planidtype",
        "version":"1.3.1"}`)

	mrf, err := parseMrfRoot(doc, -1)
	assert.NoError(t, err)
	assert.Equal(t, "Aetna Health Insurance Company", mrf.ReportingEntityName)
	assert.Equal(t, "Health Insurance Issuer", mrf.ReportingEntityType)
	assert.Equal(t, "2022-11-05", mrf.LastUpdatedOn)
	assert.Equal(t, "1.3.1", mrf.Version)
	assert.Equal(t, "group", mrf.PlanMarketType)
	assert.Equal(t, "1234-5678-9012-3456", mrf.PlanID)
	assert.Equal(t, "planidtype", mrf.PlanIDType)
	assert.Equal(t, "root", mrf.RecordType)
}

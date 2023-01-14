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
	"github.com/danielchalef/mrfparse/pkg/mrfparse/models"
	"github.com/danielchalef/mrfparse/pkg/mrfparse/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test parsePRObject function
func TestParsePRObject(t *testing.T) {
	var j = []byte(`{
		"provider_group_id": 62.0003430048,
		"provider_groups": [
		  { "npi": [1821198789], "tin": { "type": "ein", "value": "1821198789" } },
		  { "npi": [1770512915], "tin": { "type": "npi", "value": "1770512915" } }
		]
	}`)

	var providerList = NewProviderList()

	providerList.Add("62.0003430048", "2342423423")

	jp, err := utils.ParseJSON(&j, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	mrfList, err := parsePRObject(&iter, providerList, "1234-5678-9012-3456")
	assert.NoError(t, err)

	// 1 provider_group, 2 provider, 2 tin
	assert.Equal(t, 5, len(mrfList))

	pgMrf := utils.Filter(mrfList, func(mrf *models.Mrf) bool {
		return mrf.RecordType == "provider_group"
	})
	assert.Equal(t, 1, len(pgMrf))
	assert.Equal(t, "1234-5678-9012-3456", pgMrf[0].ParentUUID)
	assert.Equal(t, "62.0003430048", pgMrf[0].ProviderGroupID)

	pgUUID := pgMrf[0].UUID

	tinMrf := utils.Filter(mrfList, func(mrf *models.Mrf) bool {
		return mrf.RecordType == "tin"
	})
	assert.Equal(t, 2, len(tinMrf))
	assert.Equal(t, "ein", tinMrf[0].TinType)
	assert.Equal(t, "1821198789", tinMrf[0].Value)
	assert.Equal(t, pgUUID, tinMrf[0].ParentUUID)
	assert.Equal(t, "npi", tinMrf[1].TinType)
	assert.Equal(t, "1770512915", tinMrf[1].Value)
	assert.Equal(t, pgUUID, tinMrf[1].ParentUUID)

	providerMrf := utils.Filter(mrfList, func(mrf *models.Mrf) bool {
		return mrf.RecordType == "provider"
	})
	assert.Equal(t, 2, len(providerMrf))
	assert.Equal(t, int64(1821198789), providerMrf[0].NpiList[0])
	assert.Equal(t, pgUUID, providerMrf[0].ParentUUID)
	assert.Equal(t, int64(1770512915), providerMrf[1].NpiList[0])
	assert.Equal(t, pgUUID, providerMrf[1].ParentUUID)

}

// test parseTin
func TestParseTin(t *testing.T) {
	var j = []byte(`{ "npi": [1821198789], "tin": { "type": "npi", "value": "1821198789" } }`)

	jp, err := utils.ParseJSON(&j, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	mrf, err := parseTin(&iter, "1234-5678-9012-3456")
	assert.NoError(t, err)

	assert.Equal(t, "1234-5678-9012-3456", mrf.ParentUUID)
	assert.Equal(t, "npi", mrf.Tin.TinType)
	assert.Equal(t, "1821198789", mrf.Tin.Value)
	assert.Equal(t, "tin", mrf.RecordType)
}

// test parsePRRoot
func TestParsePRRootPgIdPresent(t *testing.T) {
	var j = []byte(`{
		"provider_group_id": 62.0003430048,
		"provider_groups": [
			{
				"npi": [1821198789, 987654321],
				"tin": { "type": "npi", "value": "1407989569" }
			},
			{
				"npi": [1174123582],
				"tin": { "type": "ein", "value": "850645536" }
			}
		]
	}`)

	var providerList = NewProviderList()

	providerList.Add("62.0003430048", "2342423423")

	jp, err := utils.ParseJSON(&j, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	mrfList, err := parsePRRoot(providerList, "1234-5678-9012-3456", &iter)
	assert.NoError(t, err)

	assert.Equal(t, "1234-5678-9012-3456", mrfList.ParentUUID)
	assert.Equal(t, "62.0003430048", mrfList.ProviderGroupID)
	assert.Equal(t, "provider_group", mrfList.RecordType)
}

func TestParsePRRootPgIdNotPresent(t *testing.T) {
	var j = []byte(`{
		"provider_group_id": 62.0003430048,
		"provider_groups": [
			{
				"npi": [1821198789, 987654321],
				"tin": { "type": "npi", "value": "1407989569" }
			},
			{
				"npi": [1174123582],
				"tin": { "type": "ein", "value": "850645536" }
			}
		]
	}`)

	var providerList = NewProviderList()

	providerList.Add("65.0003430048", "2342423423")

	jp, err := utils.ParseJSON(&j, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	mrfList, err := parsePRRoot(providerList, "1234-5678-9012-3456", &iter)
	if assert.Error(t, err) {
		assert.Equal(t, &NotInListError{"62.0003430048"}, err)
	}

	assert.Nil(t, mrfList)
}

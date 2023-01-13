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
	"mrfparse/pkg/mrfparse/models"
	"mrfparse/pkg/mrfparse/utils"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/stretchr/testify/assert"
)

func TestParseInObject(t *testing.T) {
	var j = []byte(`{
		"negotiation_arrangement": "ffs",
		"name": "INJECTION, TRASTUZUMAB-QYYP, BIOSIMILAR, (TRAZIMERA), 10 MG",
		"billing_code_type": "HCPCS",
		"billing_code_type_version": "2022",
		"billing_code": "Q5116",
		"negotiated_rates": [
		  {
			"provider_references": [62.0004808658],
			"negotiated_prices": [
			  {
				"negotiated_type": "fee schedule",
				"negotiated_rate": 45.48,
				"expiration_date": "9999-12-31",
				"service_code": ["11", "22"],
				"billing_class": "professional"
			  }
			]
		  },
		  {
			"provider_references": [62.0000565525],
			"negotiated_prices": [
			  {
				"negotiated_type": "fee schedule",
				"negotiated_rate": 69.14,
				"expiration_date": "9999-12-31",
				"service_code": ["22"],
				"billing_class": "institutional"
			  }
			]
		  }
		]
	  }`)

	jp, err := utils.ParseJSON(&j, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	serviceList := mapset.NewSet("Q5116")

	mrfList, err := parseInObject(&iter, "rootUUID", serviceList)
	assert.NoError(t, err)

	mrf := mrfList[0]

	assert.Equal(t, "ffs", mrf.NegotiationArrangement)
	assert.Equal(t, "INJECTION, TRASTUZUMAB-QYYP, BIOSIMILAR, (TRAZIMERA), 10 MG", mrf.Name)
	assert.Equal(t, "HCPCS", mrf.BillingCodeType)
	assert.Equal(t, "2022", mrf.BillingCodeTypeVersion)
	assert.Equal(t, "Q5116", mrf.BillingCode)
	assert.Equal(t, "rootUUID", mrf.ParentUUID)

	prMrf := utils.Filter(mrfList, func(mrf *models.Mrf) bool {
		return mrf.RecordType == "negotiated_rate"
	})

	assert.Equal(t, 2, len(prMrf))
	assert.Equal(t, models.ProviderReferences{"62.0004808658"}, prMrf[0].PRList)
	assert.Equal(t, models.ProviderReferences{"62.0000565525"}, prMrf[1].PRList)

	npMrf := utils.Filter(mrfList, func(mrf *models.Mrf) bool {
		return mrf.RecordType == "negotiated_prices"
	})

	assert.Equal(t, 2, len(npMrf))
	assert.Equal(t, "fee schedule", npMrf[0].NegotiatedType)
	assert.Equal(t, 45.48, npMrf[0].NegotiatedRateValue)
	assert.Equal(t, "9999-12-31", npMrf[0].ExpirationDate)
	assert.Equal(t, models.ServiceCodes{"11", "22"}, npMrf[0].ServiceCodes)
	assert.Equal(t, "professional", npMrf[0].BillingClass)
	assert.Equal(t, "fee schedule", npMrf[1].NegotiatedType)
	assert.Equal(t, 69.14, npMrf[1].NegotiatedRateValue)
	assert.Equal(t, "9999-12-31", npMrf[1].ExpirationDate)
	assert.Equal(t, models.ServiceCodes{"22"}, npMrf[1].ServiceCodes)
	assert.Equal(t, "institutional", npMrf[1].BillingClass)
	assert.Equal(t, prMrf[0].UUID, npMrf[0].ParentUUID)
	assert.Equal(t, prMrf[1].UUID, npMrf[1].ParentUUID)

}

// test isServiceInList
func TestIsServiceInListCPT(t *testing.T) {
	var j = []byte(`{
		"negotiation_arrangement": "ffs",
		"name": "REV 204 & ICD10DX F12.10",
		"billing_code_type": "CPT",
		"billing_code_type_version": "2021",
		"billing_code": "2021",
		"description": "REV 204 & ICD10DX F12.10"}`)

	jp, err := utils.ParseJSON(&j, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	serviceList := mapset.NewSet("2025", "2021", "53")

	bt, bc, ok := isServiceInList(&iter, serviceList)
	assert.Equal(t, true, ok)
	assert.Equal(t, "CPT", bt)
	assert.Equal(t, "2021", bc)

	serviceList = mapset.NewSet("1", "2", "3")

	bt, bc, ok = isServiceInList(&iter, serviceList)
	assert.Equal(t, false, ok)
	assert.Equal(t, "CPT", bt)
	assert.Equal(t, "2021", bc)
}

func TestIsServiceInListHCPCS(t *testing.T) {
	var j = []byte(`{
		"negotiation_arrangement": "ffs",
		"name": "REV 204 & ICD10DX F12.10",
		"billing_code_type": "HCPCS",
		"billing_code_type_version": "2021",
		"billing_code": "2021",
		"description": "REV 204 & ICD10DX F12.10"}`)

	jp, err := utils.ParseJSON(&j, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	serviceList := mapset.NewSet("2025", "2021", "53")

	bt, bc, ok := isServiceInList(&iter, serviceList)
	assert.Equal(t, true, ok)
	assert.Equal(t, "HCPCS", bt)
	assert.Equal(t, "2021", bc)

	serviceList = mapset.NewSet("1", "2", "3")

	bt, bc, ok = isServiceInList(&iter, serviceList)
	assert.Equal(t, false, ok)
	assert.Equal(t, "HCPCS", bt)
	assert.Equal(t, "2021", bc)
}

// Test parseInRoot
func TestParseInRoot(t *testing.T) {
	var mrf *models.Mrf
	var j = []byte(`{
		"negotiation_arrangement": "ffs",
		"name": "REV 204",
		"billing_code_type": "CPT",
		"billing_code_type_version": "1.0",
		"billing_code": "999",
		"description": "REV 204 & ICD10DX F12.10"}`)

	jp, err := utils.ParseJSON(&j, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	serviceList := mapset.NewSet("2025", "999", "53")

	rootUUID := "1234"

	mrf, err = parseInRoot(&iter, rootUUID, serviceList)

	assert.NoError(t, err)
	assert.Equal(t, "REV 204", mrf.Name)
	assert.Equal(t, "CPT", mrf.InNetwork.BillingCodeType)
	assert.Equal(t, "1.0", mrf.InNetwork.BillingCodeTypeVersion)
	assert.Equal(t, "999", mrf.InNetwork.BillingCode)
	assert.Equal(t, "REV 204 & ICD10DX F12.10", mrf.InNetwork.Description)
	assert.Equal(t, "ffs", mrf.NegotiationArrangement)
	assert.Equal(t, "1234", mrf.ParentUUID)
	assert.Equal(t, "in_network", mrf.RecordType)
}

func TestParseInRootNotInServiceListError(t *testing.T) {
	var j = []byte(`{
		"negotiation_arrangement": "ffs",
		"name": "REV 204",
		"billing_code_type": "CPT",
		"billing_code_type_version": "1.0",
		"billing_code": "888",
		"description": "REV 204 & ICD10DX F12.10"}`)

	jp, err := utils.ParseJSON(&j, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	serviceList := mapset.NewSet("2025", "999", "53")

	rootUUID := "1234"

	_, err = parseInRoot(&iter, rootUUID, serviceList)

	_, ok := err.(*NotInListError)

	assert.Equal(t, true, ok)
}

func TestParseBundledCodes(t *testing.T) {
	var j = []byte(`{"bundled_codes":[
		{"billing_code_type":"RC","billing_code_type_version":"2022","billing_code":"0636","description":"DRUGS REQUIRING DETAILED CODING"},
		{"billing_code_type":"YC","billing_code_type_version":"2021","billing_code":"0450","description":"EMERGENCY ROOM GENERAL CLASSIFICATION"}
		]}`)

	jp, err := utils.ParseJSON(&j, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	mrfList, err := parseBundledCodes(&iter, "987978hjh")
	assert.NoError(t, err)

	assert.Equal(t, 2, len(mrfList))
	assert.Equal(t, "987978hjh", mrfList[0].ParentUUID)
	assert.Equal(t, "987978hjh", mrfList[1].ParentUUID)
	assert.Equal(t, "RC", mrfList[0].BundledCodes.BCBillingCodeType)
	assert.Equal(t, "YC", mrfList[1].BundledCodes.BCBillingCodeType)
	assert.Equal(t, "2022", mrfList[0].BundledCodes.NCBillingCodeTypeVersion)
	assert.Equal(t, "2021", mrfList[1].BundledCodes.NCBillingCodeTypeVersion)
	assert.Equal(t, "0636", mrfList[0].BundledCodes.BCBillingCode)
	assert.Equal(t, "0450", mrfList[1].BundledCodes.BCBillingCode)
	assert.Equal(t, "DRUGS REQUIRING DETAILED CODING", mrfList[0].BundledCodes.BCDescription)
	assert.Equal(t, "EMERGENCY ROOM GENERAL CLASSIFICATION", mrfList[1].BundledCodes.BCDescription)
}

func TestParseBundledCodesMissing(t *testing.T) {
	var j = []byte(`{"field":"value"}`)

	jp, err := utils.ParseJSON(&j, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	mrfList, err := parseBundledCodes(&iter, "987978hjh")
	assert.NoError(t, err)

	assert.Equal(t, 0, len(mrfList))
}

func TestParseNeServiceCodesProfessional(t *testing.T) {
	var j = []byte(`{"negotiated_type":"fee schedule","negotiated_rate":410.82,"expiration_date":"9999-12-31",
	"service_code":["23","41","26","21","52","42","24","22","56","31","51","53","61","19","34"],"billing_class":"professional"}`)

	var scsExpected = []string{"23", "41", "26", "21", "52", "42", "24", "22", "56", "31", "51", "53", "61", "19", "34"}

	jp, err := utils.ParseJSON(&j, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	scs, err := parseNPServiceCodes(&iter, "professional")
	assert.NoError(t, err)

	assert.Equal(t, 15, len(scs))
	assert.Equal(t, scsExpected, scs)
}

func TestParseNeServiceCodesProfessionalMissing(t *testing.T) {
	var j = []byte(`{"negotiated_type":"fee schedule","negotiated_rate":410.82,"expiration_date":"9999-12-31",
	"billing_class":"professional"}`)

	jp, err := utils.ParseJSON(&j, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	_, err = parseNPServiceCodes(&iter, "professional")
	assert.ErrorContains(t, err, "service_code is missing")
}

func TestParseNeServiceCodesInstitutional(t *testing.T) {
	var j = []byte(`{"negotiated_type":"fee schedule","negotiated_rate":410.82,"expiration_date":"9999-12-31",
	"billing_class":"institutional"}`)

	jp, err := utils.ParseJSON(&j, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	scs, err := parseNPServiceCodes(&iter, "institutional")
	assert.True(t, utils.TestElementNotPresent(err, "negotiated_rate"))
	assert.Nil(t, scs)
}

func TestParseNegotiatedPrices(t *testing.T) {
	var j = []byte(`{
		"provider_references": [62.0000005757],
		"negotiated_prices": [
		  {
			"negotiated_type": "fee schedule",
			"negotiated_rate": 34.54,
			"expiration_date": "9999-12-31",
			"service_code": ["81", "11", "22"],
			"billing_class": "professional"
		  },
		  {
			"negotiated_type": "fee schedule",
			"negotiated_rate": 14.8,
			"expiration_date": "9999-12-31",
			"service_code": ["67", "55", "44"],
			"billing_code_modifier": ["12", "14"],
			"billing_class": "institutional"
		  }]}`)

	jp, err := utils.ParseJSON(&j, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	mrf, err := parseNegotiatedPrices(&iter, "nrUUID")
	assert.NoError(t, err)

	assert.Equal(t, 2, len(mrf))
	assert.Equal(t, "nrUUID", mrf[0].ParentUUID)
	assert.Equal(t, "nrUUID", mrf[1].ParentUUID)
	assert.Equal(t, "fee schedule", mrf[0].NegotiatedPrices.NegotiatedType)
	assert.Equal(t, "fee schedule", mrf[1].NegotiatedPrices.NegotiatedType)
	assert.Equal(t, 34.54, mrf[0].NegotiatedPrices.NegotiatedRateValue)
	assert.Equal(t, 14.8, mrf[1].NegotiatedPrices.NegotiatedRateValue)
	assert.Equal(t, "9999-12-31", mrf[0].NegotiatedPrices.ExpirationDate)
	assert.Equal(t, "9999-12-31", mrf[1].NegotiatedPrices.ExpirationDate)
	assert.Equal(t, "professional", mrf[0].NegotiatedPrices.BillingClass)
	assert.Equal(t, "institutional", mrf[1].NegotiatedPrices.BillingClass)
	assert.Equal(t, models.ServiceCodes{"81", "11", "22"}, mrf[0].NegotiatedPrices.ServiceCodes)
	assert.Equal(t, models.BillingCodeModifiers{"12", "14"}, mrf[1].NegotiatedPrices.BillingCodeModifiers)
	assert.Equal(t, "negotiated_prices", mrf[0].RecordType)
	assert.Equal(t, "negotiated_prices", mrf[1].RecordType)
	assert.NotZero(t, mrf[0].UUID)
	assert.NotZero(t, mrf[1].UUID)

}

// test parseNegotiatedRates function
func TestParseNegotiatedRates(t *testing.T) {
	var j = []byte(`{
		"negotiation_arrangement": "ffs",
		"negotiated_rates": [
		  {
			"provider_references": [492089],
			"negotiated_prices": [
			  {
				"negotiated_type": "per diem",
				"negotiated_rate": 781.0,
				"expiration_date": "9999-12-31",
				"service_code": [
				  "21",
				  "31",
				  "32",
				  "33"
				],
				"billing_class": "institutional"
			  }
			]
		  },
		  {
			"provider_references": [11925, 403819],
			"negotiated_prices": [
			  {
				"negotiated_type": "per diem",
				"negotiated_rate": 4793.0,
				"expiration_date": "9999-12-31",
				"service_code": [
				  "21",
				  "31",
				  "32",
				  "33",
				  "34",
				  "51",
				  "54",
				  "55",
				  "56",
				  "61"
				],
				"billing_class": "institutional"
			  }
			]
		  },
		  {
			"provider_references": [62.0004643342],
			"negotiated_prices": [
			  {
				"negotiated_type": "fee schedule",
				"negotiated_rate": 199.06,
				"expiration_date": "9999-12-31",
				"service_code": ["12", "11"],
				"billing_class": "professional",
				"billing_code_modifier": ["NU", ""]
			  }
			]
		  }
		]
	  }
	  `)

	jp, err := utils.ParseJSON(&j, nil)
	assert.NoError(t, err)

	iter := jp.Iter()

	mrf, err := parseNegotiatedRates(&iter, "inUUID")
	assert.NoError(t, err)

	assert.Equal(t, 6, len(mrf))
	assert.Equal(t, "per diem", mrf[0].NegotiatedType)
	assert.Equal(t, 781.0, mrf[0].NegotiatedPrices.NegotiatedRateValue)
	assert.Equal(t, "9999-12-31", mrf[0].ExpirationDate)
	assert.Equal(t, "institutional", mrf[0].BillingClass)
	assert.Equal(t, models.ServiceCodes{"21", "31", "32", "33"}, mrf[0].ServiceCodes)
	assert.Equal(t, models.ProviderReferences{"492089"}, mrf[1].PRList)
	assert.Equal(t, "inUUID", mrf[1].ParentUUID)
	assert.Equal(t, models.ServiceCodes{"21", "31", "32", "33", "34", "51", "54", "55", "56", "61"}, mrf[2].ServiceCodes)
	assert.Equal(t, models.ProviderReferences{"11925", "403819"}, mrf[3].PRList)
	assert.Equal(t, "fee schedule", mrf[4].NegotiatedType)
	assert.Equal(t, 199.06, mrf[4].NegotiatedPrices.NegotiatedRateValue)
	assert.Equal(t, "9999-12-31", mrf[4].ExpirationDate)
	assert.Equal(t, "professional", mrf[4].BillingClass)
	assert.Equal(t, models.ServiceCodes{"12", "11"}, mrf[4].ServiceCodes)
	assert.Equal(t, models.BillingCodeModifiers{"NU", ""}, mrf[4].BillingCodeModifiers)
	// test for RecordType
	assert.Equal(t, "negotiated_rate", mrf[1].RecordType)
	assert.Equal(t, "negotiated_rate", mrf[3].RecordType)
}

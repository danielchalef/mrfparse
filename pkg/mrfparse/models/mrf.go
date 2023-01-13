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
package models

type ServiceCodes []string
type BillingCodeModifiers []string
type ProviderReferences []string
type NpiList []int64

// We use plain encoding for all fields to increase compatibility with parquetlibraries.
type Mrf struct {
	MrfRoot
	InNetwork

	BundledCodes
	Tin
	ProviderGroup
	UUID       string `parquet:"uuid,plain"`
	ParentUUID string `parquet:"parent_uuid,plain"`
	RecordType string `parquet:"record_type,enum,plain"`

	Provider

	NegotiatedRate

	NegotiatedPrices
}

type MrfRoot struct {
	ReportingEntityName string `json:"reporting_entity_name" parquet:"reporting_entity_name,plain"`
	ReportingEntityType string `json:"reporting_entity_type" parquet:"reporting_entity_type,plain"`
	LastUpdatedOn       string `json:"last_updated_on" parquet:"last_updated_on,plain"`
	Version             string `json:"version" parquet:"version,plain"`
	PlanMarketType      string `json:"plan_market_type,omitempty" parquet:"plan_market_type,enum,plain"`
	PlanName            string `json:"plan_name,omitempty" parquet:"plan_name,plain"`
	PlanIDType          string `json:"plan_id_type,omitempty" parquet:"plan_id_type,plain"`
	PlanID              string `json:"plan_id,omitempty" parquet:"plan_id,plain"`
}

type ProviderGroup struct {
	ProviderGroupID string `parquet:"provider_group_id,enum,plain"`
}

type Provider struct {
	Parent  string  `parquet:"provider_parent,plain"`
	NpiList NpiList `parquet:"provider_npi_list,list,plain"`
}

type Tin struct {
	Value   string `parquet:"provider_tin_value,plain"`
	TinType string `parquet:"provider_tin_type,enum,plain"`
}

type InNetwork struct {
	Name                   string `parquet:"in_name,plain"`
	Description            string `parquet:"in_description,plain"`
	NegotiationArrangement string `parquet:"in_negotiation_arrangement,enum,plain"`
	BillingCodeType        string `parquet:"in_billing_code_type,enum,plain"`
	BillingCode            string `parquet:"in_billing_code,plain"`
	BillingCodeTypeVersion string `parquet:"in_billing_code_type_version,plain"`
}

type BundledCodes struct {
	BCDescription            string `parquet:"in_bc_description,plain"`
	BCBillingCodeType        string `parquet:"in_bc_billing_code_type,enum,plain"`
	BCBillingCode            string `parquet:"in_bc_billing_code,plain"`
	NCBillingCodeTypeVersion string `parquet:"in_bc_billing_code_type_version,plain"`
}

type NegotiatedPrices struct {
	NegotiatedType        string               `parquet:"in_np_negotiated_type,enum,plain"`
	BillingClass          string               `parquet:"in_np_billing_class,plain"`
	ExpirationDate        string               `parquet:"in_np_expiration_date,plain"`
	AdditionalInformation string               `parquet:"in_np_additional_information,plain"`
	ServiceCodes          ServiceCodes         `parquet:"in_np_service_codes,list,plain"`
	BillingCodeModifiers  BillingCodeModifiers `parquet:"in_np_billing_code_modifiers,list,plain"`
	NegotiatedRateValue   float64              `parquet:"in_np_negotiated_rate,plain"`
}

type NegotiatedRate struct {
	PRList ProviderReferences `parquet:"in_nr_provider_references,list,plain"`
}

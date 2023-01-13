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
	"bufio"
	"context"
	"fmt"
	"github.com/danielchalef/mrfparse/pkg/mrfparse/cloud"
	"github.com/danielchalef/mrfparse/pkg/mrfparse/models"
	"github.com/danielchalef/mrfparse/pkg/mrfparse/utils"
	"io"
	"strings"

	"github.com/minio/simdjson-go"
)

func parseInNetworkRates(filename, rootUUUID string, serviceList StringSet) {
	const LinesAtATime int = 100

	var line string
	var lineCount = 0
	var totalLineCount = 0

	var strBuilder strings.Builder

	log.Info("Parsing in_network_rates: ", filename)

	f, err := cloud.NewReader(context.TODO(), filename)
	utils.ExitOnError(err)

	defer func(f io.ReadCloser) {
		err := f.Close()
		if err != nil {
			utils.ExitOnError(err)
		}
	}(f)

	scanner := bufio.NewScanner(f)

	buf := make([]byte, MaxLineLength)
	scanner.Buffer(buf, MaxLineLength)

	for scanner.Scan() {
		line = scanner.Text()
		strBuilder.WriteString(line)
		strBuilder.WriteString("\n")

		if lineCount == LinesAtATime {
			lines := strBuilder.String()

			inPoolGroup.Submit(func() {
				parseInLines(&lines, rootUUUID, serviceList)
			})

			lineCount = 0

			strBuilder.Reset()
		} else {
			lineCount++
		}

		if totalLineCount%5000 == 0 {
			log.Debug("Read ", totalLineCount, " lines")
		}

		totalLineCount++
	}

	if lineCount > 0 {
		lines := strBuilder.String()

		inPoolGroup.Submit(func() {
			parseInLines(&lines, rootUUUID, serviceList)
		})
	}

	log.Info("Completed reading negotiated_rates: ", filename)
}

func parseInLines(lines *string, rootUUID string, serviceList StringSet) {
	parsed, err := utils.ParseJSON(lines, nil)
	utils.ExitOnError(err)

	var iter = parsed.Iter()
	var tmpIter *simdjson.Iter

	var mrfList []*models.Mrf

	for {
		typ := iter.Advance()

		if typ == simdjson.TypeRoot {
			_, tmpIter, err = iter.Root(nil)
			utils.ExitOnError(err)

			_, err := tmpIter.FindElement(nil, "covered_services")
			if err == nil {
				// This is a covered_services record. We don't yet support these.
				utils.ExitOnError(fmt.Errorf("covered_services records are not supported"))
			}

			// Parse in_network_rates object
			mrfList, err = parseInObject(tmpIter, rootUUID, serviceList)
			// if we get a NotInListError, skip this record as it's not in the serviceList
			if e, ok := err.(*NotInListError); ok {
				log.Tracef("Skipping in_network_rates record. %s", e.Error())
				continue
			}

			// if it's another error, exit
			utils.ExitOnError(err)

			err = WriteRecords(mrfList)
			utils.ExitOnError(err)
		} else if typ == simdjson.TypeNone {
			break
		}
	}
}

func parseInObject(iter *simdjson.Iter, rootUUID string, serviceList StringSet) ([]*models.Mrf, error) {
	var (
		err                 error
		mrf                 *models.Mrf
		mrfList, mrfListTmp []*models.Mrf
		inUUID              string
	)

	// Parse the root of the in_network_rates record
	mrf, err = parseInRoot(iter, rootUUID, serviceList)
	if err != nil {
		return nil, err
	}

	inUUID = mrf.UUID

	mrfList = append(mrfList, mrf)

	// Parse bundled_codes, if present
	mrfListTmp, err = parseBundledCodes(iter, inUUID)
	if err != nil {
		return nil, err
	}

	mrfList = append(mrfList, mrfListTmp...)

	// Parse negotiated_rates
	mrfListTmp, err = parseNegotiatedRates(iter, inUUID)
	if err != nil {
		return nil, err
	}

	mrfList = append(mrfList, mrfListTmp...)

	return mrfList, nil
}

func parseBundledCodes(iter *simdjson.Iter, inUUID string) ([]*models.Mrf, error) {
	var mrfList []*models.Mrf

	path := "bundled_codes"
	bc, err := iter.FindElement(nil, path)

	// If bundled_codes is not present, return an empty list
	// bundled_codes is optional
	if utils.TestElementNotPresent(err, path) {
		return mrfList, nil
	}

	bcIter := bc.Iter

	for {
		typ := bcIter.Advance()

		if typ == simdjson.TypeObject {
			bcUUID := utils.GetUniqueID()

			bcType, err := utils.GetElementValue[string]("billing_code_type", &bcIter)
			if err != nil {
				return nil, err
			}

			bcCode, err := utils.GetElementValue[string]("billing_code", &bcIter)
			if err != nil {
				return nil, err
			}

			bcTypeVersion, err := utils.GetElementValue[string]("billing_code_type_version", &bcIter)
			if err != nil {
				return nil, err
			}

			path = "description"
			bcDescription, err := utils.GetElementValue[string]("description", &bcIter)

			r := utils.TestElementNotPresent(err, path)
			if r {
				bcDescription = ""
			}

			mrfList = append(mrfList,
				&models.Mrf{UUID: bcUUID, ParentUUID: inUUID, RecordType: "bundled_codes",
					BundledCodes: models.BundledCodes{BCBillingCodeType: bcType, BCBillingCode: bcCode,
						NCBillingCodeTypeVersion: bcTypeVersion, BCDescription: bcDescription}})
		} else if typ == simdjson.TypeNone {
			break
		}
	}

	return mrfList, nil
}

func parseNegotiatedRates(iter *simdjson.Iter, inUUID string) ([]*models.Mrf, error) {
	const prParent = "negotiated_rates"

	var (
		err                           error
		mrfList, npMrfList, prMrfList []*models.Mrf
		nr                            *simdjson.Array
		neIter                        simdjson.Iter
		uuid                          string
		pr                            []string
	)

	nr, err = utils.GetArrayForElement("negotiated_rates", iter)
	if err != nil {
		return nil, err
	}

	neIter = nr.Iter()

	for {
		typ := neIter.Advance()

		if typ == simdjson.TypeObject {
			uuid = utils.GetUniqueID()

			// Parse negotiated_prices
			npMrfList, err = parseNegotiatedPrices(&neIter, uuid)
			if err != nil {
				return nil, err
			}

			mrfList = append(mrfList, npMrfList...)

			// We should have one of provider_references or provider_groups
			pr, err = utils.GetArrayElementAsSlice[string]("provider_references", &neIter)
			// if provider_references is missing, parse provider_groups
			if utils.TestElementNotPresent(err, "provider_references") {
				// Add a record to capture the NR / parent relationship
				mrfList = append(mrfList, &models.Mrf{UUID: uuid, ParentUUID: inUUID})

				prMrfList, err = parseProviderGroups(iter, uuid, prParent)
				if err != nil {
					return nil, err
				}

				mrfList = append(mrfList, prMrfList...)
			} else {
				// if provider_references not missing, add to providersFilter and write record
				providersFilter.Add(pr...)

				mrfList = append(mrfList, &models.Mrf{UUID: uuid, ParentUUID: inUUID, RecordType: "negotiated_rate",
					NegotiatedRate: models.NegotiatedRate{PRList: pr}})
			}
		} else if typ == simdjson.TypeNone {
			break
		}
	}

	return mrfList, nil
}

// parseNegotiatedPrices parses the negotiated_prices array, and returns a list of MRFs.
func parseNegotiatedPrices(iter *simdjson.Iter, nrUUID string) ([]*models.Mrf, error) {
	var (
		err        error
		np         *simdjson.Array
		mrfList    []*models.Mrf
		scs        []string
		bcs        []string
		ai, t      string
		uuid, path string
	)

	np, err = utils.GetArrayForElement("negotiated_prices", iter)
	if err != nil {
		return nil, err
	}

	npIter := np.Iter()

	for {
		typ := npIter.Advance()
		if typ == simdjson.TypeObject {
			uuid = utils.GetUniqueID()

			t, err = utils.GetElementValue[string]("negotiated_type", &npIter)
			if err != nil {
				return nil, err
			}

			bc, err := utils.GetElementValue[string]("billing_class", &npIter)
			if err != nil {
				return nil, err
			}

			ed, err := utils.GetElementValue[string]("expiration_date", &npIter)
			if err != nil {
				return nil, err
			}

			nr, err := utils.GetElementValue[float64]("negotiated_rate", &npIter)
			if err != nil {
				return nil, err
			}

			path = "additional_information"
			ai, err = utils.GetElementValue[string](path, &npIter)
			if utils.TestElementNotPresent(err, path) {
				ai = ""
			} else if err != nil {
				return nil, err
			}

			scs, err = parseNPServiceCodes(&npIter, bc)
			if err != nil {
				return nil, err
			}

			path = "billing_code_modifier"
			bcs, err = utils.GetArrayElementAsSlice[string](path, &npIter)
			if utils.TestElementNotPresent(err, path) {
				bcs = []string{}
			} else if err != nil {
				return nil, err
			}

			mrfList = append(mrfList, &models.Mrf{UUID: uuid, ParentUUID: nrUUID, RecordType: "negotiated_prices",
				NegotiatedPrices: models.NegotiatedPrices{NegotiatedType: t, BillingClass: bc, ExpirationDate: ed,
					NegotiatedRateValue: nr, AdditionalInformation: ai, ServiceCodes: scs,
					BillingCodeModifiers: bcs}})
		} else if typ == simdjson.TypeNone {
			break
		}
	}

	return mrfList, nil
}

// parseNPServiceCodes parses Negotiated Prices service_codes, which should be present if billing_class is professional
// Returns an empty slice if billing_class is not professional
func parseNPServiceCodes(iter *simdjson.Iter, billingClass string) ([]string, error) {
	var scs []string
	var err error

	scs, err = utils.GetArrayElementAsSlice[string]("service_code", iter)

	if billingClass == "professional" && utils.TestElementNotPresent(err, "service_code") {
		return nil, fmt.Errorf("service_code is missing from negotiated_prices for billing_class == professional")
	} else if err != nil {
		return nil, err
	}

	return scs, nil
}

// isServiceInList gets the billing_code_type and code and determines if the service is in serviceList
func isServiceInList(tmpIter *simdjson.Iter, serviceList StringSet) (billingCodeType, billingCode string, ok bool) {
	bct, err := utils.GetElementValue[string]("billing_code_type", tmpIter)
	if err != nil {
		utils.ExitOnError(err)
	}

	bc, err := utils.GetElementValue[string]("billing_code", tmpIter)
	if err != nil {
		utils.ExitOnError(err)
	}

	return bct, bc, ((bct == "HCPCS" || bct == "CPT") && serviceList.Contains(bc))
}

// parseInRoot parses the root of the in_network file, returning an Mrf record.
// If the service is not in the serviceList, it returns a NotInServiceListError
func parseInRoot(iter *simdjson.Iter, rootUUID string, serviceList StringSet) (*models.Mrf, error) {
	var uuid = utils.GetUniqueID()

	// Get the billing_code_type and code and determine if in serviceList
	inBillingCodeType, inBillingCode, ok := isServiceInList(iter, serviceList)
	if !ok {
		// This is not a service we care about. Skip it.
		return nil, &NotInListError{inBillingCode}
	}

	name, err := utils.GetElementValue[string]("name", iter)
	if err != nil {
		return nil, err
	}

	bcv, err := utils.GetElementValue[string]("billing_code_type_version", iter)
	if err != nil {
		return nil, err
	}

	na, err := utils.GetElementValue[string]("negotiation_arrangement", iter)
	if err != nil {
		return nil, err
	}

	path := "description"
	desc, err := utils.GetElementValue[string](path, iter)
	r := utils.TestElementNotPresent(err, path)
	// some carriers don't have a description, despite it being a required field
	if r {
		desc = ""
	}

	return &models.Mrf{UUID: uuid, ParentUUID: rootUUID, RecordType: "in_network",
		InNetwork: models.InNetwork{Name: name, BillingCodeTypeVersion: bcv, NegotiationArrangement: na,
			Description: desc, BillingCodeType: inBillingCodeType, BillingCode: inBillingCode}}, nil
}

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
	"sync/atomic"

	"github.com/minio/simdjson-go"
)

var matchedProviderCounter = atomic.Int32{}
var totalProviderCounter = atomic.Int32{}

// parseProviderReference parses provider_references_*.jsonl files
func parseProviderReference(filename, rootUUID string) {
	const LinesAtATime int = 2_000

	var (
		line           string
		lineCount      = 0
		totalLineCount = 0
		strBuilder     strings.Builder
	)

	log.Info("Parsing provider references: ", filename)

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
		// Build a NDJSON string with LinesAtATime lines
		line = scanner.Text()
		strBuilder.WriteString(line)
		strBuilder.WriteString("\n")

		if lineCount == LinesAtATime {
			lines := strBuilder.String()

			// submit the parse job to the goroutine pool
			prPoolGroup.Submit(func() {
				parsePRLines(&lines, rootUUID)
			})

			lineCount = 0

			strBuilder.Reset()
		} else {
			lineCount++
		}

		if totalLineCount%100_000 == 0 {
			log.Debug("Read ", totalLineCount, " lines")
		}

		totalLineCount++
	}

	// Ensure we parse the last few lines if we've not yet reached LinesAtATime
	if lineCount > 0 {
		lines := strBuilder.String()

		prPoolGroup.Submit(func() {
			parsePRLines(&lines, rootUUID)
		})
	}

	log.Info("Completed reading provider references: ", filename)
}

// parsePRLines parses provider_references lines, each of which is a json object.
// It's designed to run concurrently, with parseProviderReference submitting parsePRLines jobs
// to the goroutine pool. Parsed Mrf records are written to a channel for processing by a Writer thread.
func parsePRLines(lines *string, rootUUID string) {
	parsed, err := utils.ParseJSON(lines, nil)
	utils.ExitOnError(err)

	var (
		iter    = parsed.Iter()
		tmpIter *simdjson.Iter
		mrfList []*models.Mrf
	)

	for {
		typ := iter.Advance()

		if typ == simdjson.TypeRoot {
			totalProviderCounter.Add(1)

			_, tmpIter, err = iter.Root(nil)
			utils.ExitOnError(err)

			_, err := tmpIter.FindElement(nil, "location")
			if err == nil {
				// This is a location record. We don't yet support these.
				utils.ExitOnError(fmt.Errorf("location records are not supported"))
			}

			mrfList, err = parsePRObject(tmpIter, providersFilter, rootUUID)
			// We only want to parse records where the provider_group_id is present in the in_network_rates dataset.
			// If we get a NotInListError, skip this record.
			if e, ok := err.(*NotInListError); ok {
				log.Tracef("Skipping provider_reference record. %s", e.Error())
				continue
			}

			// Exit on any other error
			utils.ExitOnError(err)

			// Count a matched provider
			matchedProviderCounter.Add(1)

			err = WriteRecords(mrfList)
			utils.ExitOnError(err)
		} else if typ == simdjson.TypeNone {
			break
		}
	}
}

// parsePRObject parses a provider_reference object. It returns a slice of Mrf records, which
// contains the root object and any provider_groups.
func parsePRObject(iter *simdjson.Iter, providersFilter *ProviderList, rootUUID string) ([]*models.Mrf, error) {
	const parent = "provider_references"

	var (
		mrf     *models.Mrf
		mrfList []*models.Mrf
		err     error
	)

	mrf, err = parsePRRoot(providersFilter, rootUUID, iter)
	if err != nil {
		return nil, err
	}

	mrfList, err = parseProviderGroups(iter, mrf.UUID, parent)
	if err != nil {
		return nil, err
	}

	mrfList = append(mrfList, mrf)

	return mrfList, nil
}

// parsePRRoot parses the root of the provider_reference file. If the provider is not in the
// providerFilter set, then it returns a NotInListError.
func parsePRRoot(providers *ProviderList, rootUUID string, iter *simdjson.Iter) (*models.Mrf, error) {
	uuid := utils.GetUniqueID()

	id, err := utils.GetElementValue[string]("provider_group_id", iter)
	if err != nil {
		return nil, err
	}

	if !providers.Contains(id) {
		return nil, &NotInListError{item: id}
	}

	return &models.Mrf{UUID: uuid, ParentUUID: rootUUID, RecordType: "provider_group",
		ProviderGroup: models.ProviderGroup{ProviderGroupID: id}}, nil
}

// parseProviderGroups parses the provider_groups element, which is an array of providers.
// It creates an MRf record for each provider.
func parseProviderGroups(iter *simdjson.Iter, parentUUID, parent string) ([]*models.Mrf, error) {
	var mrfList []*models.Mrf
	var mrf *models.Mrf
	var uuid string
	var err error
	var p *simdjson.Object
	var npi []int64

	pa, err := utils.GetArrayForElement("provider_groups", iter)
	if err != nil {
		return nil, err
	}

	paIter := pa.Iter()

	for {
		typ := paIter.Advance()

		if typ == simdjson.TypeObject {
			p, err = paIter.Object(p)
			if err != nil {
				return nil, err
			}

			uuid = utils.GetUniqueID()

			// Parse the npi array
			npi, err = utils.GetArrayElementAsSlice[int64]("npi", &paIter)
			if err != nil {
				return nil, err
			}

			mrfList = append(mrfList, &models.Mrf{UUID: uuid, ParentUUID: parentUUID, RecordType: "provider",
				Provider: models.Provider{Parent: parent, NpiList: npi}})

			// parse tin element
			mrf, err = parseTin(&paIter, parentUUID)
			if err != nil {
				return nil, err
			}

			mrfList = append(mrfList, mrf)
		} else if typ == simdjson.TypeNone {
			break
		}
	}

	return mrfList, err
}

// parseTin parses the tin element of the provider record.
// parentUUID is the UUID of the parent providers record.
func parseTin(iter *simdjson.Iter, parentUUID string) (*models.Mrf, error) {
	tin, err := iter.FindElement(nil, "tin")
	if err != nil {
		return nil, err
	}

	tt, err := utils.GetElementValue[string]("type", &tin.Iter)
	if err != nil {
		return nil, err
	}

	tv, err := utils.GetElementValue[string]("value", &tin.Iter)
	if err != nil {
		return nil, err
	}

	tinUUID := utils.GetUniqueID()

	return &models.Mrf{UUID: tinUUID, ParentUUID: parentUUID, RecordType: "tin",
		Tin: models.Tin{TinType: tt, Value: tv}}, nil
}

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
	"context"
	"github.com/danielchalef/mrfparse/pkg/mrfparse/cloud"
	"github.com/danielchalef/mrfparse/pkg/mrfparse/models"
	"github.com/danielchalef/mrfparse/pkg/mrfparse/parquet"
	"github.com/danielchalef/mrfparse/pkg/mrfparse/utils"
	"path/filepath"
	"strings"

	"github.com/alitto/pond"
	mapset "github.com/deckarep/golang-set/v2"
)

const MaxWorkers int = 5
const MaxCapacity int = 4

const MaxLineLength int = 5000000 // bytes
type StringSet mapset.Set[string]

var log = utils.GetLogger()

var processPool = pond.New(MaxWorkers, MaxCapacity)
var inPoolGroup = processPool.Group()
var prPoolGroup = processPool.Group()
var writerPoolGroup = processPool.Group()

func Parse(inputPath, outputPath string, planID int64, serviceFile string) {
	const writerChannelSize int = 4 * 1024

	// used to persist []mrf to parquet
	wc := make(chan []*models.Mrf, writerChannelSize)
	// done channel for writers
	done := make(chan bool)

	// Start the writer in a goroutine
	writerPoolGroup.Submit(func() { parquet.Writer("mrf", outputPath, wc, done) })

	// create the record writer using the new wc channel
	WriteRecords = NewRecordWriter(wc)

	// Load service list that we'll use to filter for services we care about
	serviceList := loadServiceList(serviceFile)
	log.Info("Loaded ", serviceList.Cardinality(), " services.")

	// Get list of files in inputPath. We expect to find a root file and in_network_rate and provider_references files
	filesList, err := cloud.Glob(context.TODO(), inputPath, "*.json*")
	utils.ExitOnError(err)

	log.Info("Found ", len(filesList), " files.")

	// Parse root file first as we need root uuid for the other records
	filename, err := findRootFile(filesList)
	utils.ExitOnError(err)

	rootUUID := writeRoot(filename, planID)
	log.Info("MrfRoot file parsed: ", filename)

	// Parse in_network files first
	for i := range filesList {
		f := filepath.Base(filesList[i])
		if strings.HasPrefix(f, "in_network_") {
			log.Info("Found in_network_rate file", filename)
			parseInNetworkRates(filesList[i], rootUUID, serviceList)
		}
	}

	// Wait for all in_network threads to finish
	log.Debug("Waiting for in_network_rate threads to finish.")
	inPoolGroup.Wait()

	log.Info("Found ", providersFilter.Len(), " providers in in_network_rates.")

	// Parse provider_references_ files
	for i := range filesList {
		f := filepath.Base(filesList[i])
		if strings.HasPrefix(f, "provider_references_") {
			log.Info("Found provider_references file", f)
			parseProviderReference(filesList[i], rootUUID)
		}
	}

	// Wait for all pr threads to finish
	prPoolGroup.Wait()
	// Tell writer to finish
	done <- true
	// Wait for Writers to clean up
	writerPoolGroup.Wait()
	log.Debugf("Finished waiting for writer pool group to finish.")
	// Stop the process pool
	processPool.StopAndWait()

	log.Info("Found ", totalProviderCounter.Load(), " providers. Matched on ", matchedProviderCounter.Load(), " providers.")
}

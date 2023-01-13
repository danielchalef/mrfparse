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
	"encoding/csv"
	"io"
	"mrfparse/pkg/mrfparse/cloud"
	"mrfparse/pkg/mrfparse/utils"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/spf13/viper"
)

// loadServiceList loads a list of services from a csv file and returns a stringSet of the services.
// The csv file is expected to have a header row, with first column being the
// CPT/HCPCS service code, and subsequent columns being ignored.
func loadServiceList(uri string) StringSet {
	var f io.ReadCloser
	var err error
	var services StringSet = mapset.NewSet[string]()

	// if empty, get from config file
	if uri == "" {
		uri = viper.GetString("services.file")
	}

	f, err = cloud.NewReader(context.TODO(), uri)
	utils.ExitOnError(err)

	defer func(f io.ReadCloser) {
		err = f.Close()
		if err != nil {
			utils.ExitOnError(err)
		}
	}(f)

	csvReader := csv.NewReader(f)
	serviceData, err := csvReader.ReadAll()
	utils.ExitOnError(err)

	// extract the first column, the CPT/HCPCS code, from csv,
	// skipping the header row
	for _, s := range serviceData[1:] { // skip header
		services.Add(s[0])
	}

	return services
}

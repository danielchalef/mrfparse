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
	"encoding/json"
	"errors"
	"io"
	"mrfparse/pkg/mrfparse/cloud"
	"mrfparse/pkg/mrfparse/models"
	"mrfparse/pkg/mrfparse/utils"
	"path/filepath"
	"strconv"
	"strings"
)

// parseMrfRoot parses the root json doc and returns a Mrf struct
func parseMrfRoot(doc []byte, planID int64) (*models.Mrf, error) {
	var (
		root models.MrfRoot
		mrf  models.Mrf
		uuid = utils.GetUniqueID()
	)

	err := json.Unmarshal(doc, &root)
	if err != nil {
		return nil, err
	}

	if planID != -1 {
		root.PlanID = strconv.FormatInt(planID, 10)
	}

	mrf = models.Mrf{UUID: uuid, RecordType: "root", MrfRoot: root}

	return &mrf, nil
}

// WriteRoot loads the root.json file and writes it
func writeRoot(filename string, planID int64) string {
	f, err := cloud.NewReader(context.TODO(), filename)
	utils.ExitOnError(err)

	defer func(f io.ReadCloser) {
		err = f.Close()
		if err != nil {
			log.Errorf("Unable to close file: %s", err.Error())
		}
	}(f)

	doc, err := io.ReadAll(f)
	utils.ExitOnError(err)

	mrf, err := parseMrfRoot(doc, planID)
	utils.ExitOnError(err)

	err = WriteRecords([]*models.Mrf{mrf})
	utils.ExitOnError(err)

	return mrf.UUID
}

func findRootFile(filesList []string) (string, error) {
	for _, file := range filesList {
		filename := filepath.Base(file)
		if strings.Contains(filename, "root.json") {
			return file, nil
		}
	}

	return "", errors.New("root.json file not found")
}

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
package pipeline

import (
	"os"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/spf13/viper"
)

func TestNewParsePipeline(t *testing.T) {
	inputPath := "http://server.com/somepath/input.gz?somestuff"
	outputPath := "output"
	serviceFile := "service.csv"
	planID := int64(1)

	viper.Set("tmp.path", "/tmp")

	p := NewParsePipeline(inputPath, outputPath, serviceFile, planID)
	assert.Equal(t, len(p.Steps), 4)

	downloadStep, ok := p.Steps[0].(*DownloadStep)
	assert.True(t, ok)

	assert.Equal(t, downloadStep.URL, inputPath)
	assert.True(t, strings.HasPrefix(downloadStep.OutputPath, "/tmp"))

	tmpPath := downloadStep.OutputPath
	assert.False(t, strings.Contains(tmpPath, "?"))

	splitStep, ok := p.Steps[1].(*SplitStep)
	assert.True(t, ok)
	assert.Equal(t, splitStep.InputPath, tmpPath)
	assert.True(t, splitStep.Overwrite)

	parseStep, ok := p.Steps[2].(*ParseStep)
	assert.True(t, ok)
	assert.Equal(t, parseStep.OutputPath, outputPath)
	assert.Equal(t, parseStep.ServiceFile, serviceFile)
	assert.Equal(t, parseStep.PlanID, planID)

	cleanupStep, ok := p.Steps[3].(*CleanStep)
	assert.True(t, ok)
	assert.True(t, strings.HasPrefix(tmpPath, cleanupStep.TmpPath))

	err := os.RemoveAll(tmpPath)
	assert.NoError(t, err)
}

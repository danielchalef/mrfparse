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
	"github.com/danielchalef/mrfparse/pkg/mrfparse/mrf"
	"github.com/danielchalef/mrfparse/pkg/mrfparse/split"
	"github.com/danielchalef/mrfparse/pkg/mrfparse/utils"
	"os"

	"github.com/spf13/viper"
)

// NewParsePipeline returns a pipeline that splits the input file, parses the
// split files, and then cleans up afterwards.
//
// InputPath is the path to the input JSON object file.
// OutputPath is the path to the output parquet fileset.
// ServiceFile is the path to the HCPCS/CPT service file in CSV format.
// PlanID is the plan ID to use for the parquet fileset.
//
// The pipeline uses a tmp path to store the intermediate split files. The tmp
// path ican be configured in the config file, an enrivonment variable, or a
// default system tmp path will be used.
func NewParsePipeline(inputPath, outputPath, serviceFile string, planID int64) *Pipeline {
	var (
		err        error
		tmpPath    string
		steps      []Step
		cfgTmpPath = viper.GetString("tmp.path")
	)

	if cfgTmpPath != "" {
		tmpPath, err = os.MkdirTemp(cfgTmpPath, "mrfparse")
	} else {
		tmpPath, err = os.MkdirTemp("", "mrfparse")
	}

	utils.ExitOnError(err)

	steps = []Step{
		&SplitStep{
			InputPath:  inputPath,
			OutputPath: tmpPath,
			Overwrite:  true,
		},
		&ParseStep{
			InputPath:   tmpPath,
			OutputPath:  outputPath,
			ServiceFile: serviceFile,
			PlanID:      planID,
		},
		&CleanStep{
			TmpPath: tmpPath,
		},
	}

	return New(steps...)
}

// SplitStep splits the input JSON object file into NDJSON files using split.File
type SplitStep struct {
	InputPath  string
	OutputPath string
	Overwrite  bool
}

func (s *SplitStep) Run() {
	split.File(s.InputPath, s.OutputPath, s.Overwrite)
}

func (s *SplitStep) Name() string {
	return "Split"
}

// ParseStep parses the split NDJSON files into a parquet fileset using mrf.Parse
type ParseStep struct {
	InputPath   string
	OutputPath  string
	ServiceFile string
	PlanID      int64
}

func (s *ParseStep) Run() {
	mrf.Parse(s.InputPath, s.OutputPath, s.PlanID, s.ServiceFile)
}

func (s *ParseStep) Name() string {
	return "Parse"
}

// CleanStep removes the tmp directory used to store the split files
type CleanStep struct {
	TmpPath string
}

func (s *CleanStep) Run() {
	err := os.RemoveAll(s.TmpPath)
	utils.ExitOnError(err)
}

func (s *CleanStep) Name() string {
	return "Clean"
}

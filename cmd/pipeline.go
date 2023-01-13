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
package cmd

import (
	"mrfparse/pkg/mrfparse/pipeline"
	"mrfparse/pkg/mrfparse/utils"

	"github.com/spf13/cobra"
)

// pipelineCmd represents the pipeline command
var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Parse in-network MRF files. Input is a single MRF JSON file. Output is a parquet fileset.",
	Long: `Parse in-network MRF files. Input is a single MRF JSON file. Output is a parquet fileset.

- Input is a JSON MRF file. Can be located at a local, HTTP, S3, or GCS path. 
  Supports GZIPed files.
- Output is a fileset in parquet format. See README for schema.

Requires a services file containing a list of CPT/HCPCS service codes to filter on. Typically, we'd use the CMS 500 Shoppable Services list.

Plan ID is acquired from the carrier's Index file.`,
	Run: func(cmd *cobra.Command, args []string) {
		inputPath, err := cmd.Flags().GetString("input")
		utils.ExitOnError(err)

		outputPath, err := cmd.Flags().GetString("output")
		utils.ExitOnError(err)

		serviceFile, err := cmd.Flags().GetString("services")
		utils.ExitOnError(err)

		planID, err := cmd.Flags().GetInt64("planid")
		utils.ExitOnError(err)

		p := pipeline.NewParsePipeline(inputPath, outputPath, serviceFile, planID)
		p.Run()
	},
}

func init() {
	rootCmd.AddCommand(pipelineCmd)

	pipelineCmd.Flags().StringP("input", "i", "", "Input path to JSON MRF file. Can be a local, HTTP, S3, or GCS path. Supports GZIPed files.")
	err := pipelineCmd.MarkFlagRequired("input")
	utils.ExitOnError(err)

	pipelineCmd.Flags().StringP("output", "o", "", "Output path for parsed MRF fileset in parquet format")
	err = pipelineCmd.MarkFlagRequired("output")
	utils.ExitOnError(err)

	pipelineCmd.Flags().StringVarP(&servicesFile, "services", "s", "", "Path to a CSV file containing a list of CPT/HCPCS service codes to filter on")

	pipelineCmd.Flags().Int64P("planid", "p", -1, "The planid acquired from the index file")
	err = pipelineCmd.MarkFlagRequired("planid")
	utils.ExitOnError(err)
}

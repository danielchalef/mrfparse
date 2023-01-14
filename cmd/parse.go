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
	"github.com/danielchalef/mrfparse/pkg/mrfparse/mrf"
	"github.com/danielchalef/mrfparse/pkg/mrfparse/utils"

	"github.com/spf13/cobra"
)

var servicesFile string

// parseCmd represents the parseMrf command
var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse in-network MRF files. Expects split NDJSON files as input.",
	Long: `Parse in-network MRF files. Expects split NDJSON files as input.
	
parseMrf outputs a parquet fileset. See README for schema.`,
	Run: func(cmd *cobra.Command, args []string) {
		inputPath, err := cmd.Flags().GetString("input")
		utils.ExitOnError(err)

		outputPath, err := cmd.Flags().GetString("output")
		utils.ExitOnError(err)

		serviceFile, err := cmd.Flags().GetString("services")
		utils.ExitOnError(err)

		planID, err := cmd.Flags().GetInt64("planid")
		utils.ExitOnError(err)

		fn := func() { mrf.Parse(inputPath, outputPath, planID, serviceFile) }

		elapsed := utils.Timed(fn)
		log.Infof("Completed in %d seconds", elapsed)
	},
}

func init() {
	rootCmd.AddCommand(parseCmd)

	parseCmd.Flags().StringP("input", "i", "", "input path to NDJSON files")
	err := parseCmd.MarkFlagRequired("input")
	utils.ExitOnError(err)

	parseCmd.Flags().StringP("output", "o", "", "output path for parsed MRF files in parquet format")
	err = parseCmd.MarkFlagRequired("output")
	utils.ExitOnError(err)

	parseCmd.Flags().StringVarP(&servicesFile, "services", "s", "", "path to a CSV file containing a list of CPT/HCPCS service codes to filter on")

	parseCmd.Flags().Int64P("planid", "p", -1, "the planid acquired from the index file")
	err = parseCmd.MarkFlagRequired("planid")
	utils.ExitOnError(err)
}

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
	"mrfparse/pkg/mrfparse/split"
	"mrfparse/pkg/mrfparse/utils"

	"github.com/spf13/cobra"
)

// splitCmd represents the splitFile command
var splitCmd = &cobra.Command{
	Use:   "split",
	Short: "Split JSON files.",
	Long: `Split JSON files into a root.json and a series of NDJSON files for each top-level array element.
	
The input JSON file can be gzipped and may be located on the local filesystem or in a S3/GCS bucket.`,
	Run: func(cmd *cobra.Command, args []string) {
		inputPath, err := cmd.Flags().GetString("input")
		utils.ExitOnError(err)

		outputPath, err := cmd.Flags().GetString("output")
		utils.ExitOnError(err)

		overwrite, err := cmd.Flags().GetBool("overwrite")
		utils.ExitOnError(err)

		fn := func() { split.File(inputPath, outputPath, overwrite) }

		elapsed := utils.Timed(fn)
		log.Infof("Completed in %d seconds", elapsed)
	},
}

func init() {
	rootCmd.AddCommand(splitCmd)

	splitCmd.Flags().StringP("input", "i", "", "input path to JSON file.")
	err := splitCmd.MarkFlagRequired("input")
	utils.ExitOnError(err)

	splitCmd.Flags().StringP("output", "o", "", "output path for split NDJSON files")
	err = splitCmd.MarkFlagRequired("output")
	utils.ExitOnError(err)

	splitCmd.Flags().Bool("overwrite", false, "overwrite contents of output path if it exists")
}

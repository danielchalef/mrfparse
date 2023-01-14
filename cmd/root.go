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
	"github.com/danielchalef/mrfparse/pkg/mrfparse/utils"
	"os"
	"runtime/pprof"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	log            *logrus.Logger
	cfgFile        string
	memProfileFile string
	cpuProfileFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mrfparse",
	Short: "A parser for Transparency in Coverage Machine Readable Format (MRF) files",
	Long: `MRFParse is a Go parser for Transparency in Coverage Machine Readable Format (MRF) files. 
	
The parser is designed to be memory and CPU efficient, and easily containerized. It will run on any modern cloud container platform (and potentially cloud function infrastructure).

Input and Output paths can be local filesytem paths or AWS S3 and Google Cloud Storage paths.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cpuProfileFile != "" {
			f, err := os.Create(cpuProfileFile)
			if err != nil {
				log.Fatal(err)
			}
			if err := pprof.StartCPUProfile(f); err != nil {
				log.Fatal(err)
			}
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if cpuProfileFile != "" {
			pprof.StopCPUProfile()
		}
		if memProfileFile != "" {
			f, err := os.Create(memProfileFile)
			if err != nil {
				log.Fatal(err)
			}
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal(err)
			}
			f.Close()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default config.yaml)")
	rootCmd.PersistentFlags().StringVar(&memProfileFile, "memprofile", "", "Write memory profile to this file")
	rootCmd.PersistentFlags().StringVar(&cpuProfileFile, "cpuprofile", "", "Write CPU profile to this file")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix("mrf")                              // ENV variables will be prefixed with MRF_
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`)) // replaced nested . with _
	viper.AutomaticEnv()                                   // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	// Initialize or update logger with level from ENV or config
	log = utils.GetLogger()
}

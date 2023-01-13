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
package utils

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var once sync.Once
var logger *logrus.Logger

var log = GetLogger()

func GetLogger() *logrus.Logger {
	var level logrus.Level

	level = logrus.InfoLevel

	if viper.IsSet("log.level") {
		switch viper.GetString("log.level") {
		case "debug":
			level = logrus.DebugLevel
		case "warn":
			level = logrus.WarnLevel
		case "error":
			level = logrus.ErrorLevel
		case "trace":
			level = logrus.TraceLevel
		}
	}

	// Use a singleton so we can update log level once config is loaded
	once.Do(func() {
		logger = logrus.New()
	})

	logger.Out = os.Stdout
	logger.SetLevel(level)

	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	return logger
}

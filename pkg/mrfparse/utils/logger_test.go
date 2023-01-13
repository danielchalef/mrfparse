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
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// test logger config
func TestLoggerConfig(t *testing.T) {
	log := GetLogger()
	assert.Equal(t, log.Level, logrus.InfoLevel)

	viper.Set("log.level", "debug")
	log = GetLogger()
	assert.Equal(t, log.Level, logrus.DebugLevel)

	viper.Set("log.level", "info")
	log = GetLogger()
	assert.Equal(t, log.Level, logrus.InfoLevel)

	viper.Set("log.level", "error")
	log = GetLogger()
	assert.Equal(t, log.Level, logrus.ErrorLevel)

	viper.Set("log.level", "warn")
	log = GetLogger()
	assert.Equal(t, log.Level, logrus.WarnLevel)
}

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
package parquet

import (
	"context"
	"github.com/danielchalef/mrfparse/pkg/mrfparse/models"
	"github.com/danielchalef/mrfparse/pkg/mrfparse/utils"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/spf13/viper"
)

func TestPqWriteCloserURIIncrement(t *testing.T) {
	const (
		expectedURIZero = "/tmp/mrf_0000.zstd.parquet"
		expectedURIOne  = "/tmp/mrf_0001.zstd.parquet"
	)

	pwf := NewPqWriterFactory("mrf", "/tmp/")
	pwc, err := pwf.CreateWriter(context.TODO())
	assert.NoError(t, err)

	err = pwc.Close()
	assert.NoError(t, err)

	assert.Equal(t, expectedURIZero, pwc.URI())

	pwc, err = pwf.CreateWriter(context.TODO())
	assert.NoError(t, err)

	err = pwc.Close()
	assert.NoError(t, err)

	assert.Equal(t, expectedURIOne, pwc.URI())
}

// test writing data to a parquet file
func TestPqWriteCloserWrite(t *testing.T) {
	var mrfList = []*models.Mrf{{UUID: utils.GetUniqueID(), ParentUUID: utils.GetUniqueID()},
		{UUID: utils.GetUniqueID(), ParentUUID: utils.GetUniqueID()}}

	pwf := NewPqWriterFactory("mrf", "/tmp/")
	pwc, err := pwf.CreateWriter(context.TODO())
	assert.NoError(t, err)

	rows, err := pwc.Write(mrfList)
	assert.NoError(t, err)

	err = pwc.Close()
	assert.NoError(t, err)

	assert.Equal(t, 2, rows)
}

// test NewPqWriterFactory config
func TestNewPqWriterFactoryConfig(t *testing.T) {
	const (
		expectedMaxRowsPerFile  = 555
		expectedMaxRowsPerGroup = 666
		expectedOutputTemplate  = "_mrf.parquet"
	)

	viper.Set("writer.max_rows_per_file", expectedMaxRowsPerFile)
	viper.Set("writer.filename_template", expectedOutputTemplate)
	viper.Set("writer.max_rows_per_group", expectedMaxRowsPerGroup)

	pwf := NewPqWriterFactory("file", "/tmp/output")
	assert.Equal(t, expectedMaxRowsPerFile, pwf.MaxRowsPerFile)
	assert.Equal(t, expectedMaxRowsPerGroup, pwf.MaxRowsPerGroup)
	assert.Equal(t, "/tmp/output/file"+expectedOutputTemplate, pwf.filenameTemplate)
}

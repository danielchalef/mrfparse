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
	"fmt"
	"io"
	"mrfparse/pkg/mrfparse/cloud"
	"mrfparse/pkg/mrfparse/models"

	"github.com/segmentio/parquet-go"
	"github.com/spf13/viper"
)

// PqWriteCloser is a wrapper around parquet.GenericWriter and io.WriteCloser
type PqWriteCloser struct {
	closer io.WriteCloser
	ctx    context.Context
	writer *parquet.GenericWriter[*models.Mrf]
	uri    string
}

// Close closes the underlying parquet.GenericWriter and the underlying io.WriteCloser.
func (pwc *PqWriteCloser) Close() error {
	err := pwc.writer.Close()
	if err != nil {
		return err
	}

	err = pwc.closer.Close()
	if err != nil {
		return err
	}

	return nil
}

// Write writes the given data to the underlying parquet.GenericWriter.
func (pwc *PqWriteCloser) Write(rows []*models.Mrf) (int, error) {
	return pwc.writer.Write(rows)
}

// Flush flushes the underlying parquet.GenericWriter.
func (pwc *PqWriteCloser) Flush() error {
	return pwc.writer.Flush()
}

// URI returns the composed URI of the underlying io.WriteCloser.
func (pwc *PqWriteCloser) URI() string {
	return pwc.uri
}

// NewPqWriter creates a new PqWriteCloser. ctx is the context to use for the underlying io.WriteCloser, making it
// possible to cancel the write operation.
func NewPqWriter(ctx context.Context, uri string, maxRowsPerGroup int64) (*PqWriteCloser, error) {
	pqConfig := parquet.WriterConfig{Compression: &parquet.Zstd, MaxRowsPerRowGroup: maxRowsPerGroup}

	w, err := cloud.NewWriter(ctx, uri)
	if err != nil {
		return nil, err
	}

	writer := parquet.NewGenericWriter[*models.Mrf](w, &pqConfig)

	return &PqWriteCloser{uri: uri, writer: writer, closer: w, ctx: ctx}, nil
}

// PqWriterFactory is a factory for creating PqWriteClosers.
// It is used to create a new PqWriteCloser when the number of rows written to the current PqWriteCloser exceeds
// MaxRowsPerFile. The URI of the new PqWriteCloser is created by incrementing the fileIndex and formatting it
// into the filenameTemplate.
type PqWriterFactory struct {
	filenameTemplate string
	fileIndex        int
	MaxRowsPerFile   int
	MaxRowsPerGroup  int64
}

// NewPqWriterFactory creates a new PqWriterFactory. filePrefix is the prefix of the filename (e.g. "mrf"), outputURI
// is the URI of the output directory (e.g. "gs://bucket/output").
func NewPqWriterFactory(filePrefix, outputURI string) *PqWriterFactory {
	var (
		DefaultMaxRowsPerFile       = 100_000_000
		MaxRowsPerGroup       int64 = 1_000_000
		DefaultOutputTemplate       = "_%04d.zstd.parquet"
	)

	if viper.IsSet("writer.max_rows_per_file") {
		DefaultMaxRowsPerFile = viper.GetInt("writer.max_rows_per_file")
	}

	if viper.IsSet("writer.filename_template") {
		DefaultOutputTemplate = viper.GetString("writer.filename_template")
	}

	if viper.IsSet("writer.max_rows_per_group") {
		MaxRowsPerGroup = viper.GetInt64("writer.max_rows_per_group")
	}

	filenameTemplate := cloud.JoinURI(outputURI, filePrefix) + DefaultOutputTemplate

	return &PqWriterFactory{
		filenameTemplate: filenameTemplate,
		fileIndex:        0,
		MaxRowsPerFile:   DefaultMaxRowsPerFile,
		MaxRowsPerGroup:  MaxRowsPerGroup,
	}
}

// CreateWriter creates a new PqWriteCloser. fileIndex is incremented by each call to CreateWriter, and the
// filenameTemplate is formatted with the new fileIndex to create the URI of the new PqWriteCloser.
func (pwf *PqWriterFactory) CreateWriter(ctx context.Context) (*PqWriteCloser, error) {
	uri := fmt.Sprintf(pwf.filenameTemplate, pwf.fileIndex)
	pwf.fileIndex++

	w, err := NewPqWriter(ctx, uri, pwf.MaxRowsPerGroup)
	if err != nil {
		return nil, err
	}

	return w, nil
}

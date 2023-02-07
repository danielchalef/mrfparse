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
package http

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/avast/retry-go/v4"
	"github.com/danielchalef/mrfparse/pkg/mrfparse/utils"
	"github.com/spf13/viper"

	"time"
)

const MaxRetryAttempts = 10

var log = utils.GetLogger()

// DownloadFileReader downloads a file from the given URL and returns an io.ReadCloser.
// The caller is responsible for closing the returned io.ReadCloser.
// DownloadFilereader attempts to retry the download if it receives a RetryAfterDelay error.
func DownloadReader(fileURL string) (io.ReadCloser, error) {
	var (
		err         error
		r           *http.Response
		HTTPTimeOut time.Duration
	)

	HTTPTimeOut = time.Duration(viper.GetInt("pipeline.download_timeout")) * time.Minute
	log.Info("HTTP timeout set to ", HTTPTimeOut)

	var httpClient = &http.Client{
		Timeout: HTTPTimeOut,
	}

	err = retry.Do(func() error {
		r, err = httpClient.Get(fileURL) //nolint:bodyclose // Embedded in retry confusing linter
		if err != nil {
			return err
		}
		return nil
	}, retry.DelayType(RetryAfterDelay),
		retry.Attempts(MaxRetryAttempts),
	)
	if err != nil {
		r.Body.Close()

		return nil, fmt.Errorf("unable to download file from %s: %s", fileURL, errors.Unwrap(err))
	}

	if r.StatusCode != http.StatusOK {
		errorText := fmt.Errorf("bad status downloading %s: %s", fileURL, r.Status)
		log.Error(errorText)

		return nil, errorText
	}

	return r.Body, nil
}

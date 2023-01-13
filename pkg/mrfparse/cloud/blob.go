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
package cloud

import (
	"context"
	"errors"
	"io"
	"mrfparse/pkg/mrfparse/utils"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob" // required by CDK as blob driver
	_ "gocloud.dev/blob/gcsblob"  // required by CDK as blob driver
	_ "gocloud.dev/blob/s3blob"   // required by CDK as blob driver
)

var log = utils.GetLogger()

// OpenBucket opens a blob storage bucket at the URI. Context can be used to cancel any operations.
// Google CDK is used to support both AWS S3 and Google Cloud Storage. Use the correct URI scheme to
// specify the storage provider (gs:// or s3://).
func OpenBucket(ctx context.Context, uri string) (*blob.Bucket, error) {
	var (
		err error
		b   *blob.Bucket
	)

	b, err = blob.OpenBucket(ctx, uri)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// NewWriter creates a new io.WriteCloser for the given URI. Context can be used to cancel any operations.
// Google Cloud Storage, AWS S3, and local filesystem URIs are supported. Use the correct URI scheme for
// the storage provider (gs://, s3://) or no scheme for local filesystem.
func NewWriter(ctx context.Context, uri string) (io.WriteCloser, error) {
	const (
		flags = os.O_CREATE | os.O_WRONLY
		perms = 0o644
	)

	var (
		err error
		k   string
	)

	if !IsCloudURI(uri) {
		return os.OpenFile(uri, flags, perms)
	}

	_, _, k, err = ParseBlobURI(uri)
	if err != nil {
		return nil, err
	}

	b, err := OpenBucket(ctx, uri)
	if err != nil {
		return nil, err
	}

	return b.NewWriter(ctx, k, nil)
}

// NewReader creates a new io.ReadCloser for the given URI. Context can be used to cancel any operations.
// Google Cloud Storage, AWS S3, and local filesystem URIs are supported. Use the correct URI scheme for
// the storage provider (gs://, s3://) or no scheme for local filesystem.
// The URI must be a file, not a directory.
func NewReader(ctx context.Context, uri string) (io.ReadCloser, error) {
	var (
		err error
		k   string
	)

	if !IsCloudURI(uri) {
		return os.Open(uri)
	}

	_, _, k, err = ParseBlobURI(uri)
	if err != nil {
		return nil, err
	}

	b, err := OpenBucket(ctx, uri)
	if err != nil {
		return nil, err
	}

	return b.NewReader(ctx, k, nil)
}

// JoinURI joins two URI parts together, removing any trailing slashes from the left part and any
// leading slashes from the right part.
func JoinURI(left, right string) string {
	return strings.TrimRight(left, "/") + "/" + strings.TrimLeft(right, "/")
}

// Glob enumerates cloud storage objects/file names at a URI and returns a list of objects/ filename URIs that match the given pattern.
// Context can be used to cancel any cloud operations.
// Google Cloud Storage, AWS S3, and local filesystem URIs are supported. Use the correct URI scheme for
// the storage provider (gs://, s3://) or no scheme for local filesystem.
// The pattern is a glob pattern, not a regular expression.
func Glob(ctx context.Context, uri, pattern string) ([]string, error) {
	var (
		matches []string
		err     error
	)

	// The path is a local filesystem path
	if !IsCloudURI(uri) {
		matches, err = filepath.Glob(filepath.Join(uri, pattern))
		if err != nil {
			return nil, err
		}

		return matches, nil
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	b, err := OpenBucket(context.Background(), uri)
	if err != nil {
		return nil, err
	}

	iter := b.List(&blob.ListOptions{Prefix: u.Path})

	for {
		obj, err := iter.Next(ctx)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		k := filepath.Base(obj.Key)
		log.Debugf("Key is %s", k)

		if matched, _ := filepath.Match(pattern, k); matched {
			matches = append(matches, u.Scheme+u.Host+"/"+obj.Key)
		}
	}
	log.Debugf("Found %d matches for %s", len(matches), pattern)

	return matches, nil
}

// ParseBlobURI parses a URI into its scheme, bucket, and key components.
func ParseBlobURI(uri string) (scheme, bucket, key string, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", "", "", err
	}

	return u.Scheme, u.Host, strings.TrimLeft(u.Path, "/"), nil
}

// IsCloudURI returns true if the URI is a cloud storage URI (gs:// or s3://).
// It does so by attempting to parse the URI and checking if the scheme is non-empty.
func IsCloudURI(uri string) bool {
	s, _, _, err := ParseBlobURI(uri)
	if err != nil {
		return false
	}

	return s != ""
}

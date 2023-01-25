/*
MIT License

# Copyright (c) 2020 aereal

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package http

import (
	"reflect"
	"testing"
	"time"
)

func TestParseSeconds(t *testing.T) {
	cases := []struct {
		name    string
		args    string
		want    time.Duration
		wantErr bool
	}{
		{"ok", "60", time.Minute * 1, false},
		{"invalid", "", time.Duration(0), true},
		{"negative", "-10", time.Duration(0), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseSeconds(c.args)
			if (err != nil) != c.wantErr {
				t.Errorf("error = %v, wantErr %v", err, c.wantErr)
				return
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got = %v, want = %v", got, c.want)
			}
		})
	}
}

func TestParseHTTPDate(t *testing.T) {
	now := time.Now()
	// orig := nowFunc
	// nowFunc = func() time.Time { return now }
	// defer func() {
	// 	nowFunc = orig
	// }()
	aMinuteLater := now.Add(time.Minute)

	cases := []struct {
		name    string
		args    string
		want    time.Time
		wantErr bool
	}{
		{"ok", aMinuteLater.Format(time.RFC1123), aMinuteLater, false},
		{"invalid format", "2020-01-02", time.Time{}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseHTTPDate(c.args)
			if (err != nil) != c.wantErr {
				t.Errorf("error = %v, wantErr %v", err, c.wantErr)
				return
			}
			if got.Unix() != c.want.Unix() {
				t.Errorf("got = %s, want %s", got, c.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	now := time.Now()
	aMinuteLater := now.Add(time.Minute)

	cases := []struct {
		name    string
		args    string
		want    time.Time
		wantErr bool
	}{
		{"seconds/ok", "60", aMinuteLater, false},
		{"http date/ok", aMinuteLater.Format(time.RFC1123), aMinuteLater, false},
		{"invalid", "", time.Time{}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseRetryAfter(c.args)
			if (err != nil) != c.wantErr {
				t.Errorf("error = %v, wantErr %v", err, c.wantErr)
				return
			}
			if got.Unix() != c.want.Unix() {
				t.Errorf("got = %s, want %s", got, c.want)
			}
		})
	}
}

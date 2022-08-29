/*
Copyright © 2022 Scott Hawkins <scott@echoboomer.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package common

import (
	"os"
	"reflect"
	"testing"

	"github.com/spf13/afero"
)

var testFilePath afero.File

func TestDeleteFromSlice(t *testing.T) {
	type args struct {
		slice    []string
		selector string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "A slice provided as an argument should return the slice without the requested element to be deleted",
			args: args{
				slice:    []string{"foo", "bar", "baz"},
				selector: "bar",
			},
			want: []string{"foo", "baz"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeleteFromSlice(tt.args.slice, tt.args.selector); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteFromSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	// Mock Vault secret path
	appFS := afero.NewMemMapFs()
	appFS.MkdirAll("/tmp/testfile", 0755)
	afero.WriteFile(appFS, "/tmp/testfile", []byte("test value"), 0755)
	testFilePath, _ = appFS.OpenFile("/tmp/testfile", os.O_RDONLY, 0755)

	type args struct {
		fs       afero.Fs
		filename string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "If a file exists, should return true",
			args: args{
				fs:       appFS,
				filename: testFilePath.Name(),
			},
			want: true,
		},
		{
			name: "If a file does not exist, should return false",
			args: args{
				fs:       appFS,
				filename: "/some/file",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileExists(tt.args.fs, tt.args.filename); got != tt.want {
				t.Errorf("FileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindInSlice(t *testing.T) {
	type args struct {
		slice   []string
		element string
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 bool
	}{
		{
			name: "FindInSlice should return the index of an item/true if present in a slice",
			args: args{
				slice:   []string{"foo", "bar", "baz"},
				element: "bar",
			},
			want:  1,
			want1: true,
		},
		{
			name: "FindInSlice should return -1/false if an item in a slice is not present",
			args: args{
				slice:   []string{"foo", "bar", "baz"},
				element: "biz",
			},
			want:  -1,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := FindInSlice(tt.args.slice, tt.args.element)
			if got != tt.want {
				t.Errorf("FindInSlice() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("FindInSlice() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

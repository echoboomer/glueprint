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
	"fmt"

	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/afero"
)

// DeleteFromSlice removes the supplied single element from a given slice
func DeleteFromSlice(slice []string, selector string) []string {
	var result []string
	for _, element := range slice {
		if element != selector {
			result = append(result, element)
		}
	}
	return result
}

// FileExists returns whether or not the given file exists in the OS
func FileExists(fs afero.Fs, filename string) bool {
	_, err := fs.Stat(filename)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// FindInSlice returns whether or not the given slice contains the supplied element
// and, if so, returns its index
func FindInSlice(slice []string, element string) (int, bool) {
	for i, item := range slice {
		if item == element {
			return i, true
		}
	}
	return -1, false
}

// GPHeader returns a header for the app
func GPHeader() {
	myFigure := figure.NewFigure("glueprint", "small", true)
	myFigure.Print()
	fmt.Printf("\n")
}

// vim:ts=4:sw=4:noexpandtab
// © 2012-2014 Michael Stapelberg (see also: LICENSE)
package godebiancontrol_test

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stapelberg/godebiancontrol"
)

// Open and read file from path.
func readFile(path string) (*bytes.Buffer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file '%s'", path)
	}
	defer f.Close()

	finfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file '%s'", path)
	}

	buf := make([]byte, finfo.Size())
	_, err = f.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s'", path)
	}

	return bytes.NewBuffer(buf), nil
}

func TestParse(t *testing.T) {
	testCases := []struct {
		testFile      string
		expectedCount int
		title         string
	}{
		{
			"test-files/01",
			1,
			"control file for bti package",
		},
		{
			"test-files/02",
			1,
			"control file for i3-wm control file",
		},
		{
			"test-files/03",
			12,
			"mirror Packages file",
		},
		{
			"test-files/04",
			1,
			"mirror release file",
		},
	}

	for i, tc := range testCases {
		buff, err := readFile(tc.testFile)
		if err != nil {
			t.Errorf(
				"skipping test ID: %d, title: %s, failed to read test file, %s",
				i, tc.title, err,
			)
		}
		inpStr := buff.String()

		paragraphs, err := godebiancontrol.Parse(buff)
		if err != nil {
			t.Errorf(
				"failed to parse input data, test ID '%d', test title '%s', err: %s",
				i, tc.title, err,
			)
			continue
		}

		if tc.expectedCount != len(paragraphs) {
			t.Errorf(
				"wrong number of paragraph detected, expected: %d, result: %d",
				tc.expectedCount, len(paragraphs),
			)
		}

		resStr := godebiancontrol.ParagraphsToText(paragraphs)

		if resStr != inpStr {
			inpSlice := strings.Split(inpStr, "\n")
			resSlice := strings.Split(resStr, "\n")

			if len(inpSlice) != len(resSlice) {
				t.Errorf(
					"content length mismatch, input '%d', output '%d', "+
						"test ID '%d', test title '%s'",
					len(inpSlice), len(resSlice), i, tc.title,
				)
			}

			for i, v := range inpSlice {
				if i >= len(resSlice) {
					t.Errorf(
						"skipping iteration because output is shorter, test ID '%d', test title '%s'",
						i, tc.title,
					)
					break
				}
				vRes := resSlice[i]
				fmt.Printf("u: %d, input '%s', output: '%s'\n", i, v, vRes)
				if v != vRes {
					t.Errorf(
						"line mismatch:\n\tinput: '%s'\n\toutput: '%s'\n",
						v, vRes,
					)
				}
			}
			t.Errorf(
				"paragraph converted to string is different as test data, test ID: '%d', test title: '%s'",
				i, tc.title,
			)
		}
	}

}

func ExampleParse() {
	file, err := os.Open("debian-mirror/dists/sid/main/source/Sources")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	paragraphs, err := godebiancontrol.Parse(file)
	if err != nil {
		log.Fatal(err)
	}

	// Print a list of which source package uses which package format.
	for _, pkg := range paragraphs {
		fmt.Printf("%s uses %s\n", pkg.Get("Package"), pkg.Get("Format"))
	}
}

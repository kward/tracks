package actions

import (
	"fmt"
	"io/ioutil"
)

var (
	fnReadDir = ioutil.ReadDir
)

type DiscoverFilesFn func(dir string, filters ...Filter) ([]string, error)

// DiscoverFiles looks for track names in a given directory, and returns them
// as a slice.
func DiscoverFiles(dir string, filters ...Filter) ([]string, error) {
	fileInfos, err := fnReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := []string{}
	for _, fi := range fileInfos {
		files = append(files, fi.Name())
	}
	for _, filter := range filters {
		files = filter(files)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in %q", dir)
	}

	return files, nil
}

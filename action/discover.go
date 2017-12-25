package action

import (
	"fmt"
	"io/ioutil"
)

var (
	fnReadDir = ioutil.ReadDir
)

// DiscoverFiles looks for track names in a given directory, and returns them
// as a slice.
func DiscoverFiles(dir string, filters ...Filter) ([]string, error) {
	fileInfos, err := fnReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error discovering files in %q; %s", dir, err)
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

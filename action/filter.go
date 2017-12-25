package action

import "strings"

// Filter defines a file filter function.
type Filter func(unfiltered []string) (filtered []string)

func FilterWaves(unfiltered []string) []string {
	filtered := []string{}
	for _, f := range unfiltered {
		if strings.HasSuffix(f, ".wav") {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

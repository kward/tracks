package commands

import (
	"fmt"
	"testing"

	"github.com/kward/tracks/actions"
)

func TestIssue7(t *testing.T) {
	setup()

	// Prepare track names.
	discoverFilesFn = func(_ string, _ ...actions.Filter) ([]string, error) {
		files := []string{}
		for i := 1; i <= 64; i++ {
			files = append(files, fmt.Sprintf("Track %02d-1.wav", i))
		}
		return files, nil
	}

	// Read patch list.
	names, err := venueNames(VenueFlags{
		dryRun:    true,
		patchFile: "../testdata/20180128 Avid S3L-X Patch List.html",
	})
	if err != nil {
		t.Fatalf("%s", err)
	}

	// Create map of names for easy lookup.
	nameMap := make(map[string]string)
	for _, name := range names {
		nameMap[name.orig] = name.dest
	}

	// Test.
	for _, tt := range []struct {
		src, dest string
	}{
		{"Track 13-1.wav", "01-13 Track 13.wav"},
		{"Track 14-1.wav", "01-14 Track 14.wav"},
		{"Track 16-1.wav", "01-16 Track 16.wav"},
		{"Track 18-1.wav", "01-18 Track 18.wav"},
		{"Track 29-1.wav", "01-29 vFlorina.wav"},
		{"Track 30-1.wav", "01-30 vLaura.wav"},
		{"Track 32-1.wav", "01-32 vGloria.wav"},
		{"Track 34-1.wav", "01-34 Producer.wav"},
		{"Track 63-1.wav", "01-63 Left -23 LUFS (direct out).wav"},
		{"Track 64-1.wav", "01-64 Right (direct out).wav"},
	} {
		if got, want := nameMap[tt.src], tt.dest; got != want {
			t.Errorf("%q: incorrect file name %q, want %q", tt.src, got, want)
		}
	}
}

func setup() {
	resetDiscoverFiles()
}

package actions

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	k8os "github.com/kward/golib/os"
)

func init() {
	fnReadDir = mockReadDir
}

func TestDiscoverFiles(t *testing.T) {
	for _, tt := range []struct {
		desc    string
		dir     string
		files   []string
		filters []Filter
		ok      bool
	}{
		{"no filter",
			"somedir",
			[]string{"Track 01-1.wav", "Track 02-1.wav", "Track 03-1.mp3", "Track 01-2.wav", "Track 02-2.wav", "Track 01-3.mp3"},
			[]Filter{},
			true},
		{"wave filter",
			"somedir",
			[]string{"Track 01-1.wav", "Track 02-1.wav", "Track 01-2.wav", "Track 02-2.wav"},
			[]Filter{FilterWaves},
			true},
		{"readdir error", "error", nil, []Filter{FilterWaves}, false},
	} {
		files, err := DiscoverFiles(tt.dir, tt.filters...)
		if err == nil && !tt.ok {
			t.Errorf("%s: DiscoverFiles() expected error", tt.desc)
		}
		if err != nil && tt.ok {
			t.Fatalf("%s: DiscoverFiles() unexpected error, %s", tt.desc, err)
		}
		if !tt.ok {
			continue
		}
		if got, want := files, tt.files; !reflect.DeepEqual(got, want) {
			t.Errorf("DiscoverFiles() = %q, want %q", got, want)
		}
	}
}

// mockReadDir returns a list of files found in some mock directory. The default
// file type produced by Waves Tracks is Wave (RF64), which has a `.wav` file
// extension. The other types are present to test filtering.
func mockReadDir(dir string) ([]os.FileInfo, error) {
	switch dir {
	case "error":
		return nil, fmt.Errorf("MockReadDir() error.")
	default:
		return []os.FileInfo{
			&k8os.MockFileInfo{MockName: "Track 01-1.wav"},
			&k8os.MockFileInfo{MockName: "Track 02-1.wav"},
			// General test for non-.wav filtering.
			&k8os.MockFileInfo{MockName: "Track 03-1.mp3"},
			&k8os.MockFileInfo{MockName: "Track 01-2.wav"},
			&k8os.MockFileInfo{MockName: "Track 02-2.wav"},
			// If .mp3 are filtered, no 3rd session should be present for this file.
			&k8os.MockFileInfo{MockName: "Track 01-3.mp3"},
		}, nil
	}
}

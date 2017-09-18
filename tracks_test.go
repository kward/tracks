package main

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	kos "github.com/kward/golib/os"
	"github.com/kward/tracks/tracks"
	"github.com/kward/tracks/venue"
	"github.com/kward/tracks/venue/hardware"
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
			[]Filter{filterWaves},
			true},
		{"readdir error", "error", nil, []Filter{filterWaves}, false},
	} {
		got, err := DiscoverFiles(tt.dir, tt.filters...)
		if err == nil && !tt.ok {
			t.Errorf("%s: DiscoverFiles() expected error", tt.desc)
		}
		if err != nil && tt.ok {
			t.Fatalf("%s: DiscoverFiles() unexpected error; %s", tt.desc, err)
		}
		if !tt.ok {
			continue
		}
		if want := tt.files; !reflect.DeepEqual(got, want) {
			t.Errorf("DiscoverFiles() = %q, want %q", got, want)
		}
	}
}

func TestMapTrackToChannel(t *testing.T) {
	devs := mockDevices()
	for _, tt := range []struct {
		desc    string
		track   *tracks.Track
		device  *venue.Device
		channel *venue.Channel
	}{
		{"track one",
			tracks.NewTrack("Track", 1, 1),
			devs["Stage 1"],
			venue.NewChannel("1", "iOne")},
		{"track eight",
			tracks.NewTrack("Track", 8, 1),
			devs["Stage 2"],
			venue.NewChannel("4", "iEight")},
	} {
		device, channel, err := mapTrackToChannel(tt.track, devs)
		if err != nil {
			t.Errorf("%s: unexpected error; %s", tt.desc, err)
			continue
		}
		if got, want := device.Name(), tt.device.Name(); got != want {
			t.Errorf("%s: stage box = %s, want %s", tt.desc, got, want)
			continue
		}
		if got, want := channel, tt.channel; !got.Equal(want) {
			t.Errorf("%s: channel = %s, want %s", tt.desc, got, want)
			continue
		}
	}
}

func mockDevices() venue.Devices {
	return venue.Devices{
		"Stage 1": venue.NewDevice(
			hardware.StageBox,
			"Stage 1",
			venue.Channels{
				"1": venue.NewChannel("1", "iOne"),
				"2": venue.NewChannel("2", "iTwo"),
				"3": venue.NewChannel("3", "iThree"),
				"4": venue.NewChannel("4", "iFour")},
			venue.Channels{
				"1": venue.NewChannel("1", "oOne"),
				"2": venue.NewChannel("2", "oTwo")},
		),
		"Stage 2": venue.NewDevice(
			hardware.StageBox,
			"Stage 2",
			venue.Channels{
				"1": venue.NewChannel("1", "iFive"),
				"2": venue.NewChannel("2", "iSix"),
				"3": venue.NewChannel("3", "iSeven"),
				"4": venue.NewChannel("4", "iEight")},
			venue.Channels{
				"1": venue.NewChannel("1", "oThree"),
				"2": venue.NewChannel("2", "oFour")},
		),
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
			&kos.MockFileInfo{MockName: "Track 01-1.wav"},
			&kos.MockFileInfo{MockName: "Track 02-1.wav"},
			// General test for non-.wav filtering.
			&kos.MockFileInfo{MockName: "Track 03-1.mp3"},
			&kos.MockFileInfo{MockName: "Track 01-2.wav"},
			&kos.MockFileInfo{MockName: "Track 02-2.wav"},
			// If .mp3 are filtered, no 3rd session should be present for this file.
			&kos.MockFileInfo{MockName: "Track 01-3.mp3"},
		}, nil
	}
}

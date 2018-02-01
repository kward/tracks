package actions

import (
	"testing"

	"github.com/kward/tracks/tracks"
	"github.com/kward/tracks/venue"
	"github.com/kward/tracks/venue/hardware"
)

func TestMapTrackToChannel(t *testing.T) {
	devs := mockDevices()
	for _, tt := range []struct {
		desc    string
		track   *tracks.Track
		channel *venue.Channel
	}{
		{"track one",
			tracks.NewTrack("Track", 1, 1),
			venue.NewChannel("1", "iOne")},
		{"track eight",
			tracks.NewTrack("Track", 8, 1),
			venue.NewChannel("4", "iEight")},
	} {
		channel, err := mapTrackToChannel(tt.track, devs)
		if err != nil {
			t.Errorf("%s: unexpected error; %s", tt.desc, err)
			continue
		}
		if got, want := channel, tt.channel; !got.Equal(want) {
			t.Errorf("%s: channel = %s, want %s", tt.desc, got, want)
			continue
		}
	}
}

func TestIssue7(t *testing.T) {
	// In Issue #7, tracks from Stage 2 were mapped into positions on Stage 1 if
	// the track was unnamed on Stage 1. Strangely, they were also mapped into the
	// correct Stage 2 position too.
	devs := venue.Devices{
		"Stage 1": venue.NewDevice(
			hardware.StageBox,
			"Stage 1",
			venue.Channels{
				"1": venue.NewChannel("1", "iOne"),
				"2": venue.NewChannel("2", "")},
			venue.Channels{},
		),
		"Stage 2": venue.NewDevice(
			hardware.StageBox,
			"Stage 2",
			venue.Channels{
				"1": venue.NewChannel("1", "iThree"),
				"2": venue.NewChannel("2", "iFour")},
			venue.Channels{},
		),
	}

	for _, tt := range []struct {
		desc    string
		track   *tracks.Track
		channel *venue.Channel
	}{
		{"track two",
			tracks.NewTrack("Track", 2, 1),
			venue.NewChannel("2", "")},
		{"track four",
			tracks.NewTrack("Track", 4, 1),
			venue.NewChannel("2", "iFour")},
	} {
		channel, err := mapTrackToChannel(tt.track, devs)
		if err != nil {
			t.Errorf("%s: unexpected error; %s", tt.desc, err)
			continue
		}
		if got, want := channel, tt.channel; !got.Equal(want) {
			t.Errorf("%s: channel = %s, want %s", tt.desc, got, want)
			continue
		}
	}
}

func TestMapTrackNameToFilename(t *testing.T) {
	for _, tt := range []struct {
		desc     string
		name     string
		filename string
	}{
		{"clean", "abc123", "abc123"},
		{"unix separator", "abc/123", "abc_123"},
		{"windows separator", "abc\\123", "abc_123"},
		{"empty", "", ""},
	} {
		if got, want := MapTrackNameToFilename(tt.name), tt.filename; got != want {
			t.Errorf("%s: MapTrackNameToFilename(%s) = %s, want %s", tt.desc, tt.name, got, want)
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

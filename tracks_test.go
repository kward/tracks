package main

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
		device  *venue.Device
		channel *venue.Channel
	}{
		{"track one", tracks.NewTrack("Track", 1, 1), devs["Stage 1"], venue.NewChannel("1", "iOne")},
		{"track eight", tracks.NewTrack("Track", 8, 1), devs["Stage 2"], venue.NewChannel("4", "iEight")},
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

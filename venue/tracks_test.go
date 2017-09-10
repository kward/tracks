package venue

import (
	"testing"

	"github.com/kward/tracks/venue/hardware"
)

func TestMapTrackToChannel(t *testing.T) {
	devs := mockDevices()
	for _, tt := range []struct {
		desc    string
		track   *Track
		device  *Device
		channel *Channel
	}{
		{"track one", NewTrack("Track", 1, 1), devs["Stage 1"], &Channel{"1", "iOne"}},
		{"track eight", NewTrack("Track", 8, 1), devs["Stage 2"], &Channel{"4", "iEight"}},
	} {
		device, channel, err := mapTrackToChannel(tt.track, devs)
		if err != nil {
			t.Errorf("%s: unexpected error; %s", tt.desc, err)
			continue
		}
		if got, want := device.name, tt.device.name; got != want {
			t.Errorf("%s: stage box = %s, want %s", tt.desc, got, want)
			continue
		}
		if got, want := channel, tt.channel; !got.Equal(want) {
			t.Errorf("%s: channel = %s, want %s", tt.desc, got, want)
			continue
		}
	}
}

func mockDevices() Devices {
	return Devices{
		"Stage 1": &Device{
			hardware.StageBox,
			"Stage 1",
			Channels{
				"1": &Channel{"1", "iOne"},
				"2": &Channel{"2", "iTwo"},
				"3": &Channel{"3", "iThree"},
				"4": &Channel{"4", "iFour"}},
			Channels{
				"1": &Channel{"1", "oOne"},
				"2": &Channel{"2", "oTwo"}},
		},
		"Stage 2": &Device{
			hardware.StageBox,
			"Stage 2",
			Channels{
				"1": &Channel{"1", "iFive"},
				"2": &Channel{"2", "iSix"},
				"3": &Channel{"3", "iSeven"},
				"4": &Channel{"4", "iEight"}},
			Channels{
				"1": &Channel{"1", "oThree"},
				"2": &Channel{"2", "oFour"}},
		},
	}
}

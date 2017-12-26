package action

import (
	"fmt"
	"strings"

	"github.com/kward/tracks/tracks"
	"github.com/kward/tracks/venue"
	"github.com/kward/tracks/venue/hardware"
)

// MapTracksToNames based on their channel name.
func MapTracksToNames(ts tracks.Tracks, ds venue.Devices) (tracks.Tracks, error) {
	for i, t := range ts {
		_, ch, err := mapTrackToChannel(t, ds)
		if err != nil {
			return nil, fmt.Errorf("error mapping track to channel, %s", err)
		}
		ts[i].SetName(ch.CleanName())
	}
	return ts, nil
}

func mapTrackToChannel(t *tracks.Track, devs venue.Devices) (*venue.Device, *venue.Channel, error) {
	// Walk the stage boxes in order, counting channels as we go.
	offset := 0
	// Track whether we've found an empty channel name.
	empty := false

	// Search the devices based on their order in the slice.
	// Note: the stage boxes must be in sorted order.
	for _, name := range []string{venue.ProTools, venue.Stage1, venue.Stage2, venue.Stage3, venue.Stage4} {
		// Check that stage box was configured.
		dev, ok := devs[name]
		if !ok {
			continue
		}
		switch dev.Type() {
		case hardware.StageBox:
			// Check whether track is on this device.
			if t.TrackNum() > offset+dev.NumInputs() {
				offset += dev.NumInputs()
				continue
			}
			moniker := fmt.Sprintf("%d", t.TrackNum()-offset)
			ch := dev.Input(moniker)
			if ch == nil {
				continue
			}
			if ch.Name() == "" {
				empty = true
				continue
			}
			return dev, ch, nil
		case hardware.Local, hardware.ProTools:
			// Check whether the current device has enough inputs for the track
			// number.
			if t.TrackNum() > dev.NumOutputs() {
				continue
			}
			moniker := fmt.Sprintf("%d", t.TrackNum())
			ch := dev.Output(moniker)
			if ch == nil {
				continue
			}
			if ch.Name() == "" {
				empty = true
				continue
			}
			return dev, ch, nil
		default:
			return nil, nil, fmt.Errorf("unrecognized hardware")
		}
	}
	// No valid channel name found, so return Pro Tools device and empty channel.
	if empty {
		return devs[venue.ProTools], &venue.Channel{}, nil
	}

	return nil, nil, fmt.Errorf("channel not found")
}

// MapTrackNameToFilename returns a valid filename for a track name.
func MapTrackNameToFilename(name string) string {
	name = strings.Replace(name, "/", "_", -1)  // Unix path separator.
	name = strings.Replace(name, "\\", "_", -1) // Windows path separator.
	return name
}

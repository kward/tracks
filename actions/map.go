package actions

import (
	"fmt"
	"strings"

	"github.com/kward/tracks/tracks"
	"github.com/kward/tracks/venue"
)

// MapTracksToNames based on their channel name.
func MapTracksToNames(ts tracks.Tracks, devs venue.Devices) (tracks.Tracks, error) {
	for i, t := range ts {
		ch, err := mapTrackToChannel(t, devs)
		if err != nil {
			return nil, fmt.Errorf("error mapping track %q to channel; %s", t.Src(), err)
		}
		ts[i].SetName(ch.CleanName())
	}
	return ts, nil
}

// mapTrackToChannel maps a track name to the appropriate channel name.
//
// Venue only maps the stage box inputs directly to output files. Other inputs
// such as the "Engine AES 1" input are not mapped. To record them, the must
// be explicitly mapped as a Pro Tools output or direct out.
func mapTrackToChannel(t *tracks.Track, devs venue.Devices) (*venue.Channel, error) {
	ins := devs.Inputs()
	ch := ins[t.TrackNum()]
	if ch == nil {
		return nil, fmt.Errorf("channel not found")
	}
	if ch.Name() != "" {
		return ch, nil
	}

	// See if the ProTools device is available.
	dev, ok := devs[venue.ProTools]
	if !ok {
		return ch, nil
	}
	m := venue.Moniker(t.TrackNum())
	ptch := dev.Output(m)
	fmt.Printf("pro tools moniker: %s channel: %s\n", m, ptch)
	if ptch.Name() != "" {
		return ptch, nil
	}

	return ch, nil
}

// MapTrackNameToFilename returns a valid filename for a track name.
func MapTrackNameToFilename(name string) string {
	name = strings.Replace(name, "/", "_", -1)  // Unix path separator.
	name = strings.Replace(name, "\\", "_", -1) // Windows path separator.
	return name
}

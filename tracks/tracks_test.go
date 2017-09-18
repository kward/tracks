package tracks

import (
	"testing"
)

func TestTrackEqual(t *testing.T) {
	for _, tt := range []struct {
		desc   string
		t1, t2 *Track
		equal  bool
	}{
		{"equal", NewTrack("Track", 1, 1), &Track{name: "Track", tnum: 1, snum: 1}, true},
		{"unequal", NewTrack("Track", 1, 2), &Track{name: "Track", tnum: 100, snum: 2}, false},
		{"one nil", NewTrack("Track", 1, 3), nil, false},
		{"both nil", nil, nil, true},
	} {
		if got, want := tt.t1.Equal(tt.t2), tt.equal; got != want {
			t.Errorf("%s: Equal() = %v, want %v", tt.desc, got, want)
		}
	}
}

func TestExtractTrack(t *testing.T) {
	for _, tt := range []struct {
		desc  string
		file  string
		track *Track
	}{
		{"track-1 session-1", "Track 01-1.wav",
			&Track{src: "Track 01-1.wav", name: "Track", tnum: 1, snum: 1}},
		{"track-9 session-2", "Track 09-2.wav",
			&Track{src: "Track 09-2.wav", name: "Track", tnum: 9, snum: 2}},
	} {
		got, err := extractTrack(tt.file)
		if err != nil {
			t.Errorf("%s: extractTrack() unexpected error; %s", tt.desc, err)
			continue
		}
		if want := tt.track; !got.Equal(want) {
			t.Errorf("%s: extractTrack() = %s, want %s", tt.desc, got, want)
		}
	}
}

package tracks

import (
	"regexp"
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
		re    *regexp.Regexp
		track *Track
	}{
		// Avid Pro Tools
		{"pro tools s2 t1", "Audio 1_02.wav", proToolsRE,
			&Track{src: "Audio 1_02.wav", name: "Audio", snum: 2, tnum: 1}},
		{"pro tools s32 t29", "Audio 29_32.wav", proToolsRE,
			&Track{src: "Audio 29_32.wav", name: "Audio", snum: 32, tnum: 29}},
		// Waves Tracks
		{"tracks s1 t3", "Track 03-1.wav", tracksRE,
			&Track{src: "Track 03-1.wav", name: "Track", tnum: 3, snum: 1}},
		{"tracks s2 t9", "Track 09-2.wav", tracksRE,
			&Track{src: "Track 09-2.wav", name: "Track", tnum: 9, snum: 2}},
	} {
		got, err := extractTrack(tt.re, tt.file)
		if err != nil {
			t.Errorf("%s: extractTrack() unexpected error; %s", tt.desc, err)
			continue
		}
		if want := tt.track; !got.Equal(want) {
			t.Errorf("%s: extractTrack() = %s, want %s", tt.desc, got, want)
		}
	}
}

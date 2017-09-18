package tracks

import (
	"testing"
)

func TestExtractSessions(t *testing.T) {
	for _, tt := range []struct {
		desc     string
		files    []string
		sessions Sessions
		ok       bool
	}{
		{"empty files", []string{}, nil, false},
		{"one session",
			[]string{"Track 01-1.wav", "Track 02-1.wav"},
			Sessions{1: NewSession(1).SetTracks(
				Tracks{
					1: NewTrack("Track", 1, 1).SetSrc("Track 01-1.wav"),
					2: NewTrack("Track", 2, 1).SetSrc("Track 02-1.wav"),
				}),
			},
			true,
		},
	} {
		sessions, err := ExtractSessions(tt.files)
		if err == nil && !tt.ok {
			t.Errorf("%s: ExtractSessions() expected error", tt.desc)
		}
		if err != nil && tt.ok {
			t.Fatalf("%s: ExtractSessions() unexpected error; %s", tt.desc, err)
		}
		if !tt.ok {
			continue
		}
		if got, want := sessions, tt.sessions; !got.Equal(want) {
			t.Errorf("ExtractSession() = %q, want %q", got, want)
		}
	}
}

package vt

import "testing"

func TestMapTrackToChannel(t *testing.T) {
	sbs := mockStageBoxes()
	for _, tt := range []struct {
		desc     string
		track    *Track
		stageBox *StageBox
		channel  *Channel
	}{
		{"track one", NewTrack("Track", 1, 1), sbs["Stage 1"], &Channel{1, "iOne"}},
		{"track eight", NewTrack("Track", 8, 1), sbs["Stage 2"], &Channel{4, "iEight"}},
	} {
		sb, ch, err := mapTrackToChannel(tt.track, sbs)
		if err != nil {
			t.Errorf("%s: unexpected error; %s", tt.desc, err)
			continue
		}
		if got, want := sb.name, tt.stageBox.name; got != want {
			t.Errorf("%s: stage box = %s, want %s", tt.desc, got, want)
			continue
		}
		if got, want := ch, tt.channel; !got.Equal(want) {
			t.Errorf("%s: channel = %s, want %s", tt.desc, got, want)
			continue
		}
	}
}

func mockStageBoxes() StageBoxes {
	return StageBoxes{
		"Stage 1": &StageBox{
			"Stage 1",
			Channels{
				1: &Channel{1, "iOne"},
				2: &Channel{2, "iTwo"},
				3: &Channel{3, "iThree"},
				4: &Channel{4, "iFour"}},
			Channels{
				1: &Channel{1, "oOne"},
				2: &Channel{2, "oTwo"}},
		},
		"Stage 2": &StageBox{
			"Stage 2",
			Channels{
				1: &Channel{1, "iFive"},
				2: &Channel{2, "iSix"},
				3: &Channel{3, "iSeven"},
				4: &Channel{4, "iEight"}},
			Channels{
				1: &Channel{1, "oThree"},
				2: &Channel{2, "oFour"}},
		},
	}
}

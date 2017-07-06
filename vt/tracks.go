package vt

import (
	"fmt"
	"sort"
	"strings"
)

type Tracks map[int]*Track

func NewTracks() Tracks {
	return make(Tracks)
}

func (t Tracks) Slice() []*Track {
	slice := []*Track{}
	for _, v := range t {
		slice = append(slice, v)
	}
	sort.Slice(slice, func(i, j int) bool { return slice[i].num < slice[j].num })
	return slice
}

// Track holds metadata about a track.
type Track struct {
	name              string
	origName, newName string
	num               int
	sessionNum        int
}

func NewTrack(origName string, num, sessionNum int) *Track {
	return &Track{
		origName:   origName,
		num:        num,
		sessionNum: sessionNum,
	}
}

// String returns a string representation of a track.
func (t *Track) String() string {
	return fmt.Sprintf("{name: %s num: %d session_num: %d}",
		t.name, t.num, t.sessionNum)
}

func (t *Track) Name() string           { return t.name }
func (t *Track) SetName(name string)    { t.name = name }
func (t *Track) OrigName() string       { return t.origName }
func (t *Track) NewName() string        { return t.newName }
func (t *Track) SetNewName(name string) { t.newName = name }
func (t *Track) Num() int               { return t.num }

// filterTracks from a list of file names.
func filterTracks(names []string) []string {
	filtered := []string{}
	for _, name := range names {
		if strings.HasSuffix(name, ".wav") {
			filtered = append(filtered, name)
		}
	}
	return filtered
}

func mapTrackToChannel(t *Track, sbs StageBoxes) (*StageBox, *Channel, error) {
	// Walk the stage boxes in order, counting channels as we go.
	chOffset := 0

	for _, i := range stageBoxList {
		sbName := fmt.Sprintf("Stage %s", i)
		// Check that stage box was configured.
		sb, ok := sbs[sbName]
		if !ok {
			continue
		}
		// Check whether track is on this stage box.
		if t.num > chOffset+len(sb.inputs) {
			chOffset += len(sb.inputs)
			continue
		}
		// Found it.
		chNum := t.num - chOffset
		return sb, sb.inputs[chNum], nil
	}
	return nil, nil, fmt.Errorf("channel not found")
}

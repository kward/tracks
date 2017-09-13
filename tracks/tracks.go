package tracks

import (
	"fmt"
	"sort"
	"strings"
)

// Tracks is a map of tracks.
type Tracks map[int]*Track

// Track holds metadata about a track.
type Track struct {
	name               string
	origName, destName string
	num                int
	sessionNum         int
}

// NewTrack returns an instantiated Track object.
func NewTrack(origName string, num, sessionNum int) *Track {
	return &Track{
		origName:   origName,
		num:        num,
		sessionNum: sessionNum,
	}
}

// Slice returns the Tracks as a slice.
func (t Tracks) Slice() []*Track {
	slice := []*Track{}
	for _, v := range t {
		slice = append(slice, v)
	}
	sort.Slice(slice, func(i, j int) bool { return slice[i].num < slice[j].num })
	return slice
}

// String implements the fmt.Stringer interface.
func (t *Track) String() string {
	return fmt.Sprintf("{name: %s num: %d session_num: %d}",
		t.name, t.num, t.sessionNum)
}

func (t *Track) Name() string        { return t.name }
func (t *Track) SetName(name string) { t.name = name }

func (t *Track) OrigName() string { return t.origName }

func (t *Track) DestName() string        { return t.destName }
func (t *Track) SetDestName(name string) { t.destName = name }

func (t *Track) Num() int { return t.num }

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

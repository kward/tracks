package tracks

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
)

var (
	trackRE *regexp.Regexp
)

func init() {
	trackRE = regexp.MustCompile("(?P<name>[a-zA-Z]+) (?P<channel>[0-9]+)-(?P<session>[0-9]+).wav")
}

// Tracks is a map of tracks.
type Tracks map[int]*Track

// TrackSlice is a slice of tracks.
type TrackSlice []*Track

// Verify proper interface implementation.
var _ sort.Interface = (*TrackSlice)(nil)

// Sort tracks based on their track number.
func (ts TrackSlice) Len() int           { return len(ts) }
func (ts TrackSlice) Less(i, j int) bool { return ts[i].tnum < ts[j].tnum }
func (ts TrackSlice) Swap(i, j int)      { ts[i], ts[j] = ts[j], ts[i] }

// Slice returns the Tracks as a slice.
func (ts Tracks) Slice() []*Track {
	slice := TrackSlice{}
	for _, t := range ts {
		slice = append(slice, t)
	}
	sort.Sort(slice)
	return slice
}

// Equal returns true if the two Tracks are equivalent.
func (ts Tracks) Equal(ts2 Tracks) bool {
	if len(ts) != len(ts2) {
		return false
	}
	for tnum, t := range ts {
		t2, ok := ts2[tnum]
		if !ok {
			return false
		}
		if !t.Equal(t2) {
			return false
		}
	}
	return true
}

// Track holds metadata about a track.
type Track struct {
	src, dest string // Source and destination files.
	name      string // Extracted name.
	tnum      int    // Track number.
	snum      int    // Session number.
}

// NewTrack returns an instantiated Track object.
func NewTrack(name string, tnum, snum int) *Track {
	return &Track{
		name: name,
		tnum: tnum,
		snum: snum,
	}
}

// Equal returns true if the tracks are equal.
func (t *Track) Equal(t2 *Track) bool {
	return reflect.DeepEqual(t, t2)
}

// String implements the fmt.Stringer interface.
func (t *Track) String() string {
	return fmt.Sprintf("{src: %q dest: %q name: %q tnum: %d snum: %d}",
		t.src, t.dest, t.name, t.tnum, t.snum)
}

func (t *Track) Name() string               { return t.name }
func (t *Track) SetName(name string) *Track { t.name = name; return t }

func (t *Track) Src() string              { return t.src }
func (t *Track) SetSrc(src string) *Track { t.src = src; return t }

func (t *Track) Dest() string               { return t.dest }
func (t *Track) SetDest(dest string) *Track { t.dest = dest; return t }

func (t *Track) TrackNum() int   { return t.tnum }
func (t *Track) SessionNum() int { return t.snum }

// extractTrack returns a populated Track from a file name.
func extractTrack(file string) (*Track, error) {
	if !trackRE.MatchString(file) {
		return nil, fmt.Errorf("error matching file %q", file)
	}
	name := trackRE.ReplaceAllString(file, "${name}")

	tnum, err := strconv.Atoi(trackRE.ReplaceAllString(file, "${channel}"))
	if err != nil {
		return nil, fmt.Errorf("error converting file %q channel; %s", file, err)
	}

	snum, err := strconv.Atoi(trackRE.ReplaceAllString(file, "${session}"))
	if err != nil {
		return nil, fmt.Errorf("error converting file %q session; %s", file, err)
	}

	return &Track{src: file, name: name, tnum: tnum, snum: snum}, nil
}

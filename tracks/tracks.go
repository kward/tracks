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

// Slice returns the Tracks as a slice.
func (t Tracks) Slice() []*Track {
	slice := []*Track{}
	for _, v := range t {
		slice = append(slice, v)
	}
	sort.Slice(slice, func(i, j int) bool { return slice[i].tnum < slice[j].tnum })
	return slice
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
	return fmt.Sprintf("{name: %s tnum: %d snum: %d}",
		t.name, t.tnum, t.snum)
}

func (t *Track) Name() string        { return t.name }
func (t *Track) SetName(name string) { t.name = name }

func (t *Track) Src() string { return t.src }

func (t *Track) Dest() string        { return t.dest }
func (t *Track) SetDest(dest string) { t.dest = dest }

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

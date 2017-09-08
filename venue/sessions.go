package venue

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"sort"
	"strconv"
)

// Sessions holds a map of session numbers to Session data.
type Sessions map[int]*Session

// NewSessions returns a new Sessions object.
func NewSessions() Sessions {
	return make(Sessions)
}

func (ss Sessions) Slice() []*Session {
	slice := []*Session{}
	for _, v := range ss {
		slice = append(slice, v)
	}
	sort.Slice(slice, func(i, j int) bool { return slice[i].num < slice[j].num })
	return slice
}

// Session returns the pointer to session `num`. If the session doesn't yet
// exist, it is created.
func (ss Sessions) Session(num int) *Session {
	s, ok := ss[num]
	if !ok {
		s = NewSession(num)
		ss[num] = s
	}
	return s
}

// Sessions as a list.
func (ss Sessions) SessionList() []int {
	l := []int{}
	for s := range ss {
		l = append(l, s)
	}
	return l
}

// Discover tracks and sessions based on files in a directory.
func (ss Sessions) Discover(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("error discovering files in %q; %s", dir, err)
	}
	names := []string{}
	for _, file := range files {
		names = append(names, file.Name())
	}

	tracks := filterTracks(names)
	if len(tracks) == 0 {
		return fmt.Errorf("no tracks found in %q", dir)
	}
	return ss.extract(tracks)
}

// extract track names, numbers, and session numbers.
func (ss Sessions) extract(tracks []string) error {
	re := regexp.MustCompile("(?P<name>[a-zA-Z]+) (?P<channel>[0-9]+)-(?P<session>[0-9]+).wav")

	for _, track := range tracks {
		if !re.MatchString(track) {
			return fmt.Errorf("error matching track %q", track)
		}
		name := re.ReplaceAllString(track, "${name}")

		num, err := strconv.Atoi(re.ReplaceAllString(track, "${channel}"))
		if err != nil {
			return fmt.Errorf("error converting track %q channel; %s", track, err)
		}

		sessionNum, err := strconv.Atoi(re.ReplaceAllString(track, "${session}"))
		if err != nil {
			return fmt.Errorf("error converting track %q session; %s", track, err)
		}

		t := NewTrack(track, num, sessionNum)
		t.SetName(name)

		session := ss.Session(sessionNum)
		session.tracks[num] = t
	}

	return nil
}

// Session maps channel numbers to track info.
type Session struct {
	num    int
	tracks Tracks
}

func NewSession(num int) *Session {
	return &Session{
		num:    num,
		tracks: NewTracks(),
	}
}

func (s *Session) Num() int { return s.num }

func (s *Session) Tracks() Tracks      { return s.tracks }
func (s *Session) SetTracks(ts Tracks) { s.tracks = ts }

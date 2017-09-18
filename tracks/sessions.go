package tracks

import "fmt"

// Sessions holds a map of session numbers to Session data.
type Sessions map[int]*Session

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

// Equal returns true if the two Sessions are equivalent.
func (ss Sessions) Equal(ss2 Sessions) bool {
	if len(ss) != len(ss2) {
		return false
	}
	for snum, s := range ss {
		s2, ok := ss2[snum]
		if !ok {
			return false
		}
		if !s.Equal(s2) {
			return false
		}
	}
	return true
}

// ExtractSessions from a slice of track names.
func ExtractSessions(files []string) (Sessions, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	sessions := make(Sessions)
	for _, file := range files {
		t, err := extractTrack(file)
		if err != nil {
			return nil, err
		}
		s := sessions.Session(t.SessionNum())
		s.tracks[t.TrackNum()] = t
	}
	return sessions, nil
}

// Session maps channel numbers to track info.
type Session struct {
	num    int
	tracks Tracks
}

func NewSession(num int) *Session {
	return &Session{
		num:    num,
		tracks: make(Tracks),
	}
}

// Equal returns true if the two Session objects are equal.
func (s *Session) Equal(s2 *Session) bool {
	if s.num != s2.num {
		return false
	}
	return s.tracks.Equal(s2.tracks)
}

// String implements the fmt.Stringer interface.
func (s *Session) String() string {
	return fmt.Sprintf("{num: %d tracks: %v}", s.num, s.tracks)
}

func (s *Session) Num() int { return s.num }

func (s *Session) Tracks() Tracks               { return s.tracks }
func (s *Session) SetTracks(ts Tracks) *Session { s.tracks = ts; return s }

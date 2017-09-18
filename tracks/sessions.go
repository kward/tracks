package tracks

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

// ExtractSessions from a slice of track names.
func ExtractSessions(files []string) (Sessions, error) {
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

func (s *Session) Num() int { return s.num }

func (s *Session) Tracks() Tracks      { return s.tracks }
func (s *Session) SetTracks(ts Tracks) { s.tracks = ts }

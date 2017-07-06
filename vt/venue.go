package vt

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	xmlpath "gopkg.in/xmlpath.v2"
)

type xPath struct {
	name  string
	xpath string
	path  *xmlpath.Path
}

var xpaths = map[string]xPath{
	"console":               xPath{xpath: `//meta[@name='description']/@content`},
	"version":               xPath{xpath: `//meta[@name='author']/@content`},
	"show":                  xPath{xpath: `//table//td[contains(span,'Show')]/../td[2]`},
	"stageBoxInputs":        xPath{xpath: `//table//tr[contains(td/span,'Stage') and contains(td/span,'Inputs')]`},
	"stageBoxOutputs":       xPath{xpath: `//table//tr[contains(td/span,'Stage') and contains(td/span,'Outputs')]`},
	"stageBoxChannel":       xPath{xpath: `../tr`},
	"stageBoxChannelDetail": xPath{xpath: `td`},
}

func init() {
	for k, v := range xpaths {
		v.name = k
		v.path = xmlpath.MustCompile(v.xpath)
		xpaths[k] = v
	}
}

type Venue struct {
	console    string
	version    string
	show       string
	stageBoxes StageBoxes
}

func NewVenue() *Venue {
	return &Venue{
		stageBoxes: make(StageBoxes),
	}
}

func (v *Venue) String() string {
	s := fmt.Sprintf("{console: %q version: %q show: %q stageBoxes:{",
		v.console, v.version, v.show)
	for _, stageBox := range v.stageBoxes {
		s += stageBox.String()
	}
	s += "}"
	return s
}

// Parse a Venue patch file.
func (v *Venue) Parse(data []byte) error {
	root, err := xmlpath.ParseHTML(bytes.NewReader(data))
	if err != nil {
		return err
	}
	if err := v.parseMetadata(root); err != nil {
		return err
	}
	if err := v.discoverStageBoxes(root); err != nil {
		return err
	}

	return nil
}

func (v *Venue) parseMetadata(root *xmlpath.Node) error {
	for _, row := range []struct {
		name string
		val  *string
	}{
		{"console", &v.console},
		{"version", &v.version},
		{"show", &v.show},
	} {
		str, ok := xpaths[row.name].path.String(root)
		if !ok {
			return fmt.Errorf("%s xpath returned no values", row.name)
		}

		val := row.val
		*val = trim(str)
		row.val = val
	}

	return nil
}

func (v *Venue) discoverStageBoxes(root *xmlpath.Node) error {
	// Find stage box inputs. The stage box name is derived based on what is
	// discovered here.
	iter := xpaths["stageBoxInputs"].path.Iter(root)
	for iter.Next() {
		node := iter.Node()
		name := trim(node.String())
		name = strings.TrimSuffix(name, " Inputs")

		chs, err := probeStageBox(node)
		if err != nil {
			return fmt.Errorf("error probing stage box inputs; %s", err)
		}

		stageBox := v.stageBox(name)
		stageBox.inputs = chs
	}

	iter = xpaths["stageBoxOutputs"].path.Iter(root)
	for iter.Next() {
		node := iter.Node()
		name := trim(node.String())
		name = strings.TrimSuffix(name, " Outputs")

		chs, err := probeStageBox(node)
		if err != nil {
			return fmt.Errorf("error probing stage box outputs; %s", err)
		}

		stageBox := v.stageBox(name)
		stageBox.outputs = chs
	}

	return nil
}

func (v *Venue) stageBox(name string) *StageBox {
	sb, ok := v.stageBoxes[name]
	if !ok {
		sb = NewStageBox(name)
		v.stageBoxes[name] = sb
	}
	return sb
}

func probeStageBox(root *xmlpath.Node) (Channels, error) {
	chs := Channels{}

	chIter := xpaths["stageBoxChannel"].path.Iter(root)
	first := true
	for chIter.Next() {
		if first { // Skip the stage box description.
			first = false
			continue
		}

		ch := &Channel{}
		state := "number"

		dIter := xpaths["stageBoxChannelDetail"].path.Iter(chIter.Node())
		for dIter.Next() {
			node := dIter.Node()
			str := trim(node.String())

			switch state {
			case "number":
				number, err := strconv.Atoi(str)
				if err != nil {
					return nil, fmt.Errorf("error converting stage box channel %q; %s", str, err)
				}
				ch.number = number
				state = "name"
			case "name":
				ch.name = sanitize(str)
				state = "number2"
			case "number2":
				// Do nothing.
			}

			chs[ch.number] = ch
		}
	}

	return chs, nil
}

// MapTrack to a StageBox.
func (v *Venue) NameTracks(ts Tracks) (Tracks, error) {
	for i, t := range ts {
		var sb *StageBox
		sb, ch, err := mapTrackToChannel(t, v.stageBoxes)
		_ = sb
		if err != nil {
			return nil, fmt.Errorf("error mapping track to channel; %s", err)
		}
		ts[i].name = ch.name
	}
	return ts, nil
}

// StageBoxList is an ordered list of stage boxes. Taking the lazy route and
// using a string instead of an int for the stage box numbs.
var stageBoxList = []string{"1", "2", "3", "4"}

type StageBoxes map[string]*StageBox

type StageBox struct {
	name            string   // StageBox name.
	inputs, outputs Channels // Channel data.
}

func NewStageBox(name string) *StageBox {
	return &StageBox{
		name:    name,
		inputs:  make(Channels),
		outputs: make(Channels),
	}
}

func (b *StageBox) String() string {
	s := fmt.Sprintf("{name: %q inputs:{", b.name)
	for _, ch := range b.inputs {
		s += ch.String()
	}
	s += "} outputs:{"
	for _, ch := range b.outputs {
		s += ch.String()
	}
	s += "}"
	return s
}

type Channels map[int]*Channel
type Channel struct {
	number int
	name   string
}

func (c *Channel) Equal(c2 *Channel) bool {
	if c.number != c2.number {
		return false
	}
	return c.name == c2.name
}

func (c *Channel) String() string {
	s := fmt.Sprintf("{number: %d", c.number)
	if c.name != "" {
		s += fmt.Sprintf(" name: %q", c.name)
	}
	s += "}"
	return s
}

func sanitize(text string) string {
	// Remove &nbsp; equivalent chars.
	return strings.Replace(text, "\u00a0", "", -1)
}
func trim(text string) string {
	return strings.Trim(text, "\r\n")
}

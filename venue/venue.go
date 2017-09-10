/*
Package venue provides functionality for parsing and understanding the Venue
Info (Options > System > Info) HTML output.
*/
package venue

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/kward/tracks/venue/hardware"

	xmlpath "gopkg.in/xmlpath.v2"
)

func init() {
	for k, v := range xpaths {
		v.name = k
		v.path = xmlpath.MustCompile(v.xpath)
		xpaths[k] = v
	}
}

//-----------------------------------------------------------------------------
// Venue

type Venue struct {
	console string
	version string
	show    string

	devices Devices
}

func NewVenue() *Venue {
	return &Venue{
		devices: make(Devices),
	}
}

func (v *Venue) String() string {
	s := fmt.Sprintf("{console: %s version: %s show: %s devices:{",
		v.console, v.version, v.show)
	for _, d := range v.devices {
		s += d.String()
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

	sbs, err := discoverStageBoxes(root)
	if err != nil {
		return err
	}
	for k, sb := range sbs {
		v.devices[k] = sb
	}

	pt, err := discoverProTools(root)
	if err != nil {
		return err
	}
	v.devices[pt.name] = pt

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

// NameTracks based on their channel name.
func (v *Venue) NameTracks(ts Tracks) (Tracks, error) {
	for i, t := range ts {
		_, ch, err := mapTrackToChannel(t, v.devices)
		if err != nil {
			return nil, fmt.Errorf("error mapping track to channel; %s", err)
		}
		ts[i].name = ch.name
	}
	return ts, nil
}

// discoverStageBoxes parses the XML for stage box references.
func discoverStageBoxes(root *xmlpath.Node) (Devices, error) {
	devs := make(Devices)

	// Find stage box inputs. The stage box name is derived based on what is
	// discovered here.
	iter := xpaths["stageBoxInputs"].path.Iter(root)
	for iter.Next() {
		name, chs, err := discover(iter.Node(), "Inputs")
		if err != nil {
			return nil, fmt.Errorf("error probing stage box inputs; %s", err)
		}
		devs[name] = &Device{
			typ:    hardware.StageBox,
			name:   name,
			inputs: chs,
		}
	}

	iter = xpaths["stageBoxOutputs"].path.Iter(root)
	for iter.Next() {
		name, chs, err := discover(iter.Node(), "Outputs")
		if err != nil {
			return nil, fmt.Errorf("error probing stage box outputs; %s", err)
		}
		if devs[name] == nil {
			return nil, fmt.Errorf("found outputs without corresponding inputs for %s", name)
		}
		devs[name].outputs = chs
	}

	return devs, nil
}

func discoverProTools(root *xmlpath.Node) (*Device, error) {
	dev := &Device{
		typ:  hardware.ProTools,
		name: "Pro Tools",
	}

	iter := xpaths["proToolsInputs"].path.Iter(root)
	if !iter.Next() {
		return nil, fmt.Errorf("Pro Tools inputs not found")
	}
	_, chs, err := discover(iter.Node(), "Inputs")
	if err != nil {
		return nil, err
	}
	dev.inputs = chs

	iter = xpaths["proToolsOutputs"].path.Iter(root)
	if !iter.Next() {
		return nil, fmt.Errorf("Pro Tools outputs not found")
	}
	_, chs, err = discover(iter.Node(), "Outputs")
	if err != nil {
		return nil, err
	}
	dev.outputs = chs

	return dev, nil
}

func discover(node *xmlpath.Node, title string) (string, Channels, error) {
	name := trim(node.String())
	name = strings.TrimSuffix(name, " "+title)

	chs, err := probeChannels(node)
	if err != nil {
		return "", nil, fmt.Errorf("error probing for %s; %s", title, err)
	}
	return name, chs, nil
}

//-----------------------------------------------------------------------------
// Device

type Device struct {
	typ             hardware.Hardware
	name            string
	inputs, outputs Channels
}

type Devices map[string]*Device

func (d *Device) String() string {
	s := fmt.Sprintf("{name: %s inputs:{", d.name)
	for _, ch := range d.inputs {
		s += ch.String()
	}
	s += "} outputs:{"
	for _, ch := range d.outputs {
		s += ch.String()
	}
	s += "}"
	return s
}

//-----------------------------------------------------------------------------
// Channel

type Channels map[string]*Channel
type Channel struct {
	moniker string // The channel number (e.g. "1") or IO name (e.g. "FWx 1").
	name    string
}

// Equal returns true if the channels are equal.
func (c *Channel) Equal(c2 *Channel) bool {
	if c.moniker != c2.moniker {
		return false
	}
	return c.name == c2.name
}

// String provides human readable output.
func (c *Channel) String() string {
	s := fmt.Sprintf("{moniker: %d", c.moniker)
	if c.name != "" {
		s += fmt.Sprintf(" name: %s", c.name)
	}
	s += "}"
	return s
}

func probeChannels(root *xmlpath.Node) (Channels, error) {
	chs := Channels{}

	chIter := xpaths["channel"].path.Iter(root)
	first := true
	for chIter.Next() {
		if first { // Skip the stage box description.
			first = false
			continue
		}

		ch := &Channel{}
		state := "number"

		dIter := xpaths["channelDetail"].path.Iter(chIter.Node())
		for dIter.Next() {
			node := dIter.Node()
			moniker := trim(node.String())

			switch state {
			case "number":
				ch.moniker = moniker
				state = "name"
			case "name":
				ch.name = sanitize(moniker)
				state = "number2"
			case "number2":
				// Do nothing.
			}

			chs[ch.moniker] = ch
		}
	}

	return chs, nil
}

//-----------------------------------------------------------------------------
// XPath

type XPath struct {
	name  string
	xpath string
	path  *xmlpath.Path
}

var xpaths = map[string]XPath{
	"console": XPath{
		xpath: `//meta[@name='description']/@content`},
	"version": XPath{
		xpath: `//meta[@name='author']/@content`},
	"show": XPath{
		xpath: `//table//td[contains(span,'Show:')]/../td[2]`},
	"stageBoxInputs": XPath{
		xpath: `//table//tr[contains(td/span,'Stage') and contains(td/span,'Inputs')]`},
	"stageBoxOutputs": XPath{
		xpath: `//table//tr[contains(td/span,'Stage') and contains(td/span,'Outputs')]`},
	"channel": XPath{
		xpath: `../tr`},
	"channelDetail": XPath{
		xpath: `td`},
	"proToolsInputs": XPath{
		xpath: `//table//tr[contains(td/span,'Pro Tools') and contains(td/span,'Inputs')]`},
	"proToolsOutputs": XPath{
		xpath: `//table//tr[contains(td/span,'Pro Tools') and contains(td/span,'Outputs')]`},
}

//-----------------------------------------------------------------------------
// Miscellaneous

func sanitize(text string) string {
	// Remove &nbsp; equivalent chars.
	return strings.Replace(text, "\u00a0", "", -1)
}

func trim(text string) string {
	return strings.Trim(text, "\r\n")
}

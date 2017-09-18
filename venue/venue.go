/*
Package venue provides functionality for parsing and understanding the Venue
Info (Options > System > Info) HTML output.
*/
package venue

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/kward/golib/errors"
	"github.com/kward/tracks/venue/hardware"
	"google.golang.org/grpc/codes"
	xmlpath "gopkg.in/xmlpath.v2"
)

const (
	Console  = "Console"
	Engine   = "Engine"
	Local    = "Local"
	ProTools = "Pro Tools"
	Stage1   = "Stage 1"
	Stage2   = "Stage 2"
	Stage3   = "Stage 3"
	Stage4   = "Stage 4"
)

func init() {
	for k, v := range xpaths {
		v.name = k
		if !v.dynamic { // Only compile the non-dynamic paths.
			v.path = xmlpath.MustCompile(v.xpath)
			xpaths[k] = v
		}
	}
}

//-----------------------------------------------------------------------------
// Venue

// Venue describes an Avid Venue device as found in an exported patch list.
type Venue struct {
	console string
	version string
	show    string

	devices Devices
}

// NewVenue returns a pointer to an instantiated Venue struct.
func NewVenue() *Venue {
	return &Venue{
		devices: make(Devices),
	}
}

// String implements the fmt.Stringer interface.
func (v *Venue) String() string {
	s := fmt.Sprintf("{console: %s version: %s show: %s devices:{",
		v.console, v.version, v.show)
	for _, d := range v.devices {
		s += d.String()
	}
	s += "}"
	return s
}

// Devices returns the known devices.
func (v *Venue) Devices() Devices {
	return v.devices
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

	devs, err := discoverDevices(root)
	if err != nil {
		return err
	}
	v.devices = devs

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

//-----------------------------------------------------------------------------
// Device

// Devices is a map of devices.
type Devices map[string]*Device

// Device describes a Venue IO device.
type Device struct {
	typ             hardware.Hardware
	name            string
	inputs, outputs Channels
}

// NewDevice returns a pointer to an instantiated Device struct.
func NewDevice(typ hardware.Hardware, name string, inputs, outputs Channels) *Device {
	return &Device{
		typ:     typ,
		name:    name,
		inputs:  inputs,
		outputs: outputs,
	}
}

// Type returns the device hardware type.
func (d *Device) Type() hardware.Hardware { return d.typ }

// Name returns the device name.
func (d *Device) Name() string { return d.name }

// Input returns a copy of the named input channel.
func (d *Device) Input(ch string) *Channel {
	prefix := ""
	if d.typ == hardware.ProTools {
		prefix = "FWx "
	}
	return d.inputs[prefix+ch]
}

// NumInputs returns the number of input channels.
func (d *Device) NumInputs() int { return len(d.inputs) }

// Output returns a copy of the named output channel.
func (d *Device) Output(ch string) *Channel {
	prefix := ""
	if d.typ == hardware.ProTools {
		prefix = "FWx "
	}
	return d.outputs[prefix+ch]
}

// NumOutputs returns the number of output channels.
func (d *Device) NumOutputs() int { return len(d.outputs) }

// String implements the fmt.Stringer interface.
func (d *Device) String() string {
	s := fmt.Sprintf("{name: %s inputs:{", d.name)
	for _, ch := range d.inputs.Sorted() {
		s += ch.String()
	}
	s += "} outputs:{"
	for _, ch := range d.outputs.Sorted() {
		s += ch.String()
	}
	s += "}"
	return s
}

// discoverDevices walks the XML, looking for known Venue devices.
func discoverDevices(root *xmlpath.Node) (Devices, error) {
	devs := make(Devices)

	for _, name := range []string{
		Console, Engine, Local, ProTools, Stage1, Stage2, Stage3, Stage4,
	} {
		dev, err := discoverDevice(root, name)
		switch errors.Code(err) {
		case codes.OK: // Do nothing.
		case codes.NotFound:
			continue
		default:
			return nil, err
		}
		devs[name] = dev
	}

	return devs, nil
}

// discoverDevice walks the XML, looking for specific device inputs and outputs.
func discoverDevice(root *xmlpath.Node, name string) (*Device, error) {
	dev := &Device{
		typ:  hardware.ProTools,
		name: name,
	}

	iter := xmlpath.MustCompile(fmt.Sprintf(xpaths["devices"].xpath, name, "Inputs")).Iter(root)
	if !iter.Next() {
		return nil, errors.Errorf(codes.NotFound, "%s inputs not found", name)
	}
	_, chs, err := probeDevice(iter.Node(), "Inputs")
	if err != nil {
		return nil, err
	}
	dev.inputs = chs

	iter = xmlpath.MustCompile(fmt.Sprintf(xpaths["devices"].xpath, name, "Outputs")).Iter(root)
	if !iter.Next() {
		return nil, errors.Errorf(codes.NotFound, "%s outputs not found", name)
	}
	_, chs, err = probeDevice(iter.Node(), "Outputs")
	if err != nil {
		return nil, err
	}
	dev.outputs = chs

	return dev, nil
}

// probeDevice walks the XML, probing a device for info.
func probeDevice(node *xmlpath.Node, title string) (string, Channels, error) {
	name := trim(node.String())
	name = strings.TrimSuffix(name, " "+title)

	chs, err := probeChannels(node)
	if err != nil {
		return "", nil, errors.Errorf(codes.Internal, "error probing for %s; %s", title, err)
	}
	return name, chs, nil
}

//-----------------------------------------------------------------------------
// Channel

// Channels is a map of channels.
type Channels map[string]*Channel

func (cs Channels) Sorted() ChannelsByMoniker {
	chs := ChannelsByMoniker{}
	for _, ch := range cs {
		chs = append(chs, ch)
	}
	sort.Sort(chs)
	return chs
}

type ChannelsByMoniker []*Channel

// Verify proper interface implementation.
var _ sort.Interface = new(ChannelsByMoniker)

// TODO(kward) Make this a numeric sort, instead of alphanumeric.
func (d ChannelsByMoniker) Len() int           { return len(d) }
func (d ChannelsByMoniker) Less(i, j int) bool { return d[i].moniker < d[j].moniker }
func (d ChannelsByMoniker) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }

// Channel describes a device channel.
type Channel struct {
	moniker string // The channel number (e.g. "1") or IO name (e.g. "FWx 1").
	name    string
}

// NewChannel returns an instantiated Channel.
func NewChannel(moniker, name string) *Channel {
	return &Channel{
		moniker: moniker,
		name:    name,
	}
}

// Equal returns true if the channels are equal.
func (c *Channel) Equal(c2 *Channel) bool {
	if c.moniker != c2.moniker {
		return false
	}
	return c.name == c2.name
}

// Name returns the channel name.
func (c *Channel) Name() string { return c.name }

// String implements the fmt.Stringer interface.
func (c *Channel) String() string {
	s := fmt.Sprintf("{moniker: %s", c.moniker)
	if c.name != "" {
		s += fmt.Sprintf(" name: %s", c.name)
	}
	s += "}"
	return s
}

// probeChannels walks XML, looking for channel info.
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

// XPath describes an xpath, and holds a compiled version.
type XPath struct {
	name    string
	xpath   string
	path    *xmlpath.Path
	dynamic bool // Compile dynamically?
}

var xpaths = map[string]XPath{
	// Static paths.
	"console": XPath{
		xpath: `//meta[@name='description']/@content`},
	"version": XPath{
		xpath: `//meta[@name='author']/@content`},
	"show": XPath{
		xpath: `//table//td[contains(span,'Show:')]/../td[2]`},
	"channel": XPath{
		xpath: `../tr`},
	"channelDetail": XPath{
		xpath: `td`},
	// Dynamic paths.
	"devices": XPath{
		xpath:   `//table//tr[contains(td/span,'%s') and contains(td/span,'%s')]`,
		dynamic: true},
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

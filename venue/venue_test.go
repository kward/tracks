package venue

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/kward/tracks/venue/hardware"

	xmlpath "gopkg.in/xmlpath.v2"
)

type TestData struct {
	name       string
	root       *xmlpath.Node
	console    string
	version    string
	show       string
	devNames   []string
	hardware   []hardware.Hardware
	numInputs  []int
	numOutputs []int
}

var testdata []*TestData

func init() {
	testdata = []*TestData{
		// Avid S3L-X console doing recording.
		{
			name:       "20170526 ICF Conference Worship Night.html",
			console:    "Avid VENUE",
			version:    "VENUE 4.5.3",
			show:       "ICF Zurich\\20170526 Conf WN",
			devNames:   []string{"Console", "Engine", "Pro Tools", "Stage 1", "Stage 2", "Stage 3", "Stage 4"},
			hardware:   []hardware.Hardware{hardware.Local, hardware.Local, hardware.ProTools, hardware.StageBox, hardware.StageBox, hardware.StageBox, hardware.StageBox},
			numInputs:  []int{4, 11, 64, 16, 16, 16, 16},
			numOutputs: []int{4, 10, 64, 12, 12, 12, 12}},
		// Avid Profile console doing FoH and in-ear monitoring.
		{
			name:       "20170906 ICF Ladies Night.html",
			console:    "Avid VENUE",
			version:    "VENUE 4.5.3",
			show:       "ICF Zurich\\20170906 Ladies Night",
			devNames:   []string{"Local", "Pro Tools", "Stage 1"},
			hardware:   []hardware.Hardware{hardware.Local, hardware.ProTools, hardware.StageBox},
			numInputs:  []int{31, 32, 48},
			numOutputs: []int{28, 32, 48}},
		// Avid D-Show console doing FoH.
		{
			name:       "20170910 Avid D-Show Patch List.html",
			console:    "Avid VENUE",
			version:    "D-Show 3.1.1",
			show:       "GenX\\2017_09_10PM",
			devNames:   []string{"Pro Tools", "Stage 1"},
			hardware:   []hardware.Hardware{hardware.ProTools, hardware.StageBox},
			numInputs:  []int{32, 48},
			numOutputs: []int{32, 48}},
		{
			name:       "20170910 Avid D-Show System Info.html",
			console:    "Avid VENUE",
			version:    "D-Show 3.1.1",
			show:       "GenX\\2017_09_10PM",
			devNames:   []string{"Pro Tools", "Stage 1"},
			hardware:   []hardware.Hardware{hardware.ProTools, hardware.StageBox},
			numInputs:  []int{32, 48},
			numOutputs: []int{32, 48}},
		// Avid S3L-X console doing recording.
		{
			name:       "20170910 Avid S3L-X Patch List.html",
			console:    "Avid VENUE",
			version:    "VENUE 4.5.3",
			show:       "01 ICF ZH Celebrations\\2017-09-10 Rec PM",
			devNames:   []string{"Console", "Engine", "Pro Tools", "Stage 1", "Stage 2", "Stage 3", "Stage 4"},
			hardware:   []hardware.Hardware{hardware.Local, hardware.Local, hardware.ProTools, hardware.StageBox, hardware.StageBox, hardware.StageBox, hardware.StageBox},
			numInputs:  []int{4, 11, 64, 16, 16, 16, 16},
			numOutputs: []int{4, 10, 64, 12, 12, 12, 12}},
		{
			name:       "20170910 Avid S3L-X System Info.html",
			console:    "Avid VENUE",
			version:    "VENUE 4.5.3",
			show:       "01 ICF ZH Celebrations\\2017-09-10 Rec PM",
			devNames:   []string{"Console", "Engine", "Pro Tools", "Stage 1", "Stage 2", "Stage 3", "Stage 4"},
			hardware:   []hardware.Hardware{hardware.Local, hardware.Local, hardware.ProTools, hardware.StageBox, hardware.StageBox, hardware.StageBox, hardware.StageBox},
			numInputs:  []int{4, 11, 64, 16, 16, 16, 16},
			numOutputs: []int{4, 10, 64, 12, 12, 12, 12}},
	}

	for _, td := range testdata {
		path := "../testdata/" + td.name
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("error reading testdata file %q; %s", path, err)
			os.Exit(1)
		}
		node, err := xmlpath.ParseHTML(bytes.NewReader(data))
		if err != nil {
			log.Printf("error parsing HTML; %s", err)
			os.Exit(1)
		}
		td.root = node
	}
}

//-----------------------------------------------------------------------------
// Venue
//

func TestParseMetadata(t *testing.T) {
	v := NewVenue()
	for _, td := range testdata {
		if err := v.parseMetadata(td.root); err != nil {
			t.Errorf("error parsing metadata for %s; %s", td.name, err)
		}

		if got, want := v.console, td.console; got != want {
			t.Errorf("%s: console = %s, want %s", td.name, got, want)
		}
		if got, want := v.version, td.version; got != want {
			t.Errorf("%s: version = %s, want %s", td.name, got, want)
		}
		if got, want := v.show, td.show; got != want {
			t.Errorf("%s: show = %s, want %s", td.name, got, want)
		}
	}
}

func TestDiscoverDevices(t *testing.T) {
	for _, td := range testdata {
		devs, err := discoverDevices(td.root)
		if err != nil {
			t.Fatalf("discoverDevices(): unexpected error; %s", err)
		}
		for i, dn := range td.devNames {
			dev := devs[dn]
			if dev == nil {
				t.Errorf("%s: discoverDevices(): missing device %s", td.name, dn)
				continue
			}
			if got, want := dev.Hardware(), td.hardware[i]; got != want {
				t.Errorf("%s: discoverDevices(): %s Type() = %s, want %s", td.name, dn, got, want)
			}
			if got, want := dev.NumInputs(), td.numInputs[i]; got != want {
				t.Errorf("%s: discoverDevices(): %s NumInputs() = %d, want %d", td.name, dn, got, want)
			}
			if got, want := dev.NumOutputs(), td.numOutputs[i]; got != want {
				t.Errorf("%s: discoverDevices(): %s NumOutputs() = %d, want %d", td.name, dn, got, want)
			}
		}
	}
}

//-----------------------------------------------------------------------------
// Device
//

func TestDevicesInputs(t *testing.T) {
	data, err := ioutil.ReadFile("../testdata/20180128 Avid S3L-X Patch List.html")
	if err != nil {
		t.Fatalf("error reading patch list; %s", err)
	}

	v := NewVenue()
	if err := v.Parse(data); err != nil {
		t.Fatalf("error parsing data; %s", err)
	}

	chs := v.Devices().Inputs()
	for _, tt := range []struct {
		num     int
		moniker string
		name    string
	}{
		{1, "1", "Kick 91"},
		{17, "1", "ePatrick-L, ePatrick-R"},
		{64, "16", ""},
	} {
		if got, want := chs[tt.num], NewChannel(tt.moniker, tt.name); !got.Equal(want) {
			t.Errorf("InputChannels() = %s, want %s", got, want)
		}
	}

	if got, want := len(chs), 64; got != want {
		t.Errorf("InputChannels() length = %d, want %d", got, want)
	}
}

//-----------------------------------------------------------------------------
// Channel
//

func TestChannelCleanName(t *testing.T) {
	for _, tt := range []struct {
		desc      string
		name      string
		cleanName string
	}{
		{"mono", "eGit", "eGit"},
		{"stereo-as-mono", "eGit-L, eGit-R", "eGit"},
		{"track with comma", "v1, v2", "v1, v2"},
	} {
		ch := &Channel{name: tt.name}
		if got, want := ch.CleanName(), tt.cleanName; got != want {
			t.Errorf("%s: CleanName() = %s, want %s", tt.desc, got, want)
		}
	}
}

package venue

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"sort"
	"testing"

	xmlpath "gopkg.in/xmlpath.v2"
)

const testdataPath = "testdata"

type TestData struct {
	name     string
	root     *xmlpath.Node
	console  string
	version  string
	show     string
	devNames []string
}

var testdata []*TestData

func init() {
	testdata = []*TestData{
		&TestData{
			name:     "20170526 ICF Conference Worship Night.html",
			console:  "Avid VENUE",
			version:  "VENUE 4.5.3",
			show:     "ICF Zurich\\20170526 Conf WN",
			devNames: []string{"Console", "Engine", "Pro Tools", "Stage 1", "Stage 2", "Stage 3", "Stage 4"}},
		&TestData{
			name:     "20170906 ICF Ladies Night.html",
			console:  "Avid VENUE",
			version:  "VENUE 4.5.3",
			show:     "ICF Zurich\\20170906 Ladies Night",
			devNames: []string{"Local", "Pro Tools", "Stage 1"}},
		&TestData{
			name:     "20170910 Avid D-Show Patch List.html",
			console:  "Avid VENUE",
			version:  "D-Show 3.1.1",
			show:     "GenX\\2017_09_10PM",
			devNames: []string{"Pro Tools", "Stage 1"}},
		&TestData{
			name:     "20170910 Avid D-Show System Info.html",
			console:  "Avid VENUE",
			version:  "D-Show 3.1.1",
			show:     "GenX\\2017_09_10PM",
			devNames: []string{"Pro Tools", "Stage 1"}},
		&TestData{
			name:     "20170910 Avid S3L-X Patch List.html",
			console:  "Avid VENUE",
			version:  "VENUE 4.5.3",
			show:     "01 ICF ZH Celebrations\\2017-09-10 Rec PM",
			devNames: []string{"Console", "Engine", "Pro Tools", "Stage 1", "Stage 2", "Stage 3", "Stage 4"}},
		&TestData{
			name:     "20170910 Avid S3L-X System Info.html",
			console:  "Avid VENUE",
			version:  "VENUE 4.5.3",
			show:     "01 ICF ZH Celebrations\\2017-09-10 Rec PM",
			devNames: []string{"Console", "Engine", "Pro Tools", "Stage 1", "Stage 2", "Stage 3", "Stage 4"}},
	}

	for _, td := range testdata {
		path := testdataPath + "/" + td.name
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("error reading testdata file %r; %s", path, err)
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

// TODO(Kate): Validate that channels are probed properly.
// func TestProbeChannels(t *testing.T) {
// }

func TestDiscoverDevices(t *testing.T) {
	for _, td := range testdata {
		devs, err := discoverDevices(td.root)
		if err != nil {
			t.Fatalf("discoverDevices(): unexpected error; %s", err)
		}
		devNames := []string{}
		for k, _ := range devs {
			devNames = append(devNames, k)
		}
		sort.Slice(devNames, func(i, j int) bool { return devNames[i] < devNames[j] })
		if got, want := devNames, td.devNames; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: discoverDevices(): device names = %q, want %q", td.name, got, want)
		}
	}
}

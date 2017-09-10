package venue

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"testing"

	xmlpath "gopkg.in/xmlpath.v2"
)

const testdataPath = "testdata"

type TestData struct {
	name    string
	root    *xmlpath.Node
	console string
	version string
	show    string
}

var testdata []*TestData

func init() {
	testdata = []*TestData{
		&TestData{
			name:    "20170526 ICF Conference Worship Night.html",
			console: "Avid VENUE",
			version: "VENUE 4.5.3",
			show:    "ICF Zurich\\20170526 Conf WN"},
		&TestData{
			name:    "20170906 ICF Ladies Night.html",
			console: "Avid VENUE",
			version: "VENUE 4.5.3",
			show:    "ICF Zurich\\20170906 Ladies Night"},
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

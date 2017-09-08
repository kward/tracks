package venue

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"testing"

	xmlpath "gopkg.in/xmlpath.v2"
)

const testdataFile = "testdata/20170526 ICF Conference Worship Night.html"

var root *xmlpath.Node

func init() {
	data, err := ioutil.ReadFile(testdataFile)
	if err != nil {
		log.Printf("error reading testdata file %r; %s", testdataFile, err)
		os.Exit(1)
	}
	node, err := xmlpath.ParseHTML(bytes.NewReader(data))
	if err != nil {
		log.Printf("error parsing HTML; %s", err)
		os.Exit(1)
	}
	root = node
}

func TestParseMetadata(t *testing.T) {
	v := NewVenue()
	if err := v.parseMetadata(root); err != nil {
		t.Fatalf("error parsing metadata; %s", err)
	}

	for _, tt := range []struct {
		desc    string
		element string
		want    string
	}{
		{"console", v.console, "Avid VENUE"},
		{"version", v.version, "VENUE 4.5.3"},
		{"show", v.show, "ICF Zurich\\20170526 Conf WN"},
	} {
		if got, want := tt.element, tt.want; got != want {
			t.Errorf("%s: element = %s; want %s", tt.desc, got, want)
		}
	}
}

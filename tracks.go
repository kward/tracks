package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	kioutil "github.com/kward/golib/io/ioutil"
	"github.com/kward/tracks/tracks"
	"github.com/kward/tracks/venue"
)

const (
	infoFileFlag = "info_file"
)

var (
	// TODO(Kate) Add support for copy.
	behaviors = map[string]bool{
		"copy": true, "cp": true,
		"link": true, "ln": true,
		"move": true, "mv": true,
		"rename": true,
	}

	behavior string
	dryRun   = flag.Bool("dry_run", false, "Do a dry run.")
	infoFile = flag.String(infoFileFlag, "", "Venue info file.")
	srcDir   = flag.String("src_dir", ".", "Source directory.")
	destDir  = flag.String("dest_dir", "", "Destination directory. Leave empty to rename in place.")

	fnReadDir = ioutil.ReadDir
)

func init() {
	bs := []string{}
	for b := range behaviors {
		bs = append(bs, b)
	}
	flag.StringVar(&behavior, "behavior", "move", fmt.Sprintf("Behavior; one of %s", bs))

	flag.Parse()
}

func flags() {
	if _, ok := behaviors[behavior]; !ok {
		fmt.Printf("unrecognized behavior %s\n", behavior)
		os.Exit(1)
	}
	switch behavior {
	case "cp":
		behavior = "copy"
	case "ln":
		behavior = "link"
	case "mv":
		behavior = "move"
	}

	if *destDir == "" {
		*destDir = *srcDir
	}
	if *infoFile == "" {
		fmt.Printf("empty %s flag\n", infoFileFlag)
		os.Exit(1)
	}
}

type Names struct {
	orig, dest string
}

func main() {
	flags()

	// Read Venue file.
	data, err := ioutil.ReadFile(*infoFile)
	if err != nil {
		fmt.Printf("error reading Venue info file; %s\n", err)
		os.Exit(1)
	}
	v := venue.NewVenue()
	if err := v.Parse(data); err != nil {
		fmt.Printf("error parsing the Venue data; %s\n", err)
		os.Exit(1)
	}

	files, err := DiscoverFiles(*srcDir, filterWaves)
	if err != nil {
		fmt.Printf("error discovering files; %s\n", err)
		os.Exit(1)
	}

	sessions, err := tracks.ExtractSessions(files)
	if err != nil {
		fmt.Printf("error extracting sessions; %s\n", err)
		os.Exit(1)
	}

	// Map tracks to stage boxes.
	for _, s := range sessions {
		ts, err := NameTracks(s.Tracks(), v.Devices())
		if err != nil {
			fmt.Printf("error mapping tracks; %s\n", err)
			os.Exit(1)
		}
		s.SetTracks(ts)
	}

	// Map tracks to new names.
	names := []Names{}
	for _, s := range sessions {
		for _, t := range s.Tracks().Slice() {
			name := t.Name()
			if name == "" {
				name = fmt.Sprintf("Track %02d", t.TrackNum())
			}
			dest := fmt.Sprintf("%02d-%02d %s.wav", s.Num(), t.TrackNum(), name)
			t.SetDest(dest)
			names = append(names, Names{t.Src(), t.Dest()})
		}
	}

	// Do work.
	if *dryRun {
		fmt.Println("Doing a dry run.")
	}

	switch behavior {
	case "copy":
		fmt.Println("Copying:")
	case "link":
		fmt.Println("Linking:")
	case "move":
		fmt.Println("Moving:")
	case "rename":
		fmt.Println("Renaming (moving):")
	}

	for _, name := range names {
		origPath := fmt.Sprintf("%s/%s", *srcDir, name.orig)
		destPath := fmt.Sprintf("%s/%s", *destDir, name.dest)
		fmt.Printf("  %q --> %q\n", origPath, destPath)
		if *dryRun {
			continue
		}
		switch behavior {
		case "copy":
			if _, err := kioutil.CopyFile(destPath, origPath); err != nil {
				fmt.Printf("error copying file; %s\n", err)
				os.Exit(1)
			}
		case "link":
			if err := os.Link(origPath, destPath); err != nil {
				fmt.Printf("error linking file; %s\n", err)
				os.Exit(1)
			}
		case "move", "rename":
			if err := os.Rename(origPath, destPath); err != nil {
				fmt.Printf("error moving file; %s\n", err)
				os.Exit(1)
			}
		}
	}
}

// DiscoverFiles looks for track names in a given directory, and returns them
// as a slice.
func DiscoverFiles(dir string, filters ...Filter) ([]string, error) {
	fileInfos, err := fnReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error discovering files in %q; %s", dir, err)
	}

	files := []string{}
	for _, fi := range fileInfos {
		files = append(files, fi.Name())
	}
	for _, filter := range filters {
		files = filter(files)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in %q", dir)
	}

	return files, nil
}

// Filter defines a file filter function.
type Filter func(unfiltered []string) (filtered []string)

func filterWaves(unfiltered []string) []string {
	filtered := []string{}
	for _, f := range unfiltered {
		if strings.HasSuffix(f, ".wav") {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// NameTracks based on their channel name.
func NameTracks(ts tracks.Tracks, ds venue.Devices) (tracks.Tracks, error) {
	for i, t := range ts {
		_, ch, err := mapTrackToChannel(t, ds)
		if err != nil {
			return nil, fmt.Errorf("error mapping track to channel; %s", err)
		}
		ts[i].SetName(ch.Name())
	}
	return ts, nil
}

func mapTrackToChannel(t *tracks.Track, devs venue.Devices) (*venue.Device, *venue.Channel, error) {
	// Walk the stage boxes in order, counting channels as we go.
	offset := 0

	for _, name := range []string{"Stage 1", "Stage 2", "Stage 3", "Stage 4"} {
		// Check that stage box was configured.
		sb, ok := devs[name]
		if !ok {
			continue
		}
		// Check whether track is on this stage box.
		if t.TrackNum() > offset+sb.NumInputs() {
			offset += sb.NumInputs()
			continue
		}
		// Found it.
		moniker := fmt.Sprintf("%d", t.TrackNum()-offset)
		return sb, sb.Input(moniker), nil
	}
	return nil, nil, fmt.Errorf("channel not found")
}

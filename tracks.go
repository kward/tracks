package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

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
		"move": true, "mv": true,
		"rename": true,
	}

	behavior string
	dryRun   = flag.Bool("dry_run", false, "Do a dry run.")
	infoFile = flag.String(infoFileFlag, "", "Venue info file.")
	srcDir   = flag.String("src_dir", ".", "Source directory.")
	destDir  = flag.String("dest_dir", "", "Destination directory. Leave empty to rename in place.")
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

	// Discover sessions and tracks.
	sessions := tracks.NewSessions()
	if err := sessions.Discover(*srcDir); err != nil {
		fmt.Printf("error discovering sessions; %s\n", err)
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
				name = fmt.Sprintf("Track %02d", t.Num())
			}
			dest := fmt.Sprintf("%02d-%02d %s.wav", s.Num(), t.Num(), name)
			t.SetDestName(dest)
			names = append(names, Names{t.OrigName(), t.DestName()})
		}
	}

	// Do work.
	switch behavior {
	case "copy":
		fmt.Println("Copying:")
	case "move":
		fmt.Println("Moving:")
	case "rename":
		fmt.Println("Renaming (moving):")
	}

	for _, name := range names {
		origPath := fmt.Sprintf("%s/%s", *srcDir, name.orig)
		destPath := fmt.Sprintf("%s/%s", *destDir, name.dest)
		fmt.Printf("  %s --> %s\n", origPath, destPath)
		if *dryRun {
			continue
		}
		switch behavior {
		case "copy":
			if _, err := kioutil.CopyFile(destPath, origPath); err != nil {
				fmt.Printf("error copying file; %s", err)
				os.Exit(1)
			}
		case "move", "rename":
			if err := os.Rename(origPath, destPath); err != nil {
				fmt.Printf("error moving file; %s", err)
				os.Exit(1)
			}
		}
	}
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
		if t.Num() > offset+sb.NumInputs() {
			offset += sb.NumInputs()
			continue
		}
		// Found it.
		moniker := fmt.Sprintf("%d", t.Num()-offset)
		return sb, sb.Input(moniker), nil
	}
	return nil, nil, fmt.Errorf("channel not found")
}

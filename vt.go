package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kward/venue-tracks/vt"
)

var (
	behaviors = map[string]bool{"rename": true, "move": true}

	behavior  string
	dryRun    = flag.Bool("dry_run", false, "Do a dry run.")
	patchFile = flag.String("patchfile", "", "Venue patch file.")
	srcDir    = flag.String("srcdir", ".", "Source directory.")
	destDir   = flag.String("destdir", "", "Destination directory. Leave empty to rename in place.")
)

func init() {
	bs := []string{}
	for b := range behaviors {
		bs = append(bs, b)
	}
	flag.StringVar(&behavior, "behavior", "rename", fmt.Sprintf("Behavior; one of %s", bs))

	flag.Parse()

	if _, ok := behaviors[behavior]; !ok {
		fmt.Printf("unrecognized behavior %s\n", behavior)
		os.Exit(1)
	}
	if *destDir == "" {
		*destDir = *srcDir
	}
	if *patchFile == "" {
		fmt.Printf("empty patchfile flag\n")
		os.Exit(1)
	}
}

func main() {
	// Read Venue file.
	data, err := ioutil.ReadFile(*patchFile)
	if err != nil {
		fmt.Printf("error reading venue patch file %q; %s\n", *patchFile, err)
		os.Exit(1)
	}
	venue := vt.NewVenue()
	if err := venue.Parse(data); err != nil {
		fmt.Printf("error parsing the Venue data; %s\n", err)
		os.Exit(1)
	}
	//fmt.Println("venue:", venue)

	// Discover sessions and tracks.
	sessions := vt.NewSessions()
	if err := sessions.Discover(*srcDir); err != nil {
		fmt.Printf("error discovering sessions; %s\n", err)
		os.Exit(1)
	}
	//fmt.Println("sessions:", sessions)

	// Map tracks to stage boxes.
	for _, s := range sessions {
		ts, err := venue.NameTracks(s.Tracks())
		if err != nil {
			fmt.Printf("error mapping tracks; %s\n", err)
			os.Exit(1)
		}
		//fmt.Printf("session: %s tracks: %v\n", s.Name(), ts)
		s.SetTracks(ts)
	}

	// Map tracks to new names.
	type rename struct{ oldName, newName string }
	names := []rename{}
	for _, s := range sessions {
		for _, t := range s.Tracks().Slice() {
			name := t.Name()
			if name == "" {
				name = fmt.Sprintf("Track %02d", t.Num())
			}
			newName := fmt.Sprintf("%02d-%02d %s.wav", s.Num(), t.Num(), name)
			t.SetNewName(newName)
			names = append(names, rename{t.OrigName(), t.NewName()})
		}
	}

	// Do renames.
	fmt.Println("Renaming:")
	for _, name := range names {
		oldPath := fmt.Sprintf("%s/%s", *srcDir, name.oldName)
		newPath := fmt.Sprintf("%s/%s", *destDir, name.newName)
		fmt.Printf("  %s --> %s\n", oldPath, newPath)
		if *dryRun {
			continue
		}
		if err := os.Rename(oldPath, newPath); err != nil {
			fmt.Printf("error renaming %s to %s; %s", oldPath, newPath, err)
			os.Exit(1)
		}
	}
}

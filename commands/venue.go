package commands

import (
	"fmt"
	"io/ioutil"
	"os"

	k8os "github.com/kward/golib/os"
	"github.com/kward/golib/os/sysexits"
	"github.com/kward/tracks/action"
	"github.com/kward/tracks/tracks"
	"github.com/kward/tracks/venue"
	"github.com/urfave/cli"
)

func init() {
	c := "venue"
	f := []cli.Flag{
		cli.StringFlag{
			Name:  "patch_file,p",
			Usage: "Venue patch or info file",
		},
		cli.StringFlag{
			Name:  "src_dir,s",
			Usage: "source directory",
		},
		cli.StringFlag{
			Name:  "dest_dir,d",
			Usage: "destination directory (leave empty if renaming in-place)",
		},
	}
	commands = append(commands, []cli.Command{
		{
			Name:     "copy",
			Usage:    "copy tracks with new names",
			Category: c,
			Flags:    f,
			Action:   VenueCopyAction,
			After:    VenueDryRunAction,
		}, {
			Name:     "link",
			Usage:    "make links with new names, without removing original files",
			Category: c,
			Flags:    f,
			Action:   VenueLinkAction,
			After:    VenueDryRunAction,
		}, {
			Name:     "move",
			Usage:    "move or rename tracks",
			Category: c,
			Flags:    f,
			Action:   VenueMoveAction,
			After:    VenueDryRunAction,
		},
	}...)
}

// VenueFlags holds the values of user-defined flags.
type VenueFlags struct {
	dryRun          bool
	patchFile       string
	srcDir, destDir string
}

func venueFlags(ctx *cli.Context) (VenueFlags, error) {
	// Validate flags.
	for _, f := range []string{"patch_file", "src_dir"} {
		if !ctx.IsSet(f) {
			return VenueFlags{}, fmt.Errorf("missing %s flag", f)
		}
	}
	if !ctx.IsSet("dest_dir") {
		ctx.Set("dest_dir", ctx.String("src_dir"))
	}

	// Parse flags.
	return VenueFlags{
		dryRun:    ctx.GlobalBool("dry_run"),
		patchFile: ctx.String("patch_file"),
		srcDir:    ctx.String("src_dir"),
		destDir:   ctx.String("dest_dir"),
	}, nil
}

// VenueCopyAction implements cli.ActionFunc.
func VenueCopyAction(ctx *cli.Context) error {
	flags, err := venueFlags(ctx)
	if err != nil {
		return cli.NewExitError(err, sysexits.Usage.Int())
	}
	names, err := venueNames(flags)
	if err != nil {
		return cli.NewExitError(err, sysexits.Software.Int())
	}
	fmt.Println("Copying:")
	if err := venueBatch(flags, k8os.Copy, names); err != nil {
		return cli.NewExitError(fmt.Sprintf("error copying file, %s", err), sysexits.Software.Int())
	}
	return nil
}

// VenueLinkAction implements cli.ActionFunc.
func VenueLinkAction(ctx *cli.Context) error {
	flags, err := venueFlags(ctx)
	if err != nil {
		return cli.NewExitError(err, sysexits.Usage.Int())
	}
	names, err := venueNames(flags)
	if err != nil {
		return cli.NewExitError(err, sysexits.Software.Int())
	}
	fmt.Println("Linking:")
	if err := venueBatch(flags, os.Link, names); err != nil {
		return cli.NewExitError(fmt.Sprintf("error copying file, %s", err), sysexits.Software.Int())
	}
	return nil
}

// VenueMoveAction implements cli.ActionFunc.
func VenueMoveAction(ctx *cli.Context) error {
	flags, err := venueFlags(ctx)
	if err != nil {
		return cli.NewExitError(err, sysexits.Usage.Int())
	}
	names, err := venueNames(flags)
	if err != nil {
		return cli.NewExitError(err, sysexits.Software.Int())
	}
	fmt.Println("Moving:")
	if err := venueBatch(flags, os.Rename, names); err != nil {
		return cli.NewExitError(fmt.Sprintf("error copying file, %s", err), sysexits.Software.Int())
	}
	return nil
}

// VenueDryRunAction implements cli.ActionFunc.
func VenueDryRunAction(ctx *cli.Context) error {
	if ctx.GlobalBool("dry_run") {
		fmt.Fprintf(os.Stderr, "-- dry run mode --\n")
	}
	return nil
}

type VenueNames struct {
	orig, dest string
}

func venueNames(flags VenueFlags) ([]VenueNames, error) {
	// Read Venue file.
	data, err := ioutil.ReadFile(flags.patchFile)
	if err != nil {
		return nil, fmt.Errorf("error reading Venue patch file, %s", err)
	}
	v := venue.NewVenue()
	if err := v.Parse(data); err != nil {
		return nil, fmt.Errorf("error parsing the Venue data, %s", err)
	}

	files, err := action.DiscoverFiles(flags.srcDir, action.FilterWaves)
	if err != nil {
		return nil, fmt.Errorf("error discovering wave files, %s", err)
	}

	sessions, err := tracks.ExtractSessions(files)
	if err != nil {
		return nil, fmt.Errorf("error extracting sessions, %s", err)
	}

	// Map tracks to stage boxes.
	// TODO(20171225 kward): Move to action package.
	for _, s := range sessions {
		ts, err := action.MapTracksToNames(s.Tracks(), v.Devices())
		if err != nil {
			return nil, fmt.Errorf("error mapping tracks, %s", err)
		}
		s.SetTracks(ts)
	}

	// Map tracks to new names.
	names := []VenueNames{}
	for _, s := range sessions {
		for _, t := range s.Tracks().Slice() {
			name := t.Name()
			if name == "" {
				name = fmt.Sprintf("Track %02d", t.TrackNum())
			}
			dest := fmt.Sprintf("%02d-%02d %s.wav", s.Num(), t.TrackNum(), action.MapTrackNameToFilename(name))
			t.SetDest(dest)
			names = append(names, VenueNames{t.Src(), t.Dest()})
		}
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("no tracks found")
	}
	return names, nil
}

func venueBatch(flags VenueFlags, fn func(src, dest string) error, names []VenueNames) error {
	for _, name := range names {
		origPath := fmt.Sprintf("%s/%s", flags.srcDir, name.orig)
		destPath := fmt.Sprintf("%s/%s", flags.destDir, name.dest)
		fmt.Printf("  %q --> %q\n", origPath, destPath)
		if flags.dryRun {
			continue
		}
		if err := fn(origPath, destPath); err != nil {
			return err
		}
	}
	return nil
}

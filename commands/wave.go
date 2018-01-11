package commands

import (
	"fmt"
	"os"

	"github.com/kward/golib/os/sysexits"
	"github.com/kward/tracks/actions"
	"github.com/urfave/cli"
)

func init() {
	c := "wave"
	commands = append(commands, []cli.Command{
		{
			Name:     "check",
			Usage:    "check wave files for known errors",
			Category: c,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "file,f", Usage: "wave filename"},
			},
			Action: WaveCheckAction,
		},
		{
			Name:     "dump",
			Usage:    "dump raw wave sample data",
			Category: c,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "file,f", Usage: "wave filename"},
				cli.DurationFlag{Name: "offset,o", Usage: "offset duration"},
				cli.DurationFlag{Name: "length,l", Usage: "sample length"},
			},
			Action: WaveDumpAction,
		},
		{
			Name:     "info",
			Usage:    "output info about wave file",
			Category: c,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "file,f", Usage: "wave filename"},
			},
			Action: WaveInfoAction,
		},
	}...)
}

// WaveCheckAction implements cli.ActionFunc.
func WaveCheckAction(ctx *cli.Context) error {
	if !ctx.IsSet("file") {
		return cli.NewExitError(fmt.Errorf("--file flag missing"), sysexits.Usage.Int())
	}

	filename := ctx.String("file")
	f, err := os.Open(filename)
	if err != nil {
		return cli.NewExitError(err, sysexits.IOError.Int())
	}
	defer f.Close()

	if err := actions.WaveCheck(filename); err != nil {
		return cli.NewExitError(err, sysexits.DataError.Int())
	}
	return nil
}

// WaveDumpAction implements cli.ActionFunc.
func WaveDumpAction(ctx *cli.Context) error {
	if !ctx.IsSet("file") {
		return cli.NewExitError(fmt.Errorf("missing %s flag", "file"), sysexits.Usage.Int())
	}

	block, frames, err := actions.WaveDump(ctx.String("file"), ctx.Duration("offset"), ctx.Duration("length"))
	if err != nil {
		return err
	}
	fmt.Printf("frames: %d\n", frames)

	for o := 0; o < frames; o += 4 {
		d := block[o : o+4]
		fmt.Printf("%08x  %g %g %g %g\n", o, d[0], d[1], d[2], d[3])
	}
	return nil
}

// WaveInfoAction implements cli.ActionFunc.
func WaveInfoAction(ctx *cli.Context) error {
	if !ctx.IsSet("file") {
		return cli.NewExitError(fmt.Errorf("missing %s flag", "file"), sysexits.Usage.Int())
	}
	info, err := actions.WaveInfo(ctx.String("file"))
	if err != nil {
		return err
	}
	fmt.Println(info)
	return nil
}

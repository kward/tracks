package commands

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kward/golib/os/sysexits"
	"github.com/kward/tracks/actions"
	"github.com/loov/audio/codec/wav"
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
				cli.StringFlag{
					Name:  "dir,s",
					Usage: "directory containing wave files",
				},
			},
			Action: WaveCheckAction,
		}, {
			Name:     "info",
			Usage:    "output info about wave file",
			Category: c,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file,f",
					Usage: "file to get info for",
				},
			},
			Action: WaveInfoAction,
		},
	}...)
}

// WaveCheckAction implements cli.ActionFunc.
func WaveCheckAction(ctx *cli.Context) error {
	if !ctx.IsSet("dir") {
		return cli.NewExitError(fmt.Errorf("missing %s flag", "dir"), sysexits.Usage.Int())
	}
	dir := ctx.String("dir")

	files, err := actions.DiscoverFiles(dir, actions.FilterWaves)
	if err != nil {
		return cli.NewExitError(err, sysexits.Software.Int())
	}

	for _, file := range files {
		fqFile := dir + string(os.PathSeparator) + file

		f, err := os.Open(fqFile)
		if err != nil {
			fmt.Printf("%s : %s\n", file, err.Error())
		}
		defer f.Close()

		if err := waveCheck(fqFile); err != nil {
			fmt.Printf("%s : %s\n", file, err.Error())
			continue
		}
		fmt.Printf("%s : OK\n", file)
	}

	return nil
}

type WaveCheckResult struct {
	file   string
	result string
}

func waveCheck(file string) error {
	d, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	r, err := wav.NewBytesReader(d)
	if err != nil {
		return err
	}
	if r.SampleRate() <= 0 {
		return fmt.Errorf("invalid sample rate")
	}
	return nil
}

// WaveInfoAction implements cli.ActionFunc.
func WaveInfoAction(ctx *cli.Context) error {
	if !ctx.IsSet("file") {
		return cli.NewExitError(fmt.Errorf("missing %s flag", "file"), sysexits.Usage.Int())
	}
	file := ctx.String("file")
	info, err := waveInfo(file)
	if err != nil {
		return err
	}
	fmt.Println(info)
	return nil
}

func waveInfo(file string) (string, error) {
	d, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	r, err := wav.NewBytesReader(d)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sample_rate: %d channels: %d duration: %s",
		r.SampleRate(), r.ChannelCount(), r.Duration()), nil
}

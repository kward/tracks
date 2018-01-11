package actions

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/kward/goaudio/codec/wav"
)

func WaveCheck(file string) error {
	r, err := waveReader(file)
	if err != nil {
		return err
	}

	cap := int(1 * r.SampleRate() * r.ChannelCount()) // 1 sec of data
	blk := make([]float32, cap, cap)
	start := 0 * time.Second
	o := 0
	for ; o < r.FrameCount(); o += cap {
		frames := r.ReadBlock(blk)
		zeros := 0
		for f := 0; f < frames; f++ {
			if blk[f] != 0 {
				zeros = 0
				if start > 0 {
					fmt.Printf("%s - %s\n", start, time.Duration(o/cap)*time.Second)
					start = 0 * time.Second
				}
				continue
			}
			zeros++
			// React if we have 10 consecutive zeros.
			if zeros == 10 && start == 0 {
				start = time.Duration(o/cap) * time.Second
			}
		}
	}
	if start > 0 {
		fmt.Printf("%s - %s\n", start, time.Duration(o)*time.Second)
	}

	return nil
}

func WaveDump(file string, offset, length time.Duration) ([]float32, int, error) {
	r, err := waveReader(file)
	if err != nil {
		return []float32{}, 0, err
	}

	cap := int(length.Seconds() * float64(r.SampleRate()))
	block := make([]float32, cap, cap)
	r.Seek(offset)
	frames := r.ReadBlock(block)

	return block, frames, nil
}

func WaveInfo(file string) (string, error) {
	r, err := waveReader(file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("sample_rate: %d Hz, channels: %d bits_per_sample: %d frame_count: %d duration: %s",
		r.SampleRate(), r.ChannelCount(), r.BitsPerSample(), r.FrameCount(), r.Duration()), nil
}

func waveReader(file string) (*wav.Reader, error) {
	d, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	r, err := wav.NewBytesReader(d)
	if err != nil {
		return nil, err
	}

	if r.SampleRate() <= 0 {
		return nil, fmt.Errorf("invalid sample rate of %d", r.SampleRate())
	}

	return r, nil
}

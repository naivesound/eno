package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/zserge/rtaudio/contrib/go/rtaudio"
)

type eno struct {
	m metronome
}

func startAudio(ctx context.Context, sampleRate int, cb func([]int16)) error {
	audio, err := rtaudio.Create(rtaudio.APIUnspecified)
	if err != nil {
		return err
	}
	defer audio.Destroy()

	params := &rtaudio.StreamParams{
		DeviceID:     uint(audio.DefaultOutputDevice()),
		NumChannels:  uint(2),
		FirstChannel: 0,
	}
	err = audio.Open(params, nil, rtaudio.FormatInt16, uint(sampleRate), 128,
		func(out, in rtaudio.Buffer, dur time.Duration, status rtaudio.StreamStatus) int {
			if status == rtaudio.StatusOutputUnderflow {
				return 0
			}
			b := out.Int16()
			for i := 0; i < len(b); i++ {
				b[i] = 0
			}
			cb(b)
			return 0
		}, nil)
	if err != nil {
		return err
	}
	defer audio.Close()

	if err := audio.Start(); err != nil {
		return err
	}
	<-ctx.Done()
	return nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	defer signal.Stop(sigc)

	sampleRate := 44100

	synth := NewSynth(sampleRate)
	synth.Load("font.sf2")
	synth.SetGain(0.4)

	m := NewMetronome(sampleRate)
	m.SetBPM(120)
	m.SetGain(0.1)

	looper := NewLooper(sampleRate)

	go startAudio(ctx, sampleRate, func(out []int16) {
		synth.MixStereo(out)
		looper.MixStereo(out)
		m.MixStereo(out)
	})

	<-sigc
}

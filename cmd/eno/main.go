package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"reflect"
	"time"

	"github.com/hypebeast/go-osc/osc"
	"github.com/naivesound/eno/pkg/looper"
	"github.com/naivesound/eno/pkg/metronome"
	"github.com/naivesound/eno/pkg/synth"
	"github.com/thestk/rtaudio/contrib/go/rtaudio"
	"github.com/thestk/rtmidi/contrib/go/rtmidi"
)

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

func startMIDI(ctx context.Context, cb func([]byte)) error {
	midi, err := rtmidi.NewMIDIInDefault()
	if err != nil {
		return err
	}
	_, _ = midi.PortCount()

	<-ctx.Done()
	return nil
}

func OSCHandler(f interface{}) osc.HandlerFunc {
	return func(msg *osc.Message) {
		call := reflect.ValueOf(f)
		if len(msg.Arguments) != call.Type().NumIn() {
			log.Println("bad number of OSC arguments:", len(msg.Arguments), "!=", call.Type().NumIn())
			return
		}
		args := make([]reflect.Value, call.Type().NumIn())
		for i := 0; i < len(args); i++ {
			arg := reflect.ValueOf(msg.Arguments[i])
			if !arg.Type().ConvertibleTo(call.Type().In(i)) {
				log.Println("argument", i, "is not convertible")
				return
			}
			args[i] = arg.Convert(call.Type().In(i))
		}
		call.Call(args)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	defer signal.Stop(sigc)

	sampleRate := 44100

	synth := synth.New(sampleRate)
	synth.Load("font.sf2")
	synth.SetGain(0.4)

	m := metronome.New(sampleRate)
	m.SetBPM(0)
	m.SetGain(0.2)

	looper := looper.New(sampleRate)

	go startAudio(ctx, sampleRate, func(out []int16) {
		synth.MixStereo(out)
		looper.MixStereo(out)
		m.MixStereo(out)
	})

	go startMIDI(ctx, func(cmd []byte) {

	})

	go func() {
		srv := &osc.Server{}

		srv.Handle("/eno/synth/on", OSCHandler(synth.NoteOn))
		srv.Handle("/eno/synth/off", OSCHandler(synth.NoteOff))
		srv.Handle("/eno/synth/pitch", OSCHandler(synth.PitchBend))
		srv.Handle("/eno/synth/cc", OSCHandler(synth.ControlChange))
		srv.Handle("/eno/synth/load", OSCHandler(synth.Load))
		srv.Handle("/eno/synth/gain", OSCHandler(synth.SetGain))

		srv.Handle("/eno/metronome/tap", OSCHandler(m.Tap))
		srv.Handle("/eno/metronome/gain", OSCHandler(m.SetGain))
		srv.Handle("/eno/metronome/bpm", OSCHandler(m.SetBPM))

		for port := 7027; port < 7030; port++ {
			ln, err := net.ListenPacket("udp", fmt.Sprintf("127.0.0.1:%d", port))
			if err != nil {
				continue
			}
			log.Println("starting OSC server on UDP port", port)
			srv.Serve(ln)
			return
		}
		log.Println("failed to find a port in the range 7027..7029")
	}()

	<-sigc
}

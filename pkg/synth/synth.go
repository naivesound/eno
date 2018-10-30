package synth

import (
	"sync"

	"github.com/naivesound/tsf"
)

type Synth interface {
	Load(filename string)
	SetGain(gain float32)
	NoteOn(ch, note, vel int)
	NoteOff(ch, note int)
	PitchBend(ch, bend int)
	ControlChange(ch, cc, value int)
	MixStereo(out []int16)
}

type synth struct {
	sync.Mutex
	sf         *tsf.TSF
	gain       float32
	sampleRate int
}

func New(sampleRate int) Synth {
	return &synth{
		sampleRate: sampleRate,
		gain:       1,
	}
}

func (s *synth) Load(filename string) {
	s.Lock()
	defer s.Unlock()
	if s.sf != nil {
		s.sf.Close()
		s.sf = nil
	}
	s.sf = tsf.NewFile(filename)
	if s.sf != nil {
		s.sf.SetOutput(tsf.ModeStereoInterleaved, s.sampleRate, s.gain)
		s.sf.SetChannelPresetNumber(0, 0)
	}
}

func (s *synth) SetGain(gain float32) {
	s.Lock()
	defer s.Unlock()
	s.gain = gain
	if s.sf != nil {
		s.sf.SetOutput(tsf.ModeStereoInterleaved, s.sampleRate, s.gain)
	}
}

func (s *synth) NoteOn(channel, note, velocity int) {
	s.Lock()
	defer s.Unlock()
	if s.sf != nil {
		s.sf.ChannelNoteOn(channel, note, float32(velocity)/128)
	}
}

func (s *synth) NoteOff(channel, note int) {
	s.Lock()
	defer s.Unlock()
	if s.sf != nil {
		s.sf.ChannelNoteOff(channel, note)
	}
}

func (s *synth) PitchBend(channel, bend int) {
	s.Lock()
	defer s.Unlock()
	if s.sf != nil {
		s.sf.SetChannelPitchWheel(channel, bend)
	}
}

func (s *synth) ControlChange(channel, cc, value int) {
	s.Lock()
	defer s.Unlock()
	if s.sf != nil {
		s.sf.ChannelMIDIControl(channel, cc, value)
	}
}

func (s *synth) MixStereo(out []int16) {
	s.Lock()
	defer s.Unlock()
	if s.sf != nil {
		s.sf.RenderInt16(out)
	}
}

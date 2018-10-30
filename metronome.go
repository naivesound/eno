package main

import (
	"sync"
)

type metronome struct {
	sync.Mutex
	sampleRate int
	frame      int
	taps       []int
	gain       float32
	bpm        int
}

func NewMetronome(sampleRate int) *metronome {
	return &metronome{
		sampleRate: sampleRate,
		gain:       1,
		bpm:        120,
	}
}

const (
	maxBPMIdleSeconds = 2
	maxBPMTaps        = 8
)

func (m *metronome) Tap() {
	m.Lock()
	defer m.Unlock()

	// Check if there was no taps for a long time - if so, reset the metronome
	if len(m.taps) > 0 {
		maxInterval := m.sampleRate * maxBPMIdleSeconds
		if m.bpm > 0 {
			maxInterval = m.sampleRate * 2 * 60 / m.bpm
		}
		if m.taps[len(m.taps)-1] > maxInterval {
			m.taps = nil
		}
	}

	// Append tap
	m.taps = append(m.taps, 0)
	m.frame = 0

	// Trim slice of taps if needed
	for len(m.taps) > maxBPMTaps {
		m.taps = m.taps[1:]
	}

	// Calculate average BPM
	if len(m.taps) > 1 {
		sum := 0
		for i := 0; i < len(m.taps); i++ {
			sum = sum + m.taps[i]
		}
		spb := sum / len(m.taps)
		m.bpm = 60 * m.sampleRate / spb
	}
}

func (m *metronome) SetBPM(bpm int) {
	m.Lock()
	defer m.Unlock()
	m.taps = nil
	m.bpm = bpm
}

func (m *metronome) SetGain(gain float32) {
	m.Lock()
	defer m.Unlock()
	m.gain = gain
}

func (m *metronome) MixStereo(out []int16) {
	m.Lock()
	defer m.Unlock()

	frames := len(out) / 2

	if len(m.taps) > 0 {
		m.taps[len(m.taps)-1] += frames
	}

	if m.bpm > 0 {
		beatDuration := int(float32(m.sampleRate) * 60 / float32(m.bpm))
		for i := 0; i < frames; i++ {
			if m.frame == 0 {
				// Report events: Beat
			}
			frame := int(m.frame * 44100 / m.sampleRate)
			if frame < len(metronomeAcousticSample)/2 {
				v := s16le(metronomeAcousticSample[frame*2+1], metronomeAcousticSample[frame*2])
				v = int16(float32(v) * m.gain)
				out[i*2] = mix(out[i*2], v)
				out[i*2+1] = mix(out[i*2+1], v)
			}
			m.frame = (m.frame + 1) % beatDuration
		}
	}
}

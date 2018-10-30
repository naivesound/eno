package looper

import "sync"

type looperState int

const (
	looperStateClear looperState = iota
	looperStateRecording
	looperStateOverdubbing
	looperStatePlaying
	looperStateStopped
)

type Looper interface {
	Tap()
	Cancel()
	SetGain(gain float32)
	SetDecay(decay float32)
	MixStereo(out []int16)
}

type looper struct {
	sync.Mutex
	state   looperState
	gain    float32
	decay   float32
	pos     int
	frames  int
	overdub []int16
	loop    []int16
}

func New(sampleRate int) *looper {
	return &looper{
		gain:  1,
		decay: 0.5,
	}
}

func (l *looper) Tap() {
	l.Lock()
	defer l.Unlock()
}

func (l *looper) Cancel() {
	l.Lock()
	defer l.Unlock()
}

func (l *looper) SetGain(gain float32) {
	l.Lock()
	defer l.Unlock()
}

func (l *looper) SetDecay(gain float32) {
	l.Lock()
	defer l.Unlock()
}

func (l *looper) MixStereo(out []int16) {
	l.Lock()
	defer l.Unlock()
}

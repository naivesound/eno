package main

type looper struct {
}

func NewLooper(sampleRate int) *looper {
	return &looper{}
}

func (l *looper) Start() {
}

func (l *looper) End() {
}

func (l *looper) Stop() {
}

func (l *looper) Clear() {
}

func (l *looper) SetGain(gain float32) {
}

func (l *looper) MixStereo(out []int16) {

}

package main

import (
	"math"
	"math/rand"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

var SampleRate = beep.SampleRate(44100)

func Init() {
	rand.Seed(time.Now().UnixNano())
	speaker.Init(SampleRate, SampleRate.N(time.Second/10))
}

type EnvelopedSine struct {
	Freq    float64
	Phase   float64
	SR      beep.SampleRate
	Pos     int
	Total   int
	FadeLen int
	Amp     float64
}

func (s *EnvelopedSine) Stream(samples [][2]float64) (n int, ok bool) {
	for i := range samples {
		if s.Pos >= s.Total {
			return i, false
		}
		v := math.Sin(2 * math.Pi * s.Phase)

		env := 1.0
		if s.Pos < s.FadeLen {
			env = float64(s.Pos) / float64(s.FadeLen)
		} else if s.Pos > s.Total-s.FadeLen {
			env = float64(s.Total-s.Pos) / float64(s.FadeLen)
		}

		v *= env * s.Amp
		samples[i][0] = v
		samples[i][1] = v

		s.Phase += s.Freq / float64(s.SR)
		s.Pos++
	}
	return len(samples), true
}

func (s *EnvelopedSine) Err() error {
	return nil
}

type Chord struct {
	Tones []*EnvelopedSine
}

func (c *Chord) Stream(samples [][2]float64) (n int, ok bool) {
	buf := make([][2]float64, len(samples))
	anyAlive := false

	for _, t := range c.Tones {
		for i := range buf {
			buf[i] = [2]float64{0, 0}
		}
		tn, tok := t.Stream(buf)
		if tok || tn > 0 {
			anyAlive = true
		}
		for i := 0; i < tn; i++ {
			samples[i][0] += buf[i][0]
			samples[i][1] += buf[i][1]
		}
		if tn > n {
			n = tn
		}
	}
	if !anyAlive {
		return n, false
	}
	return n, true
}

func (c *Chord) Err() error {
	return nil
}

func PlayChord(freqs []float64, amps []float64, d time.Duration) {
	total := SampleRate.N(d)
	fadeLen := SampleRate.N(20 * time.Millisecond)

	tones := make([]*EnvelopedSine, len(freqs))
	for i, f := range freqs {
		amp := 0.5
		if i < len(amps) {
			amp = amps[i]
		}
		tones[i] = &EnvelopedSine{
			Freq:    f,
			SR:      SampleRate,
			Total:   total,
			FadeLen: fadeLen,
			Amp:     amp,
		}
	}

	speaker.Play(&Chord{Tones: tones})
}

func PlayBounceNote(index int) {
	melody := []struct {
		root float64
		name string
	}{
		{739.99, "F#5"},
		{659.25, "E5"},
		{587.33, "D5"},
		{493.88, "B4"},
		{587.33, "D5"},
		{659.25, "E5"},
		{739.99, "F#5"},
		{880.00, "A5"},
		{739.99, "F#5"},
		{587.33, "D5"},
		{493.88, "B4"},
		{440.00, "A4"},
	}

	note := melody[index%len(melody)]
	root := note.root
	fifth := root * 1.5      // Perfect fifth
	octaveDown := root * 0.5 // Octave below for dept

	PlayChord(
		[]float64{root, fifth, octaveDown},
		[]float64{0.35, 0.15, 0.1},
		180*time.Millisecond,
	)
}

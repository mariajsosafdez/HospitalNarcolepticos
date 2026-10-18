package gameview

import (
	"testing"

	"costenos-narcolepsia/game"
)

func TestEntersFromLeft(t *testing.T) {
	tests := []struct {
		x    float64
		want bool
	}{{10, true}, {83.9, true}, {84, false}, {150, false}}
	for _, tt := range tests {
		if got := entersFromLeft(tt.x); got != tt.want {
			t.Errorf("entersFromLeft(%v) = %v, want %v", tt.x, got, tt.want)
		}
	}
}

func TestStandX(t *testing.T) {
	tests := []struct {
		name                string
		left, width, staffW int
		fromLeft            bool
		want                float64
	}{
		{"left of a patient in the middle", 60, 32, 16, true, 56},
		{"right of a patient in the middle", 60, 32, 16, false, 96},
		{"clamped at the left edge", 2, 32, 16, true, 8},
		{"clamped at the right edge", 140, 32, 16, false, 160},
	}
	for _, tt := range tests {
		if got := standX(tt.left, tt.width, tt.fromLeft, tt.staffW); got != tt.want {
			t.Errorf("%s: standX = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestStagedX(t *testing.T) {
	const stand, edge = 56.0, -8.0
	tests := []struct {
		kind     game.BeatKind
		progress float64
		want     float64
		walking  bool
	}{
		{game.BeatArrive, 0, edge, true},
		{game.BeatArrive, 1, stand, true},
		{game.BeatArrive, 0.5, 24, true},
		{game.BeatNoBed, 0.5, stand, false},
		{game.BeatLeave, 0, stand, true},
		{game.BeatLeave, 1, edge, true},
	}
	for _, tt := range tests {
		x, walking := stagedX(tt.kind, tt.progress, stand, edge)
		if x != tt.want || walking != tt.walking {
			t.Errorf("stagedX(%v, %v) = %v, %v; want %v, %v", tt.kind, tt.progress, x, walking, tt.want, tt.walking)
		}
	}
}

func TestStretcherX(t *testing.T) {
	// Sale por la izquierda: de x = 40 a fuera de la cámara (centro en -22 para 44 de ancho).
	if got := stretcherX(40, true, 44, 0); got != 40 {
		t.Errorf("start = %v", got)
	}
	if got := stretcherX(40, true, 44, 1); got != -22 {
		t.Errorf("end left = %v, want -22", got)
	}
	if got := stretcherX(120, false, 44, 1); got != feedW+22 {
		t.Errorf("end right = %v, want %v", got, float64(feedW+22))
	}
}

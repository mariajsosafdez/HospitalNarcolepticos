package gameview

import (
	"math/rand"
	"testing"
)

func TestZzzSpot(t *testing.T) {
	tests := []struct {
		name       string
		x, y, w, h int
		wantX      int
	}{
		{"acostado: sobre la cabeza (izquierda)", 50, 80, 34, 14, 51},
		{"acostado Wilfrido", 20, 80, 32, 14, 21},
		{"acostado pegado al borde izquierdo", -3, 80, 47, 14, 0},
		{"silla (Marelvis): hombro derecho", 50, 80, 20, 24, 65},
		{"de pie (Yeimy)", 50, 80, 16, 28, 61},
		{"silla pegada a la derecha", 150, 80, 20, 24, 168 - 11},
	}
	for _, tt := range tests {
		x, y := zzzSpot(tt.x, tt.y, tt.w, tt.h, 11)
		if x != tt.wantX || y != tt.y-16 {
			t.Errorf("%s: zzzSpot = (%d,%d), quiero (%d,%d)", tt.name, x, y, tt.wantX, tt.y-16)
		}
	}
}

func TestIsLying(t *testing.T) {
	tests := []struct {
		w, h int
		want bool
	}{
		{34, 14, true}, {44, 14, true}, {47, 14, true}, {32, 14, true},
		{24, 18, false}, {20, 24, false}, {16, 28, false},
	}
	for _, tt := range tests {
		if got := isLying(tt.w, tt.h); got != tt.want {
			t.Errorf("isLying(%d,%d) = %v", tt.w, tt.h, got)
		}
	}
}

func TestPickSleepSpotAvoidsOthers(t *testing.T) {
	floor := floorRect
	// Otro dormido justo donde cae el paciente.
	taken := []rect{footBox(80, 90, 40, 24)}
	for seed := int64(1); seed <= 20; seed++ {
		x, y := pickSleepSpot(80, 90, 34, 14, floor, taken, nil, rand.New(rand.NewSource(seed)))
		if a := overlapArea(footBox(x, y, 34, 14), taken[0]); a > 0 {
			t.Errorf("semilla %d: se solapa %d px2 en (%v,%v)", seed, a, x, y)
		}
	}
}

func TestPickSleepSpotKeepsFreeSpot(t *testing.T) {
	x, y := pickSleepSpot(80, 90, 34, 14, floorRect, nil, nil, rand.New(rand.NewSource(1)))
	if x != 80 || y != 90 {
		t.Errorf("un lugar libre no se mueve: (%v,%v)", x, y)
	}
}

func TestPickSleepSpotInsideMargins(t *testing.T) {
	// Pegado al borde izquierdo: se corre hacia adentro.
	x, _ := pickSleepSpot(20, 90, 47, 14, floorRect, nil, nil, rand.New(rand.NewSource(1)))
	if box := footBox(x, 90, 47, 14); box.x < sleepMargin {
		t.Errorf("sale del margen: %+v", box)
	}
	x, _ = pickSleepSpot(165, 90, 47, 14, floorRect, nil, nil, rand.New(rand.NewSource(1)))
	if box := footBox(x, 90, 47, 14); box.x+box.w > feedW-sleepMargin {
		t.Errorf("sale del margen derecho: %+v", box)
	}
}

func TestPickSleepSpotAvoidsPalm(t *testing.T) {
	obs := sleepObstacles["6"]
	for seed := int64(1); seed <= 30; seed++ {
		x, y := pickSleepSpot(44, 104, 47, 14, floorRect, nil, obs, rand.New(rand.NewSource(seed)))
		box := footBox(x, y, 47, 14)
		for _, o := range obs {
			if overlapArea(box, o) > 0 {
				t.Errorf("semilla %d: cae sobre %+v en %+v", seed, o, box)
			}
		}
	}
}

func TestPickSleepSpotDeterministic(t *testing.T) {
	taken := []rect{footBox(80, 90, 40, 24)}
	x1, y1 := pickSleepSpot(80, 90, 34, 14, floorRect, taken, nil, rand.New(rand.NewSource(7)))
	x2, y2 := pickSleepSpot(80, 90, 34, 14, floorRect, taken, nil, rand.New(rand.NewSource(7)))
	if x1 != x2 || y1 != y2 {
		t.Errorf("no es determinista: (%v,%v) y (%v,%v)", x1, y1, x2, y2)
	}
}

func TestPickSleepSpotFullFloorDoesNotPanic(t *testing.T) {
	// Piso lleno: devuelve algo dentro del piso en vez de fallar.
	taken := []rect{{0, 0, feedW, feedH}}
	x, y := pickSleepSpot(80, 90, 34, 14, floorRect, taken, nil, rand.New(rand.NewSource(1)))
	if int(y) < floorRect.y || int(y) > floorRect.y+floorRect.h || x < 0 {
		t.Errorf("fuera del piso: (%v,%v)", x, y)
	}
}

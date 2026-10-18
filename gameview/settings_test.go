package gameview

import (
	"testing"

	"costenos-narcolepsia/game"
)

func TestStepVolume(t *testing.T) {
	tests := []struct{ v, dir, want int }{
		{70, 1, 80}, {70, -1, 60}, {100, 1, 100}, {0, -1, 0}, {95, 1, 100}, {5, -1, 0},
	}
	for _, tt := range tests {
		if got := stepVolume(tt.v, tt.dir); got != tt.want {
			t.Errorf("stepVolume(%d,%d) = %d, want %d", tt.v, tt.dir, got, tt.want)
		}
	}
}

func TestVolumeAt(t *testing.T) {
	track := rect{262, 66, 60, 4}
	tests := []struct{ x, want int }{
		{262, 0}, {292, 50}, {322, 100}, {304, 70}, {200, 0}, {400, 100}, {264, 0}, {268, 10},
	}
	for _, tt := range tests {
		if got := volumeAt(tt.x, track); got != tt.want {
			t.Errorf("volumeAt(%d) = %d, want %d", tt.x, got, tt.want)
		}
	}
}

func TestStepRowsWrap(t *testing.T) {
	if stepRow(0, -1) != setBack || stepRow(setBack, 1) != 0 || stepRow(2, 1) != 3 {
		t.Error("stepRow must wrap around")
	}
	if stepPause(0, -1) != pauseExit || stepPause(pauseExit, 1) != pauseContinue || stepPause(0, 1) != pauseSettings {
		t.Error("stepPause must wrap around")
	}
}

func TestPauseItemAt(t *testing.T) {
	for i := 0; i < pauseItems; i++ {
		r := pauseItemRect(i)
		if got := pauseItemAt(r.x+r.w/2, r.y+r.h/2); got != i {
			t.Errorf("pauseItemAt over item %d = %d", i, got)
		}
	}
	if pauseItemAt(0, 0) != -1 {
		t.Error("a point outside the list must give -1")
	}
}

func TestBlinkAlpha(t *testing.T) {
	if blinkAlpha(-1) != 0 || blinkAlpha(blinkTotalTicks) != 0 || blinkAlpha(blinkTotalTicks+50) != 0 {
		t.Error("no blink before it starts or after 5 s")
	}
	if blinkAlpha(0) != 0 {
		t.Errorf("the blink starts transparent, got %v", blinkAlpha(0))
	}
	peak := blinkAlpha(blinkPeriodTicks / 2)
	if peak != blinkMaxAlpha || peak >= 1 {
		t.Errorf("peak = %v, want the soft maximum %v", peak, blinkMaxAlpha)
	}
	// Es un triángulo: sube hasta la mitad del periodo y baja.
	if !(blinkAlpha(5) < blinkAlpha(10) && blinkAlpha(30) > blinkAlpha(35)) {
		t.Error("the blink must rise and then fall within each period")
	}
	for e := 0; e < blinkTotalTicks; e++ {
		if a := blinkAlpha(e); a < 0 || a > blinkMaxAlpha {
			t.Fatalf("alpha(%d) = %v out of range", e, a)
		}
	}
}

func TestStateEvent(t *testing.T) {
	tests := []struct {
		prev, cur   game.PatientState
		slept, woke bool
	}{
		{game.Wandering, game.AsleepWaiting, true, false},
		{game.Arriving, game.AsleepWaiting, true, false},
		{game.AsleepWaiting, game.AsleepWaiting, false, false}, // sigue dormido: no es un evento nuevo
		{game.InBed, game.AwakeInBed, false, true},
		{game.AwakeInBed, game.AwakeInBed, false, false},
		{game.AsleepWaiting, game.Carried, false, false},
		{game.NoBedWaiting, game.Carried, false, false},
	}
	for _, tt := range tests {
		slept, woke := stateEvent(tt.prev, tt.cur)
		if slept != tt.slept || woke != tt.woke {
			t.Errorf("stateEvent(%v,%v) = %v,%v want %v,%v", tt.prev, tt.cur, slept, woke, tt.slept, tt.woke)
		}
	}
}

func TestOnOff(t *testing.T) {
	if onOff(true) != "Prendido" || onOff(false) != "Apagado" {
		t.Error("onOff")
	}
}

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

// Izquierda y Derecha cambian el volumen de la fila marcada, y el arrastre con
// el mouse lo sigue mientras el botón esté abajo.
func TestVolumeRowsAndDrag(t *testing.T) {
	env := &Env{Save: game.NewSaveData(), Opts: Options{Agent: true}}
	v := newSettingsView(env)
	left, right := &Input{keys: map[string]bool{"Left": true}}, &Input{keys: map[string]bool{"Right": true}}
	v.Update(right)
	v.Update(right)
	if env.Save.Settings.MusicVolume != 90 {
		t.Errorf("music = %d, want 90", env.Save.Settings.MusicVolume)
	}
	v.sel = setSfx
	v.Update(left)
	if env.Save.Settings.SfxVolume != 70 || env.Save.Settings.MusicVolume != 90 {
		t.Errorf("settings = %+v", env.Save.Settings)
	}

	track := settingsTrack(setMusic)
	y := track.y + 1
	v.Update(&Input{MouseX: track.x + 30, MouseY: y, Click: true, Down: true, Moved: true})
	if env.Save.Settings.MusicVolume != 50 || v.drag != setMusic {
		t.Fatalf("click: music = %d, drag = %d", env.Save.Settings.MusicVolume, v.drag)
	}
	// Arrastra fuera de la fila: sigue al mouse.
	v.Update(&Input{MouseX: track.x + 60, MouseY: y + 80, Down: true})
	if env.Save.Settings.MusicVolume != 100 {
		t.Errorf("drag: music = %d, want 100", env.Save.Settings.MusicVolume)
	}
	v.Update(&Input{MouseX: track.x, MouseY: y})
	if v.drag != -1 || env.Save.Settings.MusicVolume != 100 {
		t.Errorf("after release: drag = %d, music = %d", v.drag, env.Save.Settings.MusicVolume)
	}
}

func TestSettingsLayout(t *testing.T) {
	// Las seis filas y "Volver" caben en la caja sin tocarse.
	prev := settingsRowRect(0).y
	for i := 1; i < setBack; i++ {
		r, p := settingsRowRect(i), settingsRowRect(i-1)
		if r.y < p.y+p.h+3 {
			t.Errorf("row %d is too close to row %d", i, i-1)
		}
		prev = r.y + r.h
	}
	if settingsBackRect.y < prev+3 || settingsBackRect.y+settingsBackRect.h >= 200 {
		t.Errorf("Volver at y=%d overlaps rows (end %d) or the description line", settingsBackRect.y, prev)
	}
}

func TestStepRowsWrap(t *testing.T) {
	if stepRow(0, -1) != setBack || stepRow(setBack, 1) != 0 || stepRow(0, 1) != 1 {
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
	if blinkAlpha(-1) != 0 {
		t.Error("no blink before it starts")
	}
	if blinkAlpha(100000+blinkPeriodTicks/2) != blinkMaxAlpha {
		t.Error("the blink never stops: it must still peak after a long time")
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
	for e := 0; e < 10*blinkPeriodTicks; e++ {
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

// El interruptor de sonido: nace prendido y se guarda al revés (SoundOff).
func TestSoundSettingToggle(t *testing.T) {
	env := &Env{Save: game.NewSaveData()}
	v := newSettingsView(env)
	if env.Save.Settings.SoundOff {
		t.Fatal("sound must start on")
	}
	v.toggle(setSound)
	if !env.Save.Settings.SoundOff || onOff(!env.Save.Settings.SoundOff) != "Apagado" {
		t.Errorf("after toggling the row must read Apagado: %+v", env.Save.Settings)
	}
	v.toggle(setSound)
	if env.Save.Settings.SoundOff {
		t.Error("toggling twice must turn the sound back on")
	}
	if settingsToggle(setSound).y+settingsToggle(setSound).h >= settingsBackRect.y {
		t.Error("the Sonido row must not touch the Volver button")
	}
}

// Tabla de blinkReasons y updateSeen: simulan la partida tick a tick.
func TestBlinkReasons(t *testing.T) {
	const cam2, cam1b, room = game.CameraID("2"), game.CameraID("1B"), game.CameraID("7A")
	sleeper := blinkPatient{"P-001", cam2, game.AsleepWaiting}
	waker := blinkPatient{"P-002", room, game.AwakeInBed}
	tests := []struct {
		name            string
		ps              []blinkPatient
		current         game.CameraID
		seen            map[string]bool
		sleepOn, wakeOn bool
		want            map[game.CameraID]bool
	}{
		{"sleeper elsewhere blinks mango", []blinkPatient{sleeper}, cam1b, nil, true, true, map[game.CameraID]bool{cam2: false}},
		{"no-bed patient counts as sleep", []blinkPatient{{"P-003", cam2, game.NoBedWaiting}}, cam1b, nil, true, false, map[game.CameraID]bool{cam2: false}},
		{"waker blinks mint", []blinkPatient{waker}, cam1b, nil, true, true, map[game.CameraID]bool{room: true}},
		{"viewed reason does not blink", []blinkPatient{sleeper}, cam1b, map[string]bool{"s:P-001": true}, true, true, map[game.CameraID]bool{}},
		{"new sleeper blinks next to a seen one", []blinkPatient{sleeper, {"P-004", cam2, game.AsleepWaiting}}, cam1b, map[string]bool{"s:P-001": true}, true, true, map[game.CameraID]bool{cam2: false}},
		{"current camera never blinks", []blinkPatient{sleeper}, cam2, nil, true, true, map[game.CameraID]bool{}},
		{"sleep setting off", []blinkPatient{sleeper}, cam1b, nil, false, true, map[game.CameraID]bool{}},
		{"wake setting off", []blinkPatient{waker}, cam1b, nil, true, false, map[game.CameraID]bool{}},
		{"reason gone (attended)", []blinkPatient{{"P-001", cam2, game.Carried}}, cam1b, nil, true, true, map[game.CameraID]bool{}},
		{"both reasons: mint wins", []blinkPatient{{"P-001", room, game.AsleepWaiting}, waker}, cam1b, nil, true, true, map[game.CameraID]bool{room: true}},
		{"both reasons, wake off: mango", []blinkPatient{{"P-001", room, game.AsleepWaiting}, waker}, cam1b, nil, true, false, map[game.CameraID]bool{room: false}},
	}
	for _, tt := range tests {
		got := blinkReasons(tt.ps, tt.current, tt.seen, tt.sleepOn, tt.wakeOn)
		if len(got) != len(tt.want) {
			t.Errorf("%s: got %v want %v", tt.name, got, tt.want)
			continue
		}
		for cam, wake := range tt.want {
			if v, ok := got[cam]; !ok || v != wake {
				t.Errorf("%s: got %v want %v", tt.name, got, tt.want)
			}
		}
	}
}

func TestBlinkLifecycle(t *testing.T) {
	const cam2, cam1b = game.CameraID("2"), game.CameraID("1B")
	seen := map[string]bool{}
	p1 := blinkPatient{"P-001", cam2, game.AsleepWaiting}
	step := func(ps []blinkPatient, cur game.CameraID) map[game.CameraID]bool {
		updateSeen(ps, cur, seen)
		return blinkReasons(ps, cur, seen, true, true)
	}
	if len(step([]blinkPatient{p1}, cam1b)) != 1 {
		t.Fatal("a sleeper on another camera must blink")
	}
	if len(step([]blinkPatient{p1}, cam1b)) != 1 {
		t.Fatal("it keeps blinking while nobody looks")
	}
	step([]blinkPatient{p1}, cam2) // el jugador mira la cámara
	if len(step([]blinkPatient{p1}, cam1b)) != 0 {
		t.Fatal("after looking and leaving, the blink must stay off")
	}
	p2 := blinkPatient{"P-002", cam2, game.AsleepWaiting}
	if len(step([]blinkPatient{p1, p2}, cam1b)) != 1 {
		t.Fatal("a second sleeper is a new event: blink again")
	}
	step([]blinkPatient{{"P-001", cam2, game.Carried}, {"P-002", cam2, game.Carried}}, cam1b)
	if len(seen) != 0 {
		t.Fatalf("gone reasons must be forgotten, seen = %v", seen)
	}
	if len(step([]blinkPatient{p1}, cam1b)) != 1 {
		t.Fatal("the same patient falling asleep again must blink again")
	}
}

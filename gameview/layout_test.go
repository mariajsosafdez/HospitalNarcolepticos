package gameview

import (
	"image"
	"testing"

	"costenos-narcolepsia/game"
)

func TestLightFrame(t *testing.T) {
	tests := []struct {
		tick, want int
	}{
		{0, 0},
		{ticksPerLightPhase - 1, 0}, // 13:59
		{ticksPerLightPhase, 1},     // 14:00
		{2*ticksPerLightPhase - 1, 1},
		{2 * ticksPerLightPhase, 2}, // 16:00
		{game.ShiftTicks, 2},        // 18:00 no pasa del último cuadro
		{-5, 0},
	}
	for _, tt := range tests {
		if got := lightFrame(tt.tick); got != tt.want {
			t.Errorf("lightFrame(%d) = %d, want %d", tt.tick, got, tt.want)
		}
	}
}

func TestBackgroundFrame(t *testing.T) {
	tests := []struct{ width, light, want int }{
		{504, 0, 0}, {504, 2, 2}, // 3 cuadros
		{168, 2, 0}, // un solo cuadro: siempre el 0
		{336, 2, 1}, // dos cuadros: el último
	}
	for _, tt := range tests {
		if got := backgroundFrame(tt.width, tt.light); got != tt.want {
			t.Errorf("backgroundFrame(%d,%d) = %d, want %d", tt.width, tt.light, got, tt.want)
		}
	}
}

func TestBackgroundName(t *testing.T) {
	names := []string{"fondo_cam2_cafeteria", "fondo_habitacion", "fondo_cam3_pasillo_este", "menu_fondo"}
	tests := []struct {
		cam  game.CameraID
		want string
	}{
		{"2", "fondo_cam2_cafeteria"},
		{"3", "fondo_cam3_pasillo_este"},
		{"7A", "fondo_habitacion"},
		{"7F", "fondo_habitacion"},
		{"1A", ""}, // todavía no está dibujado
		{"6", ""},
	}
	// Con fondo propio por habitación se usa ese; si no, el común.
	own := append([]string{"fondo_habitacion_102"}, names...)
	if got := backgroundName("7B", own); got != "fondo_habitacion_102" {
		t.Errorf("backgroundName(7B) = %q, want its own background", got)
	}
	if got := backgroundName("7A", own); got != "fondo_habitacion" {
		t.Errorf("backgroundName(7A) = %q, want the shared background", got)
	}
	for _, tt := range tests {
		if got := backgroundName(tt.cam, names); got != tt.want {
			t.Errorf("backgroundName(%q) = %q, want %q", tt.cam, got, tt.want)
		}
	}
}

func TestPersonKey(t *testing.T) {
	tests := []struct{ name, want string }{
		{"Dra. Karen Ospina", "karen"},
		{"Dr. Álvaro Mendoza", "alvaro"},
		{"Dr. Alvaro Mendoza", "alvaro"},
		{"Chepe Barrios", "chepe"},
		{"Yesid Pacheco", "yesid"},
		{"Abelardo \"Mañe\" Julio", "abelardo"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := personKey(tt.name); got != tt.want {
			t.Errorf("personKey(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestPatientSprite(t *testing.T) {
	tests := []struct {
		state   game.PatientState
		walking bool
		suffix  string
		ok      bool
	}{
		{game.Wandering, true, "camina", true},
		{game.Wandering, false, "quieto", true},
		{game.Arriving, true, "camina", true},
		{game.AsleepWaiting, false, "dormido", true},
		{game.NoBedWaiting, false, "dormido", true},
		{game.InBed, false, "cama", true},
		{game.AwakeInBed, false, "sentado", true},
		{game.Carried, false, "", false},
		{game.Gone, false, "", false},
	}
	for _, tt := range tests {
		spec, ok := patientSprite(tt.state, tt.walking)
		if ok != tt.ok || spec.suffix != tt.suffix {
			t.Errorf("patientSprite(%v,%v) = %q,%v want %q,%v", tt.state, tt.walking, spec.suffix, ok, tt.suffix, tt.ok)
		}
	}
}

func TestAnimFrame(t *testing.T) {
	tests := []struct{ tick, ticks, frames, offset, want int }{
		{0, 9, 4, 0, 0}, {8, 9, 4, 0, 0}, {9, 9, 4, 0, 1}, {36, 9, 4, 0, 0},
		{0, 24, 2, 1, 1}, {24, 24, 2, 1, 0}, {5, 0, 2, 0, 0},
	}
	for _, tt := range tests {
		if got := animFrame(tt.tick, tt.ticks, tt.frames, tt.offset); got != tt.want {
			t.Errorf("animFrame(%d,%d,%d,%d) = %d, want %d", tt.tick, tt.ticks, tt.frames, tt.offset, got, tt.want)
		}
	}
}

func TestFrameRect(t *testing.T) {
	tests := []struct {
		w, h, count, i int
		want           image.Rectangle
	}{
		{64, 24, 4, 0, image.Rect(0, 0, 16, 24)},
		{64, 24, 4, 3, image.Rect(48, 0, 64, 24)},
		{64, 24, 4, 4, image.Rect(0, 0, 16, 24)}, // da la vuelta
		{504, 123, 3, 2, image.Rect(336, 0, 504, 123)},
		{18, 9, 2, -1, image.Rect(9, 0, 18, 9)},
	}
	for _, tt := range tests {
		if got := frameRect(tt.w, tt.h, tt.count, tt.i); got != tt.want {
			t.Errorf("frameRect(%d,%d,%d,%d) = %v, want %v", tt.w, tt.h, tt.count, tt.i, got, tt.want)
		}
	}
}

func TestMapButtonAt(t *testing.T) {
	tests := []struct {
		x, y int
		want game.CameraID
		ok   bool
	}{
		{344 + 108, 156 + 16, "2", true}, // esquina de arriba del botón 2
		{344 + 123, 156 + 26, "2", true}, // esquina de abajo
		{344 + 124, 156 + 16, "", false}, // justo afuera
		{344 + 8, 156 + 16, "6", true},
		{344 + 58, 156 + 60, "1A", true},
		{344 + 112, 156 + 60, "7F", true},
		{10, 10, "", false},
	}
	for _, tt := range tests {
		got, ok := mapButtonAt(tt.x, tt.y)
		if got != tt.want || ok != tt.ok {
			t.Errorf("mapButtonAt(%d,%d) = %q,%v want %q,%v", tt.x, tt.y, got, ok, tt.want, tt.ok)
		}
	}
	// Los 14 botones existen y no se pisan.
	for i, a := range mapCameras {
		for _, b := range mapCameras[i+1:] {
			ra, rb := mapButtonRect(a), mapButtonRect(b)
			if ra.x < rb.x+rb.w && rb.x < ra.x+ra.w && ra.y < rb.y+rb.h && rb.y < ra.y+ra.h {
				t.Errorf("buttons %s and %s overlap", a, b)
			}
		}
	}
}

func TestCameraForKey(t *testing.T) {
	unlocked := []game.CameraID{"1A", "1B", "2", "3", "7A", "7B"}
	tests := []struct {
		digit   byte
		current game.CameraID
		want    game.CameraID
		ok      bool
	}{
		{'1', "2", "1A", true},
		{'1', "1A", "1B", true},
		{'1', "1B", "1A", true}, // da la vuelta
		{'2', "2", "2", true},
		{'7', "3", "7A", true},
		{'7', "7A", "7B", true},
		{'4', "2", "", false}, // la 4 está bloqueada
	}
	for _, tt := range tests {
		got, ok := cameraForKey(tt.digit, tt.current, unlocked)
		if got != tt.want || ok != tt.ok {
			t.Errorf("cameraForKey(%c,%q) = %q,%v want %q,%v", tt.digit, tt.current, got, ok, tt.want, tt.ok)
		}
	}
}

func TestStaffCardRect(t *testing.T) {
	tests := []struct {
		i, total int
		want     rect
	}{
		{0, 2, rect{344, 34, 132, 36}},
		{2, 3, rect{344, 114, 132, 36}}, // 3 personas: tarjeta normal
		{0, 4, rect{344, 34, 132, 24}},  // 4 o 5: tarjeta compacta
		{4, 5, rect{344, 130, 132, 24}},
	}
	for _, tt := range tests {
		if got := staffCardRect(tt.i, tt.total); got != tt.want {
			t.Errorf("staffCardRect(%d,%d) = %v, want %v", tt.i, tt.total, got, tt.want)
		}
	}
	// Las 5 tarjetas compactas terminan antes de que empiece el mapa (y = 156).
	if last := staffCardRect(4, 5); last.y+last.h > mapOrigin.y {
		t.Errorf("the 5th compact card ends at %d, over the map (%d)", last.y+last.h, mapOrigin.y)
	}
}

func TestFrameByDurations(t *testing.T) {
	d := sittingDurations // 2,0 s y un parpadeo de 0,2 s
	tests := []struct{ tick, want int }{
		{0, 0}, {119, 0}, {120, 1}, {131, 1}, {132, 0}, {132 + 120, 1}, {-1, 1},
	}
	for _, tt := range tests {
		if got := frameByDurations(tt.tick, d); got != tt.want {
			t.Errorf("frameByDurations(%d) = %d, want %d", tt.tick, got, tt.want)
		}
	}
	if frameByDurations(5, nil) != 0 {
		t.Error("without durations the frame is 0")
	}
	spec, _ := patientSprite(game.AwakeInBed, false)
	if spriteFrame(spec, 125) != 1 || spriteFrame(spec, 10) != 0 {
		t.Error("the sitting sprite must blink on the second frame")
	}
}

func TestLabelPosition(t *testing.T) {
	area := feedRect
	// Sprite en el centro: la etiqueta va a la izquierda.
	got := labelPosition(rect{170, 140, 40, 48}, area)
	if got.x+got.w > 170 || got.w != labelW || got.h != labelH {
		t.Errorf("centered sprite: label %+v should be left of the sprite", got)
	}
	// Sprite pegado a la izquierda: va a la derecha.
	got = labelPosition(rect{10, 140, 40, 48}, area)
	if got.x < 50 {
		t.Errorf("left-edge sprite: label %+v should be right of the sprite", got)
	}
	// Siempre queda dentro del área.
	for _, s := range []rect{{10, 30, 40, 48}, {300, 200, 40, 48}, {150, 25, 40, 48}} {
		l := labelPosition(s, area)
		if l.x < area.x || l.y < area.y || l.x+l.w > area.x+area.w || l.y+l.h > area.y+area.h {
			t.Errorf("label %+v for sprite %+v leaves the area", l, s)
		}
	}
}

func TestScreenText(t *testing.T) {
	tests := []struct{ in, want string }{
		{"¡No hay cama libre!", "No hay cama libre!"},
		{"¿Quién lo atiende?", "¿Quién lo atiende?"},
		{"Flechas: elegir · Enter: dale", "Flechas: elegir · Enter: dale"},
		{"¡Así quedó el turno!", "Así quedó el turno!"},
	}
	for _, tt := range tests {
		if got := screenText(tt.in); got != tt.want {
			t.Errorf("screenText(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFeedToScreen(t *testing.T) {
	x, y := feedToScreen(60, 74) // pac_marelvis_camina de la nota de Figma
	if x != 124 || y != 168 {
		t.Errorf("feedToScreen(60,74) = %d,%d, want 124,168", x, y)
	}
}

func TestMenuStep(t *testing.T) {
	all := func(int) bool { return true }
	noContinue := func(i int) bool { return i != optContinue }
	tests := []struct {
		current, dir int
		enabled      func(int) bool
		want         int
	}{
		{0, +1, all, 1},
		{0, -1, all, 5}, // da la vuelta
		{5, +1, all, 0},
		{0, +1, noContinue, 2}, // se salta Continuar
		{2, -1, noContinue, 0},
	}
	for _, tt := range tests {
		if got := menuStep(tt.current, tt.dir, tt.enabled); got != tt.want {
			t.Errorf("menuStep(%d,%d) = %d, want %d", tt.current, tt.dir, got, tt.want)
		}
	}
}

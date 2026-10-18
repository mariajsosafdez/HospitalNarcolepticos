package gameview

import (
	"strings"
	"testing"

	"costenos-narcolepsia/game"
	"costenos-narcolepsia/hospital"
)

// measure falso: 5 píxeles por letra, así no hace falta abrir una ventana.
func measure(s string) int { return 5 * len([]rune(s)) }

func TestWrapText(t *testing.T) {
	tests := []struct {
		text string
		maxW int
		want []string
	}{
		{"uno dos tres", 100, []string{"uno dos tres"}},
		{"uno dos tres", 40, []string{"uno dos", "tres"}},
		{"uno dos tres", 15, []string{"uno", "dos", "tres"}},
		{"palabralarguisima corta", 20, []string{"palabralarguisima", "corta"}},
		{"", 50, nil},
	}
	for _, tt := range tests {
		got := wrapText(tt.text, tt.maxW, measure)
		if strings.Join(got, "|") != strings.Join(tt.want, "|") {
			t.Errorf("wrapText(%q,%d) = %q, want %q", tt.text, tt.maxW, got, tt.want)
		}
	}
}

func TestFitText(t *testing.T) {
	if got := fitText("hola", 100, measure); got != "hola" {
		t.Errorf("short text changed: %q", got)
	}
	got := fitText("Dra. Luz Dary Fontalvo", 60, measure)
	if measure(got) > 60 || !strings.HasSuffix(got, "...") {
		t.Errorf("fitText = %q (%d px), want <= 60 px ending in ...", got, measure(got))
	}
	if got := fitText("abc", 3, measure); got != "..." {
		t.Errorf("tiny width = %q", got)
	}
}

func TestFriendlyError(t *testing.T) {
	err := hospital.ErrNoRoomAvailable
	title, lines := friendlyError(err, "Yeimy Padilla")
	if title != "¡No hay cama libre!" || lines[0] != "Yeimy se queda dormida aquí mismo." || lines[1] != "Cuando se libere una cama, mándela usted." {
		t.Errorf("no room, female: %q %q", title, lines)
	}
	_, lines = friendlyError(err, "Wilfrido Berrio")
	if lines[0] != "Wilfrido se queda dormido aquí mismo." || lines[1] != "Cuando se libere una cama, mándelo usted." {
		t.Errorf("no room, male: %q", lines)
	}
	if title, _ := friendlyError(game.ErrStaffBusy, "X"); title != "Un momentico" {
		t.Errorf("busy title = %q", title)
	}
}

func TestDemoStepTexts(t *testing.T) {
	// El texto aprobado del paso 4 no debe cambiar sin querer.
	if demoStepTexts[3] != "El reporte con las 4 consultas. ¡Así quedó el turno!" {
		t.Errorf("step 4 text changed: %q", demoStepTexts[3])
	}
	total := 0
	for _, n := range demoStepActions {
		total += n
	}
	// Los 4 pasos de la vista tienen que gastar exactamente el guion de game.Demo.
	d, err := game.NewDemo()
	if err != nil {
		t.Fatal(err)
	}
	if total != d.StepCount() {
		t.Errorf("view steps run %d demo actions, game.Demo has %d", total, d.StepCount())
	}
}

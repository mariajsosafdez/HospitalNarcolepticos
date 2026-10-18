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

func TestDemoPartTexts(t *testing.T) {
	// El texto aprobado de la última parte no debe cambiar sin querer.
	if got := demoParts[7].text; got != "El reporte con las 4 consultas. ¡Así quedó el turno!" {
		t.Errorf("part 8 text changed: %q", got)
	}
	// Los "Paso N de 4" van de 1 a 4 y suben de a poco: 1, 2 x4, 3 x2, 4.
	want := []int{1, 2, 2, 2, 2, 3, 3, 4}
	for i, p := range demoParts {
		if p.step != want[i] {
			t.Errorf("part %d is step %d, want %d", i+1, p.step, want[i])
		}
		if p.text == "" || !strings.HasPrefix(p.title, "Paso ") {
			t.Errorf("part %d has no title or text: %+v", i+1, p)
		}
	}
	// Cada parte de la vista tiene un beat en el guion de game.Demo: se abre la Demo y se juegan todas.
	d, err := game.NewDemo()
	if err != nil {
		t.Fatal(err)
	}
	if len(demoParts) != game.DemoParts || d.Part() != 1 {
		t.Errorf("view has %d parts, game.Demo has %d", len(demoParts), game.DemoParts)
	}
}

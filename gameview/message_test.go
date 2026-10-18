package gameview

import (
	"strings"
	"testing"

	"costenos-narcolepsia/game"
)

func TestMessageBoxAdvance(t *testing.T) {
	m := newMessage(msgTutorial, game.TutorialLines[:], 3)
	if m.index != 3 || m.counter() != "4 / 9" {
		t.Errorf("start at line 4: index %d counter %q", m.index, m.counter())
	}
	for i := 0; i < 5; i++ {
		if m.advance() {
			t.Fatalf("advance %d finished too early", i)
		}
	}
	if m.counter() != "9 / 9" || !m.advance() {
		t.Error("the last advance must end the message")
	}
	one := newMessage(msgDayIntro, []string{"hola"}, 0)
	if one.counter() != "" || !one.advance() {
		t.Error("a single message has no counter and ends at once")
	}
	if got := newMessage(msgTutorial, game.TutorialLines[:], 99).index; got != 8 {
		t.Errorf("a start past the end is clamped to the last line, got %d", got)
	}
}

// Todas las líneas del Director caben en las 3 líneas de la caja con una
// medida aproximada de 4 px por letra.
func TestWrapMessageFitsThreeLines(t *testing.T) {
	measure := func(s string) int { return len([]rune(s)) * 4 }
	all := append([]string{}, game.TutorialLines[:]...)
	for d := 2; d <= game.MaxDay; d++ {
		text, _ := game.DayIntro(d)
		all = append(all, text)
	}
	all = append(all, game.OverworkLine("Dra. Luz Dary Fontalvo"), game.EndLine(1, 0))
	for _, text := range all {
		lines := wrapMessage(text, measure)
		if len(lines) > 3 {
			t.Errorf("%d lines for %q", len(lines), text)
		}
		if strings.Join(lines, " ") != text {
			t.Errorf("wrapping lost words: %q", lines)
		}
	}
	// Con la medida angosta la primera prueba (200 px) es la que se usa.
	for _, l := range wrapMessage(game.TutorialLines[3], measure) {
		if measure(l) > 200 {
			t.Errorf("line %q is wider than 200 px", l)
		}
	}
}

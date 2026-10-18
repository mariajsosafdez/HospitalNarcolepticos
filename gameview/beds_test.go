package gameview

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
)

var (
	bedComment   = regexp.MustCompile(`(?m)^//\s*cama:\s*(\d+)\s*,\s*(\d+)`)
	plateComment = regexp.MustCompile(`(?m)^//\s*placa:\s*(\d+)\s*,\s*(\d+)`)
)

// parsePoint lee "x,y" de la primera coincidencia de re (ok = false si no hay).
func parsePoint(re *regexp.Regexp, raw []byte) (point, bool) {
	m := re.FindSubmatch(raw)
	if m == nil {
		return point{}, false
	}
	x, _ := strconv.Atoi(string(m[1]))
	y, _ := strconv.Atoi(string(m[2]))
	return point{x, y}, true
}

// La tabla roomArts tiene que decir lo mismo que los comentarios "// cama: x,y"
// y "// placa: x,y" de cada art/sprites/fondo_habitacion_<n>.txt. Si el arte
// mueve una cama y la tabla no se actualiza (o al revés), este test falla.
func TestRoomArtsMatchSprites(t *testing.T) {
	for n := 101; n <= 106; n++ {
		path := filepath.Join("..", "art", "sprites", fmt.Sprintf("fondo_habitacion_%d.txt", n))
		raw, err := os.ReadFile(path)
		if err != nil {
			if _, ok := roomArts[n]; ok {
				t.Errorf("room %d is in roomArts but %s does not exist", n, path)
			}
			continue // el arte todavía no hizo este cuarto: se usa el por defecto
		}
		bed, ok1 := parsePoint(bedComment, raw)
		plate, ok2 := parsePoint(plateComment, raw)
		if !ok1 || !ok2 {
			t.Errorf("%s needs \"// cama: x,y\" and \"// placa: x,y\" lines", path)
			continue
		}
		if got := roomLayout(n); got != (roomArt{bed, plate}) {
			t.Errorf("room %d: roomArts says %v, the sprite comments say bed %v plate %v", n, got, bed, plate)
		}
	}
}

func TestRoomLayoutDefault(t *testing.T) {
	if got := roomLayout(999); got != defaultRoom || defaultRoom.bed != (point{34, 72}) {
		t.Errorf("roomLayout(999) = %v", got)
	}
}

// El botón "Dar de alta" va debajo de la cama (sin tocarla) y dentro de la
// cámara, en las seis habitaciones.
func TestDischargeRectBelowBed(t *testing.T) {
	for n, art := range roomArts {
		r := dischargeRect(art.bed)
		_, bedBottom := feedToScreen(0, art.bed.y+bedH)
		if r.y < bedBottom+8 {
			t.Errorf("habitación %d: el botón (y=%d) está muy pegado a la cama (termina en %d)", n, r.y, bedBottom)
		}
		if r.x < feedRect.x || r.x+r.w > feedRect.x+feedRect.w || r.y+r.h > feedRect.y+feedRect.h {
			t.Errorf("habitación %d: el botón %+v se sale de la cámara", n, r)
		}
	}
}

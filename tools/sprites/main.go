// Command sprites convierte los sprites de texto del juego en imágenes PNG.
//
// Uso (desde la raíz del repo):
//
//	go run ./tools/sprites
//
// Lee la paleta art/palette.hex y cada archivo art/sprites/*.txt, y escribe
// art/png/<nombre>.png (tamaño real, 1×) y art/png/<nombre>@4x.png (cada píxel
// copiado en un bloque de 4×4, sin suavizar: "vecino más cercano").
//
// Formato de la paleta (art/palette.hex), una entrada por línea:
//
//	m #f2a33a mango
//
// La clave es un solo carácter, luego el color y luego un nombre. Las líneas
// vacías y las que empiezan con "#" se ignoran. "." no se puede usar como clave.
//
// Formato de un sprite (art/sprites/<nombre>.txt):
//
//	..00..
//	.0mm0.
//	---
//	..00..
//	.0MM0.
//
// Cada carácter es un píxel: una clave de la paleta, o "." = transparente.
// Todas las filas de un cuadro deben tener el mismo ancho. Una línea "---"
// separa los cuadros de una animación; en el PNG quedan uno al lado del otro,
// de izquierda a derecha (una "tira" de cuadros). Las líneas vacías y las que
// empiezan con "//" se ignoran (sirven de comentarios).
package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// scale es cuántas veces se agranda la copia @4x.
const scale = 4

func main() {
	palette, err := loadPalette("art/palette.hex")
	if err != nil {
		fail(err)
	}
	files, err := filepath.Glob("art/sprites/*.txt")
	if err != nil {
		fail(err)
	}
	if err := os.MkdirAll("art/png", 0o755); err != nil {
		fail(err)
	}
	for _, file := range files {
		frames, err := loadFrames(file)
		if err != nil {
			fail(err)
		}
		img, err := paint(frames, palette)
		if err != nil {
			fail(fmt.Errorf("%s: %w", file, err))
		}
		name := strings.TrimSuffix(filepath.Base(file), ".txt")
		if err := savePNG("art/png/"+name+".png", img); err != nil {
			fail(err)
		}
		if err := savePNG("art/png/"+name+"@4x.png", enlarge(img, scale)); err != nil {
			fail(err)
		}
		fmt.Printf("%s -> art/png/%s.png (%dx%d, %d cuadros)\n",
			file, name, img.Bounds().Dx(), img.Bounds().Dy(), len(frames))
	}
}

// fail imprime el error y termina el programa con código 1.
func fail(err error) {
	fmt.Fprintln(os.Stderr, "sprites:", err)
	os.Exit(1)
}

// loadPalette lee art/palette.hex y devuelve un mapa clave -> color.
func loadPalette(path string) (map[rune]color.RGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	palette := map[rune]color.RGBA{}
	scanner := bufio.NewScanner(f)
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		fields := strings.Fields(text)
		key := []rune(fields[0])
		if len(fields) < 2 || len(key) != 1 || key[0] == '.' {
			return nil, fmt.Errorf("%s:%d: se esperaba \"clave #rrggbb nombre\"", path, line)
		}
		var c color.RGBA
		if _, err := fmt.Sscanf(fields[1], "#%02x%02x%02x", &c.R, &c.G, &c.B); err != nil {
			return nil, fmt.Errorf("%s:%d: color inválido %q", path, line, fields[1])
		}
		c.A = 255
		palette[key[0]] = c
	}
	return palette, scanner.Err()
}

// loadFrames lee un sprite y devuelve sus cuadros; cada cuadro es una lista de filas.
func loadFrames(path string) ([][]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var frames [][]string
	var current []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, " \r\t")
		switch {
		case line == "" || strings.HasPrefix(line, "//"):
			// comentario o línea vacía: no aporta píxeles
		case line == "---":
			frames = append(frames, current)
			current = nil
		default:
			current = append(current, line)
		}
	}
	frames = append(frames, current)
	for i, frame := range frames {
		if len(frame) == 0 {
			return nil, fmt.Errorf("%s: el cuadro %d está vacío", path, i+1)
		}
	}
	return frames, nil
}

// paint dibuja los cuadros uno al lado del otro en una sola imagen.
func paint(frames [][]string, palette map[rune]color.RGBA) (*image.RGBA, error) {
	// Primero se mide: ancho total = suma de anchos; alto = el del cuadro más alto.
	width, height := 0, 0
	for i, frame := range frames {
		w := len([]rune(frame[0]))
		for r, row := range frame {
			if len([]rune(row)) != w {
				return nil, fmt.Errorf("cuadro %d, fila %d: mide %d y debería medir %d", i+1, r+1, len([]rune(row)), w)
			}
		}
		width += w
		height = max(height, len(frame))
	}

	// Después se pinta. Los píxeles "." se dejan sin tocar (transparentes).
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	offsetX := 0
	for i, frame := range frames {
		for y, row := range frame {
			for x, key := range []rune(row) {
				if key == '.' {
					continue
				}
				c, ok := palette[key]
				if !ok {
					return nil, fmt.Errorf("cuadro %d, fila %d: la clave %q no está en la paleta", i+1, y+1, key)
				}
				img.SetRGBA(offsetX+x, y, c)
			}
		}
		offsetX += len([]rune(frame[0]))
	}
	return img, nil
}

// enlarge agranda la imagen n veces copiando cada píxel en un bloque n×n.
func enlarge(src *image.RGBA, n int) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx()*n, b.Dy()*n))
	for y := 0; y < b.Dy()*n; y++ {
		for x := 0; x < b.Dx()*n; x++ {
			dst.SetRGBA(x, y, src.RGBAAt(x/n, y/n))
		}
	}
	return dst
}

// savePNG guarda la imagen en un archivo PNG.
func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

package gameview

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png" // registra el decodificador de PNG
	"io/fs"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"costenos-narcolepsia/art"
)

// Assets guarda los PNG y la paleta, cargados una sola vez al abrir el juego.
// Todo viene de art/png (embebido), nunca se dibuja arte en el código.
type Assets struct {
	names   []string                 // nombres de los PNG, sin ".png"
	raw     map[string]image.Image   // PNG decodificados (para sacar siluetas)
	images  map[string]*ebiten.Image // los mismos PNG listos para dibujar
	frames  map[frameKey]*ebiten.Image
	silhs   map[silhouetteKey]*ebiten.Image
	masks   map[frameKey]*ebiten.Image // dígitos en blanco (ver digits.go)
	palette map[byte]color.RGBA
	pixel   *ebiten.Image // 1x1 blanco: sirve para rellenar rectángulos
}

type frameKey struct {
	name  string
	count int
	index int
}

type silhouetteKey struct {
	frame frameKey
	color byte
}

// LoadAssets lee todos los PNG a 1x de art/png y la paleta.
func LoadAssets() (*Assets, error) {
	a := &Assets{
		raw:     make(map[string]image.Image),
		images:  make(map[string]*ebiten.Image),
		frames:  make(map[frameKey]*ebiten.Image),
		silhs:   make(map[silhouetteKey]*ebiten.Image),
		masks:   make(map[frameKey]*ebiten.Image),
		palette: make(map[byte]color.RGBA),
	}
	entries, err := fs.ReadDir(art.Files, "png")
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		file := e.Name()
		// Los @4x son solo para revisar en Figma.
		if strings.Contains(file, "@4x") || !strings.HasSuffix(file, ".png") {
			continue
		}
		data, err := fs.ReadFile(art.Files, "png/"+file)
		if err != nil {
			return nil, err
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("decoding %s: %w", file, err)
		}
		name := strings.TrimSuffix(file, ".png")
		a.names = append(a.names, name)
		a.raw[name] = img
		a.images[name] = ebiten.NewImageFromImage(img)
	}
	if err := a.loadPalette(); err != nil {
		return nil, err
	}
	a.pixel = ebiten.NewImage(1, 1)
	a.pixel.Fill(color.White)
	return a, nil
}

// loadPalette lee art/palette.hex: "clave #rrggbb nombre" por línea.
func (a *Assets) loadPalette() error {
	file, err := art.Files.Open("palette.hex")
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var key string
		var hex string
		if _, err := fmt.Sscanf(line, "%s %s", &key, &hex); err != nil || len(key) != 1 {
			return fmt.Errorf("bad palette line %q", line)
		}
		c, err := parseHexColor(hex)
		if err != nil {
			return err
		}
		a.palette[key[0]] = c
	}
	return scanner.Err()
}

// parseHexColor convierte "#rrggbb" en un color.
func parseHexColor(hex string) (color.RGBA, error) {
	var r, g, b uint8
	if _, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b); err != nil {
		return color.RGBA{}, fmt.Errorf("bad color %q: %w", hex, err)
	}
	return color.RGBA{R: r, G: g, B: b, A: 0xff}, nil
}

// Color devuelve el color de una clave de la paleta ('0', 'M', 'w'...).
func (a *Assets) Color(key byte) color.RGBA { return a.palette[key] }

// Names devuelve los nombres de todos los PNG cargados.
func (a *Assets) Names() []string { return a.names }

// Has dice si existe el PNG. Así una pantalla puede usar un arte que todavía
// no está dibujado y recogerlo solo cuando aparezca.
func (a *Assets) Has(name string) bool { _, ok := a.images[name]; return ok }

// Image devuelve el PNG completo (nil si no existe).
func (a *Assets) Image(name string) *ebiten.Image { return a.images[name] }

// Size devuelve el tamaño del PNG completo (0, 0 si no existe).
func (a *Assets) Size(name string) (int, int) {
	img, ok := a.raw[name]
	if !ok {
		return 0, 0
	}
	b := img.Bounds()
	return b.Dx(), b.Dy()
}

// Frame devuelve el cuadro i de un sprite que trae count cuadros del mismo
// ancho puestos en horizontal. Devuelve nil si el PNG no existe.
func (a *Assets) Frame(name string, count, i int) *ebiten.Image {
	img, ok := a.images[name]
	if !ok {
		return nil
	}
	key := frameKey{name, count, i}
	if f, ok := a.frames[key]; ok {
		return f
	}
	b := img.Bounds()
	f := img.SubImage(frameRect(b.Dx(), b.Dy(), count, i)).(*ebiten.Image)
	a.frames[key] = f
	return f
}

// FrameSize devuelve el tamaño de UN cuadro de un sprite de count cuadros.
func (a *Assets) FrameSize(name string, count int) (int, int) {
	w, h := a.Size(name)
	return w / max(count, 1), h
}

// Silhouette devuelve el cuadro i del sprite con todos los píxeles que no
// son transparentes pintados del color de la paleta key. Con cuatro copias
// desplazadas se arma el contorno del paciente elegido.
func (a *Assets) Silhouette(name string, count, i int, key byte) *ebiten.Image {
	raw, ok := a.raw[name]
	if !ok {
		return nil
	}
	fk := frameKey{name, count, i}
	sk := silhouetteKey{fk, key}
	if s, ok := a.silhs[sk]; ok {
		return s
	}
	b := raw.Bounds()
	r := frameRect(b.Dx(), b.Dy(), count, i)
	out := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	c := a.palette[key]
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			if _, _, _, alpha := raw.At(r.Min.X+x, r.Min.Y+y).RGBA(); alpha > 0 {
				out.SetRGBA(x, y, c)
			}
		}
	}
	s := ebiten.NewImageFromImage(out)
	a.silhs[sk] = s
	return s
}

// frameRect calcula el rectángulo del cuadro i dentro de una imagen de w x h
// con count cuadros iguales en horizontal. Es pura para poder probarla.
func frameRect(w, h, count, i int) image.Rectangle {
	count = max(count, 1)
	i = ((i % count) + count) % count
	fw := w / count
	return image.Rect(i*fw, 0, (i+1)*fw, h)
}

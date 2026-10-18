package gameview

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"

	"costenos-narcolepsia/game"
)

// Puerta de pruebas (-agent). Un tester (persona o agente) puede saltar a una
// pantalla, mandar clics y teclas, sacar capturas y volcar el estado sin
// mirar la ventana. Todo es determinista: la hora la dan los ticks, nunca el
// reloj del computador, y la semilla es la de -seed. Documentada en
// docs/HospitalCosteños/Arquitectura.md.

// startScene elige la pantalla inicial según -scene.
func startScene(env *Env) (Scene, error) {
	opts := env.Opts
	if err := opts.validate(); err != nil {
		return nil, err
	}
	switch opts.Scene {
	case "menu":
		return newMenu(env), nil
	case "day":
		cfg, err := game.ConfigForDay(opts.Day)
		if err != nil {
			return nil, err
		}
		shift, err := game.NewShift(cfg, opts.Seed)
		if err != nil {
			return nil, err
		}
		return newCams(env, shift, nil, game.CameraID(opts.Camera))
	case "cams":
		// El escenario de la Demo (5 pacientes, 3 habitaciones, sin azar)
		// sin la caja de la Demo: sirve para probar a mano con sleep/wake.
		shift, err := game.NewShift(game.DemoConfig(), opts.Seed)
		if err != nil {
			return nil, err
		}
		return newCams(env, shift, nil, game.CameraID(opts.Camera))
	default: // demo
		demo, err := game.NewDemo()
		if err != nil {
			return nil, err
		}
		scene, err := newCams(env, demo.Shift(), demo, game.CameraID(opts.Camera))
		if err != nil {
			return nil, err
		}
		for scene.demoStep < opts.Step {
			if err := scene.nextDemoStep(); err != nil {
				return nil, err
			}
		}
		if opts.Camera != "" { // -cam manda sobre la cámara que pone cada paso
			if err := scene.shift.SwitchCamera(game.CameraID(opts.Camera)); err != nil {
				return nil, err
			}
		}
		return scene, nil
	}
}

// shotRequest es una captura pendiente: se toma en Draw, cuando el cuadro
// ya está pintado.
type shotRequest struct {
	path  string
	scale int
}

// agentDriver lee el guion y le da a la aplicación su entrada de cada tick.
type agentDriver struct {
	steps    []step
	pos      int
	waitLeft int
	extra    int // ticks extra después del guion (-frames)
	quit     bool
	pending  *shotRequest
	err      error // error de una captura (se devuelve en el próximo Update)
}

// newAgentDriver lee el archivo del guion (si se pidió uno).
func newAgentDriver(opts Options) (*agentDriver, error) {
	d := &agentDriver{extra: opts.Frames, quit: opts.Quit}
	if opts.InputFile != "" {
		raw, err := os.ReadFile(opts.InputFile)
		if err != nil {
			return nil, err
		}
		steps, err := parseScript(string(raw))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", opts.InputFile, err)
		}
		d.steps = steps
	}
	return d, nil
}

// prepareWindow hace la ventana lo más discreta posible: con -quit se saca de
// la pantalla, y siempre corre lo más rápido que pueda (el tiempo del juego
// son ticks, no segundos).
func (d *agentDriver) prepareWindow() {
	ebiten.SetTPS(ebiten.SyncWithFPS)
	ebiten.SetVsyncEnabled(false)
	ebiten.SetRunnableOnUnfocused(true)
	if d.quit {
		ebiten.SetWindowPosition(-32000, -32000)
	}
}

// update procesa pasos hasta que uno gasta el tick (wait, click, key...) o
// pide esperar el dibujo (shot).
func (d *agentDriver) update(a *App) error {
	if d.err != nil {
		return d.err
	}
	if d.pending != nil {
		return nil // Draw todavía no tomó la captura: no avanza nada
	}
	a.in.reset()
	for {
		if d.waitLeft > 0 {
			d.waitLeft--
			a.step()
			return nil
		}
		if d.pos >= len(d.steps) {
			if d.extra > 0 {
				d.waitLeft, d.extra = d.extra, 0
				continue
			}
			if d.quit {
				return ebiten.Termination
			}
			a.step() // sin -quit la ventana sigue viva
			return nil
		}
		s := d.steps[d.pos]
		d.pos++
		switch s.kind {
		case stepWait:
			d.waitLeft = s.n
		case stepMove:
			a.in.moveTo(s.x, s.y)
			a.step()
			return nil
		case stepClick:
			a.in.moveTo(s.x, s.y)
			a.in.Click = true
			a.step()
			return nil
		case stepKey:
			a.in.keys[s.text] = true
			a.step()
			return nil
		case stepShot:
			d.pending = &shotRequest{s.text, 1}
			return nil
		case stepShot4:
			d.pending = &shotRequest{s.text, 4}
			return nil
		case stepDump:
			if err := d.dump(a, s.text); err != nil {
				return fmt.Errorf("script line %d: %w", s.line, err)
			}
		case stepSleep, stepWake:
			if err := d.force(a, s); err != nil {
				return fmt.Errorf("script line %d: %w", s.line, err)
			}
		}
	}
}

// force duerme o despierta a un paciente con las ayudas de la Demo.
func (d *agentDriver) force(a *App, s step) error {
	shift := a.scene.Shift()
	if shift == nil {
		return errors.New("sleep/wake need a scene with a shift")
	}
	if s.kind == stepWake {
		return shift.ForceWake(s.text)
	}
	cam := game.CameraID(s.cam)
	info, ok := game.CameraByID(cam)
	if !ok {
		return fmt.Errorf("unknown camera %q", s.cam)
	}
	return shift.ForceSleep(s.text, cam, info.Name)
}

// afterDraw toma la captura pendiente, si hay una.
func (d *agentDriver) afterDraw(a *App) {
	if d.pending == nil {
		return
	}
	req := d.pending
	d.pending = nil
	d.err = saveShot(a.canvas, req)
}

// saveShot guarda el lienzo como PNG, a 1x o ampliado (vecino más cercano).
func saveShot(canvas *ebiten.Image, req *shotRequest) error {
	pix := make([]byte, 4*screenW*screenH)
	canvas.ReadPixels(pix)
	small := &image.RGBA{Pix: pix, Stride: 4 * screenW, Rect: image.Rect(0, 0, screenW, screenH)}
	out := image.Image(small)
	if req.scale > 1 {
		big := image.NewRGBA(image.Rect(0, 0, screenW*req.scale, screenH*req.scale))
		for y := 0; y < screenH*req.scale; y++ {
			for x := 0; x < screenW*req.scale; x++ {
				big.SetRGBA(x, y, small.RGBAAt(x/req.scale, y/req.scale))
			}
		}
		out = big
	}
	return writeFile(req.path, func(f *os.File) error { return png.Encode(f, out) })
}

// dumpFile es el JSON del volcado: la interfaz y el turno.
type dumpFile struct {
	UI   UIState        `json:"ui"`
	Game *game.Snapshot `json:"game,omitempty"`
}

// dump escribe el volcado JSON.
func (d *agentDriver) dump(a *App, path string) error {
	ui := a.scene.State()
	ui.UITick = a.env.Tick
	ui.MouseX, ui.MouseY = a.in.MouseX, a.in.MouseY
	out := dumpFile{UI: ui}
	if shift := a.scene.Shift(); shift != nil {
		snap := shift.Snapshot()
		out.Game = &snap
	}
	raw, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(path, func(f *os.File) error { _, err := f.Write(raw); return err })
}

// writeFile crea la carpeta y el archivo y le pasa el archivo a write.
func writeFile(path string, write func(*os.File) error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := write(f); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

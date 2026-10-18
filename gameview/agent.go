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
		return newShiftScene(env, opts.Day, opts.Seed, opts.Intro, game.CameraID(opts.Camera))
	case "dayintro":
		return newShiftScene(env, opts.Day, opts.Seed, true, game.CameraID(opts.Camera))
	case "tutorial":
		// Cómo jugar, empezando en la línea -step.
		return newPracticeScene(env, opts.Seed, opts.Step-1, game.CameraID(opts.Camera))
	case "report":
		// Juega el día entero con el jugador automático (game.AutoStep) y abre
		// el reporte, igual que cuando termina un turno de verdad: se anota el
		// resultado y se guarda (si hay -savepath).
		scene, err := newShiftScene(env, opts.Day, opts.Seed, false, "")
		if err != nil {
			return nil, err
		}
		for !scene.shift.Finished() {
			scene.shift.Tick()
			game.AutoStep(scene.shift)
		}
		scene.finishShift()
		return scene.toReport, nil
	case "pause":
		// Un Día 1 en la pausa (como si se hubiera pulsado Esc).
		scene, err := newShiftScene(env, opts.Day, opts.Seed, false, game.CameraID(opts.Camera))
		if err != nil {
			return nil, err
		}
		scene.paused, scene.pauseSel = true, pauseContinue
		scene.shift.Pause()
		return scene, nil
	case "settings":
		// El menú con los ajustes abiertos (desde la pausa: -scene pause, Down, Enter).
		menu := newMenu(env)
		menu.settings = newSettingsView(env)
		return menu, nil
	case "credits":
		// Los Créditos, con el menú detrás para volver (como al elegir la opción).
		menu := newMenu(env)
		menu.selected = optCredits
		return newCredits(env, menu), nil
	case "director":
		// El regaño por exceso de trabajo, con la Dra. Karen, sobre un Día 1.
		scene, err := newShiftScene(env, 1, opts.Seed, false, game.CameraID(opts.Camera))
		if err != nil {
			return nil, err
		}
		line := game.OverworkLine(displayName("D-001", "Dra. Karen Ospina"))
		scene.showMessage(msgOverwork, []string{line}, 0)
		return scene, nil
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
		if scene.toReport != nil { // paso 4: el reporte de la Demo
			return scene.toReport, nil
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
	scale int // 1 o 4: tamaño final como múltiplo de 480x270; 0 = tal cual la ventana
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
			a.in.moveToGameCell(s.x, s.y, a.viewport())
			a.step()
			return nil
		case stepClick:
			a.in.moveToGameCell(s.x, s.y, a.viewport())
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
		case stepShotWin:
			d.pending = &shotRequest{s.text, 0}
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

// takeShot toma la captura pendiente, si hay una. Lee los píxeles de la
// imagen que Draw acaba de pintar: el lienzo del juego (480x270 por Env.Scale,
// sin bordes) y, si se pidió otro tamaño, la reescala sin suavizar. Con
// shotwin (scale 0) sale la ventana ENTERA, con sus bordes oscuros si los hay.
// Así una captura nunca muestra algo distinto de lo que dibuja el juego de verdad.
func (d *agentDriver) takeShot(canvas, window *ebiten.Image) {
	if d.pending == nil {
		return
	}
	req := d.pending
	d.pending = nil
	screen := canvas
	if req.scale == 0 {
		screen = window
	}
	b := screen.Bounds()
	pix := make([]byte, 4*b.Dx()*b.Dy())
	screen.ReadPixels(pix)
	out := &image.RGBA{Pix: pix, Stride: 4 * b.Dx(), Rect: image.Rect(0, 0, b.Dx(), b.Dy())}
	if req.scale > 0 {
		out = resizeNearest(out, screenW*req.scale, screenH*req.scale)
	}
	d.err = writeFile(req.path, func(f *os.File) error { return png.Encode(f, out) })
}

// resizeNearest cambia el tamaño de una imagen tomando, para cada píxel nuevo,
// el píxel del centro del bloque que le toca de la original (sin suavizar).
func resizeNearest(src *image.RGBA, w, h int) *image.RGBA {
	sb := src.Bounds()
	if sb.Dx() == w && sb.Dy() == h {
		return src
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		sy := min((2*y+1)*sb.Dy()/(2*h), sb.Dy()-1)
		for x := 0; x < w; x++ {
			sx := min((2*x+1)*sb.Dx()/(2*w), sb.Dx()-1)
			dst.SetRGBA(x, y, src.RGBAAt(sx, sy))
		}
	}
	return dst
}

// dumpFile es el JSON del volcado: la interfaz y el turno.
type dumpFile struct {
	UI   UIState        `json:"ui"`
	Save game.SaveData  `json:"save"` // el guardado en memoria (ajustes, días desbloqueados, estrellas)
	Game *game.Snapshot `json:"game,omitempty"`
}

// dump escribe el volcado JSON.
func (d *agentDriver) dump(a *App, path string) error {
	ui := a.scene.State()
	ui.UITick = a.env.Tick
	ui.MouseX, ui.MouseY = a.in.MouseX, a.in.MouseY
	out := dumpFile{UI: ui, Save: a.env.Save}
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

// Package gameview es la VISTA del juego: abre la ventana, dibuja y lee la
// entrada con Ebitengine. Es el ÚNICO paquete del proyecto que importa
// Ebitengine; las reglas viven en el paquete game, que solo se lee y se
// manda acciones desde aquí.
package gameview

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"costenos-narcolepsia/game"
)

const (
	// Resolución lógica del juego (pixel art) y cuánto se agranda en pantalla.
	logicalWidth  = screenW
	logicalHeight = screenH
	windowScale   = 2 // 960x540: cabe cómodo en pantallas de portátil

	windowTitle = "Hospital de los Costeños con Narcolepsia"
)

// Env es lo que comparten todas las pantallas: el arte, las fuentes, el
// guardado y el reloj de la interfaz.
type Env struct {
	Assets *Assets
	Fonts  *Fonts
	Opts   Options
	Sound  *Sound // la mesa de sonido (en el modo pruebas no suena: solo anota los cues)

	// Tick cuenta los ticks de la INTERFAZ (parpadeos, animaciones). No es el
	// reloj del turno: sigue corriendo aunque el juego esté en pausa.
	Tick int

	// Scale es cuántas veces se agranda el lienzo de 480x270 en pantalla.
	// Puede ser fraccionario (1,5; 2,25...): la pantalla se llena sea cual
	// sea su resolución. Todo lo que se dibuja lo multiplica (y lo redondea a
	// píxeles enteros de pantalla), y el texto usa letras de ese tamaño: así
	// queda nítido en cualquier ventana.
	Scale float64

	Save    game.SaveData
	Store   *game.SaveStore // nil en la puerta de pruebas: nunca escribe el guardado real
	HasSave bool            // hay partida guardada (activa "Continuar")
}

// newSeed da la semilla de un turno nuevo. En el juego de verdad sale del
// reloj para que cada partida sea distinta; con -seed (modo pruebas) es fija.
func (e *Env) newSeed() int64 {
	if e.Opts.Agent && e.Opts.Seed != 0 {
		return e.Opts.Seed
	}
	return time.Now().UnixNano()
}

// persist escribe el guardado (si hay almacén). Si no se puede escribir, el
// juego sigue: es solo el progreso. En la puerta de pruebas no hay almacén
// (o es uno de prueba que se pasa con -savepath), nunca el guardado real.
func (e *Env) persist() {
	if e.Store != nil {
		_ = e.Store.Save(e.Save)
	}
}

// recordShift anota el resultado de un turno terminado y lo autoguarda.
func (e *Env) recordShift(day, stars int) {
	e.Save.RecordShift(day, stars)
	e.HasSave = true
	e.persist()
}

// UIState es lo que el volcado de pruebas dice de la interfaz (además del
// game.Snapshot del turno).
type UIState struct {
	Scene           string       `json:"scene"`
	Selected        string       `json:"selected,omitempty"`        // menú: opción elegida
	Camera          string       `json:"camera,omitempty"`          // cámara que se mira
	SelectedPatient string       `json:"selectedPatient,omitempty"` // paciente elegido para atender
	Boxes           []PatientBox `json:"boxes,omitempty"`           // dónde se puede dar clic a cada paciente (píxeles del juego)
	StaffCursor     string       `json:"staffCursor,omitempty"`     // personal marcado con el teclado
	HoveredButton   string       `json:"hoveredButton,omitempty"`   // botón bajo el mouse
	Notice          string       `json:"notice,omitempty"`          // título del aviso visible
	Overlay         string       `json:"overlay,omitempty"`         // pausa, reporte, mensaje, próximamente
	Blinking        []string     `json:"blinking,omitempty"`        // botones del mapa que están parpadeando
	Message         string       `json:"message,omitempty"`         // mensaje del Director abierto
	MessageIndex    int          `json:"messageIndex,omitempty"`    // línea (desde 1) del mensaje
	MessageCount    int          `json:"messageCount,omitempty"`    // cuántas líneas tiene
	Tutorial        bool         `json:"tutorial,omitempty"`        // está en "Cómo jugar"
	DemoPart        int          `json:"demoPart,omitempty"`        // parte de la Demo (1 a 8)
	DemoStep        int          `json:"demoStep,omitempty"`        // "Paso N de 4" del recuadro
	DemoPlaying     bool         `json:"demoPlaying,omitempty"`     // una parte se está animando (el reloj corre)
	DemoBeat        string       `json:"demoBeat,omitempty"`        // momento de la animación (sleep, arrive, carry...)
	DemoAuto        bool         `json:"demoAuto,omitempty"`
	Sounds          []string     `json:"sounds,omitempty"` // modo pruebas: los efectos que se pidieron (no suena nada)
	Music           string       `json:"music,omitempty"`  // modo pruebas: la pista que sonaría
	UITick          int          `json:"uiTick"`
	MouseX          int          `json:"mouseX"`
	MouseY          int          `json:"mouseY"`
}

// PatientBox es el rectángulo (en píxeles del juego, 480x270) donde el mouse
// toca a un paciente. La puerta de pruebas lo usa para hacer clic en el sitio
// exacto sin adivinar por la captura.
type PatientBox struct {
	ID string `json:"id"`
	X  int    `json:"x"`
	Y  int    `json:"y"`
	W  int    `json:"w"`
	H  int    `json:"h"`
}

// Scene es una pantalla del juego: menú, cámaras...
type Scene interface {
	// Update avanza la pantalla un tick. Devuelve otra pantalla si hay que
	// cambiar (nil = seguir en esta).
	Update(in *Input) Scene
	// Draw pinta la pantalla en el lienzo de 480x270.
	Draw(dst *ebiten.Image)
	// State describe la pantalla para el volcado de pruebas.
	State() UIState
	// Shift devuelve el turno que se está jugando (nil en el menú).
	Shift() *game.Shift
}

// App implementa ebiten.Game: una pantalla activa y la entrada del tick.
type App struct {
	env   *Env
	scene Scene
	in    Input
	agent *agentDriver // nil en el juego de verdad

	// Tamaño de la imagen de la ventana, tamaño del lienzo del juego (480x270
	// por Scale, redondeado) y dónde cae dentro de la ventana (ver fitScreen).
	// Layout los actualiza antes de cada Draw.
	winW, winH   int
	canvW, canvH int
	offX, offY   int
	canvas       *ebiten.Image // lienzo del juego, cuando hay bordes
}

// viewport es la ubicación actual del lienzo en la ventana.
func (a *App) viewport() viewport {
	return viewport{a.env.Scale, a.offX, a.offY}
}

// Update lo llama Ebitengine 60 veces por segundo: cada llamada es un tick
// del juego, igual que el reloj de game.Clock.
func (a *App) Update() error {
	if a.agent != nil {
		return a.agent.update(a)
	}
	a.in.readReal(a.viewport())
	a.step()
	return nil
}

// step avanza la interfaz un tick con la entrada actual.
func (a *App) step() {
	a.env.Tick++
	sound, on := a.env.Sound, !a.env.Save.Settings.SoundOff
	sound.SetEnabled(on)
	// Todo clic o tecla de aceptar/volver suena como botón.
	if a.in.Click || a.in.Pressed("Enter") || a.in.Pressed("Space") || a.in.Pressed("Escape") {
		sound.Play(CueClick)
	}
	if next := a.scene.Update(&a.in); next != nil {
		a.scene = next
	}
	track, paused := soundFor(a.scene)
	sound.Update(track, paused, !a.env.Save.Settings.SoundOff)
}

// soundFor dice qué música toca en la pantalla actual y si el juego está en
// pausa (la música baja): el menú, los créditos y los ajustes tienen la del
// menú; el turno, la Demo y el reporte, la del turno.
func soundFor(s Scene) (track Track, paused bool) {
	switch v := s.(type) {
	case *camsScene:
		return TrackTurno, v.paused
	case *reportScene:
		return TrackTurno, false
	}
	return TrackMenu, false
}

// Draw pinta un cuadro. El lienzo del juego mide 480x270 por Env.Scale (un
// factor que puede ser fraccionario). Si la imagen de la ventana tiene justo
// ese tamaño se pinta directo; si no (sobran franjas porque la pantalla no es
// 16:9) se pinta en un lienzo aparte y se pega centrado, sobre bordes oscuros,
// sin agrandar ni suavizar: el lienzo y la ventana tienen la misma resolución,
// así que no hay nada que se emborrone.
func (a *App) Draw(screen *ebiten.Image) {
	canvas := screen
	if b := screen.Bounds(); b.Dx() != a.canvW || b.Dy() != a.canvH {
		if a.canvas == nil || a.canvas.Bounds().Dx() != a.canvW || a.canvas.Bounds().Dy() != a.canvH {
			a.canvas = ebiten.NewImage(a.canvW, a.canvH)
		}
		canvas = a.canvas
		canvas.Clear()
	}
	a.scene.Draw(canvas)
	if canvas != screen {
		screen.Fill(a.env.Assets.Color('0')) // los bordes: el marrón más oscuro de la paleta
		op := &ebiten.DrawImageOptions{}     // filtro por defecto: vecino más cercano
		op.GeoM.Translate(float64(a.offX), float64(a.offY))
		screen.DrawImage(canvas, op)
	}
	if a.agent != nil {
		a.agent.takeShot(canvas, screen)
	}
}

// fitScreen reparte una imagen de ventana de w x h: la escala s (la mayor con
// la que 480x270 cabe, sin exigir que sea entera), el tamaño del lienzo
// (cw x ch, en píxeles de pantalla) y el corrimiento (ox, oy) que lo centra.
// Lo que sobra son los bordes oscuros: ninguno si la pantalla es 16:9. Es pura
// para poder probarla.
func fitScreen(w, h int) (s float64, cw, ch, ox, oy int) {
	s = math.Min(float64(w)/logicalWidth, float64(h)/logicalHeight)
	s = math.Max(s, 0.1) // ventana de 0 píxeles (minimizada): que no divida entre 0
	cw = min(int(math.Round(logicalWidth*s)), max(w, 1))
	ch = min(int(math.Round(logicalHeight*s)), max(h, 1))
	return s, cw, ch, (w - cw) / 2, (h - ch) / 2
}

// setWindowSize recibe el tamaño de la imagen de la ventana y calcula la
// escala, el lienzo y el corrimiento. En el modo pruebas -scale fija la escala
// (aunque la ventana sea de otro tamaño).
func (a *App) setWindowSize(w, h int) {
	s, cw, ch, ox, oy := fitScreen(w, h)
	if a.env.Opts.Agent && a.env.Opts.Scale > 0 {
		s = float64(a.env.Opts.Scale)
		cw, ch = logicalWidth*a.env.Opts.Scale, logicalHeight*a.env.Opts.Scale
		ox, oy = (w-cw)/2, (h-ch)/2
	}
	a.env.Scale = s
	a.winW, a.winH, a.canvW, a.canvH, a.offX, a.offY = w, h, cw, ch, ox, oy
}

// Layout decide el tamaño de la imagen de la ventana. En el juego de verdad es
// el tamaño real de la ventana (uno a uno: Ebitengine no la reescala, que es
// lo que la volvía borrosa), y el lienzo de 480x270 se agranda dentro de ella
// por un factor fraccionario (ver Draw). En el modo pruebas el tamaño sale de
// -screen o, si no, de 480x270 por -scale.
func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	w, h := outsideWidth, outsideHeight
	if o := a.env.Opts; o.Agent {
		switch {
		case o.ScreenW > 0:
			w, h = o.ScreenW, o.ScreenH
		case o.Scale > 0:
			w, h = logicalWidth*o.Scale, logicalHeight*o.Scale
		}
	}
	a.setWindowSize(w, h)
	return w, h
}

// LayoutF es Layout para pantallas con zoom del sistema (150 % en Windows, por
// ejemplo): Ebitengine da el tamaño en puntos y la ventana tiene más píxeles
// reales. Se multiplica por el factor para que la imagen tenga un píxel por
// cada píxel real, y así tampoco se reescala. Ebitengine usa esta función
// cuando existe.
func (a *App) LayoutF(outsideWidth, outsideHeight float64) (float64, float64) {
	if a.env.Opts.Agent {
		w, h := a.Layout(int(outsideWidth), int(outsideHeight))
		return float64(w), float64(h)
	}
	s := ebiten.Monitor().DeviceScaleFactor()
	w, h := int(math.Round(outsideWidth*s)), int(math.Round(outsideHeight*s))
	a.setWindowSize(w, h)
	return float64(w), float64(h)
}

// newApp arma la aplicación: carga arte y fuentes y elige la primera pantalla.
func newApp(opts Options, store *game.SaveStore) (*App, error) {
	assets, err := LoadAssets()
	if err != nil {
		return nil, fmt.Errorf("loading art: %w", err)
	}
	fonts, err := LoadFonts()
	if err != nil {
		return nil, fmt.Errorf("loading fonts: %w", err)
	}
	startScale := windowScale
	if opts.Agent && opts.Scale > 0 {
		startScale = opts.Scale
	}
	env := &Env{Scale: float64(startScale), Assets: assets, Fonts: fonts, Opts: opts, Store: store, Save: game.NewSaveData()}
	env.Sound = NewSound(opts.Agent) // el modo pruebas es siempre silencioso
	if store != nil {
		if data, err := store.Load(); err == nil {
			env.Save = data
		}
		env.HasSave = store.Exists() && env.Save.HasProgress() // solo ajustes guardados no cuentan
	}
	if opts.Agent && opts.SavedDay > 0 {
		env.Save.LastUnlockedDay = min(opts.SavedDay, game.MaxDay)
		env.HasSave = true
	}
	app := &App{env: env}
	app.in.reset()
	return app, nil
}

// Run abre la ventana y bloquea hasta que se cierra. Ebitengine exige
// llamarlo desde la goroutine principal. En el juego de verdad empieza en el
// menú; con -agent (opts.Agent) sigue la puerta de pruebas.
func Run(opts Options) error {
	var store *game.SaveStore
	switch {
	case !opts.Agent:
		path, err := game.DefaultSavePath()
		if err != nil {
			return err
		}
		store = game.NewSaveStore(path)
	case opts.SavePath != "":
		// Guardado de prueba: nunca el guardado real de la autora.
		real, err := game.DefaultSavePath()
		if err == nil && filepath.Clean(opts.SavePath) == filepath.Clean(real) {
			return errors.New("-savepath must not be the real save file")
		}
		store = game.NewSaveStore(opts.SavePath)
	}
	app, err := newApp(opts, store)
	if err != nil {
		return err
	}
	if opts.Agent && opts.ScreenW > 0 {
		ebiten.SetWindowSize(opts.ScreenW, opts.ScreenH)
	} else {
		ebiten.SetWindowSize(logicalWidth*int(app.env.Scale), logicalHeight*int(app.env.Scale))
	}
	if !opts.Agent {
		ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled) // la ventana se puede agrandar y el juego se adapta
	}
	ebiten.SetWindowTitle(windowTitle)
	if opts.Agent {
		scene, err := startScene(app.env)
		if err != nil {
			return err
		}
		app.scene = scene
		driver, err := newAgentDriver(opts)
		if err != nil {
			return err
		}
		app.agent = driver
		driver.prepareWindow()
	} else {
		app.scene = newMenu(app.env)
	}
	return ebiten.RunGame(app)
}

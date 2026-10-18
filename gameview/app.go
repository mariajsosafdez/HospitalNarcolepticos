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

	// Tick cuenta los ticks de la INTERFAZ (parpadeos, animaciones). No es el
	// reloj del turno: sigue corriendo aunque el juego esté en pausa.
	Tick int

	// Scale es cuántas veces se agranda el lienzo de 480x270 en pantalla (1,
	// 2, 3...). Todo lo que se dibuja lo multiplica, y el texto usa letras de
	// ese tamaño: así queda nítido en cualquier ventana.
	Scale int

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
	DemoStep        int          `json:"demoStep,omitempty"`        // 1 a 4
	DemoAuto        bool         `json:"demoAuto,omitempty"`
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

	// Tamaño de la imagen de la ventana y dónde cae el lienzo dentro de ella
	// (ver fitScreen). Layout los actualiza antes de cada Draw.
	winW, winH int
	offX, offY int
	canvas     *ebiten.Image // lienzo de 480x270 por Scale, cuando hay bordes
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
	if next := a.scene.Update(&a.in); next != nil {
		a.scene = next
	}
}

// Draw pinta un cuadro. El lienzo del juego mide 480x270 por Env.Scale (un
// factor ENTERO). Si la imagen de la ventana tiene justo ese tamaño se pinta
// directo; si no (pantalla completa o ventana que no es múltiplo exacto) se
// pinta en un lienzo aparte y se pega centrado, sobre bordes oscuros, sin
// agrandar ni suavizar: así cada píxel del juego sigue siendo un bloque
// exacto y el texto no se ve borroso.
func (a *App) Draw(screen *ebiten.Image) {
	k := a.env.Scale
	canvas := screen
	if b := screen.Bounds(); b.Dx() != logicalWidth*k || b.Dy() != logicalHeight*k {
		if a.canvas == nil || a.canvas.Bounds().Dx() != logicalWidth*k || a.canvas.Bounds().Dy() != logicalHeight*k {
			a.canvas = ebiten.NewImage(logicalWidth*k, logicalHeight*k)
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

// layoutScale elige el mayor factor entero (mínimo 1) con el que el lienzo de
// 480x270 cabe en una ventana de w x h. Es pura para poder probarla.
func layoutScale(w, h int) int {
	return max(1, min(w/logicalWidth, h/logicalHeight))
}

// fitScreen reparte una imagen de ventana de w x h: el factor entero k y el
// corrimiento (ox, oy) que centra el lienzo de 480k x 270k. Lo que sobra son
// los bordes oscuros. Es pura para poder probarla.
func fitScreen(w, h int) (k, ox, oy int) {
	k = layoutScale(w, h)
	return k, (w - logicalWidth*k) / 2, (h - logicalHeight*k) / 2
}

// setWindowSize recibe el tamaño de la imagen de la ventana y calcula el
// factor y el corrimiento. En el modo pruebas -scale fija el factor (aunque
// la ventana sea de otro tamaño).
func (a *App) setWindowSize(w, h int) {
	k, ox, oy := fitScreen(w, h)
	if a.env.Opts.Agent && a.env.Opts.Scale > 0 {
		k = a.env.Opts.Scale
		ox, oy = (w-logicalWidth*k)/2, (h-logicalHeight*k)/2
	}
	a.env.Scale = k
	a.winW, a.winH, a.offX, a.offY = w, h, ox, oy
}

// Layout decide el tamaño de la imagen de la ventana. En el juego de verdad es
// el tamaño real de la ventana (uno a uno: Ebitengine no la reescala, que es
// lo que la volvía borrosa), y el lienzo de 480x270 se agranda dentro de ella
// por un factor entero (ver Draw). En el modo pruebas el tamaño sale de
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
	env := &Env{Scale: startScale, Assets: assets, Fonts: fonts, Opts: opts, Store: store, Save: game.NewSaveData()}
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
		ebiten.SetWindowSize(logicalWidth*app.env.Scale, logicalHeight*app.env.Scale)
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

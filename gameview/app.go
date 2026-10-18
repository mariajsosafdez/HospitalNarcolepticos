// Package gameview es la VISTA del juego: abre la ventana, dibuja y lee la
// entrada con Ebitengine. Es el ÚNICO paquete del proyecto que importa
// Ebitengine; las reglas viven en el paquete game, que solo se lee y se
// manda acciones desde aquí.
package gameview

import (
	"fmt"
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

// recordShift anota el resultado de un turno y lo guarda (si hay almacén).
func (e *Env) recordShift(day, stars int) {
	e.Save.RecordShift(day, stars)
	e.HasSave = true
	if e.Store != nil {
		// Si no se puede guardar, el juego sigue: es solo el progreso.
		_ = e.Store.Save(e.Save)
	}
}

// UIState es lo que el volcado de pruebas dice de la interfaz (además del
// game.Snapshot del turno).
type UIState struct {
	Scene           string `json:"scene"`
	Selected        string `json:"selected,omitempty"`        // menú: opción elegida
	Camera          string `json:"camera,omitempty"`          // cámara que se mira
	SelectedPatient string `json:"selectedPatient,omitempty"` // paciente elegido para atender
	HoveredButton   string `json:"hoveredButton,omitempty"`   // botón bajo el mouse
	Notice          string `json:"notice,omitempty"`          // título del aviso visible
	Overlay         string `json:"overlay,omitempty"`         // pausa, reporte, próximamente
	StaffPage       int    `json:"staffPage,omitempty"`
	DemoStep        int    `json:"demoStep,omitempty"` // 1 a 4
	DemoAuto        bool   `json:"demoAuto,omitempty"`
	UITick          int    `json:"uiTick"`
	MouseX          int    `json:"mouseX"`
	MouseY          int    `json:"mouseY"`
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
}

// Update lo llama Ebitengine 60 veces por segundo: cada llamada es un tick
// del juego, igual que el reloj de game.Clock.
func (a *App) Update() error {
	if a.agent != nil {
		return a.agent.update(a)
	}
	a.in.readReal()
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

// Draw pinta un cuadro. La lógica no se toca aquí.
func (a *App) Draw(screen *ebiten.Image) {
	if a.agent != nil {
		a.agent.takeShot(a) // captura pendiente, con su propio tamaño
	}
	a.drawAt(screen, a.env.Scale)
}

// drawAt pinta la pantalla actual en dst con el factor dado.
func (a *App) drawAt(dst *ebiten.Image, scale int) {
	saved := a.env.Scale
	a.env.Scale = scale
	a.scene.Draw(dst)
	a.env.Scale = saved
}

// Layout elige el mayor factor entero que cabe en la ventana y devuelve ese
// tamaño: el juego se dibuja directo a resolución de pantalla (sin agrandar
// una imagen chica), que es lo que mantiene el texto nítido.
func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	k := max(1, min(outsideWidth/logicalWidth, outsideHeight/logicalHeight))
	a.env.Scale = k
	return logicalWidth * k, logicalHeight * k
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
	env := &Env{Scale: windowScale, Assets: assets, Fonts: fonts, Opts: opts, Store: store, Save: game.NewSaveData()}
	if store != nil {
		if data, err := store.Load(); err == nil {
			env.Save = data
		}
		env.HasSave = store.Exists()
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
	if !opts.Agent {
		path, err := game.DefaultSavePath()
		if err != nil {
			return err
		}
		store = game.NewSaveStore(path)
	}
	app, err := newApp(opts, store)
	if err != nil {
		return err
	}
	ebiten.SetWindowSize(logicalWidth*windowScale, logicalHeight*windowScale)
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

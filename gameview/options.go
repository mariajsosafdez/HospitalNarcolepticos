package gameview

import (
	"flag"
	"fmt"

	"costenos-narcolepsia/game"
)

// Options son las banderas de la puerta de pruebas (-agent). Se registran
// aquí, en gameview, para que main.go no tenga lógica: solo pregunta
// opts.Agent y llama a Run (ver Arquitectura.md).
type Options struct {
	Agent     bool   // modo pruebas: silencioso, sin guardado real
	Scene     string // menu, cams, demo, day
	Day       int
	Seed      int64
	Camera    string
	Step      int    // pasos de la Demo (1 a 4) con los que empieza
	InputFile string // guion de pasos (ver script.go)
	Frames    int    // ticks extra que se corren al final del guion
	Quit      bool   // cerrar al terminar el guion
	SavedDay  int    // finge una partida guardada con este último día
	Scale     int    // factor de dibujo fijo (1 a 4); 0 = el de la ventana (2 por defecto)
	Intro     bool   // -scene day: abre con el mensaje del Director del día (como el juego de verdad)
	SavePath  string // archivo de guardado de PRUEBA (nunca el real); vacío = sin guardado
}

// RegisterFlags define las banderas en fs y devuelve las opciones que se
// llenan cuando se hace fs.Parse.
func RegisterFlags(fs *flag.FlagSet) *Options {
	o := &Options{}
	fs.BoolVar(&o.Agent, "agent", false, "test mode: silent, never writes the real save file")
	fs.StringVar(&o.Scene, "scene", "menu", "with -agent: menu | cams | demo | day | dayintro | tutorial | director")
	fs.IntVar(&o.Day, "day", 1, "with -agent: day to play (1-5)")
	fs.Int64Var(&o.Seed, "seed", 1, "with -agent: random seed (same seed, same game)")
	fs.StringVar(&o.Camera, "cam", "", "with -agent: camera to look at, e.g. 2 or 7A")
	fs.IntVar(&o.Step, "step", 1, "with -agent: demo step to start at (1-4), or tutorial line (1-9) in -scene tutorial")
	fs.StringVar(&o.InputFile, "input", "", "with -agent: script file (wait, click, move, key, shot, shot4, dump, sleep, wake)")
	fs.IntVar(&o.Frames, "frames", 0, "with -agent: extra ticks to run after the script")
	fs.BoolVar(&o.Quit, "quit", false, "with -agent: close when the script ends")
	fs.IntVar(&o.SavedDay, "saved", 0, "with -agent: pretend there is a save with this last unlocked day")
	fs.IntVar(&o.Scale, "scale", 0, "with -agent: fixed drawing scale 1-4 (window is 480x270 times this)")
	fs.BoolVar(&o.Intro, "intro", false, "with -agent -scene day: open with the Director's message of that day")
	fs.StringVar(&o.SavePath, "savepath", "", "with -agent: scratch save file to read and write (never the real one)")
	return o
}

// validate revisa las opciones del modo pruebas y falla con un mensaje claro.
func (o Options) validate() error {
	switch o.Scene {
	case "menu", "cams", "demo", "day", "dayintro", "tutorial", "director":
	default:
		return fmt.Errorf("unknown scene %q (menu, cams, demo, day, dayintro, tutorial, director)", o.Scene)
	}
	if o.Day < 1 || o.Day > 5 {
		return fmt.Errorf("day %d out of range (1-5)", o.Day)
	}
	maxStep := 4
	if o.Scene == "tutorial" {
		maxStep = len(game.TutorialLines)
	}
	if o.Step < 1 || o.Step > maxStep {
		return fmt.Errorf("step %d out of range (1-%d)", o.Step, maxStep)
	}
	if o.Scene == "dayintro" && o.Day < 2 {
		return fmt.Errorf("dayintro needs -day 2 to 5 (day 1 opens with the tutorial)")
	}
	if o.Scale < 0 || o.Scale > 4 {
		return fmt.Errorf("scale %d out of range (0-4)", o.Scale)
	}
	if o.Frames < 0 {
		return fmt.Errorf("frames %d must not be negative", o.Frames)
	}
	return nil
}

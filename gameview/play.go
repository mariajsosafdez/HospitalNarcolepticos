package gameview

import "costenos-narcolepsia/game"

// Cómo se arma un turno jugable. Lo usan el menú, el reporte ("Siguiente
// día") y la puerta de pruebas, así los tres arman el día igual.

// newShiftScene arma la pantalla de cámaras del día con la semilla dada. Si
// intro es true, el turno abre con el mensaje del Director que le toca (el
// tutorial en el Día 1 la primera vez, el mensaje de inicio en los días 2 a
// 5). cam es la cámara inicial ("" = Recepción).
func newShiftScene(env *Env, day int, seed int64, intro bool, cam game.CameraID) (*camsScene, error) {
	return buildShiftScene(env, day, seed, intro, cam, true)
}

// buildShiftScene hace el trabajo de newShiftScene. Con eggs en true el turno
// puede traer un easter egg (15 %, o el que fuerce -egg); "Cómo jugar" lo
// llama con false porque es solo el tutorial.
func buildShiftScene(env *Env, day int, seed int64, intro bool, cam game.CameraID, eggs bool) (*camsScene, error) {
	cfg, err := game.ConfigForDay(day)
	if err != nil {
		return nil, err
	}
	if eggs {
		cfg = withEgg(env, cfg, seed)
	}
	shift, err := game.NewShift(cfg, seed)
	if err != nil {
		return nil, err
	}
	scene, err := newCams(env, shift, nil, cam)
	if err != nil {
		return nil, err
	}
	if intro {
		scene.queueIntro()
	}
	return scene, nil
}

// newPracticeScene es "Cómo jugar": un Día 1 con el tutorial completo (desde
// la línea start, 0 = la primera) y el turno detenido mientras se lee. Al
// terminar el último mensaje se vuelve al menú.
func newPracticeScene(env *Env, seed int64, start int, cam game.CameraID) (*camsScene, error) {
	scene, err := buildShiftScene(env, 1, seed, false, cam, false)
	if err != nil {
		return nil, err
	}
	scene.practice = true
	scene.showMessage(msgTutorial, game.TutorialLines[:], start)
	return scene, nil
}

// withEgg le agrega a la configuración del día el easter egg que toque: el
// que pida la puerta de pruebas (-egg N, que además llega en el tick 0 para
// verlo sin esperar) o, si no, el del sorteo de la semilla (game.RollEgg).
func withEgg(env *Env, cfg game.DayConfig, seed int64) game.DayConfig {
	if env.Opts.Agent && env.Opts.Egg > 0 {
		if egg, ok := game.EggByNumber(env.Opts.Egg); ok {
			cfg.ArriveAtStart = egg.ID
			return cfg.WithEgg(egg)
		}
	}
	if egg, ok := game.RollEgg(seed); ok {
		return cfg.WithEgg(egg)
	}
	return cfg
}

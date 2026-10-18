package gameview

import "costenos-narcolepsia/game"

// Cómo se arma un turno jugable. Lo usan el menú, el reporte ("Siguiente
// día") y la puerta de pruebas, así los tres arman el día igual.

// newShiftScene arma la pantalla de cámaras del día con la semilla dada. Si
// intro es true, el turno abre con el mensaje del Director que le toca (el
// tutorial en el Día 1 la primera vez, el mensaje de inicio en los días 2 a
// 5). cam es la cámara inicial ("" = Recepción).
func newShiftScene(env *Env, day int, seed int64, intro bool, cam game.CameraID) (*camsScene, error) {
	cfg, err := game.ConfigForDay(day)
	if err != nil {
		return nil, err
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
	scene, err := newShiftScene(env, 1, seed, false, cam)
	if err != nil {
		return nil, err
	}
	scene.practice = true
	scene.showMessage(msgTutorial, game.TutorialLines[:], start)
	return scene, nil
}

package gameview

import (
	"fmt"
	"strconv"
	"strings"
)

// Guion de la puerta de pruebas: un archivo de texto con un paso por línea
// (o varios separados por ";"). Las líneas que empiezan con # son comentarios.
//
//	wait N          corre N ticks
//	move X Y        pone el mouse en (X, Y) del lienzo de 480x270
//	click X Y       mueve el mouse y hace clic (dura 1 tick)
//	key Nombre      pulsa una tecla (Up Down Left Right Enter Escape Space M 1..8)
//	shot ruta       captura a 1x (480x270)
//	shot4 ruta      captura a 4x (1920x1080)
//	dump ruta       JSON con game.Snapshot y el estado de la interfaz
//	sleep ID CAM    duerme al paciente ID en la cámara CAM (ForceSleep)
//	wake ID         despierta al paciente ID que está en cama (ForceWake)
//
// Un paso desconocido, o con argumentos mal escritos, hace fallar el
// arranque con un mensaje claro: así un error de escritura no pasa
// desapercibido.

// stepKind es el tipo de paso.
type stepKind int

const (
	stepWait stepKind = iota
	stepMove
	stepClick
	stepKey
	stepShot
	stepShot4
	stepDump
	stepSleep
	stepWake
)

// step es un paso ya leído.
type step struct {
	kind stepKind
	n    int    // wait: ticks
	x, y int    // move, click
	text string // key: nombre; shot, shot4, dump: ruta; sleep, wake: ID
	cam  string // sleep: cámara
	line int    // número de línea del guion (para los errores)
}

// parseScript lee el texto de un guion y devuelve sus pasos.
func parseScript(src string) ([]step, error) {
	var steps []step
	for i, raw := range strings.Split(src, "\n") {
		line := i + 1
		raw = strings.TrimSpace(raw)
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		for _, part := range strings.Split(raw, ";") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			s, err := parseStep(part)
			if err != nil {
				return nil, fmt.Errorf("script line %d: %w", line, err)
			}
			s.line = line
			steps = append(steps, s)
		}
	}
	return steps, nil
}

// parseStep lee un solo paso, por ejemplo "click 120 80".
func parseStep(text string) (step, error) {
	f := strings.Fields(text)
	name, args := f[0], f[1:]
	need := func(n int) error {
		if len(args) != n {
			return fmt.Errorf("%q needs %d argument(s), got %d", name, n, len(args))
		}
		return nil
	}
	ints := func() ([]int, error) {
		out := make([]int, len(args))
		for i, a := range args {
			v, err := strconv.Atoi(a)
			if err != nil {
				return nil, fmt.Errorf("%q: %q is not a number", name, a)
			}
			out[i] = v
		}
		return out, nil
	}
	switch name {
	case "wait":
		if err := need(1); err != nil {
			return step{}, err
		}
		v, err := ints()
		if err != nil {
			return step{}, err
		}
		if v[0] < 0 {
			return step{}, fmt.Errorf("wait %d must not be negative", v[0])
		}
		return step{kind: stepWait, n: v[0]}, nil
	case "move", "click":
		if err := need(2); err != nil {
			return step{}, err
		}
		v, err := ints()
		if err != nil {
			return step{}, err
		}
		kind := stepMove
		if name == "click" {
			kind = stepClick
		}
		return step{kind: kind, x: v[0], y: v[1]}, nil
	case "key":
		if err := need(1); err != nil {
			return step{}, err
		}
		if err := checkKeyName(args[0]); err != nil {
			return step{}, err
		}
		return step{kind: stepKey, text: args[0]}, nil
	case "shot", "shot4", "dump":
		if err := need(1); err != nil {
			return step{}, err
		}
		kind := map[string]stepKind{"shot": stepShot, "shot4": stepShot4, "dump": stepDump}[name]
		return step{kind: kind, text: args[0]}, nil
	case "sleep":
		if err := need(2); err != nil {
			return step{}, err
		}
		return step{kind: stepSleep, text: args[0], cam: args[1]}, nil
	case "wake":
		if err := need(1); err != nil {
			return step{}, err
		}
		return step{kind: stepWake, text: args[0]}, nil
	}
	return step{}, fmt.Errorf("unknown step %q", name)
}

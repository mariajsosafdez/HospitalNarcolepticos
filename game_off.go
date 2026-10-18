//go:build nogame

package main

import (
	"errors"
	"fmt"
)

// agentMode es siempre false: sin Ebitengine no hay puerta de pruebas.
func agentMode() bool { return false }

// runAgent no se usa con la etiqueta nogame.
func runAgent() error { return errors.New("the game was compiled out (-tags nogame)") }

// runGame es la versión del tag nogame: el juego se compiló SIN Ebitengine,
// así que solo avisa. Sirve para correr el escenario obligatorio en equipos
// sin las librerías gráficas (por ejemplo, un Linux sin X11/OpenGL).
func runGame() error {
	fmt.Println("\n== Game ==\n  The game window was compiled out (built with -tags nogame). Run without the tag to play.")
	return nil
}

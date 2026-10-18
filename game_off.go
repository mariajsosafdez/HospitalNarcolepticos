//go:build nogame

package main

import "fmt"

// runGame es la versión del tag nogame: el juego se compiló SIN Ebitengine,
// así que solo avisa. Sirve para correr el escenario obligatorio en equipos
// sin las librerías gráficas (por ejemplo, un Linux sin X11/OpenGL).
func runGame() error {
	fmt.Println("\n== Game ==\n  The game window was compiled out (built with -tags nogame). Run without the tag to play.")
	return nil
}

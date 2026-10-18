//go:build !nogame

package main

import (
	"flag"
	"fmt"

	"costenos-narcolepsia/gameview"
)

// agentOpts son las banderas de la puerta de pruebas (-agent, -scene...).
// Se registran al iniciar el programa, antes de flag.Parse() de main.
var agentOpts = gameview.RegisterFlags(flag.CommandLine)

// agentMode dice si se pidió la puerta de pruebas (-agent).
func agentMode() bool { return agentOpts.Agent }

// runAgent corre el juego en modo pruebas: sin escenario por consola, en
// silencio y sin tocar el guardado real.
func runAgent() error { return gameview.Run(*agentOpts) }

// runGame abre la ventana del juego en el menú principal.
func runGame() error {
	fmt.Println("\n== Game ==\n  Opening the game window (close it to exit)")
	return gameview.Run(gameview.Options{})
}

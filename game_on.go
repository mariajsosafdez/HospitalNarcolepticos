//go:build !nogame

package main

import (
	"fmt"
	"time"

	"costenos-narcolepsia/game"
	"costenos-narcolepsia/gameview"
)

// runGame abre la ventana del juego (Día 1). La semilla viene del reloj para
// que cada partida sea distinta; la lógica de game nunca lee el reloj.
func runGame() error {
	cfg, err := game.ConfigForDay(1)
	if err != nil {
		return err
	}
	shift, err := game.NewShift(cfg, time.Now().UnixNano())
	if err != nil {
		return err
	}
	fmt.Println("\n== Game ==\n  Opening the game window (close it to exit)")
	return gameview.Run(shift)
}

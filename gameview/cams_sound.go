package gameview

import "costenos-narcolepsia/game"

// Sonidos de la pantalla de cámaras que dependen de lo que se ve en la cámara
// actual: las ruedas de la camilla y el ronquido de quien duerme. Los demás
// (alta, despierta, sin cama) se disparan donde ocurren: trackBlinks y
// showError.

// carriedInView dice si la cámara actual muestra una camilla en camino.
func (c *camsScene) carriedInView() bool {
	for _, g := range c.shift.Patients() {
		if g.Camera() == c.shift.CurrentCamera() && g.State() == game.Carried {
			return true
		}
	}
	return false
}

// sleeperInView dice si hay alguien durmiendo en la cámara actual (en el piso,
// sin cama, o ya en su cama).
func (c *camsScene) sleeperInView() bool {
	for _, g := range c.shift.Patients() {
		if g.Camera() != c.shift.CurrentCamera() {
			continue
		}
		switch g.State() {
		case game.AsleepWaiting, game.NoBedWaiting, game.InBed:
			return true
		}
	}
	return false
}

// updateSounds corre una vez por tick (con el turno andando, no en pausa).
func (c *camsScene) updateSounds() {
	if c.carriedInView() {
		c.env.Sound.Roll()
	}
	if !c.sleeperInView() {
		return
	}
	if c.snoreIn <= 0 {
		// La primera vez (o tras roncar) se sortea cuánto esperar; el ronquido
		// suena al acabarse la espera.
		c.snoreIn = snoreGap(c.snores.Float64())
	}
	c.snoreIn--
	if c.snoreIn == 0 {
		c.env.Sound.Play(CueSnore)
	}
}

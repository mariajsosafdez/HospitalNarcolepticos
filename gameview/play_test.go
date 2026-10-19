package gameview

import (
	"testing"

	"costenos-narcolepsia/game"
)

// withEgg: sin -egg manda el sorteo de la semilla (y nunca más de uno); con
// -egg N en la puerta de pruebas entra ese easter egg y llega en el tick 0.
func TestWithEgg(t *testing.T) {
	cfg, _ := game.ConfigForDay(2)
	base := len(cfg.Patients)

	// Semillas 1 y 39: la 1 no trae easter egg y la 39 trae a Misifu.
	env := &Env{}
	if got := withEgg(env, cfg, 1); len(got.Patients) != base {
		t.Errorf("seed 1: %d patients, want %d", len(got.Patients), base)
	}
	got := withEgg(env, cfg, 39)
	if len(got.Patients) != base+1 || got.Patients[base].ID != game.CatID {
		t.Errorf("seed 39: want Misifu added, got %+v", got.Patients)
	}

	// -egg solo vale en la puerta de pruebas.
	env = &Env{Opts: Options{Agent: true, Egg: 4}}
	got = withEgg(env, cfg, 1)
	if len(got.Patients) != base+1 || got.Patients[base].ID != game.CoteID || got.ArriveAtStart != game.CoteID {
		t.Errorf("-egg 4: got %+v (arrive %q)", got.Patients[base:], got.ArriveAtStart)
	}
	env = &Env{Opts: Options{Egg: 4}}
	if got := withEgg(env, cfg, 1); len(got.Patients) != base {
		t.Error("-egg must be ignored outside -agent")
	}
	if len(cfg.Patients) != base {
		t.Error("withEgg changed the original config")
	}
}

func TestEggOptionRange(t *testing.T) {
	for egg, wantErr := range map[int]bool{0: false, 1: false, 5: false, 6: true, -1: true} {
		o := Options{Scene: "day", Day: 1, Step: 1, Egg: egg}
		if err := o.validate(); (err != nil) != wantErr {
			t.Errorf("egg %d: err = %v, wantErr %v", egg, err, wantErr)
		}
	}
}

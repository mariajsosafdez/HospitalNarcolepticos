package hospital

import (
	"fmt"
	"sync/atomic"
	"time"
)

// EpisodeRecord es el registro de un ataque de sueño: quién se durmió,
// cuándo, dónde, quién lo atendió y en qué cuarto terminó.
//
// Es una "foto" del momento del episodio: una vez creado no cambia nunca.
// Por eso guarda el NÚMERO del cuarto y no un puntero a Room (el cuarto
// sigue cambiando después, pero el registro debe mostrar lo que pasó).
type EpisodeRecord struct {
	id         string
	time       time.Time
	patient    *Patient
	attender   Attender // puede ser un *Doctor o un *Orderly (polimorfismo); nil = nadie atendió
	location   string
	roomNumber int // 0 significa "no se le pudo asignar cuarto"
}

// episodeCounter genera IDs únicos (E-001, E-002...).
//
// atomic.Int64 es un contador seguro para concurrencia: si dos goroutines
// llaman Add(1) al mismo tiempo, cada una recibe un número distinto. Con un
// int normal (counter++) ambas podrían leer el mismo valor y repetir el ID.
var episodeCounter atomic.Int64

// newEpisodeRecord crea un registro con la hora actual. Lee el cuarto que el
// paciente tenga en ese momento (si no tiene, queda en 0).
func newEpisodeRecord(p *Patient, a Attender, location string) EpisodeRecord {
	roomNumber := 0
	if p.room != nil {
		roomNumber = p.room.number
	}
	return EpisodeRecord{
		id:         fmt.Sprintf("E-%03d", episodeCounter.Add(1)),
		time:       time.Now(),
		patient:    p,
		attender:   a,
		location:   location,
		roomNumber: roomNumber,
	}
}

// Getters con receptor por VALOR: EpisodeRecord es inmutable y se maneja como
// valor (se copia), así que no hace falta puntero.

// ID devuelve el identificador del episodio.
func (e EpisodeRecord) ID() string { return e.id }

// Time devuelve la fecha y hora del episodio.
func (e EpisodeRecord) Time() time.Time { return e.time }

// Patient devuelve el paciente que se durmió.
func (e EpisodeRecord) Patient() *Patient { return e.patient }

// Attender devuelve quién atendió el episodio (nil si nadie pudo).
func (e EpisodeRecord) Attender() Attender { return e.attender }

// Location devuelve el lugar donde el paciente se quedó dormido.
func (e EpisodeRecord) Location() string { return e.location }

// RoomNumber devuelve el número de cuarto asignado, o 0 si no hubo cuarto.
func (e EpisodeRecord) RoomNumber() int { return e.roomNumber }

// HasRoom dice si al paciente se le asignó un cuarto en este episodio.
func (e EpisodeRecord) HasRoom() bool { return e.roomNumber != 0 }

// Summary devuelve una línea de texto que resume el episodio.
// (Es el único método "de texto" que pide el enunciado; no imprime nada,
// solo arma el string y quien lo llame decide si lo imprime.)
func (e EpisodeRecord) Summary() string {
	destination := "stays in the hallway"
	if e.HasRoom() {
		destination = fmt.Sprintf("room %d", e.roomNumber)
	}
	attendedBy := "nobody"
	if e.attender != nil {
		attendedBy = fmt.Sprintf("%s (%s)", e.attender.Name(), e.attender.Role())
	}
	return fmt.Sprintf("%s  %s  %s %-18s at %-26s -> %-20s attended by %s",
		e.time.Format("2006-01-02 15:04"), e.id,
		e.patient.ID(), e.patient.Name(), e.location, destination, attendedBy)
}

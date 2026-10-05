package hospital

import (
	"fmt"
	"slices"
)

// Orderly es un camillero: puede llevar a un paciente dormido hasta una cama,
// pero NO puede diagnosticar (por eso no tiene método DiagnosePatient).
//
// Es el segundo tipo concreto que cumple la interfaz Attender. Gracias a
// eso el hospital lo guarda en el mismo []Attender que a los doctores.
type Orderly struct {
	Person
	episodes []EpisodeRecord // traslados que hizo
}

// NewOrderly crea un camillero validando sus datos.
func NewOrderly(id, name string, age int) (*Orderly, error) {
	person, err := newPerson(id, name, age)
	if err != nil {
		return nil, fmt.Errorf("new orderly %q: %w", id, err)
	}
	return &Orderly{Person: person}, nil
}

// Role cumple la interfaz Attender.
func (o *Orderly) Role() string { return "Orderly" }

// IsAvailable cumple la interfaz Attender: el camillero no tiene pacientes
// a cargo, así que siempre está disponible para un traslado.
func (o *Orderly) IsAvailable() bool { return true }

// Attend cumple la interfaz Attender. A diferencia del doctor, el camillero
// solo verifica que el paciente esté dormido y registra el traslado;
// no diagnostica ni toma al paciente a su cargo.
func (o *Orderly) Attend(p *Patient, location string) (EpisodeRecord, error) {
	if p == nil {
		return EpisodeRecord{}, fmt.Errorf("orderly attend: %w", ErrInvalidData)
	}
	if p.state == Awake {
		return EpisodeRecord{}, fmt.Errorf("orderly attend %s: %w", p.id, ErrPatientAwake)
	}
	record := newEpisodeRecord(p, o, location)
	o.episodes = append(o.episodes, record)
	return record, nil
}

// MyEpisodes devuelve una copia de los episodios atendidos por el camillero.
func (o *Orderly) MyEpisodes() []EpisodeRecord {
	return slices.Clone(o.episodes)
}

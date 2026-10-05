package hospital

import (
	"fmt"
	"slices"
)

// RoomState indica si un cuarto todavía tiene camas libres.
// Igual que los otros estados, es un tipo propio con constantes (no strings).
type RoomState int

// Available = 0 (tiene al menos una cama libre), Occupied = 1 (lleno).
const (
	Available RoomState = iota
	Occupied
)

// String devuelve el estado del cuarto en texto legible.
func (s RoomState) String() string {
	switch s {
	case Available:
		return "Available"
	case Occupied:
		return "Occupied"
	default:
		return "Unknown"
	}
}

// Room es un cuarto del hospital con una cantidad fija de camas (capacity).
type Room struct {
	number    int
	capacity  int
	state     RoomState
	occupants []*Patient // punteros: son los MISMOS pacientes del hospital, no copias
}

// NewRoom crea un cuarto vacío y disponible.
func NewRoom(number, capacity int) (*Room, error) {
	if number <= 0 || capacity <= 0 {
		return nil, fmt.Errorf("new room %d: number and capacity must be positive: %w", number, ErrInvalidData)
	}
	return &Room{number: number, capacity: capacity, state: Available}, nil
}

// Number devuelve el número del cuarto (por ejemplo 101).
func (r *Room) Number() int { return r.number }

// Capacity devuelve cuántas camas tiene el cuarto.
func (r *Room) Capacity() int { return r.capacity }

// State devuelve si el cuarto está Available u Occupied.
func (r *Room) State() RoomState { return r.state }

// Occupants devuelve una COPIA del slice de ocupantes. Si devolviéramos
// r.occupants directamente, quien lo reciba podría hacer append o cambiar
// posiciones y romper el estado interno del cuarto (rompería el encapsulamiento).
func (r *Room) Occupants() []*Patient {
	return slices.Clone(r.occupants)
}

// IsAvailable dice si queda al menos una cama libre.
func (r *Room) IsAvailable() bool {
	return len(r.occupants) < r.capacity
}

// Occupy acuesta a un paciente en una cama de este cuarto.
// Si el cuarto está lleno devuelve ErrRoomFull en vez de pasarse de capacidad.
func (r *Room) Occupy(p *Patient) error {
	if p == nil {
		return fmt.Errorf("occupy room %d: %w", r.number, ErrInvalidData)
	}
	if slices.Contains(r.occupants, p) {
		return fmt.Errorf("occupy room %d: patient %s is already here: %w", r.number, p.id, ErrDuplicateID)
	}
	if !r.IsAvailable() {
		return fmt.Errorf("occupy room %d: %w", r.number, ErrRoomFull)
	}
	r.occupants = append(r.occupants, p)
	p.moveToBed(r) // el paciente queda AsleepInBed y sabe en qué cuarto está
	r.updateState()
	return nil
}

// Release saca a un paciente del cuarto y libera su cama.
func (r *Room) Release(p *Patient) error {
	i := slices.Index(r.occupants, p) // posición del paciente, o -1 si no está
	if i == -1 {
		return fmt.Errorf("release room %d: %w", r.number, ErrNotInRoom)
	}
	r.occupants = slices.Delete(r.occupants, i, i+1)
	p.leaveRoom()
	r.updateState()
	return nil
}

// updateState recalcula el estado del cuarto según cuántas camas quedan.
// Es privado: el estado nunca se cambia "a mano" desde fuera.
func (r *Room) updateState() {
	if r.IsAvailable() {
		r.state = Available
	} else {
		r.state = Occupied
	}
}

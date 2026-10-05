package hospital

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

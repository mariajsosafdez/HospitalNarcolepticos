package game

import "errors"

// Errores que devuelven las acciones del jugador. La vista los convierte en
// mensajes amables. Los errores del modelo (por ejemplo
// hospital.ErrNoRoomAvailable) se devuelven envueltos con %w, así que se
// reconocen con errors.Is igual que en el paquete hospital.
var (
	ErrUnknownPatient     = errors.New("unknown patient")
	ErrUnknownStaff       = errors.New("unknown staff member")
	ErrUnknownCamera      = errors.New("camera does not exist or is locked")
	ErrWrongState         = errors.New("the patient is not in a state that allows this action")
	ErrStaffBusy          = errors.New("the staff member is in transfer or resting")
	ErrStaffNoQuota       = errors.New("the doctor already has 4 patients")
	ErrShiftOver          = errors.New("the shift is over")
	ErrInvalidDay         = errors.New("day must be between 1 and 5")
	ErrUnexpectedDispatch = errors.New("the chosen staff member was not dispatched")
)

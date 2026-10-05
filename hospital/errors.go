package hospital

import "errors"

// Errores "centinela" (sentinel errors) del dominio.
//
// En Go los errores son valores normales. Declararlos una sola vez como
// variables exportadas permite que quien llame a un método pueda preguntar
// QUÉ falló con errors.Is(err, hospital.ErrNoRoomAvailable), aunque el error
// venga envuelto con más contexto mediante fmt.Errorf("...: %w", err).
//
// Así el hospital nunca usa panic para situaciones previsibles: devuelve
// un error y el programa sigue funcionando.
var (
	// ErrNoRoomAvailable: no hay ningún cuarto libre para un paciente dormido.
	ErrNoRoomAvailable = errors.New("no room available")
	// ErrPatientNotFound: el paciente no está admitido en el hospital.
	ErrPatientNotFound = errors.New("patient not found")
	// ErrDuplicateID: ya existe alguien (o un cuarto) con ese identificador.
	ErrDuplicateID = errors.New("duplicate identifier")
	// ErrInvalidData: datos vacíos, negativos o nil al crear o registrar algo.
	ErrInvalidData = errors.New("invalid data")
	// ErrRoomFull: se intentó ocupar un cuarto que ya está lleno.
	ErrRoomFull = errors.New("room is full")
	// ErrNotInRoom: se intentó sacar de un cuarto a alguien que no está en él.
	ErrNotInRoom = errors.New("patient is not in this room")
	// ErrAlreadyAwake: se intentó despertar a un paciente que ya está despierto.
	ErrAlreadyAwake = errors.New("patient is already awake")
	// ErrPatientAwake: se intentó atender una emergencia de sueño de alguien despierto.
	ErrPatientAwake = errors.New("patient is awake, there is no sleep emergency")
	// ErrNotInHallway: se pidió cuarto para alguien que no está dormido en el pasillo.
	ErrNotInHallway = errors.New("patient is not asleep in a hallway")
	// ErrNoStaffAvailable: ningún miembro del personal puede atender ahora.
	ErrNoStaffAvailable = errors.New("no staff member available")
)

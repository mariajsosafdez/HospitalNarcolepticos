package game

import (
	"errors"
	"fmt"

	"costenos-narcolepsia/hospital"
)

// El modo Demo es un turno SIN AZAR que sigue el guion del escenario
// obligatorio (runScenario de main.go): mismos datos, mismo orden. Usa las
// mismas acciones del jugador (Assign, MoveToBed, Discharge...), así que
// también prueba la lógica real. La única ayuda es ForceSleep/ForceWake, que
// reemplazan al azar para que el guion siempre sea igual.

// DemoStep es el resultado de un paso del guion, listo para mostrarse.
type DemoStep struct {
	Title string // qué pasa en este paso
	Note  string // detalle extra, por ejemplo el error esperado de "sin cama"
}

type demoAction struct {
	title string
	run   func(s *Shift) (note string, err error)
}

// Demo avanza el guion paso a paso (botón "Siguiente paso" o modo Auto).
type Demo struct {
	shift   *Shift
	actions []demoAction
	next    int
}

// maxWaitTicks evita un bucle infinito si algo del guion nunca se cumple.
const maxWaitTicks = 5 * ShiftTicks

// NewDemo arma el turno de la Demo con los datos del escenario del enunciado
// y los diagnósticos iniciales (cada doctor toma a algunos pacientes severos).
func NewDemo() (*Demo, error) {
	shift, err := NewShift(DemoConfig(), 1)
	if err != nil {
		return nil, err
	}
	// h.Doctors() sale vacío en el hospital del juego, así que el diagnóstico
	// se hace con el *Doctor real que guarda cada StaffMember.
	diagnoses := []struct{ doctorID, patientID string }{
		{"D-001", "P-002"},
		{"D-001", "P-004"},
		{"D-002", "P-005"},
	}
	for _, dg := range diagnoses {
		member, patient := shift.FindStaff(dg.doctorID), shift.FindPatient(dg.patientID)
		if member == nil || patient == nil {
			return nil, fmt.Errorf("demo: unknown %s or %s", dg.doctorID, dg.patientID)
		}
		if err := member.Doctor().DiagnosePatient(patient.Patient()); err != nil {
			return nil, err
		}
	}
	d := &Demo{shift: shift}
	d.actions = []demoAction{
		{"Yesid se duerme en la cafetería y el Dr. Karen lo lleva a cama", sleepAndAssign("P-001", "2", "cafeteria (sancocho line)", "D-001")},
		{"Yeimy se duerme en el salón de champeta y el Dr. Alvaro la lleva a cama", sleepAndAssign("P-002", "8", "champeta dance hall", "D-002")},
		{"Rafa se duerme en la fila de radiología y Chepe lo lleva a cama", sleepAndAssign("P-003", "5", "radiology queue", "O-001")},
		{"Wilfrido se duerme en el pasillo: no hay cama y se queda ahí", sleepAndAssign("P-004", "4", "hallway 2, second floor", "D-001")},
		{"Yesid despierta y se le da de alta: su cama queda libre", dischargeFirst},
		{"Wilfrido pasa del pasillo a la cama que se liberó", moveHallwayPatient},
		{"Reporte del turno con las 4 consultas", reportStep},
	}
	return d, nil
}

// Shift devuelve el turno que la Demo está manejando (la vista lo dibuja).
func (d *Demo) Shift() *Shift { return d.shift }

// Done dice si ya se ejecutaron todos los pasos.
func (d *Demo) Done() bool { return d.next >= len(d.actions) }

// StepCount y StepIndex sirven para mostrar "paso 3 de 7".
func (d *Demo) StepCount() int { return len(d.actions) }
func (d *Demo) StepIndex() int { return d.next }

// Next ejecuta el siguiente paso del guion. Devuelve error solo si algo
// salió mal de verdad (no el error esperado de "sin cama").
func (d *Demo) Next() (DemoStep, error) {
	if d.Done() {
		return DemoStep{}, errors.New("demo: no more steps")
	}
	action := d.actions[d.next]
	d.next++
	note, err := action.run(d.shift)
	return DemoStep{Title: action.title, Note: note}, err
}

// waitUntil avanza el reloj hasta que cond sea verdadera.
func waitUntil(s *Shift, cond func() bool) error {
	for i := 0; !cond(); i++ {
		if i >= maxWaitTicks || s.Finished() {
			return errors.New("demo: the awaited condition never happened")
		}
		s.Tick()
	}
	return nil
}

// sleepAndAssign: espera a que el personal esté libre, duerme al paciente
// donde dice el guion y lo asigna. Si no hay cama, el error del modelo es
// lo esperado y se devuelve como nota.
func sleepAndAssign(patientID string, cam CameraID, location, staffID string) func(*Shift) (string, error) {
	return func(s *Shift) (string, error) {
		member := s.FindStaff(staffID)
		if err := waitUntil(s, member.Free); err != nil {
			return "", err
		}
		if err := s.ForceSleep(patientID, cam, location); err != nil {
			return "", err
		}
		err := s.Assign(patientID, staffID)
		if errors.Is(err, hospital.ErrNoRoomAvailable) {
			return err.Error(), nil
		}
		return "", err
	}
}

// dischargeFirst despierta a P-001 en su cama y le da de alta.
func dischargeFirst(s *Shift) (string, error) {
	if err := waitUntil(s, func() bool { return s.FindPatient("P-001").State() == InBed }); err != nil {
		return "", err
	}
	if err := s.ForceWake("P-001"); err != nil {
		return "", err
	}
	s.Tick() // el Tick lo pasa a AwakeInBed
	return "", s.Discharge("P-001")
}

// moveHallwayPatient manda a P-004 a la cama liberada, con el camillero.
func moveHallwayPatient(s *Shift) (string, error) {
	if err := waitUntil(s, s.FindStaff("O-001").Free); err != nil {
		return "", err
	}
	return "", s.MoveToBed("P-004", "O-001")
}

// reportStep espera a que la camilla llegue y arma el reporte.
func reportStep(s *Shift) (string, error) {
	if err := waitUntil(s, func() bool { return s.FindPatient("P-004").State() == InBed }); err != nil {
		return "", err
	}
	r := s.Report()
	return fmt.Sprintf("%d dormidos en pasillo, %d camas libres de %d", len(r.Hallway), r.FreeRooms, len(r.Rooms)), nil
}

package game

import (
	"slices"

	"costenos-narcolepsia/hospital"
)

// AutoStep es un jugador automático MUY simple: en cada llamada atiende lo
// que haya pendiente, usando solo las acciones públicas del turno (Assign,
// MoveToBed, Discharge), las mismas que usa la persona que juega. Sirve para
// los tests (jugar los 5 días enteros sin ventana) y para que la puerta de
// pruebas llegue a un reporte de fin de turno sin escribir un guion largo.
// Devuelve cuántas acciones hizo.
func AutoStep(s *Shift) int {
	if s.Finished() {
		return 0
	}
	done := 0
	for _, g := range s.Patients() {
		id := g.Patient().ID()
		switch g.State() {
		case AwakeInBed:
			if s.Discharge(id) == nil {
				done++
			}
		case AsleepWaiting:
			if m := freeStaffFor(s, g.Patient().Level() == hospital.Severe); m != nil {
				// El error "sin cama" es normal: el paciente queda esperando.
				_ = s.Assign(id, m.ID())
				done++
			}
		case NoBedWaiting:
			if hasFreeRoom(s) {
				if m := freeStaffFor(s, false); m != nil && s.MoveToBed(id, m.ID()) == nil {
					done++
				}
			}
		}
	}
	return done
}

// freeStaffFor busca a alguien que el jugador pueda elegir ahora. Para los
// severos prefiere un doctor (da bonus); para los demás, un camillero, así
// los doctores no se llenan.
func freeStaffFor(s *Shift, doctorFirst bool) *StaffMember {
	staff := s.Staff()
	if !doctorFirst {
		slices.Reverse(staff) // los camilleros se contratan después de los doctores
	}
	for _, m := range staff {
		if checkFree(m) == nil && m.IsDoctor() == doctorFirst {
			return m
		}
	}
	for _, m := range staff { // si no hay del tipo preferido, cualquiera libre
		if checkFree(m) == nil {
			return m
		}
	}
	return nil
}

// hasFreeRoom dice si hay una cama libre en el hospital.
func hasFreeRoom(s *Shift) bool {
	for _, r := range s.Hospital().Rooms() {
		if r.IsAvailable() {
			return true
		}
	}
	return false
}

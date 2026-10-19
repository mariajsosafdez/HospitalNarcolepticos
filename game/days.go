package game

import (
	"slices"

	"costenos-narcolepsia/hospital"
)

// Datos de personas que se crean al armar un turno. Son structs simples para
// que la configuración de cada día sea una "tabla" fácil de leer.
type DoctorData struct {
	ID, Name  string
	Age       int
	Specialty string
}

type OrderlyData struct {
	ID, Name string
	Age      int
}

type PatientData struct {
	ID, Name string
	Age      int
	Level    hospital.NarcolepsyLevel
}

// DayConfig describe todo lo que cambia de un día a otro (sección 6).
type DayConfig struct {
	Day       int
	Doctors   []DoctorData
	Orderlies []OrderlyData
	Rooms     []int      // números de habitación (101, 102...), todas de capacidad 1
	Cameras   []CameraID // zonas comunes desbloqueadas
	Patients  []PatientData
	Scripted  bool // true en la Demo: sin movimiento ni sueño al azar
	// ArriveAtStart es el ID de un paciente que llega en el tick 0 en vez de a
	// una hora al azar. Solo lo usa la puerta de pruebas (-egg) para ver un
	// easter egg sin esperar horas de juego; vacío = nadie.
	ArriveAtStart string
}

// Elenco fijo aprobado (docs/HospitalCosteños/Elenco.md).
var (
	castDoctors = []DoctorData{
		{"D-001", "Dra. Karen Ospina", 41, "Neurología"},
		{"D-002", "Dr. Álvaro Mendoza", 55, "Medicina del sueño"},
		{"D-003", "Dra. Luz Dary Fontalvo", 38, "Medicina general"},
	}
	castOrderlies = []OrderlyData{
		{"O-001", "Chepe Barrios", 33},
		{"O-002", "Jhon Jairo \"El Mono\" Arrieta", 29},
	}
	// Las edades de P-006 a P-010 salen de la tabla del Elenco.
	castPatients = []PatientData{
		{"P-001", "Yesid Pacheco", 34, hospital.Moderate},
		{"P-002", "Yeimy Padilla", 28, hospital.Severe},
		{"P-003", "Rafa Cantillo", 52, hospital.Mild},
		{"P-004", "Wilfrido Berrio", 45, hospital.Severe},
		{"P-005", "Marelvis Ortega", 61, hospital.Severe},
		{"P-006", "Nayeli Mercado", 19, hospital.Mild},
		{"P-007", "Abelardo \"Mañe\" Julio", 73, hospital.Moderate},
		{"P-008", "Kelly Johana Ruiz", 37, hospital.Moderate},
		{"P-009", "Eder Salcedo", 41, hospital.Mild},
		{"P-010", "Yuranis Polo", 25, hospital.Moderate},
	}
)

// dayRow es una fila de la tabla de días.
type dayRow struct {
	doctors, orderlies, rooms int
	cameras                   []CameraID
	patients                  int // cuántos del Elenco vienen ese día
}

var dayTable = [...]dayRow{
	{2, 1, 3, []CameraID{"1A", "1B", "2", "3"}, 10},
	{2, 1, 4, []CameraID{"1A", "1B", "2", "3", "4"}, 10},
	{3, 1, 4, []CameraID{"1A", "1B", "2", "3", "4", "5"}, 10},
	{3, 2, 5, []CameraID{"1A", "1B", "2", "3", "4", "5", "6"}, 10},
	{3, 2, 6, []CameraID{"1A", "1B", "2", "3", "4", "5", "6", "8"}, 10},
}

// MaxDay es el último día del juego.
const MaxDay = 5

// ConfigForDay arma la configuración del día 1 a 5.
//
// Decisión de la autora (2026-10-17): como el Elenco tiene 10 pacientes, los
// días 3 a 5 también traen 10 (la dificultad sube por cámaras, camas y
// penalizaciones). Si sobra tiempo se agregan personajes y se sube a 12.
func ConfigForDay(day int) (DayConfig, error) {
	if day < 1 || day > MaxDay {
		return DayConfig{}, ErrInvalidDay
	}
	row := dayTable[day-1]
	cfg := DayConfig{
		Day:       day,
		Doctors:   slices.Clone(castDoctors[:row.doctors]),
		Orderlies: slices.Clone(castOrderlies[:row.orderlies]),
		Cameras:   slices.Clone(row.cameras),
		Patients:  slices.Clone(castPatients[:min(row.patients, len(castPatients))]),
	}
	for i := 0; i < row.rooms; i++ {
		cfg.Rooms = append(cfg.Rooms, 101+i)
	}
	return cfg, nil
}

// DemoConfig tiene EXACTAMENTE los datos de runScenario en main.go: mismos
// IDs, nombres, edades, niveles y habitaciones. Todas las zonas comunes están
// desbloqueadas porque la Demo se duerme en la cafetería (CAM 2), el salón de
// champeta (CAM 8), radiología (CAM 5) y el pasillo (CAM 4).
func DemoConfig() DayConfig {
	return DayConfig{
		Day: 1,
		Doctors: []DoctorData{
			{"D-001", "Dr. Karen Ospina", 41, "Neurology"},
			{"D-002", "Dr. Alvaro Mendoza", 55, "Sleep Medicine"},
		},
		Orderlies: []OrderlyData{{"O-001", "Chepe Barrios", 33}},
		Rooms:     []int{101, 102, 103},
		Cameras:   []CameraID{"1A", "1B", "2", "3", "4", "5", "6", "8"},
		Patients: []PatientData{
			{"P-001", "Yesid Pacheco", 34, hospital.Moderate},
			{"P-002", "Yeimy Padilla", 28, hospital.Severe},
			{"P-003", "Rafa Cantillo", 52, hospital.Mild},
			{"P-004", "Wilfrido Berrio", 45, hospital.Severe},
			{"P-005", "Marelvis Ortega", 61, hospital.Severe},
		},
		Scripted: true,
	}
}

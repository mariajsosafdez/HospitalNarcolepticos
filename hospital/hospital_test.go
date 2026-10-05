package hospital

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// Las pruebas en Go viven en archivos *_test.go y son funciones que empiezan
// con Test y reciben *testing.T. Se ejecutan con: go test ./...
//
// t.Fatalf detiene la prueba con un mensaje; t.Errorf marca el fallo pero
// sigue revisando el resto.

// newTestHospital arma un hospital pequeño para las pruebas: 1 doctor,
// 1 camillero, roomCount cuartos de capacidad 1 y los pacientes dados.
// t.Helper() hace que, si algo falla aquí, Go reporte la línea de la prueba
// que llamó a esta función (más fácil de depurar).
func newTestHospital(t *testing.T, roomCount int, patients ...*Patient) *Hospital {
	t.Helper()
	h := NewHospital("Test Hospital")

	doctor, err := NewDoctor("D-001", "Dr. Test", 40, "Neurology")
	if err != nil {
		t.Fatalf("NewDoctor: %v", err)
	}
	orderly, err := NewOrderly("O-001", "Orderly Test", 30)
	if err != nil {
		t.Fatalf("NewOrderly: %v", err)
	}
	if err := h.HireDoctor(doctor); err != nil {
		t.Fatalf("HireDoctor: %v", err)
	}
	if err := h.HireStaff(orderly); err != nil {
		t.Fatalf("HireStaff: %v", err)
	}

	for i := 0; i < roomCount; i++ {
		room, err := NewRoom(101+i, 1)
		if err != nil {
			t.Fatalf("NewRoom: %v", err)
		}
		if err := h.AddRoom(room); err != nil {
			t.Fatalf("AddRoom: %v", err)
		}
	}
	for _, p := range patients {
		if err := h.AdmitPatient(p); err != nil {
			t.Fatalf("AdmitPatient: %v", err)
		}
	}
	return h
}

// mustPatient crea un paciente o detiene la prueba si los datos son inválidos.
func mustPatient(t *testing.T, id, name string, level NarcolepsyLevel) *Patient {
	t.Helper()
	p, err := NewPatient(id, name, 30, level)
	if err != nil {
		t.Fatalf("NewPatient: %v", err)
	}
	return p
}

// Prueba obligatoria 1: asignación de cuarto cuando SÍ hay uno libre.
func TestAssignRoomWithFreeRoom(t *testing.T) {
	p := mustPatient(t, "P-001", "Yesid Pacheco", Moderate)
	h := newTestHospital(t, 1, p)

	p.SufferSleepAttack("cafeteria")
	room, err := h.AssignRoom(p)

	if err != nil {
		t.Fatalf("expected a room, got error: %v", err)
	}
	if room.Number() != 101 {
		t.Errorf("expected room 101, got %d", room.Number())
	}
	if room.State() != Occupied {
		t.Errorf("expected room state Occupied, got %v", room.State())
	}
	if p.State() != AsleepInBed {
		t.Errorf("expected patient AsleepInBed, got %v", p.State())
	}
	if p.Room() != room {
		t.Errorf("patient should know its room")
	}
}

// Prueba obligatoria 2: asignación de cuarto cuando NO hay ninguno libre.
func TestAssignRoomNoRoomAvailable(t *testing.T) {
	first := mustPatient(t, "P-001", "Yesid Pacheco", Moderate)
	second := mustPatient(t, "P-002", "Yeimy Padilla", Severe)
	h := newTestHospital(t, 1, first, second)

	// El primero ocupa el único cuarto.
	if err := h.RegisterEpisode(first, "cafeteria"); err != nil {
		t.Fatalf("first episode should get the room: %v", err)
	}

	// El segundo debe fallar con ErrNoRoomAvailable y quedarse en el pasillo.
	err := h.RegisterEpisode(second, "radiology queue")
	if !errors.Is(err, ErrNoRoomAvailable) {
		t.Fatalf("expected ErrNoRoomAvailable, got %v", err)
	}
	if second.State() != AsleepInHallway {
		t.Errorf("expected patient to stay AsleepInHallway, got %v", second.State())
	}
	if second.Room() != nil {
		t.Errorf("no room must be invented for the patient")
	}
	// El episodio sí se registra aunque no haya cuarto.
	if got := len(h.History()); got != 2 {
		t.Errorf("expected 2 episodes in history, got %d", got)
	}
}

// Prueba obligatoria 3 (una consulta): pacientes dormidos en el pasillo.
func TestPatientsInHallway(t *testing.T) {
	inBed := mustPatient(t, "P-001", "Yesid Pacheco", Moderate)
	inHallway := mustPatient(t, "P-002", "Yeimy Padilla", Severe)
	awake := mustPatient(t, "P-003", "Rafa Cantillo", Mild)
	h := newTestHospital(t, 1, inBed, inHallway, awake)

	_ = h.RegisterEpisode(inBed, "cafeteria")            // obtiene el cuarto
	_ = h.RegisterEpisode(inHallway, "hallway 2, floor") // se queda sin cuarto

	hallway := h.PatientsInHallway()
	if len(hallway) != 1 || hallway[0] != inHallway {
		t.Fatalf("expected only P-002 in hallway, got %d patients", len(hallway))
	}
}

// Consulta 5.4: reporte de pacientes Severe, incluyendo los que tienen 0 episodios.
func TestSevereReport(t *testing.T) {
	severeTwice := mustPatient(t, "P-002", "Yeimy Padilla", Severe)
	severeNever := mustPatient(t, "P-005", "Marelvis Ortega", Severe)
	mild := mustPatient(t, "P-003", "Rafa Cantillo", Mild)
	h := newTestHospital(t, 3, severeTwice, severeNever, mild)

	_ = h.RegisterEpisode(severeTwice, "cafeteria")
	_ = h.WakeUpPatient(severeTwice)
	_ = h.RegisterEpisode(severeTwice, "champeta dance hall")
	_ = h.RegisterEpisode(mild, "radiology queue")

	report := h.SevereReport()
	if len(report) != 2 {
		t.Fatalf("expected 2 severe patients in report, got %d", len(report))
	}
	if report[severeTwice] != 2 {
		t.Errorf("expected 2 episodes for P-002, got %d", report[severeTwice])
	}
	// "comma ok" para distinguir "está con 0" de "no está en el map".
	if count, ok := report[severeNever]; !ok || count != 0 {
		t.Errorf("expected P-005 present with 0 episodes, got %d (present=%v)", count, ok)
	}
	if _, ok := report[mild]; ok {
		t.Errorf("mild patient must not appear in the severe report")
	}
}

// Consulta 5.2: el doctor conoce sus propios episodios.
func TestDoctorMyEpisodes(t *testing.T) {
	p := mustPatient(t, "P-001", "Yesid Pacheco", Moderate)
	h := newTestHospital(t, 1, p)

	// El primer turno del round-robin es el doctor (fue contratado primero).
	_ = h.RegisterEpisode(p, "cafeteria")

	doctor := h.Doctors()[0]
	episodes := doctor.MyEpisodes()
	if len(episodes) != 1 {
		t.Fatalf("expected 1 episode for the doctor, got %d", len(episodes))
	}
	if episodes[0].Patient() != p || episodes[0].RoomNumber() != 101 {
		t.Errorf("unexpected episode: %s", episodes[0].Summary())
	}
	if p.AssignedDoctor() != doctor {
		t.Errorf("doctor should have taken the patient under care")
	}
}

// Polimorfismo: el hospital despacha por turnos a tipos distintos de personal.
func TestStaffDispatchIsPolymorphic(t *testing.T) {
	a := mustPatient(t, "P-001", "Yesid Pacheco", Moderate)
	b := mustPatient(t, "P-002", "Yeimy Padilla", Severe)
	h := newTestHospital(t, 2, a, b)

	_ = h.RegisterEpisode(a, "cafeteria")
	_ = h.RegisterEpisode(b, "radiology queue")

	history := h.History()
	if history[0].Attender().Role() != "Doctor" {
		t.Errorf("first episode should be attended by the doctor, got %s", history[0].Attender().Role())
	}
	if history[1].Attender().Role() != "Orderly" {
		t.Errorf("second episode should be attended by the orderly, got %s", history[1].Attender().Role())
	}
}

// Al despertar, el paciente libera su cuarto y otro puede usarlo.
func TestWakeUpReleasesRoom(t *testing.T) {
	sleeper := mustPatient(t, "P-001", "Yesid Pacheco", Moderate)
	waiting := mustPatient(t, "P-004", "Wilfrido Berrio", Severe)
	h := newTestHospital(t, 1, sleeper, waiting)

	_ = h.RegisterEpisode(sleeper, "cafeteria")
	_ = h.RegisterEpisode(waiting, "hallway 2") // sin cuarto

	room := sleeper.Room()
	if err := h.WakeUpPatient(sleeper); err != nil {
		t.Fatalf("WakeUpPatient: %v", err)
	}
	if sleeper.State() != Awake || room.State() != Available {
		t.Fatalf("room should be Available after wake up, got %v", room.State())
	}

	got, err := h.AssignRoom(waiting)
	if err != nil || got != room {
		t.Fatalf("waiting patient should get the released room, err=%v", err)
	}
	if err := h.WakeUpPatient(sleeper); !errors.Is(err, ErrAlreadyAwake) {
		t.Errorf("expected ErrAlreadyAwake, got %v", err)
	}
}

// Bono de concurrencia: varias goroutines usan el hospital a la vez.
// Esta prueba tiene sentido sobre todo con el detector de carreras:
//
//	go test -race ./...
//
// Si algún dato se leyera o escribiera sin el candado, -race lo reportaría
// como "WARNING: DATA RACE" y la prueba fallaría.
func TestSimulationIsRaceFree(t *testing.T) {
	var patients []*Patient
	for i := 1; i <= 8; i++ {
		id := fmt.Sprintf("P-%03d", i)
		patients = append(patients, mustPatient(t, id, "Patient "+id, Severe))
	}
	h := newTestHospital(t, 3, patients...)

	// Mientras la simulación corre, otra goroutine toma fotos sin parar
	// (como hace el servidor web). done sirve para avisarle que se detenga.
	done := make(chan struct{})
	readerFinished := make(chan struct{})
	go func() {
		defer close(readerFinished)
		for {
			select {
			case <-done:
				return
			default:
				_ = h.Snapshot()
			}
		}
	}()

	result := h.Simulate(SimulationConfig{Rounds: 20, MaxDelay: time.Millisecond})
	close(done)
	<-readerFinished

	if result.Goroutines != len(patients) {
		t.Errorf("expected %d goroutines, got %d", len(patients), result.Goroutines)
	}
	if result.Episodes != len(h.History()) {
		t.Errorf("episodes counted (%d) must match history (%d)", result.Episodes, len(h.History()))
	}
	// Invariante: ningún cuarto puede tener más pacientes que camas, y todo
	// paciente "en cama" debe estar realmente dentro de un cuarto.
	snap := h.Snapshot()
	for _, r := range snap.Rooms {
		if len(r.Occupants) > r.Capacity {
			t.Errorf("room %d has %d occupants for %d beds", r.Number, len(r.Occupants), r.Capacity)
		}
	}
	for _, p := range snap.Patients {
		if p.State == AsleepInBed && p.RoomNumber == 0 {
			t.Errorf("patient %s is in bed without a room", p.ID)
		}
	}
}

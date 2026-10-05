// Programa principal del Hospital de los Costeños con Narcolepsia.
//
// Regla del enunciado: main SOLO arma el escenario y imprime. Toda la lógica
// de negocio (asignar cuartos, despachar personal, consultas) vive en el
// paquete hospital; aquí únicamente se llaman sus métodos y se muestran
// los datos que devuelven.
package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"time"

	"costenos-narcolepsia/hospital"
)

const hospitalName = "Hospital de los Costeños con Narcolepsia"

func main() {
	if _, err := runScenario(os.Stdout); err != nil {
		// Solo llegamos aquí si los DATOS del escenario están mal escritos
		// (por ejemplo, un ID repetido). No es una situación del hospital.
		log.Fatalf("could not build the scenario: %v", err)
	}

	if err := runSimulation(os.Stdout); err != nil {
		log.Fatalf("could not run the simulation: %v", err)
	}
}

// runSimulation ejecuta el bono de concurrencia sobre un hospital NUEVO (con
// los mismos datos iniciales), para no alterar el resultado del escenario.
func runSimulation(out io.Writer) error {
	h, err := buildHospital()
	if err != nil {
		return err
	}
	fmt.Fprintln(out, "\n== Bonus: concurrent simulation (one goroutine per patient) ==")

	// Simulate bloquea hasta que todas las goroutines terminen.
	result := h.Simulate(hospital.SimulationConfig{Rounds: 7, MaxDelay: 40 * time.Millisecond})

	fmt.Fprintf(out, "  goroutines launched: %d\n", result.Goroutines)
	fmt.Fprintf(out, "  sleep attacks: %d (no room available: %d)\n", result.Episodes, result.NoRoom)
	fmt.Fprintf(out, "  wake ups: %d, patients moved from hallway to bed: %d\n", result.WakeUps, result.Relocations)

	// Snapshot: lectura segura del estado final.
	snap := h.Snapshot()
	fmt.Fprintln(out, "  final state of the rooms:")
	for _, r := range snap.Rooms {
		occupants := "-"
		for i, p := range r.Occupants {
			if i == 0 {
				occupants = ""
			} else {
				occupants += ", "
			}
			occupants += p.ID + " " + p.Name
		}
		fmt.Fprintf(out, "    room %d  %-9v  %s\n", r.Number, r.State, occupants)
	}
	fmt.Fprintf(out, "  patients still asleep in a hallway: %d\n", len(snap.Hallway))
	return nil
}

// runScenario ejecuta el escenario obligatorio de la sección 6, en orden,
// y escribe todo el texto en out. Recibir un io.Writer (en vez de usar
// fmt.Println directo) permite mandar el texto a la consola, a un buffer
// o a ninguna parte (io.Discard) sin cambiar esta función.
func runScenario(out io.Writer) (*hospital.Hospital, error) {
	// ---------------------------------------------------------------- Paso 1
	h, err := buildHospital()
	if err != nil {
		return nil, err
	}
	printStep1(out, h)

	// ---------------------------------------------------------------- Paso 2
	// Cuatro pacientes se duermen en lugares distintos. Solo hay 3 cuartos,
	// así que el cuarto paciente (P-004) debe quedarse en el pasillo.
	fmt.Fprintln(out, "\n== Step 2: four sleep attacks ==")
	attacks := []struct{ patientID, location string }{
		{"P-001", "cafeteria (sancocho line)"},
		{"P-002", "champeta dance hall"},
		{"P-003", "radiology queue"},
		{"P-004", "hallway 2, second floor"},
	}
	for _, attack := range attacks {
		p, err := h.FindPatient(attack.patientID)
		if err != nil {
			return nil, err
		}
		err = h.RegisterEpisode(p, attack.location)
		printEpisodeResult(out, h, p, attack.location, err)
	}

	fmt.Fprintln(out, "\n  Patients asleep in a hallway right now:")
	printHallway(out, h.PatientsInHallway())

	// ---------------------------------------------------------------- Paso 3
	// P-001 se despierta, libera su cuarto y el paciente del pasillo lo recibe.
	fmt.Fprintln(out, "\n== Step 3: a patient wakes up and the hallway patient gets the bed ==")
	p1, err := h.FindPatient("P-001")
	if err != nil {
		return nil, err
	}
	freedRoom := p1.Room()
	if err := h.WakeUpPatient(p1); err != nil {
		fmt.Fprintf(out, "  [!!] %v\n", err)
	} else {
		fmt.Fprintf(out, "  %s %s woke up -> room %d is now %v\n",
			p1.ID(), p1.Name(), freedRoom.Number(), freedRoom.State())
	}
	for _, p := range h.PatientsInHallway() {
		room, err := h.AssignRoom(p)
		if err != nil {
			fmt.Fprintf(out, "  [!!] %s %s: %v\n", p.ID(), p.Name(), err)
			continue
		}
		fmt.Fprintf(out, "  [OK] %s %s moved from the hallway to room %d\n", p.ID(), p.Name(), room.Number())
	}

	// ---------------------------------------------------------------- Paso 4
	printQueries(out, h)
	return h, nil
}

// buildHospital crea el hospital del paso 1: 2 doctores, 1 camillero,
// 3 cuartos de capacidad 1 y 5 pacientes (3 de ellos Severe).
func buildHospital() (*hospital.Hospital, error) {
	h := hospital.NewHospital(hospitalName)

	// Slices de structs anónimos: una "tabla" de datos fácil de leer.
	doctors := []struct {
		id, name  string
		age       int
		specialty string
	}{
		{"D-001", "Dr. Karen Ospina", 41, "Neurology"},
		{"D-002", "Dr. Alvaro Mendoza", 55, "Sleep Medicine"},
	}
	for _, d := range doctors {
		doctor, err := hospital.NewDoctor(d.id, d.name, d.age, d.specialty)
		if err != nil {
			return nil, err
		}
		if err := h.HireDoctor(doctor); err != nil {
			return nil, err
		}
	}

	orderly, err := hospital.NewOrderly("O-001", "Chepe Barrios", 33)
	if err != nil {
		return nil, err
	}
	if err := h.HireStaff(orderly); err != nil {
		return nil, err
	}

	for _, number := range []int{101, 102, 103} {
		room, err := hospital.NewRoom(number, 1)
		if err != nil {
			return nil, err
		}
		if err := h.AddRoom(room); err != nil {
			return nil, err
		}
	}

	patients := []struct {
		id, name string
		age      int
		level    hospital.NarcolepsyLevel
	}{
		{"P-001", "Yesid Pacheco", 34, hospital.Moderate},
		{"P-002", "Yeimy Padilla", 28, hospital.Severe},
		{"P-003", "Rafa Cantillo", 52, hospital.Mild},
		{"P-004", "Wilfrido Berrio", 45, hospital.Severe},
		{"P-005", "Marelvis Ortega", 61, hospital.Severe},
	}
	for _, data := range patients {
		p, err := hospital.NewPatient(data.id, data.name, data.age, data.level)
		if err != nil {
			return nil, err
		}
		if err := h.AdmitPatient(p); err != nil {
			return nil, err
		}
	}

	// Diagnósticos iniciales: cada doctor toma a algunos pacientes Severe.
	diagnoses := []struct {
		doctorIndex int
		patientID   string
	}{
		{0, "P-002"},
		{0, "P-004"},
		{1, "P-005"},
	}
	for _, dg := range diagnoses {
		p, err := h.FindPatient(dg.patientID)
		if err != nil {
			return nil, err
		}
		if err := h.Doctors()[dg.doctorIndex].DiagnosePatient(p); err != nil {
			return nil, err
		}
	}
	return h, nil
}

// ---------------------------------------------------------------------------
// Funciones de impresión: solo leen datos del hospital y los muestran.
// ---------------------------------------------------------------------------

func printStep1(out io.Writer, h *hospital.Hospital) {
	fmt.Fprintf(out, "==================================================\n")
	fmt.Fprintf(out, "  %s\n", h.Name())
	fmt.Fprintf(out, "==================================================\n")
	fmt.Fprintln(out, "\n== Step 1: hospital setup ==")

	// h.Staff() devuelve []Attender: aquí hay Doctores y un Camillero mezclados
	// y los recorremos igual, sin saber qué tipo concreto es cada uno.
	fmt.Fprintln(out, "  Staff:")
	for _, a := range h.Staff() {
		fmt.Fprintf(out, "    %-6s %-20s %s\n", a.ID(), a.Name(), a.Role())
	}
	fmt.Fprintln(out, "  Rooms:")
	for _, r := range h.Rooms() {
		fmt.Fprintf(out, "    room %d  capacity %d  %v\n", r.Number(), r.Capacity(), r.State())
	}
	fmt.Fprintln(out, "  Patients:")
	for _, p := range h.Patients() {
		doctor := "-"
		if d := p.AssignedDoctor(); d != nil {
			doctor = d.Name()
		}
		fmt.Fprintf(out, "    %-6s %-17s age %-3d %-9v doctor: %s\n",
			p.ID(), p.Name(), p.Age(), p.Level(), doctor)
	}
}

// printEpisodeResult muestra qué pasó con un ataque de sueño. Usa errors.Is
// para reconocer el caso "no hay cuarto" y mostrar un mensaje legible.
func printEpisodeResult(out io.Writer, h *hospital.Hospital, p *hospital.Patient, location string, err error) {
	history := h.History()
	attendedBy := "nobody"
	if len(history) > 0 {
		if a := history[len(history)-1].Attender(); a != nil {
			attendedBy = a.Name()
		}
	}

	switch {
	case errors.Is(err, hospital.ErrNoRoomAvailable):
		fmt.Fprintf(out, "  [!!] %s %s fell asleep at %s: no room available, the patient stays in the hallway (attended by %s)\n",
			p.ID(), p.Name(), location, attendedBy)
	case err != nil:
		fmt.Fprintf(out, "  [!!] %s %s: %v\n", p.ID(), p.Name(), err)
	default:
		fmt.Fprintf(out, "  [OK] %s %s fell asleep at %s -> room %d (attended by %s)\n",
			p.ID(), p.Name(), location, p.Room().Number(), attendedBy)
	}
}

func printHallway(out io.Writer, patients []*hospital.Patient) {
	if len(patients) == 0 {
		fmt.Fprintln(out, "  (nobody is asleep in a hallway)")
		return
	}
	for _, p := range patients {
		fmt.Fprintf(out, "  %-6s %-17s (%s)\n", p.ID(), p.Name(), p.Location())
	}
}

// printQueries imprime el paso 4: el resultado de las cuatro consultas.
func printQueries(out io.Writer, h *hospital.Hospital) {
	fmt.Fprintln(out, "\n== Step 4: required queries ==")

	// Consulta 5.1
	fmt.Fprintln(out, "\n== Query 5.1: patients asleep in hallway ==")
	printHallway(out, h.PatientsInHallway())

	// Consulta 5.2
	fmt.Fprintln(out, "\n== Query 5.2: episodes attended by each doctor ==")
	for _, d := range h.Doctors() {
		fmt.Fprintf(out, "  %s (%s):\n", d.Name(), d.Specialty())
		episodes := d.MyEpisodes()
		if len(episodes) == 0 {
			fmt.Fprintln(out, "    (no episodes)")
		}
		for _, e := range episodes {
			destination := "stayed in the hallway"
			if e.HasRoom() {
				destination = fmt.Sprintf("room %d", e.RoomNumber())
			}
			fmt.Fprintf(out, "    %s  %-6s %-17s %-26s -> %s\n",
				e.Time().Format("2006-01-02 15:04"), e.Patient().ID(), e.Patient().Name(), e.Location(), destination)
		}
	}

	// Consulta 5.3 (el resultado de AssignRoom ya se vio en los pasos 2 y 3;
	// aquí se muestra cómo quedaron las camas).
	fmt.Fprintln(out, "\n== Query 5.3: bed availability ==")
	free := 0
	for _, r := range h.Rooms() {
		occupants := "-"
		for i, p := range r.Occupants() {
			if i == 0 {
				occupants = ""
			} else {
				occupants += ", "
			}
			occupants += p.ID() + " " + p.Name()
		}
		if r.IsAvailable() {
			free++
		}
		fmt.Fprintf(out, "  room %d  %-9v  %s\n", r.Number(), r.State(), occupants)
	}
	fmt.Fprintf(out, "  free rooms: %d of %d\n", free, len(h.Rooms()))

	// Consulta 5.4. El map no tiene orden, así que sacamos las claves a un
	// slice y lo ordenamos por ID para que la salida sea siempre igual.
	fmt.Fprintln(out, "\n== Query 5.4: severity report (Severe patients) ==")
	report := h.SevereReport()
	severe := make([]*hospital.Patient, 0, len(report))
	for p := range report {
		severe = append(severe, p)
	}
	slices.SortFunc(severe, func(a, b *hospital.Patient) int {
		return cmp.Compare(a.ID(), b.ID())
	})
	for _, p := range severe {
		fmt.Fprintf(out, "  %-6s %-17s %d episode(s) today\n", p.ID(), p.Name(), report[p])
	}
}

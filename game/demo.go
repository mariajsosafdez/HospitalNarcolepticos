package game

import (
	"errors"
	"fmt"

	"costenos-narcolepsia/hospital"
)

// El modo Demo es un turno SIN AZAR que sigue el guion del escenario
// obligatorio (runScenario de main.go): mismos datos, mismo orden. Usa las
// mismas acciones del jugador (Assign, MoveToBed, Discharge...), así que
// también prueba la lógica real. La única ayuda es ForceMove/ForceSleep/
// ForceWake, que reemplazan al azar para que el guion siempre sea igual.
//
// La Demo es una secuencia GUIADA de 8 partes (ver demoScript). Cada parte
// es una lista de "beats" (momentos): un paciente se duerme, el personal
// entra, se lo lleva, corte a la habitación... Cada beat dura unos ticks.
//
// REGLA DE ORO: el reloj del turno solo corre mientras una parte se anima
// (Demo.Tick). Entre parte y parte todo queda quieto esperando a "Siguiente
// paso" (o a Auto), y el turno nunca llega a las 18:00 dentro de la Demo.

// Cuántas partes tiene la Demo (1, 2.1 a 2.4, 3.1, 3.2 y 4).
const DemoParts = 8

// BeatKind dice qué pasa en un beat.
type BeatKind int

const (
	BeatMingle    BeatKind = iota // el paciente está de pie en la cámara, todavía despierto
	BeatSleep                     // se queda dormido (zzz)
	BeatArrive                    // el personal entra caminando hacia el paciente
	BeatCarry                     // lo atiende (Assign o MoveToBed) y se lo lleva en la camilla hasta la cama
	BeatNoBed                     // lo atiende pero no hay cama: aviso "No hay cama libre!"
	BeatLeave                     // el personal se va caminando a descansar
	BeatBed                       // corte a la habitación: el paciente ya está en la cama
	BeatWake                      // el paciente despierta en su cama ("!")
	BeatHighlight                 // el botón "Dar de alta" se resalta
	BeatDischarge                 // el botón se pulsa solo (Discharge): la cama queda libre
	BeatReport                    // la parte final: el reporte
)

// Beat es un momento de una parte de la Demo.
type Beat struct {
	Kind    BeatKind
	Camera  CameraID // cámara que se mira al empezar el beat ("" = la misma)
	Patient string   // ID del paciente
	Staff   string   // ID del personal
	Place   string   // BeatSleep: el lugar que se le reporta al hospital
	Frames  int      // duración mínima, en ticks de la Demo (60 por segundo)
}

// EventKind es algo que la vista quiere enterarse de que pasó.
type EventKind int

const (
	EventNone EventKind = iota
	EventNoBed
)

// DemoEvent avisa a la vista de algo puntual (hoy solo: se quedó sin cama).
type DemoEvent struct {
	Kind    EventKind
	Patient string
}

// carrySpeed: el traslado real dura 5 s (doctor) o 3 s (camillero) de juego.
// En la Demo se muestra 4 veces más rápido: por cada tick de la Demo el
// reloj del turno avanza 4 ticks mientras la camilla va camino a la cama.
const carrySpeed = 4

// maxBeatFrames evita un bucle infinito si algo del guion nunca se cumple.
const maxBeatFrames = 2000

// Demo avanza el guion parte por parte (botón "Siguiente paso" o modo Auto).
type Demo struct {
	shift   *Shift
	script  [DemoParts][]Beat
	part    int  // parte que se muestra (0 = la parte 1)
	playing bool // una parte se está animando
	beat    int  // beat actual dentro de la parte
	frames  int  // ticks de la Demo que lleva el beat actual
	done    bool // ya se llegó a la última parte (el reporte)
}

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
	return &Demo{shift: shift, script: demoScript()}, nil
}

// Duraciones de los beats, en ticks de la Demo (60 = 1 s).
const (
	mingleFrames    = 24
	sleepFrames     = 42
	arriveFrames    = 54
	noBedFrames     = 100
	leaveFrames     = 48
	bedFrames       = 54
	wakeFrames      = 54
	highlightFrames = 42
	dischargeFrames = 54
)

// sleeperPart arma una parte donde un paciente se duerme y el personal se lo
// lleva a su habitación: se duerme, entra el personal, se lo lleva, corte a
// la cámara de la habitación y ahí está en la cama.
func sleeperPart(patient, staff string, cam CameraID, place string, roomCam CameraID) []Beat {
	return []Beat{
		{Kind: BeatMingle, Camera: cam, Patient: patient, Frames: mingleFrames},
		{Kind: BeatSleep, Camera: cam, Patient: patient, Place: place, Frames: sleepFrames},
		{Kind: BeatArrive, Patient: patient, Staff: staff, Frames: arriveFrames},
		{Kind: BeatCarry, Patient: patient, Staff: staff},
		{Kind: BeatBed, Camera: roomCam, Patient: patient, Frames: bedFrames},
	}
}

// demoScript es el guion de las 8 partes, igual a la tabla de "Textos del
// Director" (Demo guiada): 1, 2.1 a 2.4, 3.1, 3.2 y 4. Los lugares y los
// quién-atiende-a-quién son los de runScenario en main.go.
func demoScript() [DemoParts][]Beat {
	return [DemoParts][]Beat{
		// 1. Se abre el turno: nada se anima, solo se presenta.
		nil,
		// 2.1 a 2.3: tres pacientes se duermen y consiguen cama.
		sleeperPart("P-001", "D-001", "2", "cafeteria (sancocho line)", "7A"),
		sleeperPart("P-002", "D-002", "8", "champeta dance hall", "7B"),
		sleeperPart("P-003", "O-001", "5", "radiology queue", "7C"),
		// 2.4: Wilfrido se duerme y la doctora lo atiende, pero ya no hay
		// cama: ella se va a descansar y él se queda en el pasillo.
		{
			{Kind: BeatMingle, Camera: "4", Patient: "P-004", Frames: mingleFrames},
			{Kind: BeatSleep, Camera: "4", Patient: "P-004", Place: "hallway 2, second floor", Frames: sleepFrames},
			{Kind: BeatArrive, Patient: "P-004", Staff: "D-001", Frames: arriveFrames},
			{Kind: BeatNoBed, Patient: "P-004", Staff: "D-001", Frames: noBedFrames},
			{Kind: BeatLeave, Patient: "P-004", Staff: "D-001", Frames: leaveFrames},
		},
		// 3.1: Yesid despierta y se le da de alta: la cama 101 queda libre.
		{
			{Kind: BeatWake, Camera: "7A", Patient: "P-001", Frames: wakeFrames},
			{Kind: BeatHighlight, Patient: "P-001", Frames: highlightFrames},
			{Kind: BeatDischarge, Patient: "P-001", Frames: dischargeFrames},
		},
		// 3.2: Chepe lleva a Wilfrido del pasillo a la cama que se liberó.
		{
			{Kind: BeatArrive, Camera: "4", Patient: "P-004", Staff: "O-001", Frames: arriveFrames},
			{Kind: BeatCarry, Patient: "P-004", Staff: "O-001"},
			{Kind: BeatBed, Camera: "7A", Patient: "P-004", Frames: bedFrames},
		},
		// 4. El reporte (lo abre la vista).
		{
			{Kind: BeatReport},
		},
	}
}

// Shift devuelve el turno que la Demo está manejando (la vista lo dibuja).
func (d *Demo) Shift() *Shift { return d.shift }

// Part devuelve la parte que se muestra (1 a 8).
func (d *Demo) Part() int { return d.part + 1 }

// Playing dice si una parte se está animando (el reloj del turno corre).
func (d *Demo) Playing() bool { return d.playing }

// Done dice si ya se llegó a la última parte (el reporte).
func (d *Demo) Done() bool { return d.done }

// Beat devuelve el beat que se está animando (ok es false si no hay uno).
func (d *Demo) Beat() (Beat, bool) {
	if !d.playing {
		return Beat{}, false
	}
	return d.script[d.part][d.beat], true
}

// BeatProgress dice cuánto va del beat actual, de 0 a 1 (según su duración mínima).
func (d *Demo) BeatProgress() float64 {
	b, ok := d.Beat()
	if !ok || b.Frames <= 0 {
		return 0
	}
	return min(float64(d.frames)/float64(b.Frames), 1)
}

// Start empieza a animar la parte siguiente. Falla si ya hay una animándose o
// si no quedan partes. Devuelve el evento del primer beat, si tiene.
func (d *Demo) Start() (DemoEvent, error) {
	if d.playing {
		return DemoEvent{}, errors.New("demo: a part is already playing")
	}
	if d.part+1 >= DemoParts {
		return DemoEvent{}, errors.New("demo: no more parts")
	}
	d.part++
	d.beat, d.frames = 0, 0
	d.playing = true
	return d.enter()
}

// Tick avanza la animación un tick de la Demo. Sin ninguna parte animándose
// NO HACE NADA: el reloj del turno queda quieto. Es el único lugar donde la
// Demo hace correr el reloj.
func (d *Demo) Tick() (DemoEvent, error) {
	if !d.playing {
		return DemoEvent{}, nil
	}
	b := d.script[d.part][d.beat]
	if d.shift.Finished() {
		return DemoEvent{}, errors.New("demo: the shift ended during the demo")
	}
	speed := 1
	if b.Kind == BeatCarry {
		speed = carrySpeed
	}
	for i := 0; i < speed; i++ {
		d.shift.Tick()
	}
	d.frames++
	if d.frames > maxBeatFrames {
		return DemoEvent{}, fmt.Errorf("demo: part %d, beat %d never finished", d.Part(), d.beat)
	}
	if !d.beatOver(b) {
		return DemoEvent{}, nil
	}
	d.beat++
	d.frames = 0
	if d.beat >= len(d.script[d.part]) {
		d.playing = false // la parte terminó: todo queda quieto
		return DemoEvent{}, nil
	}
	return d.enter()
}

// Finish corre lo que falte de la parte actual de golpe (para saltar la
// animación). Devuelve el último evento que haya pasado en el camino.
func (d *Demo) Finish() (DemoEvent, error) {
	var last DemoEvent
	for d.playing {
		ev, err := d.Tick()
		if err != nil {
			return last, err
		}
		if ev.Kind != EventNone {
			last = ev
		}
	}
	return last, nil
}

// beatOver dice si el beat actual ya terminó: pasó su tiempo mínimo y, en la
// camilla, el paciente ya llegó a la cama.
func (d *Demo) beatOver(b Beat) bool {
	if d.frames < b.Frames {
		return false
	}
	if b.Kind == BeatCarry {
		return d.shift.FindPatient(b.Patient).State() != Carried
	}
	return true
}

// enter hace lo que le toca a un beat al empezar: cambia la cámara y ejecuta
// la acción de juego. Es el único lugar donde la Demo llama a Assign,
// MoveToBed, Discharge y a las ayudas del guion.
func (d *Demo) enter() (DemoEvent, error) {
	b := d.script[d.part][d.beat]
	s := d.shift
	if b.Camera != "" {
		if err := s.SwitchCamera(b.Camera); err != nil {
			return DemoEvent{}, err
		}
	}
	switch b.Kind {
	case BeatMingle:
		return DemoEvent{}, s.ForceMove(b.Patient, b.Camera)
	case BeatSleep:
		return DemoEvent{}, s.ForceSleep(b.Patient, s.CurrentCamera(), b.Place)
	case BeatCarry:
		if s.FindPatient(b.Patient).State() == NoBedWaiting {
			return DemoEvent{}, s.MoveToBed(b.Patient, b.Staff)
		}
		return DemoEvent{}, s.Assign(b.Patient, b.Staff)
	case BeatNoBed:
		// El error del modelo "no hay cama" es lo ESPERADO en este beat.
		err := s.Assign(b.Patient, b.Staff)
		if !errors.Is(err, hospital.ErrNoRoomAvailable) {
			return DemoEvent{}, fmt.Errorf("demo: expected no free room, got %v", err)
		}
		return DemoEvent{EventNoBed, b.Patient}, nil
	case BeatWake:
		return DemoEvent{}, s.ForceWake(b.Patient)
	case BeatDischarge:
		return DemoEvent{}, s.Discharge(b.Patient)
	case BeatReport:
		d.playing = false
		d.done = true
	}
	return DemoEvent{}, nil
}

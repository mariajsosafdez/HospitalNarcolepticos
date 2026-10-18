package game

import (
	"cmp"
	"errors"
	"fmt"
	"math/rand"
	"slices"

	"costenos-narcolepsia/hospital"
)

const (
	hospitalName = "Hospital de los Costeños con Narcolepsia"

	// arrivalWindowTicks: los pacientes llegan repartidos entre las 12:00 y las 16:00.
	arrivalWindowTicks = 4 * 60 * TicksPerMinute

	maxMessages = 100

	// neverTick es un tick que nunca llega (la Demo no despierta sola a nadie).
	neverTick = 1 << 60
)

// noBedEvent recuerda un paciente que se quedó sin cama y el lugar donde
// estaba dormido en ese momento (el reporte 5.1 lo muestra aunque después
// haya pasado a una cama).
type noBedEvent struct {
	patient  *GamePatient
	location string
}

// Message es una línea del registro de eventos que la vista puede mostrar.
type Message struct {
	Tick int    `json:"tick"`
	Time string `json:"time"`
	Text string `json:"text"`
}

// Shift es UN turno de juego (un día): une el reloj, el hospital, los
// pacientes, el personal y los puntos. Recibe las acciones del jugador y
// avanza el tiempo con Tick. Toda su lógica corre en una sola goroutine (la
// del bucle del juego), así que no necesita candados propios.
type Shift struct {
	cfg   DayConfig
	seed  int64
	rng   *rand.Rand // azar con semilla: misma semilla, misma partida
	clock Clock
	world *World
	h     *hospital.Hospital // un hospital NUEVO por día

	pending  []*GamePatient // todavía no han llegado
	patients []*GamePatient // ya admitidos en el hospital
	staff    []*StaffMember

	camera       CameraID // cámara que mira el jugador
	score        ScoreBreakdown
	noBedCount   int            // veces que un paciente se quedó sin cama
	noBedEvents  []noBedEvent   // quién se quedó sin cama y dónde (para el reporte)
	overworked   []string       // IDs de doctores con exceso de trabajo
	episodeTimes map[string]int // ID de episodio -> tick de juego en que ocurrió
	messages     []Message
	finished     bool
}

// NewShift arma un turno con la configuración del día y una semilla.
func NewShift(cfg DayConfig, seed int64) (*Shift, error) {
	s := &Shift{
		cfg:          cfg,
		seed:         seed,
		rng:          rand.New(rand.NewSource(seed)),
		world:        NewWorld(cfg.Cameras, cfg.Rooms),
		h:            hospital.NewHospital(hospitalName),
		camera:       Reception,
		episodeTimes: make(map[string]int),
	}
	now := func() int { return s.clock.Tick() }

	// El personal se contrata ENVUELTO en StaffMember (ver staff.go).
	for _, d := range cfg.Doctors {
		doctor, err := hospital.NewDoctor(d.ID, d.Name, d.Age, d.Specialty)
		if err != nil {
			return nil, err
		}
		member := NewDoctorMember(doctor, now)
		if err := s.h.HireStaff(member); err != nil {
			return nil, err
		}
		s.staff = append(s.staff, member)
	}
	for _, o := range cfg.Orderlies {
		orderly, err := hospital.NewOrderly(o.ID, o.Name, o.Age)
		if err != nil {
			return nil, err
		}
		member := NewOrderlyMember(orderly, now)
		if err := s.h.HireStaff(member); err != nil {
			return nil, err
		}
		s.staff = append(s.staff, member)
	}
	for _, number := range cfg.Rooms {
		room, err := hospital.NewRoom(number, 1)
		if err != nil {
			return nil, err
		}
		if err := s.h.AddRoom(room); err != nil {
			return nil, err
		}
	}

	// Los pacientes esperan en "pending" hasta su hora de llegada: se barajan
	// y se reparten entre las 12:00 y las 16:00. En la Demo llegan todos en
	// el tick 0 y en orden de ID, sin azar.
	interval := arrivalWindowTicks
	if len(cfg.Patients) > 0 {
		interval = arrivalWindowTicks / len(cfg.Patients)
	}
	for slot, index := range s.rng.Perm(len(cfg.Patients)) {
		data := cfg.Patients[index]
		patient, err := hospital.NewPatient(data.ID, data.Name, data.Age, data.Level)
		if err != nil {
			return nil, err
		}
		arrival := slot*interval + s.rng.Intn(interval)
		if cfg.Scripted {
			arrival = 0
		}
		s.pending = append(s.pending, &GamePatient{patient: patient, arrivalTick: arrival})
	}
	if cfg.Scripted {
		slices.SortFunc(s.pending, func(a, b *GamePatient) int {
			return cmp.Compare(a.patient.ID(), b.patient.ID())
		})
	}
	s.admitArrivals(0)
	return s, nil
}

// ---------------------------------------------------------------------------
// Lectura del estado (para la vista y los tests)
// ---------------------------------------------------------------------------

func (s *Shift) Day() int                     { return s.cfg.Day }
func (s *Shift) Seed() int64                  { return s.seed }
func (s *Shift) Clock() *Clock                { return &s.clock }
func (s *Shift) World() *World                { return s.world }
func (s *Shift) Hospital() *hospital.Hospital { return s.h }
func (s *Shift) Patients() []*GamePatient     { return slices.Clone(s.patients) }
func (s *Shift) Staff() []*StaffMember        { return slices.Clone(s.staff) }
func (s *Shift) CurrentCamera() CameraID      { return s.camera }
func (s *Shift) Score() ScoreBreakdown        { return s.score }
func (s *Shift) Finished() bool               { return s.finished }
func (s *Shift) NoBedCount() int              { return s.noBedCount }
func (s *Shift) Overworked() []string         { return slices.Clone(s.overworked) }
func (s *Shift) Messages() []Message          { return slices.Clone(s.messages) }

// Stars calcula las estrellas con los puntos de este momento.
func (s *Shift) Stars() int { return Stars(s.score.Total(), s.score.Episodes) }

// FindPatient busca un paciente ya admitido por su ID.
func (s *Shift) FindPatient(id string) *GamePatient {
	for _, g := range s.patients {
		if g.patient.ID() == id {
			return g
		}
	}
	return nil
}

// FindStaff busca un miembro del personal por su ID.
func (s *Shift) FindStaff(id string) *StaffMember {
	for _, m := range s.staff {
		if m.ID() == id {
			return m
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Avance del tiempo
// ---------------------------------------------------------------------------

// Tick avanza el juego un tick. Si el reloj está en pausa o el turno ya
// terminó, no hace nada. Se llama una vez por cada Update de la ventana.
func (s *Shift) Tick() {
	if s.finished || s.clock.Paused() {
		return
	}
	s.clock.Advance()
	now := s.clock.Tick()
	s.admitArrivals(now)
	for _, g := range s.patients {
		s.updatePatient(g, now)
	}
	if s.clock.Over() {
		s.finish()
	}
}

// admitArrivals admite en el hospital a los pacientes cuya hora ya llegó.
func (s *Shift) admitArrivals(now int) {
	remaining := s.pending[:0]
	for _, g := range s.pending {
		if g.arrivalTick > now {
			remaining = append(remaining, g)
			continue
		}
		if err := s.h.AdmitPatient(g.patient); err != nil {
			// Solo puede fallar por un ID repetido en la configuración.
			s.log("No se pudo admitir a %s: %v", g.patient.Name(), err)
			continue
		}
		g.camera = Reception
		g.state = Arriving
		g.nextMoveTick = now + MinutesToTicks(arrivingMinutes)
		g.nextSleepCheck = now + MinutesToTicks(sleepCheckMinutes) + s.rng.Intn(MinutesToTicks(sleepCheckMinutes))
		if s.cfg.Scripted {
			g.state = Wandering
		}
		s.patients = append(s.patients, g)
		s.log("%s llegó al hospital.", g.patient.Name())
	}
	s.pending = remaining
}

// updatePatient hace lo que le toque a un paciente en este tick.
func (s *Shift) updatePatient(g *GamePatient, now int) {
	switch g.state {
	case Arriving:
		if now >= g.nextMoveTick {
			g.state = Wandering
			g.nextMoveTick = now + nextMoveDelay(s.rng)
		}
	case Wandering:
		if s.cfg.Scripted {
			return // la Demo no tiene azar
		}
		s.moveIfDue(g, now)
		if g.state == Wandering { // moveIfDue pudo mandarlo a Gone
			s.sleepIfDue(g, now)
		}
	case Carried:
		if now >= g.arrivesAtBed {
			s.arriveInBed(g, now)
		}
	case InBed:
		if now >= g.wakeTick {
			g.state = AwakeInBed
			g.awakeSince = now
			g.bedPenalties = 0
			s.log("%s despertó en la habitación %d.", g.patient.Name(), g.patient.Room().Number())
		}
	case AwakeInBed:
		s.chargeBedPenalty(g, now)
	}
}

// moveIfDue mueve al paciente a otra cámara cuando le toca (ticks aleatorios).
func (s *Shift) moveIfDue(g *GamePatient, now int) {
	if now < g.nextMoveTick {
		return
	}
	g.nextMoveTick = now + nextMoveDelay(s.rng)
	if g.leaving {
		if g.camera == Reception {
			g.state = Gone
			s.log("%s se fue del hospital.", g.patient.Name())
			return
		}
		g.camera = s.world.StepToward(g.camera, Reception)
	} else {
		g.camera = s.world.RandomMove(g.camera, s.rng)
	}
	g.spot = s.rng.Intn(max(cameras[g.camera].Spots, 1))
}

// sleepIfDue tira el dado del sueño cada sleepCheckMinutes.
func (s *Shift) sleepIfDue(g *GamePatient, now int) {
	if now < g.nextSleepCheck {
		return
	}
	g.nextSleepCheck = now + MinutesToTicks(sleepCheckMinutes)
	if g.leaving || now < g.noSleepUntil {
		return
	}
	if s.rng.Float64() < SleepChance(g.patient.Level(), g.camera) {
		s.fallAsleep(g, cameras[g.camera].Name, now)
	}
}

// fallAsleep deja al paciente dormido en su cámara. El modelo NO se entera:
// sigue viéndolo Awake hasta que el jugador asigne personal (Assign).
func (s *Shift) fallAsleep(g *GamePatient, location string, now int) {
	g.state = AsleepWaiting
	g.sleepLocation = location
	g.sleepTick = now
	g.sleeps++
	s.log("%s se durmió en %s.", g.patient.Name(), location)
}

// arriveInBed: la camilla llegó a la habitación. Aquí termina la espera y se
// sortea cuánto va a dormir.
func (s *Shift) arriveInBed(g *GamePatient, now int) {
	g.state = InBed
	if room := g.patient.Room(); room != nil {
		if cam, ok := RoomCamera(room.Number()); ok {
			g.camera = cam
		}
	}
	g.waitMinutes = TicksToMinutes(now - g.sleepTick)
	s.score.addWait(g.waitMinutes)
	if s.cfg.Scripted {
		g.wakeTick = neverTick
	} else {
		g.wakeTick = now + MinutesToTicks(sleepMinutes(g.patient.Level(), s.rng))
	}
	s.log("%s llegó a la cama tras esperar %d min.", g.patient.Name(), g.waitMinutes)
}

// chargeBedPenalty cobra −2 (x multiplicador del día) por cada 10 min que un
// paciente despierto sigue ocupando la cama, después de 15 min de gracia.
func (s *Shift) chargeBedPenalty(g *GamePatient, now int) {
	elapsed := TicksToMinutes(now - g.awakeSince)
	due := 0
	if elapsed > bedGraceMinutes {
		due = (elapsed - bedGraceMinutes) / bedPenaltyEvery
	}
	for g.bedPenalties < due {
		s.score.BedPenalty -= ScalePenalty(bedPenalty, s.cfg.Day)
		g.bedPenalties++
	}
}

// finish congela el turno. Decisión de la autora: los que siguen dormidos
// sin cama RESTAN puntos ("realmente no se atendieron"). Los que van en la
// camilla sí fueron atendidos: su espera se califica como si llegaran ya.
func (s *Shift) finish() {
	if s.finished {
		return
	}
	s.finished = true
	now := s.clock.Tick()
	for _, g := range s.patients {
		switch g.state {
		case AsleepWaiting, NoBedWaiting:
			s.score.Unattended -= ScalePenalty(unattendedPenalty, s.cfg.Day)
			s.score.Episodes++
		case Carried:
			s.score.addWait(TicksToMinutes(now - g.sleepTick))
		}
	}
	s.log("Terminó el turno: %d puntos, %d estrellas.", s.score.Total(), s.Stars())
}

// ---------------------------------------------------------------------------
// Acciones del jugador (sección 5)
// ---------------------------------------------------------------------------

// Pause y Resume congelan y reanudan el turno.
func (s *Shift) Pause()  { s.clock.Pause() }
func (s *Shift) Resume() { s.clock.Resume() }

// SwitchCamera cambia la cámara que mira el jugador.
func (s *Shift) SwitchCamera(id CameraID) error {
	if !s.world.Has(id) {
		return fmt.Errorf("camera %q: %w", id, ErrUnknownCamera)
	}
	s.camera = id
	return nil
}

// lookup valida que el paciente y el personal existan.
func (s *Shift) lookup(patientID, staffID string) (*GamePatient, *StaffMember, error) {
	if s.finished {
		return nil, nil, ErrShiftOver
	}
	g := s.FindPatient(patientID)
	if g == nil {
		return nil, nil, fmt.Errorf("patient %q: %w", patientID, ErrUnknownPatient)
	}
	m := s.FindStaff(staffID)
	if m == nil {
		return nil, nil, fmt.Errorf("staff %q: %w", staffID, ErrUnknownStaff)
	}
	return g, m, nil
}

// checkFree verifica que el jugador pueda elegir a esta persona ahora.
func checkFree(m *StaffMember) error {
	switch m.State() {
	case StaffTransferring, StaffResting:
		return fmt.Errorf("%s: %w", m.Name(), ErrStaffBusy)
	case StaffNoQuota:
		return fmt.Errorf("%s: %w", m.Name(), ErrStaffNoQuota)
	}
	return nil
}

// Assign manda al personal elegido a atender a un paciente dormido (§4.2).
//
// Si no hay cama devuelve un error que cumple errors.Is(err,
// hospital.ErrNoRoomAvailable): el paciente queda en NoBedWaiting, pero el
// personal SÍ lo atendió y entra en cooldown.
func (s *Shift) Assign(patientID, staffID string) error {
	g, m, err := s.lookup(patientID, staffID)
	if err != nil {
		return err
	}
	if g.state != AsleepWaiting {
		return fmt.Errorf("assign %s (%v): %w", patientID, g.state, ErrWrongState)
	}
	if err := checkFree(m); err != nil {
		return err
	}

	// El truco: solo el elegido se ve disponible mientras dura RegisterEpisode,
	// así que el round-robin del hospital lo elige a él.
	m.selected = true
	regErr := s.h.RegisterEpisode(g.patient, g.sleepLocation)
	m.selected = false

	if errors.Is(regErr, hospital.ErrNoStaffAvailable) {
		// No debería pasar porque se validó antes: sería un bug.
		return fmt.Errorf("assign %s: %w", patientID, errors.Join(ErrUnexpectedDispatch, regErr))
	}
	if regErr != nil && !errors.Is(regErr, hospital.ErrNoRoomAvailable) {
		return fmt.Errorf("assign %s: %w", patientID, regErr)
	}

	// El episodio quedó registrado: se guarda su hora de juego y se confirma
	// que el hospital despachó a quien eligió el jugador.
	history := s.h.History()
	record := history[len(history)-1]
	if record.Attender() != m.inner {
		return fmt.Errorf("assign %s: %w", patientID, ErrUnexpectedDispatch)
	}
	s.episodeTimes[record.ID()] = s.clock.Tick()
	m.startCooldown(false)
	if m.IsDoctor() && g.patient.Level() == hospital.Severe {
		s.score.SevereBonus += severeDoctorBonus
	}
	s.checkOverwork(m)

	if regErr != nil {
		g.state = NoBedWaiting
		s.noBedCount++
		s.noBedEvents = append(s.noBedEvents, noBedEvent{g, g.sleepLocation})
		s.log("%s atendió a %s pero no hay cama: se queda en %s.", m.Name(), g.patient.Name(), g.sleepLocation)
		return fmt.Errorf("assign %s: %w", patientID, regErr)
	}
	g.state = Carried
	g.arrivesAtBed = m.transferUntil
	g.carrier = m
	s.log("%s lleva a %s a la habitación %d.", m.Name(), g.patient.Name(), g.patient.Room().Number())
	return nil
}

// MoveToBed manda a cama (siempre a mano) a un paciente que se quedó sin cama
// cuando ya hay una libre. No crea un episodio nuevo: es el mismo episodio.
func (s *Shift) MoveToBed(patientID, staffID string) error {
	g, m, err := s.lookup(patientID, staffID)
	if err != nil {
		return err
	}
	if g.state != NoBedWaiting {
		return fmt.Errorf("move to bed %s (%v): %w", patientID, g.state, ErrWrongState)
	}
	if err := checkFree(m); err != nil {
		return err
	}
	room, err := s.h.AssignRoom(g.patient)
	if err != nil {
		return fmt.Errorf("move to bed %s: %w", patientID, err)
	}
	m.startCooldown(true) // cooldown corto: descanso a la mitad
	g.state = Carried
	g.arrivesAtBed = m.transferUntil
	g.carrier = m
	s.log("%s lleva a %s a la habitación %d.", m.Name(), g.patient.Name(), room.Number())
	return nil
}

// Discharge da de alta a un paciente despierto en su cama y libera la cama.
// Después el paciente se queda o se va del hospital.
func (s *Shift) Discharge(patientID string) error {
	if s.finished {
		return ErrShiftOver
	}
	g := s.FindPatient(patientID)
	if g == nil {
		return fmt.Errorf("patient %q: %w", patientID, ErrUnknownPatient)
	}
	if g.state != AwakeInBed {
		return fmt.Errorf("discharge %s (%v): %w", patientID, g.state, ErrWrongState)
	}
	if err := s.h.WakeUpPatient(g.patient); err != nil {
		return fmt.Errorf("discharge %s: %w", patientID, err)
	}
	now := s.clock.Tick()
	g.state = Wandering
	g.camera = s.exitCamera(g.camera)
	g.spot = 0
	g.leaving = s.rng.Float64() < leaveChance(g.patient.Level())
	g.noSleepUntil = now + MinutesToTicks(noSleepAfterMinute)
	g.nextMoveTick = now + nextMoveDelay(s.rng)
	g.nextSleepCheck = g.noSleepUntil
	if g.leaving {
		s.log("%s recibió el alta y se va del hospital.", g.patient.Name())
	} else {
		s.log("%s recibió el alta y se queda un rato.", g.patient.Name())
	}
	return nil
}

// exitCamera dice por dónde sale un paciente de su habitación: el pasillo
// de ese lado, o la sala de espera si ese pasillo aún está bloqueado.
func (s *Shift) exitCamera(room CameraID) CameraID {
	hallway := CameraID("4") // 7A-7C salen al Pasillo Oeste
	if room == "7D" || room == "7E" || room == "7F" {
		hallway = "3" // 7D-7F salen al Pasillo Este
	}
	if slices.Contains(s.world.common, hallway) {
		return hallway
	}
	return "1B"
}

// checkOverwork cobra el exceso de trabajo UNA sola vez por doctor, cuando
// llega a su cupo de 4 pacientes.
func (s *Shift) checkOverwork(m *StaffMember) {
	if m.doctor == nil || m.overworkCharged || m.PatientCount() < DoctorQuota {
		return
	}
	m.overworkCharged = true
	s.overworked = append(s.overworked, m.ID())
	s.score.OverworkLost -= ScalePenalty(overworkPenalty, s.cfg.Day)
	s.log("%s llegó a %d/%d pacientes: exceso de trabajo.", m.Name(), DoctorQuota, DoctorQuota)
}

// ---------------------------------------------------------------------------
// Ayudas del guion de la Demo (no son acciones del jugador)
// ---------------------------------------------------------------------------

// ForceSleep duerme a un paciente en una cámara y lugar concretos. Lo usa la
// Demo (y los tests) para reemplazar el azar por un guion.
func (s *Shift) ForceSleep(patientID string, cam CameraID, location string) error {
	g := s.FindPatient(patientID)
	if g == nil {
		return fmt.Errorf("patient %q: %w", patientID, ErrUnknownPatient)
	}
	if g.state != Wandering && g.state != Arriving {
		return fmt.Errorf("force sleep %s (%v): %w", patientID, g.state, ErrWrongState)
	}
	if !s.world.Has(cam) {
		return fmt.Errorf("camera %q: %w", cam, ErrUnknownCamera)
	}
	g.camera = cam
	s.fallAsleep(g, location, s.clock.Tick())
	return nil
}

// ForceWake despierta al paciente que duerme en su cama en el próximo Tick.
func (s *Shift) ForceWake(patientID string) error {
	g := s.FindPatient(patientID)
	if g == nil {
		return fmt.Errorf("patient %q: %w", patientID, ErrUnknownPatient)
	}
	if g.state != InBed {
		return fmt.Errorf("force wake %s (%v): %w", patientID, g.state, ErrWrongState)
	}
	g.wakeTick = s.clock.Tick()
	return nil
}

// log agrega un mensaje al registro, que guarda solo los últimos maxMessages.
func (s *Shift) log(format string, args ...any) {
	tick := s.clock.Tick()
	s.messages = append(s.messages, Message{Tick: tick, Time: FormatGameTime(tick), Text: fmt.Sprintf(format, args...)})
	if len(s.messages) > maxMessages {
		s.messages = s.messages[len(s.messages)-maxMessages:]
	}
}

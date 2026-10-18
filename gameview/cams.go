package gameview

import (
	"errors"
	"fmt"
	"math/rand"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"

	"costenos-narcolepsia/game"
	"costenos-narcolepsia/hospital"
)

// Pantalla de cámaras (turno y Demo), según las vistas aprobadas "Cam / zona
// común", "paciente seleccionado", "sin cama", "habitación", "Demo
// controles" y "Panel / estados del personal" (frames 9:205 a 9:674 de Figma).
// Este archivo tiene la lógica de la pantalla (entrada y estado); lo que se
// dibuja está en cams_draw.go.

const (
	noticeTicks  = 4 * game.TicksPerSecond // el aviso se va a los 4 s
	demoAutoTick = 4 * game.TicksPerSecond // Auto: un paso cada 4 s
)

// demoStepActions dice cuántas acciones del guion de game.Demo se ejecutan
// al llegar a cada uno de los 4 pasos que muestra la vista (el paso 1 no
// hace nada: solo presenta el turno; el 2 son los cuatro sueños; el 3 el
// alta y el traslado; el 4 el reporte).
var demoStepActions = [4]int{0, 4, 2, 1}

// demoStepCamera es la cámara que se mira al llegar a cada paso.
var demoStepCamera = [4]game.CameraID{"2", "4", "4", "4"}

// notice es el aviso amable que sale cuando una acción falla.
type notice struct {
	title string
	lines []string
	until int // tick de la interfaz en que se va
}

// camsScene es la pantalla de cámaras.
type camsScene struct {
	env   *Env
	shift *game.Shift
	demo  *game.Demo // nil si es un turno normal

	demoStep int  // 1 a 4 (0 = no es la Demo)
	demoAuto bool // modo Auto
	autoTick int  // ticks desde el último paso automático

	selected string // ID del paciente elegido ("" = ninguno)
	cursor   int    // con teclado: tarjeta del panel marcada (-1 = ninguna)
	pauseSel int    // en la pausa: 0 = Seguir, 1 = Menú
	notice   *notice
	paused   bool
	toReport *reportScene // reporte listo para abrirse (fin del turno o paso 4 de la Demo)
	msg      *messageBox  // mensaje del Director abierto (el reloj del turno se detiene)
	practice bool         // "Cómo jugar": solo el tutorial, y al terminar vuelve al menú
	recorded bool         // el resultado del turno ya se guardó

	mouseX, mouseY int
	hovered        string
	hoverLocked    game.CameraID // cámara bloqueada bajo el mouse ("" = ninguna)

	walkers map[string]*walker
	rng     *rand.Rand
	feed    *ebiten.Image // la cámara se pinta aquí a 168x123 y luego a x2
}

// newCams arma la pantalla de cámaras. cam es la cámara inicial ("" = la que
// pida la pantalla: la 2 en la Demo, la 1A en un turno normal).
func newCams(env *Env, shift *game.Shift, demo *game.Demo, cam game.CameraID) (*camsScene, error) {
	c := &camsScene{
		env:     env,
		shift:   shift,
		demo:    demo,
		cursor:  -1,
		walkers: make(map[string]*walker),
		rng:     rand.New(rand.NewSource(shift.Seed() + 1)),
		feed:    ebiten.NewImage(feedW, feedH),
	}
	if demo != nil {
		c.demoStep = 1
		if cam == "" {
			cam = demoStepCamera[0]
		}
	}
	if cam != "" {
		if err := shift.SwitchCamera(cam); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *camsScene) Shift() *game.Shift { return c.shift }

// State describe la pantalla para el volcado de pruebas.
func (c *camsScene) State() UIState {
	st := UIState{
		Scene:           "cams",
		Camera:          string(c.shift.CurrentCamera()),
		SelectedPatient: c.selected,
		StaffCursor:     c.cursorID(),
		HoveredButton:   c.hovered,
		DemoStep:        c.demoStep,
		DemoAuto:        c.demoAuto,
		Tutorial:        c.practice,
	}
	if c.demo != nil {
		st.Scene = "demo"
	}
	for _, p := range c.placements() {
		r := p.screenRect()
		st.Boxes = append(st.Boxes, PatientBox{p.patient.Patient().ID(), r.x, r.y, r.w, r.h})
	}
	if c.notice != nil {
		st.Notice = c.notice.title
	}
	if c.msg != nil {
		st.Message, st.MessageIndex, st.MessageCount = c.msg.text(), c.msg.index+1, len(c.msg.lines)
	}
	switch {
	case c.msg != nil:
		st.Overlay = "message"
	case c.paused:
		st.Overlay = "pause"
		st.Selected = [2]string{"resume", "menu"}[c.pauseSel]
	}
	return st
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

// Update avanza el turno un tick y atiende la entrada. Si el turno terminó (o
// la Demo llegó a su paso 4), devuelve la pantalla del reporte.
func (c *camsScene) Update(in *Input) Scene {
	next := c.update(in)
	if next == nil && c.toReport != nil {
		next, c.toReport = c.toReport, nil
	}
	return next
}

func (c *camsScene) update(in *Input) Scene {
	c.mouseX, c.mouseY = in.MouseX, in.MouseY
	c.hovered, c.hoverLocked = c.hoverTarget(in.MouseX, in.MouseY), ""
	if id, ok := mapButtonAt(in.MouseX, in.MouseY); ok && !c.shift.World().Has(id) {
		c.hoverLocked = id
	}

	// Con un mensaje del Director abierto el turno está detenido: solo se
	// atiende el clic para seguir.
	if c.msg != nil {
		return c.updateMessage(in)
	}

	if c.paused {
		return c.updatePause(in)
	}

	if !c.recorded && c.demo == nil && c.shift.Finished() {
		c.finishShift()
	}
	c.shift.Tick()
	c.updateWalkers()
	if c.notice != nil && c.env.Tick >= c.notice.until {
		c.notice = nil
	}
	if c.demoAuto {
		c.autoTick++
		if c.autoTick >= demoAutoTick && c.demoStep < 4 {
			c.autoTick = 0
			c.stepDemo()
		}
	}

	// Un clic mientras hay un aviso solo cierra el aviso.
	if in.Click && c.notice != nil && c.hovered != "next" && c.hovered != "auto" {
		c.notice = nil
		return nil
	}
	return c.handleInput(in)
}

// showMessage abre un mensaje del Director y detiene el reloj del turno.
// start es la línea con la que empieza (0 = la primera).
func (c *camsScene) showMessage(kind messageKind, lines []string, start int) {
	c.msg = newMessage(kind, lines, start)
	c.shift.Pause()
}

// queueIntro abre el mensaje con que empieza el turno: el tutorial en el Día
// 1 (solo la primera vez) y el mensaje del director en los días 2 a 5.
func (c *camsScene) queueIntro() {
	if c.shift.Day() == 1 {
		if !c.env.Save.TutorialSeen {
			c.showMessage(msgTutorial, game.TutorialLines[:], 0)
		}
		return
	}
	if text, ok := game.DayIntro(c.shift.Day()); ok {
		c.showMessage(msgDayIntro, []string{text}, 0)
	}
}

// updateMessage avanza el mensaje con clic, Enter o Espacio. Al terminar el
// último se cierra y el turno sigue (o, en "Cómo jugar", se vuelve al menú).
func (c *camsScene) updateMessage(in *Input) Scene {
	if !in.Click && !in.Pressed("Enter") && !in.Pressed("Space") {
		return nil
	}
	if !c.msg.advance() {
		return nil
	}
	kind := c.msg.kind
	c.msg = nil
	if !c.paused {
		c.shift.Resume()
	}
	if kind == msgTutorial {
		c.env.Save.TutorialSeen = true
		c.env.persist()
		if c.practice {
			return newMenu(c.env)
		}
	}
	return nil
}

// finishShift autoguarda el resultado del turno (desbloquea el día siguiente
// y anota las mejores estrellas) y prepara el reporte.
func (c *camsScene) finishShift() {
	c.recorded = true
	c.env.recordShift(c.shift.Day(), c.shift.Stars())
	c.toReport = newReportScene(c.env, c.shift, false)
}

// handleInput atiende teclas y clics de una pantalla en juego.
func (c *camsScene) handleInput(in *Input) Scene {
	if in.Pressed("Escape") {
		switch {
		case c.selected != "":
			c.selected = ""
		default:
			c.paused = true
			c.pauseSel = 0
			c.shift.Pause()
		}
		return nil
	}
	unlocked := c.shift.World().Unlocked()
	for digit := byte('1'); digit <= '8'; digit++ {
		if in.Pressed(string(digit)) {
			if id, ok := cameraForKey(digit, c.shift.CurrentCamera(), unlocked); ok {
				c.switchCamera(id)
			}
		}
	}
	c.handleKeys(in)
	if !in.Click {
		return nil
	}
	return c.handleClick(in.MouseX, in.MouseY)
}

// handleKeys es el camino del teclado para atender (sin mouse): Izquierda y
// Derecha eligen al paciente dormido, Arriba y Abajo marcan una tarjeta del
// personal y Enter la manda (o da de alta al de la habitación).
func (c *camsScene) handleKeys(in *Input) {
	dir := 0
	switch {
	case in.Pressed("Right"):
		dir = 1
	case in.Pressed("Left"):
		dir = -1
	}
	if dir != 0 && c.demo == nil {
		c.selected = nextPatient(c.sleeperIDs(), c.selected, dir)
		c.cursor = -1
	}
	if c.selected != "" {
		n := len(c.shift.Staff())
		switch {
		case in.Pressed("Down"):
			c.cursor = nextCursor(c.cursor, 1, n)
		case in.Pressed("Up"):
			c.cursor = nextCursor(c.cursor, -1, n)
		}
	}
	if !in.Pressed("Enter") && !in.Pressed("Space") {
		return
	}
	switch {
	case c.selected != "" && c.cursor >= 0:
		if staff := c.shift.Staff(); c.cursor < len(staff) {
			c.assign(staff[c.cursor])
		}
	case c.selected == "":
		if g := c.roomPatient(); g != nil {
			c.discharge(g)
		}
	}
}

// sleeperIDs lista, de izquierda a derecha, los pacientes dormidos que se ven
// en la cámara y se pueden elegir.
func (c *camsScene) sleeperIDs() []string {
	list := c.placements()
	slices.SortStableFunc(list, func(a, b placement) int { return a.x - b.x })
	var ids []string
	for _, p := range list {
		if st := p.patient.State(); st == game.AsleepWaiting || st == game.NoBedWaiting {
			ids = append(ids, p.patient.Patient().ID())
		}
	}
	return ids
}

// cursorID dice qué persona del personal marca el teclado ("" = ninguna).
func (c *camsScene) cursorID() string {
	staff := c.shift.Staff()
	if c.selected == "" || c.cursor < 0 || c.cursor >= len(staff) {
		return ""
	}
	return staff[c.cursor].ID()
}

// nextPatient elige el siguiente (dir = 1) o el anterior (dir = -1) de la
// lista, dando la vuelta. Sin nadie elegido, Derecha toma el primero e
// Izquierda el último. Es pura para poder probarla.
func nextPatient(ids []string, current string, dir int) string {
	if len(ids) == 0 {
		return ""
	}
	i := slices.Index(ids, current)
	switch {
	case i < 0 && dir > 0:
		return ids[0]
	case i < 0:
		return ids[len(ids)-1]
	}
	return ids[((i+dir)%len(ids)+len(ids))%len(ids)]
}

// nextCursor mueve la tarjeta marcada entre las n tarjetas, dando la vuelta.
// Desde "ninguna" (-1), Abajo marca la primera y Arriba la última.
func nextCursor(current, dir, n int) int {
	if n <= 0 {
		return -1
	}
	if current < 0 {
		if dir > 0 {
			return 0
		}
		return n - 1
	}
	return ((current+dir)%n + n) % n
}

// handleClick decide a qué le dio clic el jugador.
func (c *camsScene) handleClick(x, y int) Scene {
	// Controles de la Demo.
	if c.demo != nil {
		switch c.hovered {
		case "next":
			if c.demoStep < 4 {
				c.stepDemo()
			}
			return nil
		case "auto":
			c.demoAuto = !c.demoAuto
			c.autoTick = 0
			return nil
		}
	}
	// Mapa: cambiar de cámara (las bloqueadas no hacen nada).
	if id, ok := mapButtonAt(x, y); ok {
		if c.shift.World().Has(id) {
			c.switchCamera(id)
		}
		return nil
	}
	// Tarjetas del personal: solo sirven con un paciente elegido.
	if c.selected != "" {
		staff := c.shift.Staff()
		for i := range staff {
			if staffCardRect(i, len(staff)).contains(x, y) {
				c.assign(staff[i])
				return nil
			}
		}
	}
	// "Dar de alta" en una habitación.
	if c.hovered == "discharge" {
		if g := c.roomPatient(); g != nil {
			c.discharge(g)
		}
		return nil
	}
	// Los pacientes de la cámara (solo en un turno, no en la Demo).
	if c.demo == nil && feedRect.contains(x, y) {
		if g := c.patientAt(x, y); g != nil {
			c.selected = g.Patient().ID()
			return nil
		}
	}
	// Clic en el vacío cancela.
	c.selected = ""
	return nil
}

// switchCamera cambia de cámara y suelta la selección.
func (c *camsScene) switchCamera(id game.CameraID) {
	if err := c.shift.SwitchCamera(id); err == nil {
		c.selected = ""
		c.cursor = -1
	}
}

// assign manda al personal elegido a atender al paciente elegido.
func (c *camsScene) assign(m *game.StaffMember) {
	g := c.shift.FindPatient(c.selected)
	if g == nil {
		c.selected = ""
		c.cursor = -1
		return
	}
	overworked := len(c.shift.Overworked())
	var err error
	if g.State() == game.NoBedWaiting {
		err = c.shift.MoveToBed(c.selected, m.ID())
	} else {
		err = c.shift.Assign(c.selected, m.ID())
	}
	c.selected = ""
	c.cursor = -1
	if err != nil {
		c.showError(err, g.Patient().Name())
	}
	// Un doctor que llega a 4/4 (con o sin cama para el paciente) recibe el
	// regaño del director.
	if list := c.shift.Overworked(); len(list) > overworked {
		if doc := c.shift.FindStaff(list[len(list)-1]); doc != nil {
			line := game.OverworkLine(displayName(doc.ID(), doc.Name()))
			c.showMessage(msgOverwork, []string{line}, 0)
		}
	}
}

// discharge da de alta al paciente despierto de la habitación.
func (c *camsScene) discharge(g *game.GamePatient) {
	if err := c.shift.Discharge(g.Patient().ID()); err != nil {
		c.showError(err, g.Patient().Name())
	}
}

// showError muestra el aviso amable de un error de game.
func (c *camsScene) showError(err error, patientName string) {
	title, lines := friendlyError(err, patientName)
	c.notice = &notice{title, lines, c.env.Tick + noticeTicks}
}

// roomPatient devuelve el paciente despierto en la cama de la habitación
// que se está mirando (nil si no hay).
func (c *camsScene) roomPatient() *game.GamePatient {
	for _, g := range c.shift.Patients() {
		if g.Camera() == c.shift.CurrentCamera() && g.State() == game.AwakeInBed {
			return g
		}
	}
	return nil
}

// hoverTarget dice qué botón está bajo el mouse (para el resaltado y para el
// volcado de pruebas).
func (c *camsScene) hoverTarget(x, y int) string {
	if c.paused {
		switch {
		case pauseResumeRect.contains(x, y):
			return "resume"
		case pauseMenuRect.contains(x, y):
			return "menu"
		}
		return ""
	}
	if c.demo != nil {
		switch {
		case demoNextRect.contains(x, y):
			return "next"
		case demoAutoRect.contains(x, y):
			return "auto"
		}
	}
	if id, ok := mapButtonAt(x, y); ok {
		return "cam " + string(id)
	}
	if c.roomPatient() != nil && dischargeRect(c.bed()).contains(x, y) {
		return "discharge"
	}
	if c.selected != "" {
		staff := c.shift.Staff()
		for i := range staff {
			if staffCardRect(i, len(staff)).contains(x, y) {
				return "staff " + staff[i].ID()
			}
		}
	}
	if c.demo == nil && feedRect.contains(x, y) {
		if g := c.patientAt(x, y); g != nil {
			return "patient " + g.Patient().ID()
		}
	}
	return ""
}

// updatePause atiende la pausa. Izquierda/Arriba marcan "Seguir" y
// Derecha/Abajo "Menú"; Enter (o clic) ejecuta el marcado. Esc sigue y M sale.
// El mouse marca el botón que toca, igual que las flechas.
func (c *camsScene) updatePause(in *Input) Scene {
	switch {
	case in.Pressed("Left"), in.Pressed("Up"):
		c.pauseSel = 0
	case in.Pressed("Right"), in.Pressed("Down"):
		c.pauseSel = 1
	}
	if in.Moved {
		switch c.hovered {
		case "resume":
			c.pauseSel = 0
		case "menu":
			c.pauseSel = 1
		}
	}
	chosen := -1
	switch {
	case in.Pressed("Escape"):
		chosen = 0
	case in.Pressed("M"):
		chosen = 1
	case in.Pressed("Enter"), in.Pressed("Space"):
		chosen = c.pauseSel
	case in.Click && c.hovered == "resume":
		chosen = 0
	case in.Click && c.hovered == "menu":
		chosen = 1
	}
	switch chosen {
	case 0:
		c.paused = false
		c.shift.Resume()
	case 1:
		return newMenu(c.env)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Demo
// ---------------------------------------------------------------------------

// stepDemo pasa al siguiente de los 4 pasos de la Demo.
func (c *camsScene) stepDemo() {
	if err := c.nextDemoStep(); err != nil {
		c.notice = &notice{"Algo salió mal", []string{err.Error()}, c.env.Tick + noticeTicks}
	}
}

// nextDemoStep ejecuta las acciones del siguiente paso y, si alguien se
// quedó sin cama, muestra el aviso de la vista "Cam / sin cama".
func (c *camsScene) nextDemoStep() error {
	if c.demo == nil || c.demoStep >= 4 {
		return errors.New("the demo has no more steps")
	}
	c.demoStep++
	for i := 0; i < demoStepActions[c.demoStep-1]; i++ {
		if _, err := c.demo.Next(); err != nil {
			return fmt.Errorf("demo step %d: %w", c.demoStep, err)
		}
	}
	c.switchCamera(demoStepCamera[c.demoStep-1])
	c.notice = nil
	for _, g := range c.shift.Patients() {
		if g.State() == game.NoBedWaiting {
			c.showError(hospital.ErrNoRoomAvailable, g.Patient().Name())
			break
		}
	}
	if c.demoStep == 4 {
		c.toReport = newReportScene(c.env, c.shift, true)
	}
	return nil
}

// patientAt devuelve el paciente dormido (y elegible) bajo el punto de la
// pantalla; si hay varios, el de más abajo, que se dibuja encima.
func (c *camsScene) patientAt(x, y int) *game.GamePatient {
	var best *game.GamePatient
	bestFoot := -1.0
	for _, p := range c.placements() {
		st := p.patient.State()
		if st != game.AsleepWaiting && st != game.NoBedWaiting {
			continue
		}
		if p.screenRect().contains(x, y) && p.footY >= bestFoot {
			best, bestFoot = p.patient, p.footY
		}
	}
	return best
}

// sortPlacements ordena por la y de los pies: el de más abajo, encima.
func sortPlacements(list []placement) {
	slices.SortStableFunc(list, func(a, b placement) int {
		switch {
		case a.footY < b.footY:
			return -1
		case a.footY > b.footY:
			return 1
		}
		return 0
	})
}

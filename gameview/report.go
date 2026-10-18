package gameview

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"costenos-narcolepsia/game"
)

// Reporte de fin de turno, según la vista aprobada "Reporte / fin de turno"
// (frame 18:439 de Figma, nota 18:534; copia en preview/turno/). Muestra
// estrellas, puntos, las 4 consultas del enunciado en sus cajas y la frase del
// Director. Medidas leídas de la vista (posiciones de Figma / 4).

// Rectángulos de la pantalla (a 1x sobre 480x270).
var (
	reportPointsCard = rect{8, 42, 158, 166}
	reportNoBedCard  = rect{172, 42, 148, 80}  // consulta 5.1
	reportDoctorCard = rect{326, 42, 148, 80}  // consulta 5.2
	reportBedsCard   = rect{172, 128, 148, 80} // consulta 5.3
	reportSevereCard = rect{326, 128, 148, 80} // consulta 5.4
	reportDirector   = rect{8, 214, 358, 52}
	reportNextRect   = rect{374, 216, 98, 22}
	reportMenuRect   = rect{374, 244, 98, 22}
)

// reportCardLines es cuántas líneas de 11 px caben en una caja de 80 px de
// alto después del título (la primera línea va 20 px bajo el borde de arriba).
const reportCardLines = 5

// ---------------------------------------------------------------------------
// Funciones puras: de game.Report a las líneas de cada caja (con test)
// ---------------------------------------------------------------------------

// rline es una línea de una caja. left va a la izquierda y right alineado a
// la derecha; dim las pinta en gris; indent las mete 7 px; marker pone el
// cuadrito rojo de los severos.
type rline struct {
	left, right string
	dim         bool
	indent      bool
	marker      bool
}

// countText arma "1 paciente" o "3 pacientes".
func countText(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", one)
	}
	return fmt.Sprintf("%d %s", n, many)
}

// signedPoints escribe puntos con signo: "+200", "-4" y "0".
func signedPoints(n int) string {
	switch {
	case n > 0:
		return fmt.Sprintf("+%d", n)
	case n < 0:
		return fmt.Sprintf("-%d", -n)
	}
	return "0"
}

// attendedText es "4 pacientes atendidos" (o "1 paciente atendido").
func attendedText(n int) string {
	if n == 1 {
		return "1 paciente atendido"
	}
	return fmt.Sprintf("%d pacientes atendidos", n)
}

// reportSubtitle es la línea bajo el título: "Día 1  ·  de 12 a 6  ·  4 pacientes
// atendidos". La Demo no es un día del juego y dice "Demo".
func reportSubtitle(day int, demo bool, attended int) string {
	who := fmt.Sprintf("Día %d", day)
	if demo {
		who = "Demo"
	}
	return fmt.Sprintf("%s  ·  de 12 a 6  ·  %s", who, attendedText(attended))
}

// capWithMore deja a lo sumo max líneas; si sobran, la última se cambia por
// "y N más" (N = las que no se ven, contando la que se cambió).
func capWithMore(lines []rline, max int) []rline {
	if len(lines) <= max || max < 1 {
		return lines
	}
	out := append([]rline(nil), lines[:max-1]...)
	return append(out, rline{left: fmt.Sprintf("y %d más", len(lines)-(max-1)), dim: true})
}

// noBedLines es la caja "Atendidos sin cama" (consulta 5.1): cuántos
// siguen así al cerrar el turno, cuántas veces pasó, y quiénes (con la cama
// a la que pasaron después, si la consiguieron).
func noBedLines(r game.Report) []rline {
	closing := "Al cerrar el turno: nadie."
	if n := len(r.Hallway); n > 0 {
		closing = fmt.Sprintf("Al cerrar el turno: %s.", countText(n, "paciente", "pacientes"))
	}
	times := fmt.Sprintf("Durante el turno: %s.", countText(len(r.NoBed), "vez", "veces"))
	lines := []rline{{left: closing}, {left: times}}
	for _, p := range r.NoBed {
		lines = append(lines, rline{left: p.Name + ", " + locationText(p.Location)})
		if p.LaterRoom > 0 {
			lines = append(lines, rline{left: fmt.Sprintf("(después pasó a la %d)", p.LaterRoom)})
		}
	}
	return capWithMore(lines, reportCardLines)
}

// demoLocations traduce los cuatro lugares en inglés del escenario del
// enunciado (la Demo, que no se puede tocar) al nombre en español de la
// cámara donde ocurren. Solo es de la vista: el dato del modelo no cambia.
// En un turno normal el lugar ya es el nombre de la cámara en español.
var demoLocations = map[string]string{
	"cafeteria (sancocho line)": "Cafetería",
	"champeta dance hall":       "Salón de champeta",
	"radiology queue":           "Fila de radiología",
	"hallway 2, second floor":   "Pasillo Oeste",
}

// locationText devuelve el lugar para mostrar en el reporte.
func locationText(location string) string {
	if es, ok := demoLocations[location]; ok {
		return es
	}
	return location
}

// firstName es la primera palabra del nombre ("Wilfrido Berrio" -> "Wilfrido").
func firstName(full string) string {
	if words := strings.Fields(full); len(words) > 0 {
		return words[0]
	}
	return full
}

// doctorLines es la caja "Episodios por doctor" (consulta 5.2): el nombre y
// la cantidad, y debajo los pacientes. Un doctor sin episodios ocupa 1 sola
// línea para que quepan los 3.
func doctorLines(r game.Report) []rline {
	var lines []rline
	for _, d := range r.Doctors {
		lines = append(lines, rline{left: displayName(d.ID, d.Name), right: fmt.Sprint(len(d.Episodes))})
		if len(d.Episodes) > 0 {
			names := make([]string, len(d.Episodes))
			for i, e := range d.Episodes {
				names[i] = firstName(e.PatientName)
			}
			lines = append(lines, rline{left: strings.Join(names, ", "), indent: true})
		}
	}
	return capWithMore(lines, reportCardLines)
}

// bedCell es una cama de la caja "Camas": el número y quién la ocupa ("" = libre).
type bedCell struct {
	number int
	who    string
}

// bedCells arma la caja "Camas" (consulta 5.3). Los ocupantes del modelo
// vienen como "P-004 Wilfrido Berrio": se quita el ID.
func bedCells(r game.Report) []bedCell {
	cells := make([]bedCell, len(r.Rooms))
	for i, room := range r.Rooms {
		cells[i].number = room.Number
		if len(room.Occupants) > 0 {
			_, name, _ := strings.Cut(room.Occupants[0], " ")
			cells[i].who = name
		}
	}
	return cells
}

// severeStatus dice dónde está un paciente severo al cerrar el turno: "cama
// 102", "dormida"/"dormido" (todavía sin cama) o "despierta"/"despierto".
func severeStatus(s game.SevereLine) string {
	female := isFemale(s.Name)
	switch {
	case s.Room > 0:
		return fmt.Sprintf("cama %d", s.Room)
	case s.Asleep && female:
		return "dormida"
	case s.Asleep:
		return "dormido"
	case female:
		return "despierta"
	}
	return "despierto"
}

// severeLines es la caja "Pacientes severos" (consulta 5.4) y su pie ("3
// severos, 2 atendidos por doctor"). Hay lugar para 3 filas.
func severeLines(r game.Report) (rows []rline, footer string) {
	attended := 0
	for _, s := range r.Severe {
		rows = append(rows, rline{left: s.Name, right: severeStatus(s), marker: true})
		if s.Episodes > 0 {
			attended++
		}
	}
	footer = fmt.Sprintf("%s, %s por doctor", countText(len(r.Severe), "severo", "severos"), countText(attended, "atendido", "atendidos"))
	return capWithMore(rows, 3), footer
}

// ---------------------------------------------------------------------------
// La pantalla
// ---------------------------------------------------------------------------

// Botones del reporte.
const (
	reportBtnNext = iota
	reportBtnMenu
)

// reportScene es el reporte de fin de turno (de un día o de la Demo).
type reportScene struct {
	env     *Env
	shift   *game.Shift
	report  game.Report
	demo    bool
	line    string // la frase del Director (se escoge al armar el reporte)
	sel     int    // botón marcado
	hovered string
}

// newReportScene arma el reporte del turno. La frase del Director sale de las
// estrellas y de la semilla del turno (la misma semilla, la misma frase).
func newReportScene(env *Env, shift *game.Shift, demo bool) *reportScene {
	r := shift.Report()
	return &reportScene{
		env:    env,
		shift:  shift,
		report: r,
		demo:   demo,
		line:   game.EndLine(r.Stars, shift.Seed()),
	}
}

// canContinue dice si hay botón "Siguiente día": no en la Demo ni en el Día 5.
func (s *reportScene) canContinue() bool { return !s.demo && s.report.Day < game.MaxDay }

// buttons lista los botones que hay, en orden.
func (s *reportScene) buttons() []int {
	if s.canContinue() {
		return []int{reportBtnNext, reportBtnMenu}
	}
	return []int{reportBtnMenu}
}

func (s *reportScene) buttonRect(b int) rect {
	if b == reportBtnNext {
		return reportNextRect
	}
	return reportMenuRect
}

func (s *reportScene) Shift() *game.Shift { return s.shift }

func (s *reportScene) State() UIState {
	st := UIState{Scene: "report", Overlay: "report", HoveredButton: s.hovered, Message: s.line}
	st.Selected = map[int]string{reportBtnNext: "next", reportBtnMenu: "menu"}[s.sel]
	return st
}

// Update atiende flechas, Enter, Esc y mouse. Esc vuelve al menú.
func (s *reportScene) Update(in *Input) Scene {
	btns := s.buttons()
	if len(btns) == 1 {
		s.sel = btns[0]
	}
	s.hovered = ""
	hover := -1
	for _, b := range btns {
		if s.buttonRect(b).contains(in.MouseX, in.MouseY) {
			hover = b
			s.hovered = map[int]string{reportBtnNext: "next", reportBtnMenu: "menu"}[b]
		}
	}
	if len(btns) == 2 {
		switch {
		case in.Pressed("Up"), in.Pressed("Left"):
			s.sel = reportBtnNext
		case in.Pressed("Down"), in.Pressed("Right"):
			s.sel = reportBtnMenu
		}
		if hover >= 0 && in.Moved {
			s.sel = hover
		}
	}
	chosen := -1
	switch {
	case in.Pressed("Escape"):
		chosen = reportBtnMenu
	case in.Pressed("Enter"), in.Pressed("Space"):
		chosen = s.sel
	case in.Click && hover >= 0:
		chosen = hover
	}
	switch {
	case chosen == reportBtnNext && s.canContinue():
		scene, err := newShiftScene(s.env, s.report.Day+1, s.env.newSeed(), true, "")
		if err != nil {
			return newMenu(s.env)
		}
		return scene
	case chosen >= 0:
		return newMenu(s.env)
	}
	return nil
}

// Draw pinta el reporte entero.
func (s *reportScene) Draw(dst *ebiten.Image) {
	e := s.env
	r := s.report
	e.FillRect(dst, rect{0, 0, screenW, screenH}, '1', 1)

	// Título, subtítulo y estrellas.
	e.TextShadow(dst, "Fin del turno", 12, 7, textStyle{fontPixelBold, 18, 'w', alignLeft})
	e.Text(dst, reportSubtitle(r.Day, s.demo, r.Attended), 12, 29, textStyle{fontPixel, 8, 'k', alignLeft})
	for i := 0; i < 5; i++ {
		frame := 1
		if i < r.Stars {
			frame = 0
		}
		e.DrawSpriteScaled(dst, e.Assets.Frame("ui_estrella", 2, frame), 376+20*i, 8, 2)
	}

	s.drawPointsCard(dst)
	s.drawLinesCard(dst, reportNoBedCard, "Atendidos sin cama", noBedLines(r))
	s.drawLinesCard(dst, reportDoctorCard, "Episodios por doctor", doctorLines(r))
	s.drawBedsCard(dst)
	s.drawSevereCard(dst)
	s.drawDirectorCard(dst)

	for _, b := range s.buttons() {
		label, fill := "Menú", byte('k')
		if b == reportBtnNext {
			label, fill = "Siguiente día", 'm'
		}
		rc := s.buttonRect(b)
		e.FillRect(dst, rc, fill, 1)
		if s.sel == b {
			e.FillRect(dst, rc, 'w', 0.25) // el botón marcado (flechas o mouse)
		}
		e.StrokeRect(dst, rc, '0', 1)
		e.Text(dst, label, rc.x+rc.w/2, rc.y+5, textStyle{fontPixel, 10, '0', alignCenter})
	}
}

// drawCard pinta una caja crema con borde 0 y su título.
func (s *reportScene) drawCard(dst *ebiten.Image, r rect, title string) {
	e := s.env
	e.FillRect(dst, r, 'W', 1)
	e.StrokeRect(dst, r, '0', 1)
	padX := 5
	if r == reportPointsCard {
		padX = 6 // la caja de puntos tiene 1 px más de margen
	}
	e.Text(dst, title, r.x+padX, r.y+4, textStyle{fontPixelBold, 10, '2', alignLeft})
}

// lineKey elige el color del texto de una línea.
func lineKey(dim bool) byte {
	if dim {
		return 'x'
	}
	return '1'
}

// drawLinesCard pinta una caja de líneas simples (consultas 5.1 y 5.2).
func (s *reportScene) drawLinesCard(dst *ebiten.Image, r rect, title string, lines []rline) {
	e := s.env
	s.drawCard(dst, r, title)
	measure := func(t string) int { return e.Fonts.Width(t, fontPixel, 8) }
	for i, l := range lines {
		x, y := r.x+5, r.y+19+11*i
		if l.indent {
			x += 7
		}
		maxW := r.w - 10
		if l.right != "" {
			maxW -= 16
			e.Text(dst, l.right, r.x+r.w-7, y, textStyle{fontPixel, 8, '1', alignRight})
		}
		e.Text(dst, fitText(l.left, maxW-(x-r.x-5), measure), x, y, textStyle{fontPixel, 8, lineKey(l.dim), alignLeft})
	}
}

// drawBedsCard pinta "Camas" (consulta 5.3): una columna con hasta 3 camas, o
// dos columnas (solo el primer nombre) cuando hay más.
func (s *reportScene) drawBedsCard(dst *ebiten.Image) {
	e := s.env
	r := reportBedsCard
	s.drawCard(dst, r, "Camas")
	cells := bedCells(s.report)
	measure := func(t string) int { return e.Fonts.Width(t, fontPixel, 8) }
	two := len(cells) > 3
	for i, c := range cells {
		col, row := 0, i
		if two {
			col, row = i/3, i%3
		}
		x, y := r.x+5+74*col, r.y+19+11*row
		e.Text(dst, fmt.Sprint(c.number), x, y, textStyle{fontPixel, 8, '2', alignLeft})
		who, key, w := c.who, byte('1'), r.w-5-22-5
		if who == "" {
			who, key = "libre", 'x'
		}
		if two {
			who, w = firstName(who), 40
		}
		e.Text(dst, fitText(who, w, measure), x+22, y, textStyle{fontPixel, 8, key, alignLeft})
	}
	e.Text(dst, fmt.Sprintf("Veces sin cama: %d", s.report.NoBedCount), r.x+5, r.y+58, textStyle{fontPixel, 8, 'x', alignLeft})
}

// drawSevereCard pinta "Pacientes severos" (consulta 5.4).
func (s *reportScene) drawSevereCard(dst *ebiten.Image) {
	e := s.env
	r := reportSevereCard
	s.drawCard(dst, r, "Pacientes severos")
	rows, footer := severeLines(s.report)
	measure := func(t string) int { return e.Fonts.Width(t, fontPixel, 8) }
	for i, l := range rows {
		y := r.y + 19 + 11*i
		x := r.x + 5
		if l.marker {
			e.FillRect(dst, rect{x, y, 3, 7}, 'R', 1)
			x += 6
		}
		if l.right != "" {
			e.Text(dst, l.right, r.x+102, y, textStyle{fontPixel, 8, 'x', alignLeft})
		}
		e.Text(dst, fitText(l.left, 86, measure), x, y, textStyle{fontPixel, 8, lineKey(l.dim), alignLeft})
	}
	e.Text(dst, footer, r.x+5, r.y+58, textStyle{fontPixel, 8, 'x', alignLeft})
}

// drawPointsCard pinta la caja de puntos con su desglose.
func (s *reportScene) drawPointsCard(dst *ebiten.Image) {
	e := s.env
	r := reportPointsCard
	sc := s.report.Score
	s.drawCard(dst, r, "Puntos")
	e.Text(dst, fmt.Sprint(s.report.Total), r.x+r.w-12, r.y+1, textStyle{fontPixelBold, 18, '1', alignRight})
	e.Text(dst, fmt.Sprintf("de %d posibles (%d %%)", s.report.Possible, s.report.Percent), r.x+6, r.y+19, textStyle{fontPixel, 8, 'x', alignLeft})
	e.FillRect(dst, rect{r.x + 6, r.y + 31, 146, 1}, 'k', 1)

	valueKey := func(n int) byte {
		switch {
		case n > 0:
			return 'g'
		case n < 0:
			return 'R'
		}
		return 'x'
	}
	rows := []struct {
		label        string
		count, value int
	}{
		{"Excelente", sc.Excellent, sc.Excellent * 100},
		{"Bien", sc.Good, sc.Good * 70},
		{"Regular", sc.Regular, sc.Regular * 40},
		{"Mal", sc.Bad, sc.Bad * 10},
		{"Bonus severos", sc.SevereBonus / 20, sc.SevereBonus},
	}
	for i, row := range rows {
		y := r.y + 35 + 11*i
		e.Text(dst, row.label, r.x+6, y, textStyle{fontPixel, 8, '1', alignLeft})
		e.Text(dst, fmt.Sprint(row.count), r.x+98, y, textStyle{fontPixel, 8, '1', alignCenter})
		e.Text(dst, signedPoints(row.value), r.x+128, y, textStyle{fontPixel, 8, 'g', alignLeft}) // los puntos van alineados a la izquierda, siempre en verde
	}
	e.FillRect(dst, rect{r.x + 6, r.y + 91, 146, 1}, 'k', 1)
	e.Text(dst, "Penalizaciones", r.x+6, r.y+95, textStyle{fontPixelBold, 8, '2', alignLeft})
	penalties := []struct {
		label string
		value int
	}{
		{"Cama sin alta", sc.BedPenalty},
		{"Exceso de trabajo", sc.OverworkLost},
		{"Sin atender", sc.Unattended},
	}
	for i, p := range penalties {
		y := r.y + 105 + 11*i
		e.Text(dst, p.label, r.x+6, y, textStyle{fontPixel, 8, '1', alignLeft})
		e.Text(dst, signedPoints(p.value), r.x+148, y, textStyle{fontPixel, 8, valueKey(p.value), alignRight})
	}
	e.FillRect(dst, rect{r.x + 6, r.y + 139, 146, 1}, 'k', 1)
	e.Text(dst, "Total", r.x+6, r.y+145, textStyle{fontPixelBold, 10, '1', alignLeft})
	e.Text(dst, fmt.Sprint(s.report.Total), r.x+148, r.y+145, textStyle{fontPixelBold, 10, '1', alignRight})
}

// drawDirectorCard pinta la frase de Don Ernesto según las estrellas.
func (s *reportScene) drawDirectorCard(dst *ebiten.Image) {
	e := s.env
	r := reportDirector
	e.FillRect(dst, r, 'W', 1)
	e.StrokeRect(dst, r, '0', 1)
	e.DrawSprite(dst, e.Assets.Image("director_retrato"), r.x+4, r.y+2, false)
	e.Text(dst, game.DirectorName, r.x+56, r.y+5, textStyle{fontPixelBold, 8, '2', alignLeft})
	measure := func(t string) int { return e.Fonts.Width(t, fontPixel, 8) }
	for i, line := range wrapMessage(s.line, measure) {
		e.Text(dst, line, r.x+56, r.y+18+11*i, textStyle{fontPixel, 8, '1', alignLeft})
	}
}

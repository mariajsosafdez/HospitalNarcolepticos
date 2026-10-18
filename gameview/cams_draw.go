package gameview

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"costenos-narcolepsia/game"
	"costenos-narcolepsia/hospital"
)

// Dibujo de la pantalla de cámaras. Las capas van de atrás hacia adelante,
// en el orden de la nota de Figma "Cam / zona común" (frame 9:205).

// Rectángulos fijos de la pantalla (a 1x).
var (
	demoBoxRect     = rect{8, 200, 328, 62}
	demoNextRect    = rect{210, 244, 70, 13}
	demoAutoRect    = rect{286, 244, 42, 13}
	pauseBoxRect    = rect{140, 96, 200, 70}
	pauseResumeRect = rect{150, 142, 80, 13}
	pauseMenuRect   = rect{250, 142, 80, 13}
	noticeRect      = rect{60, 203, 220, 46}
)

// dischargeRect es el botón "Dar de alta", centrado bajo la cama (ranura A).
func dischargeRect() rect { return rect{66, 242, 60, 13} }

// backButtonRect es el botón de volver al menú del reporte.
func backButtonRect(demo bool) rect {
	if demo {
		return demoNextRect // en el paso 4 de la Demo el botón cambia de texto
	}
	return rect{feedScreenX + 336/2 - 45, feedScreenY + 246 - 20, 90, 13}
}

// reportRect es el recuadro del reporte: en la Demo deja libre la caja de
// controles de abajo.
func reportRect(demo bool) rect {
	if demo {
		return rect{feedScreenX, feedScreenY, 336, 176}
	}
	return rect{feedScreenX, feedScreenY, 336, 246}
}

// Ranura de cama A de las habitaciones (coordenadas del fondo, sprite 32x30).
const (
	bedSlotX = 30
	bedSlotY = 70
	bedW     = 32
	bedH     = 30
)

// placement es un paciente ya colocado en la cámara que se mira: qué sprite,
// dónde (coordenadas del fondo) y con qué efectos.
type placement struct {
	patient *game.GamePatient
	imgName string // PNG del sprite ("" = sin arte todavía: placeholder)
	frames  int
	index   int
	x, y    int // esquina de arriba a la izquierda, en coordenadas del fondo
	w, h    int
	flip    bool
	footY   float64
	zzz     bool
	alert   bool
	effectX int // esquina del efecto (zzz o !), coordenadas del fondo
	effectY int
}

// screenRect es el rectángulo del sprite en la pantalla (a x2).
func (p placement) screenRect() rect {
	sx, sy := feedToScreen(p.x, p.y)
	return rect{sx, sy, p.w * feedScale, p.h * feedScale}
}

// placements calcula dónde y cómo se dibuja cada paciente de la cámara
// actual. Lo usan el dibujo y el clic del mouse, así coinciden siempre.
func (c *camsScene) placements() []placement {
	a := c.env.Assets
	cam := c.shift.CurrentCamera()
	info, _ := game.CameraByID(cam)

	var out []placement

	for _, g := range c.shift.Patients() {
		if g.Camera() != cam {
			continue
		}
		state := g.State()
		inRoom := state == game.InBed || state == game.AwakeInBed
		if inRoom != (info.Kind == game.RoomCam) && state != game.Carried {
			continue
		}
		w := c.walkers[g.Patient().ID()]
		if w == nil {
			continue
		}

		// En camilla: se dibuja el personal empujándola, no el paciente.
		if state == game.Carried {
			m := g.Carrier()
			if m == nil {
				continue
			}
			out = append(out, c.carryPlacement(g, m, w))
			continue
		}

		spec, ok := patientSprite(state, w.walking)
		if !ok {
			continue
		}
		name := "pac_" + personKey(g.Patient().Name()) + "_" + spec.suffix
		p := placement{patient: g, frames: spec.frames, flip: w.left}
		p.index = animFrame(c.env.Tick, spec.frameTicks, spec.frames, 0)
		if a.Has(name) {
			p.imgName = name
			p.w, p.h = a.FrameSize(name, spec.frames)
		} else {
			p.w, p.h = 16, 24 // placeholder: una figura simple
			if inRoom {
				p.w, p.h = bedW, bedH
			}
		}
		if inRoom {
			p.x, p.y = bedSlotX, bedSlotY
			p.footY = float64(p.y + p.h)
			p.flip = false
			if state == game.InBed {
				p.zzz, p.effectX, p.effectY = true, bedSlotX+12, bedSlotY-12
			} else {
				p.alert, p.effectX, p.effectY = true, bedSlotX+8, bedSlotY-16
			}
		} else {
			p.x, p.y = int(math.Round(w.x))-p.w/2, int(math.Round(w.y))-p.h
			p.footY = w.y
			if state == game.AsleepWaiting || state == game.NoBedWaiting {
				p.flip = false
				p.zzz, p.effectX, p.effectY = true, p.x+p.w-5, p.y-16
			}
		}
		out = append(out, p)
	}
	sortPlacements(out)
	return out
}

// carryPlacement coloca al personal empujando la camilla donde estaba el
// paciente. Usa el sprite <doc|cam>_<nombre>_camilla (4 cuadros de 44x24).
func (c *camsScene) carryPlacement(g *game.GamePatient, m *game.StaffMember, w *walker) placement {
	a := c.env.Assets
	prefix := "cam_"
	if m.IsDoctor() {
		prefix = "doc_"
	}
	name := prefix + personKey(m.Name()) + "_camilla"
	p := placement{patient: g, frames: 4, flip: w.left}
	p.index = animFrame(c.env.Tick, 8, 4, 0)
	if a.Has(name) {
		p.imgName = name
		p.w, p.h = a.FrameSize(name, 4)
	} else {
		p.w, p.h = 44, 24
	}
	p.x, p.y = int(math.Round(w.x))-p.w/2, int(math.Round(w.y))-p.h
	p.footY = w.y
	return p
}

// Draw pinta toda la pantalla de cámaras.
func (c *camsScene) Draw(dst *ebiten.Image) {
	e := c.env
	tick := c.shift.Clock().Tick()
	light := lightFrame(tick)
	cam := c.shift.CurrentCamera()
	camInfo, _ := game.CameraByID(cam)

	// 1. Fondo de la pantalla y marco de la cámara.
	e.FillRect(dst, rect{0, 0, screenW, screenH}, '1', 1)
	e.FillRect(dst, rect{3, 19, 338, 248}, '0', 1)

	// 2. Barra superior.
	c.drawTopBar(dst)

	// 3 y 4. La cámara: fondo, camas y pacientes a 168x123, luego a x2.
	hasBackground := c.drawFeed(light, cam, camInfo)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(feedScale, feedScale)
	op.GeoM.Translate(feedScreenX, feedScreenY)
	dst.DrawImage(c.feed, op)
	if !hasBackground {
		e.Text(dst, "Fondo pendiente", feedScreenX+168, 70, textStyle{fontPixel, 8, '2', alignCenter})
	}
	if camInfo.Kind == game.RoomCam {
		// El número de la habitación va en la placa de la puerta.
		e.Text(dst, fmt.Sprintf("%d", camInfo.RoomNumber), 259, 46, textStyle{fontPixel, 8, '1', alignCenter})
	}

	// 5. Encima de la cámara: tinte por hora, viñeta y grano.
	t := tintForFrame(light)
	e.FillRect(dst, feedRect, t.key, t.alpha)
	c.drawOverlay(dst, "camara_vineta", 1, 0, 0.35)
	c.drawOverlay(dst, "camara_grano", 2, animFrame(e.Tick, 6, 2, 0), 0.12)

	// 6. Rótulo de la cámara: REC, sello y lugar, con borde oscuro.
	c.drawStamp(dst, cam, camInfo)

	// Etiqueta del paciente elegido y ayuda.
	if g := c.shift.FindPatient(c.selected); g != nil {
		c.drawSelectedLabel(dst, g)
	}
	if c.selected != "" {
		e.TextShadow(dst, "Elige a alguien del panel · Esc para cancelar", 90, 253, textStyle{fontPixel, 8, 'w', alignLeft})
	}
	if c.roomPatient() != nil {
		c.drawButton(dst, dischargeRect(), "Dar de alta", 'm', c.hovered == "discharge")
	}

	// 7 y 8. Panel de personal y mapa.
	c.drawStaffPanel(dst)
	c.drawMap(dst)

	if c.demo != nil {
		c.drawDemoBox(dst)
	}
	if c.report != nil {
		c.drawReport(dst)
	}
	if c.notice != nil && c.report == nil {
		c.drawNotice(dst)
	}
	if c.paused {
		c.drawPause(dst)
	}
}

// drawTopBar pinta la barra de arriba: día, turno, hora y estrellas.
func (c *camsScene) drawTopBar(dst *ebiten.Image) {
	e := c.env
	e.DrawSprite(dst, e.Assets.Image("ui_barra_superior"), 0, 0, false)
	title, subtitle := fmt.Sprintf("Día %d", c.shift.Day()), "Turno de 12 a 6"
	if c.demo != nil {
		title, subtitle = "Demo", "Escenario del enunciado"
	}
	e.Text(dst, title, 8, 2, textStyle{fontPixelBold, 10, 'M', alignLeft})
	e.Text(dst, subtitle, 50, 4, textStyle{fontPixel, 8, 'k', alignLeft})
	e.Text(dst, c.shift.Clock().Now(), 240, 2, textStyle{fontVT, 12, 'w', alignCenter})
	stars := c.shift.Stars()
	for i := 0; i < 5; i++ {
		frame := 1 // vacía
		if i < stars {
			frame = 0
		}
		e.DrawSprite(dst, e.Assets.Frame("ui_estrella", 2, frame), 430+10*i, 4, false)
	}
}

// drawFeed pinta el fondo, las camas y los pacientes en c.feed (168x123).
// Devuelve false si el fondo de esta cámara todavía no está dibujado.
func (c *camsScene) drawFeed(light int, cam game.CameraID, info game.Camera) bool {
	e := c.env
	a := e.Assets
	feed := c.feed
	feed.Clear()

	bg := backgroundName(cam, a.Names())
	hasBackground := bg != ""
	if hasBackground {
		w, _ := a.Size(bg)
		e.DrawSprite(feed, a.Frame(bg, max(w/feedW, 1), backgroundFrame(w, light)), 0, 0, false)
	} else {
		e.FillRect(feed, rect{0, 0, feedW, feedH}, 'W', 1)
		e.FillRect(feed, rect{0, floorRect.y - 4, feedW, feedH - floorRect.y + 4}, 'k', 1)
	}

	placed := c.placements()
	if info.Kind == game.RoomCam && !roomHasBedPatient(placed) {
		e.DrawSprite(feed, a.Image("cama"), bedSlotX, bedSlotY, false)
	}
	for _, p := range placed {
		c.drawPlacement(feed, p)
	}
	return hasBackground
}

// roomHasBedPatient dice si hay un paciente dibujado en la cama.
func roomHasBedPatient(list []placement) bool {
	for _, p := range list {
		if st := p.patient.State(); st == game.InBed || st == game.AwakeInBed {
			return true
		}
	}
	return false
}

// drawPlacement pinta un paciente y sus efectos en el lienzo de la cámara.
func (c *camsScene) drawPlacement(feed *ebiten.Image, p placement) {
	e := c.env
	a := e.Assets
	if p.imgName != "" {
		// Contorno del elegido: parpadea cada 0,4 s (24 ticks).
		if p.patient.Patient().ID() == c.selected && animFrame(e.Tick, 24, 2, 0) == 0 {
			sil := a.Silhouette(p.imgName, p.frames, p.index, 'y')
			for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
				e.DrawSprite(feed, sil, p.x+d[0], p.y+d[1], p.flip)
			}
		}
		e.DrawSprite(feed, a.Frame(p.imgName, p.frames, p.index), p.x, p.y, p.flip)
	} else {
		c.drawPlaceholder(feed, p)
	}
	if p.zzz {
		e.DrawSprite(feed, a.Frame("fx_zzz", 3, animFrame(e.Tick, 30, 3, 0)), p.effectX, p.effectY, false)
	}
	if p.alert {
		e.DrawSprite(feed, a.Frame("fx_alerta", 2, animFrame(e.Tick, 18, 2, 0)), p.effectX, p.effectY, false)
	}
}

// drawPlaceholder pinta una figura simple para los personajes que todavía no
// tienen sprite (P-006 a P-010). Se cambia sola cuando aparece el PNG.
func (c *camsScene) drawPlaceholder(feed *ebiten.Image, p placement) {
	e := c.env
	body := rect{p.x + 2, p.y + 2, p.w - 4, p.h - 2}
	e.FillRect(feed, rect{body.x - 1, body.y - 1, body.w + 2, body.h + 2}, '0', 1)
	e.FillRect(feed, body, 'c', 1)
	e.FillRect(feed, rect{body.x, body.y, body.w, min(8, body.h)}, 'C', 1)
	if p.patient.State() == game.AsleepWaiting || p.patient.State() == game.NoBedWaiting {
		// Dormido: acostado de lado, más bajo y ancho.
		e.FillRect(feed, rect{p.x, p.y + p.h - 10, p.w + 4, 10}, '0', 1)
		e.FillRect(feed, rect{p.x + 1, p.y + p.h - 9, p.w + 2, 8}, 'c', 1)
	}
}

// drawOverlay dibuja una capa de la cámara (viñeta, grano) con su opacidad.
func (c *camsScene) drawOverlay(dst *ebiten.Image, name string, frames, frame int, alpha float32) {
	img := c.env.Assets.Frame(name, frames, frame)
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(feedScreenX, feedScreenY)
	op.ColorScale.ScaleAlpha(alpha)
	dst.DrawImage(img, op)
}

// drawStamp pinta el rótulo de la cámara con borde oscuro de 1 px.
func (c *camsScene) drawStamp(dst *ebiten.Image, cam game.CameraID, info game.Camera) {
	e := c.env
	e.DrawSprite(dst, e.Assets.Frame("ui_rec", 2, animFrame(e.Tick, 60, 2, 0)), 10, 26, false)
	e.TextOutline(dst, "REC", 19, 24, textStyle{fontVT, stampFontSize, 'w', alignLeft})
	e.TextOutline(dst, cameraStamp(cam, c.shift.Clock().Now()), 332, 24, textStyle{fontVT, stampFontSize, 'w', alignRight})
	e.TextOutline(dst, info.Name, 10, 253, textStyle{fontVT, stampFontSize, 'w', alignLeft})
}

// levelChip devuelve el texto y el color de la manilla de un nivel.
func levelChip(l hospital.NarcolepsyLevel) (string, byte) {
	switch l {
	case hospital.Mild:
		return "Leve", 'L'
	case hospital.Moderate:
		return "Moderado", 'Y'
	default:
		return "Severo", 'R'
	}
}

// drawSelectedLabel pinta la caja con el nombre y el nivel junto al paciente.
func (c *camsScene) drawSelectedLabel(dst *ebiten.Image, g *game.GamePatient) {
	e := c.env
	var sprite rect
	found := false
	for _, p := range c.placements() {
		if p.patient == g {
			sprite, found = p.screenRect(), true
		}
	}
	if !found {
		return
	}
	box := labelPosition(sprite, feedRect)
	e.FillRect(dst, box, 'W', 1)
	e.StrokeRect(dst, box, '0', 1)
	name := g.Patient().Name()
	measure := func(s string) int { return e.Fonts.Width(s, fontPixel, 8) }
	name = fitText(name, 60, measure)
	e.Text(dst, name, box.x+4, box.y+2, textStyle{fontPixel, 8, '1', alignLeft})
	text, key := levelChip(g.Patient().Level())
	chip := rect{box.x + 64, box.y + 2, 48, 10}
	e.FillRect(dst, chip, key, 1)
	e.StrokeRect(dst, chip, '0', 1)
	e.Text(dst, text, chip.x+chip.w/2, chip.y, textStyle{fontPixel, 8, '0', alignCenter})
}

// drawButton pinta un botón: relleno, borde 0, sombra de 1 px abajo y texto.
func (c *camsScene) drawButton(dst *ebiten.Image, r rect, label string, fill byte, hover bool) {
	e := c.env
	e.FillRect(dst, rect{r.x, r.y + 1, r.w, r.h}, '0', 1)
	e.FillRect(dst, r, fill, 1)
	if hover {
		e.FillRect(dst, r, 'w', 0.25)
	}
	e.StrokeRect(dst, r, '0', 1)
	e.Text(dst, label, r.x+r.w/2, r.y+1, textStyle{fontPixel, 8, '0', alignCenter})
}

// ---------------------------------------------------------------------------
// Panel de personal
// ---------------------------------------------------------------------------

func (c *camsScene) drawStaffPanel(dst *ebiten.Image) {
	e := c.env
	a := e.Assets
	title := "Personal"
	if c.selected != "" {
		title = "¿Quién lo atiende?"
	}
	e.Text(dst, title, 346, 20, textStyle{fontPixelBold, 10, 'M', alignLeft})

	staff := c.shift.Staff()
	if len(staff) > cardsPerPage {
		label := fmt.Sprintf("Más > %d/%d", c.staffPage+1, staffPageCount(len(staff)))
		e.Text(dst, label, pageButtonRect.x+pageButtonRect.w, pageButtonRect.y+2, textStyle{fontPixel, 8, 'w', alignRight})
	}
	from, to := staffPage(len(staff), c.staffPage)
	for i := from; i < to; i++ {
		m := staff[i]
		slot := i - from
		r := staffCardRect(slot)
		state := m.State()
		if c.selected != "" && state == game.StaffAvailable {
			e.FillRect(dst, staffCardHighlight(slot), 'm', 1)
		}
		e.DrawSprite(dst, a.Image("ui_tarjeta"), r.x, r.y, false)
		c.drawHead(dst, m, r.x+4, r.y+4)

		measure := func(s string) int { return e.Fonts.Width(s, fontPixel, 8) }
		e.Text(dst, fitText(m.Name(), 82, measure), r.x+24, r.y+2, textStyle{fontPixel, 8, '1', alignLeft})
		e.Text(dst, fitText(staffRole(m), 82, measure), r.x+24, r.y+10, textStyle{fontPixel, 8, '2', alignLeft})
		if m.IsDoctor() {
			quota := fmt.Sprintf("%d/%d", m.PatientCount(), game.DoctorQuota)
			if state == game.StaffNoQuota {
				e.Text(dst, quota, r.x+108, r.y+2, textStyle{fontPixelBold, 8, 'R', alignLeft})
			} else {
				e.Text(dst, quota, r.x+108, r.y+2, textStyle{fontPixel, 8, '1', alignLeft})
			}
		}
		e.DrawSprite(dst, a.Frame("ui_chip_estado", 4, chipFrame(state)), r.x+24, r.y+22, false)
		chipKey := byte('0')
		if state == game.StaffNoQuota {
			chipKey = 'w'
		}
		e.Text(dst, chipText(state), r.x+26, r.y+22, textStyle{fontPixel, 8, chipKey, alignLeft})
		e.FillRect(dst, rect{r.x + 86, r.y + 25, 40, 4}, 'k', 1)
		e.FillRect(dst, rect{r.x + 86, r.y + 25, int(40 * m.CooldownFraction()), 4}, 'q', 1)
	}
}

// drawHead pinta la cabeza (filas 0 a 11 del cuadro 1 del sprite quieto) de
// una persona del personal; sin arte, un círculo simple.
func (c *camsScene) drawHead(dst *ebiten.Image, m *game.StaffMember, x, y int) {
	e := c.env
	a := e.Assets
	prefix := "cam_"
	if m.IsDoctor() {
		prefix = "doc_"
	}
	name := prefix + personKey(m.Name()) + "_quieto"
	if img := a.Image(name); img != nil {
		head := img.SubImage(image.Rect(0, 0, 16, 12)).(*ebiten.Image)
		e.DrawSprite(dst, head, x, y, false)
		return
	}
	e.FillRect(dst, rect{x + 3, y, 10, 12}, '0', 1)
	e.FillRect(dst, rect{x + 4, y + 1, 8, 10}, 'C', 1)
}

// ---------------------------------------------------------------------------
// Mapa
// ---------------------------------------------------------------------------

func (c *camsScene) drawMap(dst *ebiten.Image) {
	e := c.env
	a := e.Assets
	e.DrawSprite(dst, a.Image("ui_mapa"), mapOrigin.x, mapOrigin.y, false)
	e.Text(dst, "Mapa", 348, 157, textStyle{fontPixelBold, 8, 'k', alignLeft})
	current := c.shift.CurrentCamera()
	for _, id := range mapCameras {
		r := mapButtonRect(id)
		locked := !c.shift.World().Has(id)
		e.DrawSprite(dst, a.Frame("ui_boton_cam", 3, buttonFrame(id == current, locked)), r.x, r.y, false)
		key := byte('1')
		switch {
		case id == current:
			key = '0'
		case locked:
			key = 'X'
		}
		e.Text(dst, string(id), r.x+r.w/2, r.y-1, textStyle{fontVT, stampFontSize, key, alignCenter})
	}
	e.Text(dst, "Monitoreo", 391, 233, textStyle{fontPixel, 8, 'k', alignLeft})

	// Candado rojo al pasar el mouse por una cámara bloqueada. El dibujo
	// (ui_candado.png) lo hace el arte; mientras no exista, no se muestra.
	if c.hoverLocked != "" {
		if lock := a.Image("ui_candado"); lock != nil {
			r := mapButtonRect(c.hoverLocked)
			b := lock.Bounds()
			e.DrawSprite(dst, lock, r.x+r.w/2-b.Dx()/2, r.y+r.h/2-b.Dy()/2, false)
		}
	}
}

// ---------------------------------------------------------------------------
// Avisos, Demo, pausa y reporte
// ---------------------------------------------------------------------------

// drawNotice pinta el aviso amable (por ejemplo "No hay cama libre!").
func (c *camsScene) drawNotice(dst *ebiten.Image) {
	box := noticeRect
	if c.demo != nil {
		box.y = 140 // en la Demo se sube para no tapar la caja de controles
	}
	e := c.env
	drawBox(e, dst, box)
	e.Text(dst, c.notice.title, box.x+8, box.y+5, textStyle{fontPixelBold, 10, '2', alignLeft})
	for i, line := range c.notice.lines {
		e.Text(dst, line, box.x+8, box.y+20+11*i, textStyle{fontPixel, 8, '1', alignLeft})
	}
}

// drawDemoBox pinta la caja de la Demo con el paso y los dos botones.
func (c *camsScene) drawDemoBox(dst *ebiten.Image) {
	e := c.env
	drawBox(e, dst, demoBoxRect)
	e.DrawSprite(dst, e.Assets.Image("director_retrato"), 14, 207, false)
	e.Text(dst, fmt.Sprintf("Paso %d de 4", c.demoStep), 68, 206, textStyle{fontPixelBold, 10, '2', alignLeft})
	measure := func(s string) int { return e.Fonts.Width(s, fontPixel, 8) }
	lines := wrapText(demoStepTexts[c.demoStep-1], 258, measure)
	for i, line := range lines {
		if i >= 2 {
			break
		}
		e.Text(dst, line, 68, 220+11*i, textStyle{fontPixel, 8, '1', alignLeft})
	}
	next, nextFill := "Siguiente paso", byte('m')
	if c.demoStep >= 4 {
		next = "Volver al menú"
	}
	c.drawButton(dst, demoNextRect, next, nextFill, c.hovered == "next" || c.hovered == "back")
	autoLabel, autoFill := "Auto", byte('k')
	if c.demoAuto {
		autoLabel, autoFill = "Pausa", 'm'
	}
	c.drawButton(dst, demoAutoRect, autoLabel, autoFill, c.hovered == "auto")
}

// drawPause pinta la pausa mínima (la vista de Pausa todavía no está aprobada).
func (c *camsScene) drawPause(dst *ebiten.Image) {
	e := c.env
	e.FillRect(dst, rect{0, 0, screenW, screenH}, '0', 0.45)
	drawBox(e, dst, pauseBoxRect)
	e.Text(dst, "Pausa", pauseBoxRect.x+pauseBoxRect.w/2, pauseBoxRect.y+8, textStyle{fontPixelBold, 10, '2', alignCenter})
	e.Text(dst, "Esc: seguir · M: menú", pauseBoxRect.x+pauseBoxRect.w/2, pauseBoxRect.y+26, textStyle{fontPixel, 8, '1', alignCenter})
	c.drawButton(dst, pauseResumeRect, "Seguir", 'm', c.hovered == "resume")
	c.drawButton(dst, pauseMenuRect, "Menú", 'k', c.hovered == "menu")
}

// drawReport pinta el reporte con las 4 consultas. Es provisional: se cambia
// cuando la vista del Reporte esté aprobada.
func (c *camsScene) drawReport(dst *ebiten.Image) {
	e := c.env
	r := reportRect(c.demo != nil)
	e.FillRect(dst, r, 'W', 1)
	e.StrokeRect(dst, r, '0', 1)

	title := "Reporte del turno"
	if c.demo == nil {
		title = fmt.Sprintf("Fin del turno - Día %d", c.report.Day)
	}
	e.Text(dst, title, r.x+8, r.y+5, textStyle{fontPixelBold, 10, '2', alignLeft})
	for i := 0; i < 5; i++ {
		frame := 1
		if i < c.report.Stars {
			frame = 0
		}
		e.DrawSprite(dst, e.Assets.Frame("ui_estrella", 2, frame), r.x+r.w-58+10*i, r.y+5, false)
	}
	e.Text(dst, fmt.Sprintf("Puntos: %d", c.report.Total), r.x+8, r.y+19, textStyle{fontPixel, 8, '1', alignLeft})

	bottom := r.y + r.h - 4
	if c.demo == nil {
		bottom = backButtonRect(false).y - 4
	}
	maxLines := (bottom - (r.y + 32)) / 11
	for i, line := range capLines(reportLines(*c.report), maxLines) {
		family := fontPixel
		key := byte('1')
		if line.head {
			family, key = fontPixelBold, '2'
		}
		size := 8
		text := fitText(line.text, r.w-16, func(s string) int { return e.Fonts.Width(s, family, size) })
		e.Text(dst, text, r.x+8, r.y+32+11*i, textStyle{family, size, key, alignLeft})
	}
	if c.demo == nil {
		c.drawButton(dst, backButtonRect(false), "Volver al menú", 'm', c.hovered == "back")
	}
}

// stampFontSize es el tamaño de VT323 del rótulo y de los botones del mapa.
// La nota de Figma dice 10, pero en el juego a 10 px los trazos quedan a
// medio píxel y se ven borrosos; a 13 se leen nítidos (VISUAL).
const stampFontSize = 13

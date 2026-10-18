package gameview

import (
	"github.com/hajimehoshi/ebiten/v2"

	"costenos-narcolepsia/game"
)

// Pantalla del menú principal, según las vistas aprobadas "Menu / default",
// "Menu / hover Demo" y "Menu / sin partida guardada" (frames 5:9, 5:57,
// 5:95 y notas 5:134 a 5:136 de Figma). Medidas a 1x sobre 480x270.

// menuItem es una opción del menú con su descripción (Textos del Director.md).
type menuItem struct {
	label string
	desc  string
}

var menuItems = [...]menuItem{
	{"Nuevo turno", "Empieza el Día 1. Suave, sin afán."},
	{"Continuar", "Sigue en el último día que desbloqueaste."},
	{"Demo (escenario del enunciado)", "El escenario del enunciado, paso a paso."},
	{"Cómo jugar", "Te explicamos el turno en un momentico."},
	{"Ajustes", "Volumen, parpadeo de cámaras y pantalla completa."},
	{"Créditos", "Quién hizo esto y con qué música."},
}

// Índices de las opciones.
const (
	optNewShift = iota
	optContinue
	optDemo
	optHowToPlay
	optSettings
	optCredits
)

// menuHelp es la línea de ayuda (sin ¡: decisión de la autora).
const menuHelp = "Flechas: elegir · Enter: dale"

// optionRect es la barra de la opción i: (26, 85 + 18i) 172x15.
func optionRect(i int) rect { return rect{26, 85 + 18*i, 172, 15} }

// menuStep mueve la selección una opción arriba (-1) o abajo (+1) saltando
// las que no se pueden elegir (Continuar sin partida guardada). Da la vuelta
// en los extremos. Es pura para poder probarla.
func menuStep(current, dir int, enabled func(int) bool) int {
	n := len(menuItems)
	for i := 1; i <= n; i++ {
		next := ((current+dir*i)%n + n) % n
		if enabled(next) {
			return next
		}
	}
	return current
}

// menuScene es el menú.
type menuScene struct {
	env      *Env
	selected int
	overlay  string // texto de "Próximamente" mientras esté abierto
	hovered  string
}

func newMenu(env *Env) *menuScene { return &menuScene{env: env} }

// enabled dice si se puede elegir la opción i.
func (m *menuScene) enabled(i int) bool { return i != optContinue || m.env.HasSave }

func (m *menuScene) Shift() *game.Shift { return nil }

func (m *menuScene) State() UIState {
	st := UIState{Scene: "menu", Selected: menuItems[m.selected].label, HoveredButton: m.hovered}
	if m.overlay != "" {
		st.Overlay = "Próximamente: " + m.overlay
	}
	return st
}

// Update lee flechas, mouse y Enter.
func (m *menuScene) Update(in *Input) Scene {
	if m.overlay != "" {
		if in.Click || in.Pressed("Enter") || in.Pressed("Escape") || in.Pressed("Space") {
			m.overlay = ""
		}
		return nil
	}
	if in.Pressed("Up") {
		m.selected = menuStep(m.selected, -1, m.enabled)
	}
	if in.Pressed("Down") {
		m.selected = menuStep(m.selected, +1, m.enabled)
	}
	// El mouse encima de una opción la elige igual que las flechas; solo
	// cuenta cuando el mouse se mueve, para no pelear con el teclado.
	m.hovered = ""
	hoverIndex := -1
	for i := range menuItems {
		if optionRect(i).contains(in.MouseX, in.MouseY) && m.enabled(i) {
			hoverIndex = i
			m.hovered = menuItems[i].label
		}
	}
	if hoverIndex >= 0 && in.Moved {
		m.selected = hoverIndex
	}
	if in.Pressed("Enter") || in.Pressed("Space") || (in.Click && hoverIndex >= 0) {
		if in.Click && hoverIndex >= 0 {
			m.selected = hoverIndex
		}
		return m.activate()
	}
	return nil
}

// activate ejecuta la opción elegida.
func (m *menuScene) activate() Scene {
	switch m.selected {
	case optNewShift:
		return m.startDay(1)
	case optContinue:
		return m.startDay(m.env.Save.LastUnlockedDay)
	case optDemo:
		demo, err := game.NewDemo()
		if err != nil {
			m.overlay = err.Error()
			return nil
		}
		scene, err := newCams(m.env, demo.Shift(), demo, "")
		if err != nil {
			m.overlay = err.Error()
			return nil
		}
		return scene
	case optHowToPlay:
		scene, err := newPracticeScene(m.env, m.env.newSeed(), 0, "")
		if err != nil {
			m.overlay = err.Error()
			return nil
		}
		return scene
	default:
		// Créditos todavía no tiene vista aprobada.
		m.overlay = menuItems[m.selected].label
	}
	return nil
}

// startDay arma el turno de un día (con el mensaje del Director que le toca)
// y abre la pantalla de cámaras.
func (m *menuScene) startDay(day int) Scene {
	scene, err := newShiftScene(m.env, day, m.env.newSeed(), true, "")
	if err != nil {
		m.overlay = err.Error()
		return nil
	}
	return scene
}

// Draw pinta las capas de atrás hacia adelante, igual que la nota de Figma.
func (m *menuScene) Draw(dst *ebiten.Image) {
	e := m.env
	a := e.Assets

	// 1. Fondo (atardecer fijo).
	e.DrawSprite(dst, a.Image("menu_fondo"), 0, 0, false)

	// 2. Palmeras: alternan cuadro cada 0,8 s (48 ticks), desfasadas 0,4 s.
	e.DrawSprite(dst, a.Frame("palmera", 2, animFrame(e.Tick, 48, 2, 0)), 196, 114, false)
	e.DrawSprite(dst, a.Frame("palmera", 2, animFrame(e.Tick+24, 48, 2, 0)), 430, 118, false)

	// 3. Banca.
	e.DrawSprite(dst, a.Image("banca"), 278, 198, false)

	// 4. Título y subtítulo.
	e.TextShadow(dst, "Hospital de los Costeños", 20, 14, textStyle{fontPixelBold, 18, 'w', alignLeft})
	e.TextShadow(dst, "con Narcolepsia", 20, 34, textStyle{fontPixelBold, 18, 'M', alignLeft})
	e.TextShadow(dst, "El hueco más profundo de Montería", 20, 58, textStyle{fontPixel, 8, 'y', alignLeft})

	// 5. Panel con las opciones.
	e.DrawSprite(dst, a.Image("ui_panel_menu"), 20, 80, false)
	for i, item := range menuItems {
		y := 87 + 18*i
		key := byte('1')
		switch {
		case !m.enabled(i):
			key = 'x'
		case i == m.selected:
			key = '0'
			r := optionRect(i)
			e.DrawSprite(dst, a.Image("ui_seleccion"), r.x, r.y, false)
			e.DrawSprite(dst, a.Frame("ui_flecha", 2, animFrame(e.Tick, 24, 2, 0)), 29, 88+18*i, false)
		}
		e.Text(dst, item.label, 38, y, textStyle{fontPixel, 10, key, alignLeft})
		if !m.enabled(i) {
			e.Text(dst, "(aún no hay partida)", 92, y+2, textStyle{fontPixel, 8, 'x', alignLeft})
		}
	}

	// 6. Divisor y descripción de la opción elegida.
	e.FillRect(dst, rect{28, 198, 168, 1}, 'k', 1)
	// La descripción se parte en líneas que quepan en el panel (168 px).
	measure := func(s string) int { return e.Fonts.Width(s, fontPixel, 8) }
	for i, line := range wrapText(menuItems[m.selected].desc, 164, measure) {
		e.Text(dst, line, 28, 202+11*i, textStyle{fontPixel, 8, '2', alignLeft})
	}

	// 7. Línea de ayuda.
	e.TextShadow(dst, menuHelp, 24, 232, textStyle{fontPixel, 10, 'w', alignLeft})

	if m.overlay != "" {
		drawComingSoon(e, dst, m.overlay)
	}
}

// drawComingSoon pinta el aviso provisional de las pantallas sin vista
// aprobada todavía (Cómo jugar, Ajustes, Créditos).
func drawComingSoon(e *Env, dst *ebiten.Image, title string) {
	box := rect{140, 100, 200, 56}
	drawBox(e, dst, box)
	e.Text(dst, title, box.x+box.w/2, box.y+8, textStyle{fontPixelBold, 10, '2', alignCenter})
	e.Text(dst, "Próximamente", box.x+box.w/2, box.y+24, textStyle{fontPixel, 8, '1', alignCenter})
	e.Text(dst, "Enter o clic para volver", box.x+box.w/2, box.y+38, textStyle{fontPixel, 8, 'x', alignCenter})
}

// drawBox pinta una caja con borde 0 de 1 px, relleno W y sombra 0 al 60 %
// desplazada (+2,+2): el estilo de los avisos de las vistas aprobadas.
func drawBox(e *Env, dst *ebiten.Image, r rect) {
	e.FillRect(dst, rect{r.x + 2, r.y + 2, r.w, r.h}, '0', 0.6)
	e.FillRect(dst, r, 'W', 1)
	e.StrokeRect(dst, r, '0', 1)
}

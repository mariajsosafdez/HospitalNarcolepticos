package gameview

import (
	"github.com/hajimehoshi/ebiten/v2"

	"costenos-narcolepsia/game"
)

// Pantalla de Créditos, según la vista aprobada "Créditos" (frame 23:2 de
// Figma, nota 24:51; copia en preview/turno/creditos.png). Medidas a 1x sobre
// 480x270. Se abre desde la opción "Créditos" del menú y Esc, Enter, Espacio
// o un clic vuelven al menú con "Créditos" elegido.

// Desfile de personajes por la acera (encima del fondo oscurecido). Pasa TODO
// el elenco: 10 pacientes, 3 médicos, 2 camilleros y los 5 easter eggs.
const (
	paradeSpeed      = 16  // píxeles por segundo hacia la derecha
	paradeFrameTicks = 9   // 0,15 s por cuadro a 60 ticks por segundo
	paradeFrames     = 4   // cuadros de cada hoja *_camina
	paradeSlot       = 30  // lo que ocupa cada uno en la fila: 16 de ancho + 14 de aire
	paradeStretcher  = 60  // Chepe con la camilla mide 44 de ancho y ocupa más
	paradeSpan       = 630 // 19 puestos de 30 + el de la camilla: el desfile entero, de ahí da la vuelta
	paradeMargin     = 44  // lo que sobresale por cada lado antes de volver a entrar
	paradeY          = 200 // arriba de los sprites: los pies quedan en y = 224
)

// paradeWalker es uno del desfile: su hoja de cuadros, dónde empieza y con
// qué cuadro arranca (distinto para cada uno, así no pisan todos a la vez).
type paradeWalker struct {
	sprite string
	x0     int
	start  int
}

// paradeSprites es el orden del desfile: médicos, camilleros, pacientes y,
// al final, los easter eggs (Misifu camina en cuatro patas). Chepe empuja la
// camilla.
var paradeSprites = [...]string{
	"doc_karen_camina", "doc_alvaro_camina", "doc_luzdary_camina",
	"cam_chepe_camilla", "cam_mono_camina",
	"pac_yesid_camina", "pac_yeimy_camina", "pac_rafa_camina", "pac_wilfrido_camina",
	"pac_marelvis_camina", "pac_nayeli_camina", "pac_mane_camina", "pac_kelly_camina",
	"pac_eder_camina", "pac_yuranis_camina",
	"pac_egg1_camina", "pac_egg2_camina", "pac_egg3_camina", "pac_egg4_camina", "pac_egg5_camina",
}

// parade reparte a todos parejo por la fila: cada uno en su puesto, con el
// cuadro de arranque distinto (i mod 4).
var parade = buildParade()

func buildParade() []paradeWalker {
	walkers := make([]paradeWalker, 0, len(paradeSprites))
	x := 0
	for i, sprite := range paradeSprites {
		walkers = append(walkers, paradeWalker{sprite, x, (i * 3) % paradeFrames})
		if sprite == "cam_chepe_camilla" {
			x += paradeStretcher
		} else {
			x += paradeSlot
		}
	}
	return walkers
}

// paradeX da la x de un personaje del desfile en el tick dado: camina a la
// derecha y el que sale por ese lado vuelve a entrar por el izquierdo.
// x = ((x0 + 44 + 16·t) mod 630) - 44, con t = tick/60 segundos. Se calcula
// con enteros (todo multiplicado por 60) para no acumular errores.
func paradeX(x0, tick int) int {
	n := ((x0+paradeMargin)*60 + paradeSpeed*tick) % (paradeSpan * 60)
	return n/60 - paradeMargin
}

// paradeFrame da el cuadro de la hoja en el tick dado.
func paradeFrame(start, tick int) int {
	return animFrame(tick, paradeFrameTicks, paradeFrames, start)
}

// creditsScene son los Créditos. Guarda el menú para volver a él tal cual.
type creditsScene struct {
	env  *Env
	menu *menuScene
	tick int // ticks desde que se abrió (el desfile arranca en su sitio)
}

// newCredits abre los Créditos; al salir se vuelve a menu.
func newCredits(env *Env, menu *menuScene) *creditsScene {
	return &creditsScene{env: env, menu: menu}
}

func (c *creditsScene) Shift() *game.Shift { return nil }

func (c *creditsScene) State() UIState {
	return UIState{Scene: "credits", Selected: "Volver"}
}

// Update: cualquiera de estas entradas es "Volver".
func (c *creditsScene) Update(in *Input) Scene {
	c.tick++
	if in.Click || in.Pressed("Escape") || in.Pressed("Enter") || in.Pressed("Space") {
		return c.menu
	}
	return nil
}

// Draw pinta las capas de atrás hacia adelante, igual que la nota de Figma.
func (c *creditsScene) Draw(dst *ebiten.Image) {
	e := c.env
	a := e.Assets

	// 1. Fondo quieto y oscurecido al 60 %.
	e.DrawSprite(dst, a.Image("menu_fondo"), 0, 0, false)
	e.FillRect(dst, rect{0, 0, screenW, screenH}, '0', 0.6)

	// 2. Desfile: va sin oscurecer, abajo, sobre la acera.
	for _, w := range parade {
		img := a.Frame(w.sprite, paradeFrames, paradeFrame(w.start, c.tick))
		e.DrawSprite(dst, img, paradeX(w.x0, c.tick), paradeY, false)
	}

	// 3. Panel y encabezado.
	drawBox(e, dst, rect{16, 6, 448, 190})
	e.Text(dst, "Créditos", 28, 11, textStyle{fontPixelBold, 18, '1', alignLeft})
	e.Text(dst, "Hospital de los Costeños con Narcolepsia", 28, 38, textStyle{fontPixelBold, 10, '2', alignLeft})
	e.Text(dst, "El hueco más profundo de Montería", 28, 51, textStyle{fontPixel, 8, 'x', alignLeft})

	// 4. Don Ernesto con su marco.
	e.FillRect(dst, rect{403, 11, 50, 50}, '0', 1)
	e.DrawSprite(dst, a.Image("director_retrato"), 404, 12, false)
	e.Text(dst, "Don Ernesto", 294, 22, textStyle{fontPixelBold, 8, '2', alignLeft})
	e.Text(dst, "Gracias por la ayuda, vale!", 294, 33, textStyle{fontPixel, 8, '1', alignLeft})
	e.FillRect(dst, rect{24, 64, 432, 1}, 'k', 1)

	// 5. Columna izquierda (x = 28): rótulos en negrita, datos en normal.
	label := textStyle{fontPixelBold, 8, 't', alignLeft}
	data := textStyle{fontPixel, 8, '1', alignLeft}
	faint := textStyle{fontPixel, 8, 'x', alignLeft}
	e.Text(dst, "Hecho por", 28, 70, label)
	e.Text(dst, "Maria José Sosa", 28, 80, textStyle{fontPixel, 10, '1', alignLeft})
	e.Text(dst, "Clase", 28, 96, label)
	e.Text(dst, "Compiladores y Lenguajes", 28, 106, data)
	e.Text(dst, "Universidad EIA", 28, 117, data)
	e.Text(dst, "Profesor Sebastián Zapata Ramírez", 28, 128, data)
	e.Text(dst, "Herramientas", 28, 144, label)
	e.Text(dst, "Hecho en Go con Ebitengine", 28, 154, data)

	// 6. Columna derecha (x = 250).
	e.Text(dst, "Fuentes", 250, 70, label)
	e.Text(dst, "Pixelify Sans (Stefie Justprince)", 250, 80, data)
	e.Text(dst, "VT323 (Peter Hull)", 250, 91, data)
	e.Text(dst, "Licencia SIL OFL 1.1", 250, 102, faint)
	ai := textStyle{fontPixel, 8, '2', alignLeft}
	e.Text(dst, "Arte y código hechos con ayuda de IA", 250, 122, ai)
	e.Text(dst, "(Claude, de Anthropic)", 250, 133, ai)

	// 7. Botón Volver (siempre elegido: es la única opción) y ayuda.
	btn := rect{372, 174, 84, 16}
	e.FillRect(dst, btn, 'm', 1)
	e.StrokeRect(dst, btn, '0', 1)
	e.DrawSprite(dst, a.Frame("ui_flecha", 2, animFrame(e.Tick, 24, 2, 0)), 377, 177, false)
	e.Text(dst, "Volver", 400, 176, textStyle{fontPixel, 10, '0', alignLeft})
	e.Text(dst, "Esc o clic: volver al menú", 28, 178, faint)
}

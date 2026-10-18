package gameview

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"costenos-narcolepsia/game"
	"costenos-narcolepsia/hospital"
)

// Este archivo tiene las funciones PURAS de la vista: medidas, qué cuadro
// toca según la hora, a qué le da clic el mouse... Ninguna dibuja ni lee la
// entrada, así que se prueban con tablas (layout_test.go). Todas las
// medidas salen de las notas de las vistas de Figma (frames 9:205 a 9:674).

const (
	screenW = 480
	screenH = 270

	// La cámara se dibuja a x2: el fondo mide 168x123 y ocupa 336x246.
	feedW       = 168
	feedH       = 123
	feedScale   = 2
	feedScreenX = 4
	feedScreenY = 20
)

// rect es un rectángulo en píxeles del lienzo de 480x270.
type rect struct{ x, y, w, h int }

// contains dice si el punto (px, py) está dentro del rectángulo.
func (r rect) contains(px, py int) bool {
	return px >= r.x && px < r.x+r.w && py >= r.y && py < r.y+r.h
}

// feedRect es el rectángulo de la cámara en la pantalla.
var feedRect = rect{feedScreenX, feedScreenY, feedW * feedScale, feedH * feedScale}

// feedToScreen convierte coordenadas del fondo (168x123) en pantalla.
func feedToScreen(x, y int) (int, int) {
	return feedScreenX + feedScale*x, feedScreenY + feedScale*y
}

// ---------------------------------------------------------------------------
// Luz por hora
// ---------------------------------------------------------------------------

// Cada fondo trae 3 cuadros: mediodía (12-14), tarde dorada (14-16) y
// atardecer (16-18). El cambio es de golpe al cruzar la hora.
const ticksPerLightPhase = 2 * 60 * game.TicksPerMinute

// lightFrame devuelve el cuadro del fondo (0, 1 o 2) para un tick del turno.
func lightFrame(tick int) int {
	return min(max(tick/ticksPerLightPhase, 0), 2)
}

// tint es el color (clave de la paleta) y la opacidad que se pone encima de
// la cámara según la hora.
type tint struct {
	key   byte
	alpha float64
}

// tintForFrame: mediodía y 6 %, tarde dorada M 10 %, atardecer o 14 %.
func tintForFrame(frame int) tint {
	switch frame {
	case 0:
		return tint{'y', 0.06}
	case 1:
		return tint{'M', 0.10}
	default:
		return tint{'o', 0.14}
	}
}

// backgroundFrame elige el cuadro de un fondo que mide widthPx de ancho. Si
// el arte nuevo trae menos de 3 cuadros, se usa el último que exista.
func backgroundFrame(widthPx, light int) int {
	frames := max(widthPx/feedW, 1)
	return min(light, frames-1)
}

// ---------------------------------------------------------------------------
// Arte de cada cámara y de cada persona
// ---------------------------------------------------------------------------

// backgroundName busca el PNG de fondo de una cámara entre los nombres que
// hay en art/png. Las habitaciones comparten fondo_habitacion; las demás
// buscan "fondo_cam<id>_*" ("fondo_cam1a_recepcion", "fondo_cam3_...").
// Devuelve "" si todavía no está dibujado.
func backgroundName(cam game.CameraID, names []string) string {
	if c, ok := game.CameraByID(cam); ok && c.Kind == game.RoomCam {
		// Cada habitación puede tener su propio fondo (fondo_habitacion_101...);
		// si no existe se usa el fondo común.
		own := fmt.Sprintf("fondo_habitacion_%d", c.RoomNumber)
		for _, n := range []string{own, "fondo_habitacion"} {
			if slices.Contains(names, n) {
				return n
			}
		}
		return ""
	}
	prefix := "fondo_cam" + strings.ToLower(string(cam)) + "_"
	for _, n := range names {
		if strings.HasPrefix(n, prefix) {
			return n
		}
	}
	return ""
}

// point es una posición (esquina de arriba a la izquierda) dentro del fondo
// de 168x123 de una cámara.
type point struct{ x, y int }

// roomArt dice dónde están, dentro del fondo de una habitación, la cama (esquina
// del sprite de 32x30) y la placa de la puerta (rect de 9x4 donde va el número).
type roomArt struct{ bed, plate point }

// defaultRoom es la habitación del fondo común: cama en (34, 72) (nota de
// Figma 9:690) y placa en (123, 14).
var defaultRoom = roomArt{bed: point{34, 72}, plate: point{123, 14}}

// roomArts guarda lo de los fondos propios de cada habitación. Cada archivo
// art/sprites/fondo_habitacion_<n>.txt lo dice en sus dos primeras líneas
// ("// cama: x,y" y "// placa: x,y"); esta tabla se mantiene a mano igual y un
// test (TestRoomArtsMatchSprites) lee los .txt y falla si dejan de coincidir.
// Se eligió una tabla y no leer el comentario al ejecutar porque los .txt no
// van dentro del ejecutable (solo los PNG), y así la vista no depende de
// parsear arte.
var roomArts = map[int]roomArt{
	101: {point{68, 52}, point{123, 14}},
	102: {point{60, 50}, point{32, 14}},
	103: {point{76, 54}, point{129, 14}},
	104: {point{66, 56}, point{31, 14}},
	105: {point{72, 50}, point{125, 14}},
	106: {point{68, 54}, point{29, 14}},
}

// roomLayout devuelve la cama y la placa de una habitación.
func roomLayout(roomNumber int) roomArt {
	if art, ok := roomArts[roomNumber]; ok {
		return art
	}
	return defaultRoom
}

// accents quita las tildes para armar nombres de archivo.
var accents = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n")

// personKey saca la clave de los PNG de una persona: la primera palabra del
// nombre sin título, en minúsculas y sin tildes ("Dra. Karen Ospina" ->
// "karen", "Dr. Álvaro Mendoza" -> "alvaro").
func personKey(fullName string) string {
	for _, word := range strings.Fields(fullName) {
		if strings.HasSuffix(word, ".") { // Dr. Dra.
			continue
		}
		return accents.Replace(strings.ToLower(word))
	}
	return ""
}

// spriteKeys da la clave de los PNG de cada persona POR ID. No se puede sacar
// del nombre: "Abelardo "Mañe" Julio" es pac_mane_* y no pac_abelardo_*.
var spriteKeys = map[string]string{
	"P-001": "yesid", "P-002": "yeimy", "P-003": "rafa", "P-004": "wilfrido", "P-005": "marelvis",
	"P-006": "nayeli", "P-007": "mane", "P-008": "kelly", "P-009": "eder", "P-010": "yuranis",
	"D-001": "karen", "D-002": "alvaro", "D-003": "luzdary",
	"O-001": "chepe", "O-002": "mono",
}

// spriteKey devuelve la clave de los PNG de una persona: la de la tabla por
// ID, y si el ID no está (por ejemplo, un personaje nuevo) la primera palabra
// del nombre.
func spriteKey(id, fullName string) string {
	if key, ok := spriteKeys[id]; ok {
		return key
	}
	return personKey(fullName)
}

// femaleNames sirve para decir "dormida" o "dormido" en los avisos.
var femaleNames = []string{"yeimy", "marelvis", "nayeli", "kelly", "yuranis"}

// isFemale dice si el paciente es mujer, según su primer nombre.
func isFemale(fullName string) bool {
	return slices.Contains(femaleNames, personKey(fullName))
}

// spriteSpec dice qué PNG y cuántos cuadros usa un paciente en un estado.
type spriteSpec struct {
	suffix     string // pac_<nombre>_<suffix>
	frames     int    // cuadros del PNG
	frameTicks int    // cuánto dura cada cuadro, en ticks (60 por segundo)
	durations  []int  // si no es nil, cuánto dura CADA cuadro (cuadros de distinta duración)
}

// sittingDurations: el paciente despierto en la cama (`_sentado`) pasa 2,0 s
// con el cuadro 1 y parpadea 0,2 s con el cuadro 2 (nota de Figma 18:96).
var sittingDurations = []int{120, 12}

// patientSprite elige el sprite según el estado de juego (notas "Cam / zona
// común" y "Cam / habitación"). ok es false si en ese estado el paciente no
// se dibuja como persona (en camilla, o ya se fue).
func patientSprite(state game.PatientState, walking bool) (spriteSpec, bool) {
	switch state {
	case game.Arriving, game.Wandering:
		if walking {
			return spriteSpec{"camina", 4, 9, nil}, true // 0,15 s por cuadro
		}
		return spriteSpec{"quieto", 2, 30, nil}, true
	case game.AsleepWaiting, game.NoBedWaiting:
		return spriteSpec{"dormido", 2, 72, nil}, true // 1,2 s
	case game.InBed:
		return spriteSpec{"cama", 2, 90, nil}, true // 1,5 s: respira
	case game.AwakeInBed:
		return spriteSpec{"sentado", 2, 0, sittingDurations}, true // 2,0 s y un parpadeo de 0,2 s
	}
	return spriteSpec{}, false
}

// frameByDurations elige el cuadro de una animación cuyos cuadros duran
// distinto: durations[i] son los ticks del cuadro i. Después del último
// vuelve al primero. Es pura para probarla con una tabla.
func frameByDurations(tick int, durations []int) int {
	total := 0
	for _, d := range durations {
		total += d
	}
	if total <= 0 {
		return 0
	}
	t := ((tick % total) + total) % total
	for i, d := range durations {
		if t < d {
			return i
		}
		t -= d
	}
	return 0
}

// spriteFrame elige el cuadro de un sprite de paciente para un tick de la interfaz.
func spriteFrame(spec spriteSpec, tick int) int {
	if len(spec.durations) > 0 {
		return frameByDurations(tick, spec.durations)
	}
	return animFrame(tick, spec.frameTicks, spec.frames, 0)
}

// animFrame elige el cuadro de una animación que cambia cada frameTicks.
func animFrame(tick, frameTicks, frames, offset int) int {
	if frameTicks <= 0 || frames <= 0 {
		return 0
	}
	return (tick/frameTicks + offset) % frames
}

// ---------------------------------------------------------------------------
// Panel de personal
// ---------------------------------------------------------------------------

// normalCards es hasta cuántas personas caben con la tarjeta normal (132x36).
// Con 4 o 5 personas (días 3 a 5) el panel usa la tarjeta compacta (132x24),
// así caben todas sin pasar páginas (Dirección de Arte, retoques 2026-10-17).
const normalCards = 3

// compactCards dice si, con este personal, el panel usa tarjetas compactas.
func compactCards(total int) bool { return total > normalCards }

// staffCardRect es la tarjeta i del panel: normal (344, 34 + 40i) 132x36, o
// compacta (344, 34 + 24i) 132x24 cuando hay más de 3 personas.
func staffCardRect(i, total int) rect {
	if compactCards(total) {
		return rect{344, 34 + 24*i, 132, 24}
	}
	return rect{344, 34 + 40*i, 132, 36}
}

// staffCardHighlight es el borde m de 2 px por fuera de la tarjeta normal.
func staffCardHighlight(i int) rect { return rect{342, 32 + 40*i, 136, 40} }

// staffRole es lo que sale bajo el nombre: la especialidad del doctor o
// "Camillero".
func staffRole(m *game.StaffMember) string {
	if m.IsDoctor() {
		return displaySpecialty(m.ID(), m.Doctor().Specialty())
	}
	return "Camillero"
}

// chipText: texto del chip de estado.
func chipText(s game.StaffState) string {
	switch s {
	case game.StaffTransferring:
		return "En traslado"
	case game.StaffResting:
		return "Descansando"
	case game.StaffNoQuota:
		return "Sin cupo"
	default:
		return "Disponible"
	}
}

// chipFrame: cuadro de ui_chip_estado (0 Disponible, 1 En traslado,
// 2 Descansando, 3 Sin cupo).
func chipFrame(s game.StaffState) int {
	switch s {
	case game.StaffTransferring:
		return 1
	case game.StaffResting:
		return 2
	case game.StaffNoQuota:
		return 3
	default:
		return 0
	}
}

// ---------------------------------------------------------------------------
// Mapa
// ---------------------------------------------------------------------------

var mapOrigin = rect{344, 156, 132, 110}

// mapButtonOffsets son los (rx, ry) de cada botón dentro del mapa.
var mapButtonOffsets = map[game.CameraID][2]int{
	"6": {8, 16}, "8": {58, 16}, "2": {108, 16},
	"4": {8, 38}, "1B": {42, 38}, "3": {76, 38}, "5": {110, 38},
	"7A": {4, 60}, "7B": {22, 60}, "7C": {40, 60}, "1A": {58, 60},
	"7D": {76, 60}, "7E": {94, 60}, "7F": {112, 60},
}

// mapCameras es el orden en que se recorren los botones.
var mapCameras = []game.CameraID{"6", "8", "2", "4", "1B", "3", "5", "7A", "7B", "7C", "1A", "7D", "7E", "7F"}

// mapButtonRect devuelve el rectángulo (16x11) del botón de una cámara.
func mapButtonRect(id game.CameraID) rect {
	off := mapButtonOffsets[id]
	return rect{mapOrigin.x + off[0], mapOrigin.y + off[1], 16, 11}
}

// mapButtonAt devuelve el botón que está debajo del punto (px, py).
func mapButtonAt(px, py int) (game.CameraID, bool) {
	for _, id := range mapCameras {
		if mapButtonRect(id).contains(px, py) {
			return id, true
		}
	}
	return "", false
}

// buttonFrame: cuadro 0 normal, 1 cámara actual, 2 bloqueada.
func buttonFrame(current, locked bool) int {
	switch {
	case current:
		return 1
	case locked:
		return 2
	default:
		return 0
	}
}

// cameraForKey elige la cámara a la que lleva una tecla de número. La tecla
// "1" va a las cámaras que empiezan por 1 (1A, 1B); si ya estás en una, a la
// siguiente. Solo cuenta lo desbloqueado. ok es false si no hay ninguna.
func cameraForKey(digit byte, current game.CameraID, unlocked []game.CameraID) (game.CameraID, bool) {
	var candidates []game.CameraID
	for _, id := range mapCameras {
		if id[0] == digit && slices.Contains(unlocked, id) {
			candidates = append(candidates, id)
		}
	}
	slices.Sort(candidates)
	if len(candidates) == 0 {
		return "", false
	}
	if i := slices.Index(candidates, current); i >= 0 {
		return candidates[(i+1)%len(candidates)], true
	}
	return candidates[0], true
}

// cameraStamp es el texto del rótulo: "CAM 2 · 13:15".
func cameraStamp(cam game.CameraID, clock string) string {
	return fmt.Sprintf("CAM %s · %s", cam, clock)
}

// ---------------------------------------------------------------------------
// Etiqueta del paciente elegido
// ---------------------------------------------------------------------------

// Tamaño de la caja con el nombre y el nivel.
const (
	labelW = 117
	labelH = 14
)

// labelPosition coloca la caja de la etiqueta al lado del sprite donde
// quepa: izquierda, derecha, abajo y arriba, en ese orden. area es el
// rectángulo donde puede estar (la cámara). Si no cabe en ningún lado, se
// pega a la izquierda dentro del área.
func labelPosition(sprite rect, area rect) rect {
	const gap = 4
	candidates := []rect{
		{sprite.x - gap - labelW, sprite.y + sprite.h - labelH, labelW, labelH},
		{sprite.x + sprite.w + gap, sprite.y + sprite.h - labelH, labelW, labelH},
		{sprite.x + sprite.w/2 - labelW/2, sprite.y + sprite.h + gap, labelW, labelH},
		{sprite.x + sprite.w/2 - labelW/2, sprite.y - gap - labelH, labelW, labelH},
	}
	for _, c := range candidates {
		if c.x >= area.x && c.y >= area.y && c.x+c.w <= area.x+area.w && c.y+c.h <= area.y+area.h {
			return c
		}
	}
	return rect{area.x, min(sprite.y, area.y+area.h-labelH), labelW, labelH}
}

// ---------------------------------------------------------------------------
// Textos
// ---------------------------------------------------------------------------

// friendlyError convierte un error de game en el aviso de la vista: un
// título y las líneas de texto. El de "sin cama" sale de la vista aprobada.
func friendlyError(err error, patientName string) (title string, lines []string) {
	name := patientName
	if words := strings.Fields(patientName); len(words) > 0 {
		name = words[0]
	}
	switch {
	case errors.Is(err, hospital.ErrNoRoomAvailable):
		return "¡No hay cama libre!", []string{
			name + " se queda " + asleepWord(patientName) + " aquí mismo.",
			"Cuando se libere una cama, " + sendWord(patientName) + " usted.",
		}
	case errors.Is(err, game.ErrStaffBusy):
		return "Un momentico", []string{"Esa persona todavía anda ocupada."}
	case errors.Is(err, game.ErrStaffNoQuota):
		return "¡Cupo lleno!", []string{"Ese doctor ya tiene sus 4 pacientes."}
	case errors.Is(err, game.ErrShiftOver):
		return "Turno terminado", []string{"Ya no se puede hacer nada más."}
	default:
		return "Ya no se puede", []string{"Eso ya no se puede hacer ahora."}
	}
}

// asleepWord y sendWord conjugan según el género del paciente.
func asleepWord(fullName string) string {
	if isFemale(fullName) {
		return "dormida"
	}
	return "dormido"
}

func sendWord(fullName string) string {
	if isFemale(fullName) {
		return "mándela"
	}
	return "mándelo"
}

// demoStepTexts son los 4 pasos de la Demo, tal cual los aprobó la autora
// en la vista "Cam / Demo controles".
var demoStepTexts = [4]string{
	"Abrimos el turno: 2 doctores, 1 camillero, 3 habitaciones y 5 pacientes (3 severos).",
	"Cuatro pacientes se quedan dormidos en sitios distintos. Tres consiguen cama; al cuarto no le alcanza.",
	"Alguien se despierta: se le da de alta y el del pasillo pasa a esa cama.",
	"El reporte con las 4 consultas. ¡Así quedó el turno!",
}

// screenText prepara un texto para mostrarlo en pantalla: quita el "¡" de
// apertura y deja el "!" del cierre, porque en Pixelify Sans el ¡ se parece a
// una i (decisión de la autora, 2026-10-17). Las notas siguen con el español
// correcto; solo lo que se dibuja pierde el ¡.
func screenText(s string) string { return strings.ReplaceAll(s, "¡", "") }

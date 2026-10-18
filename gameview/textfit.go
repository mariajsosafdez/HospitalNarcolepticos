package gameview

import (
	"fmt"
	"strings"

	"costenos-narcolepsia/game"
)

// Ayudas puras para meter texto en cajas de tamaño fijo. Reciben la función
// que mide el texto (measure), así se prueban sin fuentes ni ventana.

// wrapText parte un texto en líneas que no pasen de maxW píxeles, cortando
// entre palabras. Si una palabra sola es más ancha que maxW, queda en su
// propia línea.
func wrapText(s string, maxW int, measure func(string) int) []string {
	var lines []string
	line := ""
	for _, word := range strings.Fields(s) {
		candidate := word
		if line != "" {
			candidate = line + " " + word
		}
		if line != "" && measure(candidate) > maxW {
			lines = append(lines, line)
			line = word
		} else {
			line = candidate
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// fitText recorta el texto con "..." para que no pase de maxW píxeles.
func fitText(s string, maxW int, measure func(string) int) string {
	if measure(s) <= maxW {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		candidate := strings.TrimRight(string(runes), " ") + "..."
		if measure(candidate) <= maxW {
			return candidate
		}
	}
	return "..."
}

// reportLine es una línea del reporte: los títulos de cada consulta van en negrita.
type reportLine struct {
	text string
	head bool
}

// roomStateText traduce el estado de una habitación del modelo.
func roomStateText(state string) string {
	switch state {
	case "Available":
		return "libre"
	case "Occupied":
		return "ocupada"
	default:
		return state
	}
}

// reportLines arma el reporte de las 4 consultas del enunciado en líneas de
// texto sencillas (provisional hasta que la vista del Reporte esté
// aprobada). Cada consulta lleva su título y sus datos.
func reportLines(r game.Report) []reportLine {
	var out []reportLine
	add := func(head bool, format string, args ...any) {
		out = append(out, reportLine{fmt.Sprintf(format, args...), head})
	}

	// Consulta 5.1: dormidos fuera de una cama (PatientsInHallway).
	add(true, "1. Dormidos fuera de una cama: %d", len(r.Hallway))
	for _, p := range r.Hallway {
		add(false, "   %s %s (%s)", p.ID, p.Name, p.Location)
	}

	// Consulta 5.2: episodios de cada doctor (MyEpisodes).
	add(true, "2. Episodios por doctor")
	for _, d := range r.Doctors {
		parts := make([]string, 0, len(d.Episodes))
		for _, e := range d.Episodes {
			where := "sin cama"
			if e.RoomNumber > 0 {
				where = fmt.Sprintf("hab. %d", e.RoomNumber)
			}
			parts = append(parts, fmt.Sprintf("%s %s", e.PatientID, where))
		}
		detail := "ninguno"
		if len(parts) > 0 {
			detail = strings.Join(parts, ", ")
		}
		add(false, "   %s (%d/%d): %s", displayName(d.ID, d.Name), d.Patients, game.DoctorQuota, detail)
	}

	// Consulta 5.3: estado de las camas.
	add(true, "3. Camas: %d libres de %d, sin cama %d veces", r.FreeRooms, len(r.Rooms), r.NoBedCount)
	for _, room := range r.Rooms {
		who := ""
		if len(room.Occupants) > 0 {
			who = " - " + strings.Join(room.Occupants, ", ")
		}
		add(false, "   Hab. %d %s%s", room.Number, roomStateText(room.State), who)
	}

	// Consulta 5.4: pacientes severos (SevereReport).
	add(true, "4. Pacientes severos")
	for _, s := range r.Severe {
		add(false, "   %s %s: %d episodio(s)", s.ID, s.Name, s.Episodes)
	}
	return out
}

// capLines deja como máximo max líneas; si sobran, la última dice cuántas.
func capLines(lines []reportLine, max int) []reportLine {
	if len(lines) <= max || max < 1 {
		return lines
	}
	out := append([]reportLine(nil), lines[:max-1]...)
	return append(out, reportLine{text: fmt.Sprintf("   ...y %d líneas más", len(lines)-(max-1))})
}

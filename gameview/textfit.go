package gameview

import "strings"

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

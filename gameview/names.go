package gameview

import "strings"

// staffText es el nombre y la especialidad que se MUESTRAN de un miembro del
// personal. El escenario del enunciado (la Demo) trae los datos en inglés
// ("Dr. Karen Ospina", "Neurology") y no se puede tocar, así que la vista
// busca el texto en español del Elenco por ID.
type staffText struct {
	name      string
	specialty string
	// shortName y shortRole son los de la tarjeta compacta del panel de 4 o
	// 5 personas (nota de Figma 18:655).
	shortName string
	shortRole string
}

var staffTexts = map[string]staffText{
	"D-001": {"Dra. Karen Ospina", "Neurología", "Dra. Karen", "Neurología"},
	"D-002": {"Dr. Álvaro Mendoza", "Medicina del sueño", "Dr. Álvaro", "Del sueño"},
	"D-003": {"Dra. Luz Dary Fontalvo", "Medicina general", "Dra. Luz Dary", "General"},
	"O-001": {"Chepe Barrios", "Camillero", "Chepe", "Camillero"},
	"O-002": {"El Mono Arrieta", "Camillero", "El Mono", "Camillero"},
}

// displayName devuelve el nombre para mostrar; si el ID no está en la tabla
// (por ejemplo un camillero) deja el nombre del modelo.
func displayName(id, modelName string) string {
	if t, ok := staffTexts[id]; ok {
		return t.name
	}
	return modelName
}

// displaySpecialty hace lo mismo con la especialidad.
func displaySpecialty(id, modelSpecialty string) string {
	if t, ok := staffTexts[id]; ok {
		return t.specialty
	}
	return modelSpecialty
}

// shortName devuelve el nombre corto de la tarjeta compacta; si el ID no
// está en la tabla, el nombre del modelo.
func shortName(id, modelName string) string {
	if t, ok := staffTexts[id]; ok {
		return t.shortName
	}
	return modelName
}

// shortRole devuelve el rol corto de la tarjeta compacta.
func shortRole(id, modelRole string) string {
	if t, ok := staffTexts[id]; ok {
		return t.shortRole
	}
	return modelRole
}

// patientShortNames son los nombres cortos de los pacientes para la etiqueta
// del paciente elegido (cabe en 60 px). El reporte usa el nombre completo.
// Los demás pacientes salen como primer nombre + primer apellido.
var patientShortNames = map[string]string{
	"P-007": "Mañe Julio",
	"P-005": "Marelvis", // "Marelvis Ortega" mide 64 px y no cabe
	"P-008": "Kelly Ruiz",
	// Easter eggs.
	"E-001": "Misifu",
	"E-002": "Alejo Urrego",
	"E-003": "Samu Giraldo",
	"E-004": "Cote",
	"E-005": "Sebs",
}

// patientLabelName devuelve el nombre corto de un paciente por su ID; si no
// está en la tabla, el primer nombre y el último apellido del nombre completo
// ("Yesid Pacheco" se queda igual y "Kelly Johana Ruiz" pasa a "Kelly Ruiz").
func patientLabelName(id, fullName string) string {
	if n, ok := patientShortNames[id]; ok {
		return n
	}
	words := strings.Fields(fullName)
	if len(words) <= 2 {
		return fullName
	}
	return words[0] + " " + words[len(words)-1]
}

package gameview

// staffText es el nombre y la especialidad que se MUESTRAN de un miembro del
// personal. El escenario del enunciado (la Demo) trae los datos en inglés
// ("Dr. Karen Ospina", "Neurology") y no se puede tocar, así que la vista
// busca el texto en español del Elenco por ID.
type staffText struct {
	name      string
	specialty string
}

var staffTexts = map[string]staffText{
	"D-001": {"Dra. Karen Ospina", "Neurología"},
	"D-002": {"Dr. Álvaro Mendoza", "Medicina del sueño"},
	"D-003": {"Dra. Luz Dary Fontalvo", "Medicina general"},
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

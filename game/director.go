package game

import (
	"math/rand"
	"strings"
)

// Textos de Don Ernesto Cassiani, el director. Salen palabra por palabra de
// docs/HospitalCosteños/Textos del Director.md. Aquí conservan los ¡ y ¿ del
// español correcto: la vista (gameview) es la que quita el ¡ al dibujar.

// DirectorName es quien habla en los mensajes.
const DirectorName = "Don Ernesto"

// TutorialLines son los 9 mensajes del tutorial del Día 1, uno por pantalla.
var TutorialLines = [...]string{
	"¡Ajá, bienvenido(a) a su primer turno! Yo soy Don Ernesto, el director de este hospital.",
	"Aquí los pacientes son costeños con narcolepsia: andan echando cuento y, ¡pum!, se quedan dormidos donde les dé la gana.",
	"Usted vigila las cámaras. Cambie de cámara con el mapa de abajo a la derecha.",
	"Cuando vea a alguien dormido, haga clic encima y después escoja a un doctor o a un camillero libre del panel.",
	"Ellos lo llevan a la primera cama libre. Después descansan un ratico, así que no se los gaste todos de una.",
	"Ojo: cada doctor puede tener máximo 4 pacientes al día. Si me lo llena, me lo quema, y eso cuesta puntos.",
	"Revise las habitaciones. Cuando alguien se despierte, dele de alta para que quede la cama libre.",
	"Si no hay cama, el paciente se queda dormido ahí mismo. Cuando se libere una, mándelo usted.",
	"El turno va de 12 del mediodía a 6 de la tarde. Mientras más rápido atienda, más estrellas. ¡Pilas pues!",
}

// dayIntros son los mensajes con que el director abre los días 2 a 5
// (índice 0 = día 2). El día 1 abre con el tutorial.
var dayIntros = [...]string{
	"Hoy abrimos el Pasillo Oeste y otra habitación. Más espacio, más gente durmiendo en él.",
	"Llegó la Dra. Luz Dary a ayudar. Y abrimos Radiología, donde la fila es larga y los párpados pesan.",
	"Contratamos al Mono Arrieta. No habla mucho, pero carga como tres. Hoy abre el Patio de las palmeras.",
	"Último día de la semana, y viene lleno. Usted ya sabe cómo es la vaina. ¡Confío en usted!",
}

// DayIntro devuelve el mensaje de inicio del día 2 a 5. Para el día 1 (que
// abre con el tutorial) o un día fuera de rango, ok es false.
func DayIntro(day int) (text string, ok bool) {
	if day < 2 || day > MaxDay {
		return "", false
	}
	return dayIntros[day-2], true
}

// Frases de fin de turno, según las estrellas (una al azar).
var (
	endLinesOneStar = []string{
		"¡Erda! Me dicen que había gente durmiendo en la cafetería encima del sancocho. ¿Usted estaba vigilando o también se durmió?",
		"Tres camas ocupadas por gente despierta pidiendo almuerzo, y usted mirando el patio. Así no, mi llave.",
		"Los doctores están más cansados que los pacientes. Si sigue así, el que va a necesitar cama soy yo.",
		"Esto parece la Batalla de Flores, pero sin flores y con gente dormida en el piso. Mañana mejoramos, ¿cierto?",
	}
	endLinesMiddle = []string{
		"Nada mal, nada mal. Pero ese paciente del pasillo todavía me tiene pensando.",
		"Buen trabajo. Con un poquito más de pilas, mañana sale perfecto.",
	}
	endLinesFiveStars = []string{
		"¡Qué nota! Ni un paciente en el piso. Usted vale más que un ventilador en agosto.",
		"Excelente turno. Ya le dije a la cocina que le guarden una presa grande del sancocho.",
	}
)

// endLinesFor devuelve la lista de frases que le toca a esa cantidad de
// estrellas: 1, de 2 a 4, o 5.
func endLinesFor(stars int) []string {
	switch {
	case stars <= 1:
		return endLinesOneStar
	case stars >= 5:
		return endLinesFiveStars
	default:
		return endLinesMiddle
	}
}

// EndLine escoge la frase del director para el reporte de fin de turno. La
// elección es al azar PERO con la semilla del turno: la misma semilla y las
// mismas estrellas dan siempre la misma frase (así se puede probar).
func EndLine(stars int, seed int64) string {
	lines := endLinesFor(stars)
	return lines[rand.New(rand.NewSource(seed)).Intn(len(lines))]
}

// OverworkLine es el mensaje cuando un doctor llega a 4/4 pacientes.
// doctorName es el nombre completo con Dr./Dra.
func OverworkLine(doctorName string) string {
	return strings.Replace("¡Oiga! {doctor} ya tiene cuatro pacientes a cargo. Repártame el trabajo, que con este calor nadie aguanta tanto.", "{doctor}", doctorName, 1)
}

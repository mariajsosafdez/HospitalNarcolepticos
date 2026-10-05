package hospital

// Attender es cualquier miembro del personal que puede atender a un paciente
// que se quedó dormido.
//
// Una interfaz en Go es un CONJUNTO DE MÉTODOS. Cualquier tipo que tenga esos
// métodos la cumple automáticamente: no existe la palabra "implements".
// Doctor y Orderly la cumplen, y por eso el hospital puede guardarlos juntos
// en un mismo []Attender y despachar a cualquiera sin saber qué tipo concreto
// es (POLIMORFISMO). Si mañana se agrega un Nurse con estos 5 métodos, el
// hospital lo puede usar sin cambiar una sola línea de su lógica.
type Attender interface {
	// ID y Name los obtienen Doctor y Orderly "gratis" del Person embebido.
	ID() string
	Name() string
	// Role describe el tipo de personal ("Doctor", "Orderly"...).
	Role() string
	// IsAvailable dice si puede atender una emergencia en este momento.
	IsAvailable() bool
	// Attend atiende al paciente que se durmió en location y devuelve el
	// registro del episodio, o un error si no lo pudo atender.
	Attend(p *Patient, location string) (EpisodeRecord, error)
}

// Verificación en tiempo de compilación: estas dos líneas no hacen nada al
// ejecutar, pero si Doctor u Orderly dejaran de cumplir Attender (por ejemplo,
// si alguien borra su método Role), el programa NO compilaría y el error
// aparecería aquí mismo. El "_" significa que la variable no se usa.
var (
	_ Attender = (*Doctor)(nil)
	_ Attender = (*Orderly)(nil)
)

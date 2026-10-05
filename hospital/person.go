// Package hospital contiene el modelo de negocio del Hospital de los Costeños
// con Narcolepsia: pacientes, personal, cuartos y episodios de sueño.
//
// Este paquete NO imprime nada ni sabe nada de interfaces gráficas: solo
// guarda el estado y responde consultas devolviendo datos. La impresión
// ocurre en el paquete main y la vista web en el paquete web.
package hospital

// Person agrupa los datos de identidad que comparten pacientes y personal.
//
// En Go no existe la herencia. En su lugar usamos COMPOSICIÓN: Person se
// "embebe" (se pone sin nombre de campo) dentro de Patient, Doctor y Orderly.
// Así esos tipos reutilizan los campos y los métodos de Person sin copiarlos.
//
// Los campos van en minúscula: son "no exportados", es decir, solo el código
// dentro del paquete hospital puede leerlos o modificarlos. Desde main (u otro
// paquete) el compilador no deja escribir p.name; hay que usar p.Name().
type Person struct {
	id   string
	name string
	age  int
}

// newPerson es un constructor privado (minúscula) que usan los constructores
// públicos NewPatient, NewDoctor y NewOrderly. Valida los datos básicos
// y devuelve un error en lugar de crear una persona inválida.
func newPerson(id, name string, age int) (Person, error) {
	if id == "" || name == "" {
		return Person{}, ErrInvalidData
	}
	if age < 0 {
		return Person{}, ErrInvalidData
	}
	return Person{id: id, name: name, age: age}, nil
}

// Los siguientes métodos son "getters": la única forma de leer los campos
// privados desde fuera del paquete.
//
// Usan receptor por VALOR (p Person) porque solo leen y Person es pequeño.
// Cuando Person está embebido en Patient, Go "promueve" estos métodos:
// se puede escribir patient.Name() directamente, como si fuera de Patient.

// ID devuelve el identificador de la persona (por ejemplo "P-001").
func (p Person) ID() string { return p.id }

// Name devuelve el nombre completo de la persona.
func (p Person) Name() string { return p.name }

// Age devuelve la edad de la persona en años.
func (p Person) Age() int { return p.age }

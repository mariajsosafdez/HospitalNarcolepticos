# Guía de estudio para la sustentación

Esta guía está pensada para repasar el proyecto antes de la defensa del 20/10/2026. Va en el mismo orden en que conviene estudiarlo:

1. Primero el mapa del proyecto.
2. Luego el recorrido de un ataque de sueño.
3. Después las preguntas típicas, con su respuesta.

---

## 1. Mapa rápido

| Archivo | Qué hay | Concepto de POO en Go |
|---|---|---|
| `hospital/person.go` | `Person` con `id`, `name`, `age` y sus getters | Encapsulamiento (minúsculas). Es la base de la composición |
| `hospital/patient.go` | `Patient`, `NarcolepsyLevel`, `PatientState` | Embebido de `Person`, constantes tipadas con `iota`, receptores puntero |
| `hospital/doctor.go` | `Doctor` | Cumple `Attender`. Guarda sus propios episodios |
| `hospital/orderly.go` | `Orderly` (camillero) | Segundo tipo que cumple `Attender`. **No** diagnostica |
| `hospital/attender.go` | `interface Attender` | Polimorfismo y verificación en compilación (`var _ Attender = ...`) |
| `hospital/room.go` | `Room`, `RoomState` | `Occupy` y `Release` mantienen sincronizados el cuarto y el paciente |
| `hospital/episode.go` | `EpisodeRecord` | Valor inmutable, `time.Time`, contador `atomic` |
| `hospital/errors.go` | `ErrNoRoomAvailable`, etc. | Errores como valores, `%w`, `errors.Is` |
| `hospital/hospital.go` | `Hospital` y las 4 consultas | `[]Attender`, round-robin, `sync.Mutex` |
| `hospital/snapshot.go` | `Snapshot()` | Lectura segura en concurrencia |
| `hospital/simulation.go` | `Simulate()` | Goroutines, `sync.WaitGroup` |
| `main.go` | Escenario e impresión | No tiene lógica de negocio |
| `web/` | API REST + HTML/JS | Separación lógica/vista |

---

## 2. Recorrido de un ataque de sueño (`RegisterEpisode`)

```
main: h.RegisterEpisode(p, "cafeteria")
  └─ Hospital (toma h.mu)
       ├─ valida: p no es nil, está admitido y está despierto
       ├─ 1. p.SufferSleepAttack(location)       → p.state = AsleepInHallway
       ├─ 2. h.assignRoomLocked(p)
       │      └─ primer cuarto con IsAvailable()
       │           └─ room.Occupy(p)              → agrega a occupants
       │                └─ p.moveToBed(room)      → p.state = AsleepInBed, p.room = room
       │      (si no hay cuarto → ErrNoRoomAvailable, el paciente sigue en el pasillo)
       ├─ 3. h.attend(p, location)
       │      └─ dispatchAttender()               → siguiente Attender disponible (round-robin)
       │           └─ attender.Attend(p, loc)     → LLAMADA POLIMÓRFICA
       │                ├─ Doctor.Attend: valida, diagnostica si hace falta, guarda el episodio
       │                └─ Orderly.Attend: valida y guarda el episodio (no diagnostica)
       ├─ 4. h.history = append(h.history, record)
       └─ return errors.Join(roomErr, staffErr)   (nil si todo salió bien)
```

Con los datos del escenario, el round-robin (Karen → Álvaro → Chepe → Karen) produce este resultado:

| Paciente | Lo atiende | Cuarto |
|---|---|---|
| P-001 | Karen | 101 |
| P-002 | Álvaro | 102 |
| P-003 | Chepe (camillero) | 103 |
| P-004 | Karen | ninguno: error y se queda en el pasillo |

Después P-001 se despierta y P-004 pasa al cuarto 101.

---

## 3. Preguntas de la sección 8.2, respondidas

### 3.1 "Agregue un Nurse que cumpla Attender y haga que el hospital lo despache, en vivo"

**Paso 1.** Crear el archivo `hospital/nurse.go`. Va dentro del paquete `hospital`, porque usa `newPerson`, `newEpisodeRecord` y `p.state`, que son privados:

```go
package hospital

import "fmt"

// Nurse es una enfermera: atiende emergencias pero no diagnostica.
type Nurse struct {
	Person              // embebido: ID() y Name() vienen gratis
	shift    string
	episodes []EpisodeRecord
}

func NewNurse(id, name string, age int, shift string) (*Nurse, error) {
	person, err := newPerson(id, name, age)
	if err != nil {
		return nil, fmt.Errorf("new nurse %q: %w", id, err)
	}
	return &Nurse{Person: person, shift: shift}, nil
}

func (n *Nurse) Role() string      { return "Nurse" }
func (n *Nurse) IsAvailable() bool { return true }

func (n *Nurse) Attend(p *Patient, location string) (EpisodeRecord, error) {
	if p == nil {
		return EpisodeRecord{}, fmt.Errorf("nurse attend: %w", ErrInvalidData)
	}
	if p.state == Awake {
		return EpisodeRecord{}, fmt.Errorf("nurse attend %s: %w", p.id, ErrPatientAwake)
	}
	record := newEpisodeRecord(p, n, location)
	n.episodes = append(n.episodes, record)
	return record, nil
}
```

**Paso 2.** En `buildHospital()` de `main.go`, contratarla:

```go
nurse, err := hospital.NewNurse("N-001", "Yuli Arrieta", 29, "night")
if err != nil {
	return nil, err
}
if err := h.HireStaff(nurse); err != nil {
	return nil, err
}
```

**Qué decir:**
- **No se tocó ni una línea de `Hospital`.** `HireStaff` recibe un `Attender`, y `dispatchAttender` solo llama `IsAvailable()` y `Attend()`. El hospital despacha a la enfermera sin saber que existe el tipo `Nurse`. Eso es polimorfismo.
- Para que el compilador verifique el tipo, se puede agregar `_ Attender = (*Nurse)(nil)` en `attender.go`.
- Este ejemplo ya se probó: compila y el tercer episodio lo atiende la `Nurse`.
- **Ojo:** si se agrega en `buildHospital`, cambia el orden del round-robin del escenario. P-003 lo atendería Chepe y P-004 la enfermera, y la salida cambia. Eso es esperado.

### 3.2 "¿Qué cambia si Person se embebe por valor o como puntero?"

Hoy está **por valor**: `type Patient struct { Person; ... }`.

| | Por valor (`Person`) | Por puntero (`*Person`) |
|---|---|---|
| Dónde viven los datos | Dentro del mismo struct `Patient` | En otro lugar de la memoria; `Patient` solo guarda la dirección |
| Valor cero | `Patient{}` tiene un `Person` vacío y usable | `Patient{}` tiene `Person == nil`, así que `p.Name()` **entra en pánico** (nil pointer dereference) |
| Al copiar un `Patient` | Cada copia tiene su propio `Person` | Las copias **comparten** el mismo `Person`; cambiar uno cambia los dos |
| Constructor | `Person: person` | `Person: &person`, y hay que acordarse de inicializarlo |
| Métodos promovidos | `ID()`, `Name()`, `Age()` | Los mismos (y también los de receptor puntero de `Person`) |

**Por qué elegimos valor:** `Person` es pequeño, nunca cambia y no tiene sentido compartirlo entre dos pacientes. Embeberlo por valor es más simple y elimina el riesgo de un puntero nil.

### 3.3 "Cambie un receptor de puntero a valor y prediga qué deja de funcionar"

**Ejemplo probado en este proyecto:** cambiar `func (p *Patient) SufferSleepAttack` por `func (p Patient) SufferSleepAttack`.

- **Compila sin errores y `go vet` no dice nada**, pero el método trabaja sobre una **copia** del paciente. Cambia `state` y `currentLocation` de la copia, y la copia se descarta al terminar. El paciente original **sigue despierto**.
- Consecuencias en cadena:
  - `assignRoomLocked` ve el paciente despierto y devuelve `ErrNotInHallway`.
  - `Doctor.Attend` devuelve `ErrPatientAwake`.
  - Nadie queda en el pasillo.
- **Fallan 5 de las 8 pruebas.** Así se comprobó, por ejemplo: `expected a room, got error: ... patient is not asleep in a hallway`.

**Regla general:**
- Receptor **puntero** (`*T`): el método puede modificar el original y no copia el struct. Se usa cuando el método cambia algo o el struct es grande.
- Receptor **valor** (`T`): el método recibe una copia. Sirve para métodos que solo leen datos pequeños, como `Person.ID()` o `EpisodeRecord.Summary()`.

**Efecto sobre las interfaces:**
- Si un método tiene receptor puntero, **solo `*T`** cumple la interfaz, no `T`. Por eso guardamos `*Doctor` en el `[]Attender`.
- Si `Role()` de `Doctor` pasara a receptor valor, `*Doctor` **seguiría** cumpliendo `Attender`, porque el conjunto de métodos de `*T` incluye también los de `T`.

### 3.4 "¿Por qué un campo está en minúscula y qué impide que el código de afuera lo toque?"

En Go la visibilidad depende de la **primera letra del nombre**:
- Mayúscula: exportado, visible desde otros paquetes.
- Minúscula: no exportado, solo visible **dentro del mismo paquete**.

Si en `main.go` alguien escribe `p.state = hospital.AsleepInBed`, **el compilador** rechaza el programa. Este es el mensaje real, probado en el proyecto:

```
p.state undefined (type *hospital.Patient has no field or method state, but does have method State)
```

Desde fuera del paquete el campo `state` es invisible, como si no existiera, y Go sugiere usar el getter `State()`. No es una convención: lo impide el compilador.

**Por qué importa:**
- Nadie puede poner un paciente en `AsleepInBed` sin darle cuarto.
- Nadie puede meter 3 pacientes en un cuarto de capacidad 1.
- Los cambios solo pasan por métodos que validan (`Occupy`, `WakeUp`, …) y que mantienen la coherencia.

**Detalle:** dentro del paquete `hospital` sí se pueden tocar los campos de otros structs. Por ejemplo, `Doctor.DiagnosePatient` hace `p.assignedDoctor = d`. El límite de la encapsulación en Go es el **paquete**, no el struct.

### 3.5 "¿Qué hace el programa si dos goroutines asignan el mismo cuarto al mismo tiempo?"

1. `AssignRoom` empieza con `h.mu.Lock()`. El mutex deja entrar a **una sola goroutine a la vez**; la otra espera en `Lock()` hasta que la primera haga `Unlock()`.
2. La primera revisa el cuarto, lo ve libre y lo ocupa: queda `Occupied`.
3. La segunda entra después, revisa de nuevo **dentro del candado** y ve el cuarto lleno. Si no hay otro libre, recibe `ErrNoRoomAvailable` y su paciente sigue en el pasillo.
4. También se revisa el estado del paciente dentro del candado. Si dos goroutines intentan acostar **al mismo paciente**, la segunda recibe `ErrNotInHallway`, así que nunca se le asignan dos camas.

**Sin el mutex** las dos podrían leer "libre" al mismo tiempo y meter dos pacientes en una cama. Eso es una **condición de carrera**, y además una *data race*: dos goroutines tocan la misma memoria sin sincronización. El detector `-race` la reportaría.

**Esto está comprobado en el proyecto:** `TestSimulationIsRaceFree` lanza 8 goroutines que compiten por 3 cuartos mientras otra goroutine toma `Snapshot()` sin parar. Al final verifica que ningún cuarto tenga más pacientes que camas.

### 3.6 "Justifique decisiones de diseño"

- **¿Por qué slice y no map para pacientes, cuartos y personal?**
  - Importa el **orden**: el "primer cuarto disponible", el orden de admisión y el turno del round-robin.
  - Un map en Go **no tiene orden**: cada recorrido puede salir distinto.
  - Con pocos elementos, buscar recorriendo el slice es simple y rápido.
- **¿Por qué `SevereReport` devuelve un map?**
  - Lo exige el enunciado, y además es natural: es una relación "paciente → cantidad".
  - Con `report[p]++` se suma por clave directamente.
  - El "comma ok" (`v, ok := report[p]`) distingue "no está" de "está con 0".
- **¿Por qué `MyEpisodes` devuelve `[]EpisodeRecord` y no punteros?**
  - El registro es inmutable. Devolver valores (y una copia del slice con `slices.Clone`) evita que alguien de afuera modifique la historia del doctor.
- **¿Por qué `AssignRoom` devuelve `(*Room, error)`?**
  - Puede fallar de forma previsible (no hay cuarto). La convención de Go es devolver el error como último valor, sin usar `panic`.
  - Devuelve el puntero para que quien llama vea el cuarto real, no una copia.
- **¿Por qué esa división de paquetes?**
  - `hospital` es el modelo: lógica pura, sin impresión, testeable.
  - `main` arma el escenario e imprime.
  - `web` es la vista.
  - Agregar la web **no cambió el paquete hospital**: `web` solo usa su API pública.
- **¿Por qué `EpisodeRecord` guarda el número de cuarto y no `*Room`?** Es una foto del momento. P-004 se durmió sin cuarto y luego recibió el 101; su registro debe seguir diciendo "stayed in the hallway".
- **¿Por qué `errors.Join` en `RegisterEpisode`?** Pueden fallar dos cosas a la vez (sin cuarto y sin personal). `Join` las devuelve juntas, y `errors.Is` sigue encontrando cada una.
- **¿Por qué un type assertion en `HireStaff`?** Solo para guardar los doctores también en `h.doctors` y poder hacer la consulta 5.2. El despacho de emergencias **no** lo usa: ahí todo es por interfaz.

---

## 4. Glosario de concurrencia (bono +5 %)

| Concepto | En el proyecto | Explicación corta |
|---|---|---|
| **goroutine** | `go func(i int, p *Patient) {...}(i, p)` en `Simulate` | Función que corre "en paralelo". Es muy liviana; se pueden lanzar miles |
| **`sync.WaitGroup`** | `wg.Add(1)`, `defer wg.Done()`, `wg.Wait()` | Contador para esperar a que terminen varias goroutines |
| **`sync.Mutex`** | `h.mu.Lock()` y `defer h.mu.Unlock()` en cada método público de `Hospital` | Candado: solo una goroutine a la vez entra a la sección protegida |
| **`defer`** | `defer h.mu.Unlock()` | Ejecuta la línea al **salir** de la función, aunque se salga por un `return` temprano. Así nunca se olvida soltar el candado |
| **Deadlock / mutex no reentrante** | Patrón público vs. `...Locked` | Si `RegisterEpisode` (que ya tiene el candado) llamara al `AssignRoom` público, este intentaría tomar el candado otra vez y se quedaría esperando para siempre. Por eso existe `assignRoomLocked` |
| **`atomic`** | `episodeCounter atomic.Int64`, `simRunning atomic.Bool` | Operaciones indivisibles sobre un número o booleano, sin candado |
| **Data race** | Lo que evitamos | Dos goroutines acceden a la misma variable a la vez y al menos una escribe |
| **`-race`** | `go test -race ./...` | Detector de Go que avisa `WARNING: DATA RACE` si encuentra una carrera. En Windows necesita gcc |
| **Snapshot** | `h.Snapshot()` | Copia de todo el estado hecha bajo el candado. La web lee la copia y no los objetos vivos |
| **Sin memoria compartida** | `results[i]` en `Simulate` | Cada goroutine escribe solo en su posición del slice; como nadie comparte esa memoria, no hace falta candado |
| **Pasar parámetros a la goroutine** | `}(i, p)` | Cada goroutine recibe su propio `i` y su propio `p` (desde Go 1.22 el `for` ya crea variables nuevas, pero así es explícito) |

**Pregunta trampa posible:** "¿El servidor web es concurrente?"
Sí. `net/http` atiende **cada petición en su propia goroutine**. Por eso el hospital necesita el mutex incluso sin la simulación: dos pestañas del navegador pueden pedir cosas al mismo tiempo.

---

## 5. Otras preguntas probables

- **¿Dónde está la herencia?** No existe en Go. Se usa **composición** (embeber `Person`) para reutilizar datos y métodos, e **interfaces** para el polimorfismo.
- **¿Cómo sabe Go que `Doctor` cumple `Attender` si nunca dice `implements`?** Las interfaces se cumplen **implícitamente**: basta con tener los métodos. La línea `var _ Attender = (*Doctor)(nil)` hace que el compilador lo verifique.
- **¿Qué es `iota`?** Un contador que vale 0, 1, 2… dentro de un bloque `const`. Con él se crean las constantes tipadas de los estados.
- **¿Para qué sirve `String()`?** Hace que el tipo cumpla `fmt.Stringer`, así que `fmt.Print(estado)` muestra `"Asleep in hallway"` en vez de `1`.
- **¿Qué es `%w`?** Envuelve un error dentro de otro con más contexto, sin perder el original. `errors.Is` lo encuentra aunque esté envuelto.
- **¿Por qué `runScenario` recibe un `io.Writer`?** Así la misma función escribe a la consola (`os.Stdout`), a un buffer para la web (`io.MultiWriter`) o a ninguna parte (`io.Discard`). Es otro ejemplo de interfaz.
- **¿Qué hace `//go:embed static`?** Mete los archivos HTML/CSS/JS dentro del ejecutable al compilar, para que `go run .` funcione sin copiar archivos aparte.
- **¿Por qué los structs de `Snapshot` tienen campos exportados si "todo debe ser privado"?** No son entidades: no tienen reglas que proteger. Son paquetes de datos de solo lectura (copias). Las entidades (`Patient`, `Room`, …) sí tienen todos sus campos privados.

---

## 6. Comandos para tener a mano

```
go run .                 # escenario + simulación + web (http://localhost:8080)
go test ./...            # pruebas
go test -v ./...         # pruebas, mostrando cada una
go test -race ./...      # pruebas con el detector de carreras (necesita gcc)
go run -race .           # el programa con el detector de carreras
gofmt -l .               # lista archivos sin formatear (debe salir vacío)
go vet ./...             # análisis estático (debe salir vacío)
```

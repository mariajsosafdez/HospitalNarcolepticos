package hospital

import (
	"errors"
	"math/rand/v2"
	"sync"
	"time"
)

// SimulationConfig configura la simulación concurrente.
type SimulationConfig struct {
	Rounds    int           // cuántas veces "actúa" cada paciente
	MaxDelay  time.Duration // espera máxima (aleatoria) antes de cada acción
	Locations []string      // lugares donde se pueden dormir; vacío = lugares por defecto
}

// SimulationResult resume lo que pasó durante la simulación.
type SimulationResult struct {
	Goroutines  int // cuántas goroutines se lanzaron (una por paciente)
	Episodes    int // ataques de sueño registrados
	NoRoom      int // veces que el paciente se quedó en el pasillo por falta de cuarto
	WakeUps     int // veces que un paciente se despertó
	Relocations int // pacientes movidos del pasillo a una cama que se liberó
}

// defaultLocations son lugares típicos donde se duermen los pacientes.
var defaultLocations = []string{
	"cafeteria (sancocho line)",
	"champeta dance hall",
	"radiology queue",
	"main entrance",
	"hallway 2, second floor",
	"waiting room hammock",
	"vending machine",
}

// Simulate lanza UNA GOROUTINE POR PACIENTE. Cada goroutine es como un
// paciente "vivo" que, a intervalos aleatorios, se duerme o se despierta.
// Todas corren AL MISMO TIEMPO y usan el mismo hospital; el candado del
// hospital (h.mu) evita que se pisen. Simulate espera a que terminen todas
// y devuelve el resumen.
//
// Conceptos de concurrencia que aparecen aquí:
//   - goroutine: una función que corre en paralelo, se lanza con "go f()".
//   - sync.WaitGroup: un contador para esperar a que terminen las goroutines.
//     Add(1) antes de lanzar cada una, Done() cuando termina, Wait() espera
//     hasta que el contador vuelva a 0.
//   - sync.Mutex (dentro de Hospital): solo una goroutine a la vez puede
//     modificar el hospital.
func (h *Hospital) Simulate(cfg SimulationConfig) SimulationResult {
	// Valores por defecto si la configuración viene incompleta.
	if cfg.Rounds <= 0 {
		cfg.Rounds = 5
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 100 * time.Millisecond
	}
	if len(cfg.Locations) == 0 {
		cfg.Locations = defaultLocations
	}

	patients := h.Patients() // copia del slice, tomada con el candado

	// Cada goroutine escribe SOLO en su propia posición results[i].
	// Como ninguna escribe en la posición de otra, no comparten memoria
	// y no hace falta candado para esto. Al final sumamos todo.
	results := make([]SimulationResult, len(patients))

	var wg sync.WaitGroup
	for i, p := range patients {
		wg.Add(1) // "hay una goroutine más por esperar"

		// Pasamos i y p como PARÁMETROS a la goroutine para que cada una
		// reciba su propio paciente y su propia posición.
		go func(i int, p *Patient) {
			defer wg.Done() // "esta goroutine terminó" (se ejecuta al salir)
			results[i] = h.simulatePatient(p, cfg)
		}(i, p)
	}
	wg.Wait() // bloquea aquí hasta que TODAS hayan llamado Done()

	total := SimulationResult{Goroutines: len(patients)}
	for _, r := range results {
		total.Episodes += r.Episodes
		total.NoRoom += r.NoRoom
		total.WakeUps += r.WakeUps
		total.Relocations += r.Relocations
	}
	return total
}

// simulatePatient es lo que hace cada goroutine: varias rondas en las que
// espera un tiempo aleatorio y luego se duerme (si está despierto) o se
// despierta (si está dormido). Solo usa métodos PÚBLICOS del hospital, que
// toman el candado; nunca toca los campos del paciente directamente.
func (h *Hospital) simulatePatient(p *Patient, cfg SimulationConfig) SimulationResult {
	var r SimulationResult
	for round := 0; round < cfg.Rounds; round++ {
		// rand.N devuelve un número aleatorio entre 0 y MaxDelay.
		// Es seguro llamarlo desde varias goroutines a la vez.
		time.Sleep(rand.N(cfg.MaxDelay))

		if h.StateOf(p) == Awake {
			location := cfg.Locations[rand.IntN(len(cfg.Locations))]
			err := h.RegisterEpisode(p, location)
			if errors.Is(err, ErrAlreadyAsleep) {
				continue // alguien más (por ejemplo la web) lo durmió primero
			}
			r.Episodes++
			if errors.Is(err, ErrNoRoomAvailable) {
				r.NoRoom++
			}
			continue
		}

		if err := h.WakeUpPatient(p); err == nil {
			r.WakeUps++
		}
		// Quizás se liberó una cama: intentamos acostar a los del pasillo.
		// Otra goroutine podría estar intentando lo mismo AHORA MISMO con el
		// mismo paciente o el mismo cuarto. No pasa nada: AssignRoom vuelve a
		// revisar todo dentro del candado, así que solo una lo logra y la otra
		// recibe un error (ErrNoRoomAvailable o ErrNotInHallway).
		for _, waiting := range h.PatientsInHallway() {
			if _, err := h.AssignRoom(waiting); err == nil {
				r.Relocations++
			}
		}
	}
	return r
}

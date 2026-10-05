// Package web es la VISTA del proyecto: un servidor HTTP (librería estándar
// net/http) que expone el hospital como una API REST en JSON y sirve una
// página HTML/CSS/JS sencilla.
//
// Este paquete solo LLAMA métodos públicos del paquete hospital. El paquete
// hospital no sabe que este existe: esa es la separación lógica/vista que
// pide el bono.
package web

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"costenos-narcolepsia/hospital"
)

// staticFiles contiene la carpeta static/ METIDA DENTRO del ejecutable.
// La línea //go:embed (sin espacio después de //) es una instrucción para el
// compilador: así `go run .` funciona sin copiar archivos aparte.
//
//go:embed static
var staticFiles embed.FS

// HospitalFactory es una función que crea un hospital nuevo con los datos
// iniciales. La recibe el servidor para poder "reiniciar" el hospital vivo.
type HospitalFactory func() (*hospital.Hospital, error)

// Server guarda los dos hospitales que muestra la página:
//   - scenario: el del escenario obligatorio (pestaña 1). Nunca se modifica.
//   - live: el de las pestañas 2 y 3, que el usuario y la simulación cambian.
type Server struct {
	scenario    *hospital.Hospital
	scenarioLog string
	factory     HospitalFactory

	// mu protege live y lastSim, porque el botón Reset reemplaza el hospital
	// y la simulación guarda su resultado desde otra goroutine.
	// (El hospital en sí ya se protege con su propio candado interno.)
	mu      sync.Mutex
	live    *hospital.Hospital
	lastSim *hospital.SimulationResult

	// simRunning es un booleano atómico: se puede leer y cambiar desde varias
	// goroutines sin candado. Evita lanzar dos simulaciones a la vez.
	simRunning atomic.Bool
}

// NewServer crea el servidor. El hospital vivo se crea con factory.
func NewServer(scenario *hospital.Hospital, scenarioLog string, factory HospitalFactory) (*Server, error) {
	live, err := factory()
	if err != nil {
		return nil, fmt.Errorf("create live hospital: %w", err)
	}
	return &Server{scenario: scenario, scenarioLog: scenarioLog, factory: factory, live: live}, nil
}

// Handler arma las rutas. Desde Go 1.22 el patrón puede incluir el método
// HTTP ("POST ...") y comodines ({id}), que luego se leen con r.PathValue.
func (s *Server) Handler() (http.Handler, error) {
	static, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServerFS(static))

	mux.HandleFunc("GET /api/scenario", s.handleScenario)
	mux.HandleFunc("GET /api/state", s.handleState)
	mux.HandleFunc("POST /api/patients", s.handleAdmitPatient)
	mux.HandleFunc("POST /api/doctors", s.handleHireDoctor)
	mux.HandleFunc("POST /api/orderlies", s.handleHireOrderly)
	mux.HandleFunc("POST /api/rooms", s.handleAddRoom)
	mux.HandleFunc("POST /api/patients/{id}/sleep", s.handleSleep)
	mux.HandleFunc("POST /api/patients/{id}/wake", s.handleWake)
	mux.HandleFunc("POST /api/patients/{id}/assign-room", s.handleAssignRoom)
	mux.HandleFunc("POST /api/simulation", s.handleSimulation)
	mux.HandleFunc("POST /api/reset", s.handleReset)
	return mux, nil
}

// Start levanta el servidor en addr (por ejemplo "localhost:8080").
// ListenAndServe se queda escuchando para siempre (hasta Ctrl+C); cada
// petición la atiende net/http en SU PROPIA goroutine, por eso el hospital
// necesita su candado.
func Start(addr string, s *Server) error {
	handler, err := s.Handler()
	if err != nil {
		return err
	}
	return http.ListenAndServe(addr, handler)
}

// liveHospital devuelve el hospital vivo actual (con el candado del servidor).
func (s *Server) liveHospital() *hospital.Hospital {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.live
}

// ---------------------------------------------------------------------------
// Lectura
// ---------------------------------------------------------------------------

func (s *Server) handleScenario(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, scenarioDTO{
		Log:   s.scenarioLog,
		State: toStateDTO(s.scenario.Snapshot()),
	})
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	state := toStateDTO(s.liveHospital().Snapshot())

	s.mu.Lock()
	if s.lastSim != nil {
		state.Simulation.Last = toResultDTO(*s.lastSim)
	}
	s.mu.Unlock()
	state.Simulation.Running = s.simRunning.Load()

	writeJSON(w, http.StatusOK, state)
}

// ---------------------------------------------------------------------------
// Registro
// ---------------------------------------------------------------------------

func (s *Server) handleAdmitPatient(w http.ResponseWriter, r *http.Request) {
	var req admitRequest
	if !readJSON(w, r, &req) {
		return
	}
	level, err := parseLevel(req.Level)
	if err != nil {
		writeError(w, err)
		return
	}
	p, err := hospital.NewPatient(req.ID, req.Name, req.Age, level)
	if err == nil {
		err = s.liveHospital().AdmitPatient(p)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeMessage(w, fmt.Sprintf("%s %s was admitted", p.ID(), p.Name()))
}

func (s *Server) handleHireDoctor(w http.ResponseWriter, r *http.Request) {
	var req staffRequest
	if !readJSON(w, r, &req) {
		return
	}
	d, err := hospital.NewDoctor(req.ID, req.Name, req.Age, req.Specialty)
	if err == nil {
		err = s.liveHospital().HireDoctor(d)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeMessage(w, fmt.Sprintf("%s was hired as a doctor", d.Name()))
}

func (s *Server) handleHireOrderly(w http.ResponseWriter, r *http.Request) {
	var req staffRequest
	if !readJSON(w, r, &req) {
		return
	}
	o, err := hospital.NewOrderly(req.ID, req.Name, req.Age)
	if err == nil {
		err = s.liveHospital().HireStaff(o)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeMessage(w, fmt.Sprintf("%s was hired as an orderly", o.Name()))
}

func (s *Server) handleAddRoom(w http.ResponseWriter, r *http.Request) {
	var req roomRequest
	if !readJSON(w, r, &req) {
		return
	}
	room, err := hospital.NewRoom(req.Number, req.Capacity)
	if err == nil {
		err = s.liveHospital().AddRoom(room)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeMessage(w, fmt.Sprintf("room %d was added", room.Number()))
}

// ---------------------------------------------------------------------------
// Acciones sobre un paciente
// ---------------------------------------------------------------------------

func (s *Server) handleSleep(w http.ResponseWriter, r *http.Request) {
	var req sleepRequest
	if !readJSON(w, r, &req) {
		return
	}
	h := s.liveHospital()
	p, err := h.FindPatient(r.PathValue("id")) // {id} de la ruta
	if err != nil {
		writeError(w, err)
		return
	}
	err = h.RegisterEpisode(p, req.Location)
	switch {
	case errors.Is(err, hospital.ErrNoRoomAvailable):
		// No es un fallo de la petición: el episodio se registró, pero el
		// paciente se quedó en el pasillo. Se responde 200 con una advertencia.
		writeJSON(w, http.StatusOK, map[string]string{
			"warning": fmt.Sprintf("%s fell asleep at %s: no room available, the patient stays in the hallway", p.Name(), req.Location),
		})
	case err != nil:
		writeError(w, err)
	default:
		writeMessage(w, fmt.Sprintf("%s fell asleep at %s and was taken to a bed", p.Name(), req.Location))
	}
}

func (s *Server) handleWake(w http.ResponseWriter, r *http.Request) {
	h := s.liveHospital()
	p, err := h.FindPatient(r.PathValue("id"))
	if err == nil {
		err = h.WakeUpPatient(p)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeMessage(w, fmt.Sprintf("%s woke up", p.Name()))
}

func (s *Server) handleAssignRoom(w http.ResponseWriter, r *http.Request) {
	h := s.liveHospital()
	p, err := h.FindPatient(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	room, err := h.AssignRoom(p)
	if err != nil {
		writeError(w, err)
		return
	}
	writeMessage(w, fmt.Sprintf("%s was moved to room %d", p.Name(), room.Number()))
}

// ---------------------------------------------------------------------------
// Simulación y reinicio
// ---------------------------------------------------------------------------

// handleSimulation lanza la simulación en SEGUNDO PLANO (en otra goroutine)
// y responde de inmediato. La página luego pide /api/state cada medio segundo
// y va viendo cómo cambian los cuartos mientras las goroutines trabajan.
func (s *Server) handleSimulation(w http.ResponseWriter, r *http.Request) {
	var req simulationRequest
	if !readJSON(w, r, &req) {
		return
	}
	cfg := hospital.SimulationConfig{
		Rounds:   clamp(req.Rounds, 1, 30),
		MaxDelay: time.Duration(clamp(req.MaxDelayMs, 50, 3000)) * time.Millisecond,
	}

	// CompareAndSwap(false, true): "si está en false, ponlo en true" en un
	// solo paso atómico. Si dos clics llegan a la vez, solo uno gana.
	if !s.simRunning.CompareAndSwap(false, true) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "a simulation is already running"})
		return
	}

	h := s.liveHospital()
	go func() {
		defer s.simRunning.Store(false) // al terminar (pase lo que pase) se libera
		result := h.Simulate(cfg)

		s.mu.Lock()
		s.lastSim = &result
		s.mu.Unlock()
	}()

	writeJSON(w, http.StatusAccepted, map[string]string{
		"message": fmt.Sprintf("simulation started: %d rounds per patient", cfg.Rounds),
	})
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	if s.simRunning.Load() {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "wait until the simulation finishes"})
		return
	}
	fresh, err := s.factory()
	if err != nil {
		writeError(w, err)
		return
	}
	s.mu.Lock()
	s.live = fresh
	s.lastSim = nil
	s.mu.Unlock()
	writeMessage(w, "the hospital was reset to its initial data")
}

// ---------------------------------------------------------------------------
// Ayudantes de JSON
// ---------------------------------------------------------------------------

// readJSON lee el cuerpo de la petición en dst. Si falla, ya responde 400
// y devuelve false para que el handler termine.
func readJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // máximo 1 MB
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body: " + err.Error()})
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v) // si el navegador cerró la conexión no hay nada más que hacer
}

func writeMessage(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusOK, map[string]string{"message": message})
}

// writeError traduce los errores del modelo a códigos HTTP con errors.Is.
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, hospital.ErrInvalidData):
		status = http.StatusBadRequest // 400: datos mal enviados
	case errors.Is(err, hospital.ErrPatientNotFound):
		status = http.StatusNotFound // 404: no existe
	case errors.Is(err, hospital.ErrDuplicateID),
		errors.Is(err, hospital.ErrNoRoomAvailable),
		errors.Is(err, hospital.ErrAlreadyAsleep),
		errors.Is(err, hospital.ErrAlreadyAwake),
		errors.Is(err, hospital.ErrNotInHallway),
		errors.Is(err, hospital.ErrNoStaffAvailable):
		status = http.StatusConflict // 409: choca con el estado actual
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func clamp(value, minimum, maximum int) int {
	return max(minimum, min(value, maximum)) // min y max son funciones nativas desde Go 1.21
}

# Hospital de los Costeños con Narcolepsia

Central management system for a hospital whose coastal patients suffer from narcolepsy and fall asleep anywhere in the facilities.
The system tracks where each patient is, which doctor attends them, records every sleep episode and assigns a free bed before the patient blocks a hallway.

Object-Oriented Programming applied exercise, Universidad EIA. Instructor: Sebastián Zapata Ramírez.

**Author:** Maria José Sosa

---

## Requirements

- Go **1.22 or later** (the enunciado asks for 1.21+. 1.22 is needed for the HTTP route patterns used by the web interface).
- The model (`hospital/`), the game rules (`game/`) and the web dashboard (`web/`) use **only the standard library**.
- The one third-party module is **Ebitengine** (`github.com/hajimehoshi/ebiten/v2`, pinned to v2.8.8 so the `go` directive stays at 1.22). Only the `gameview/` package imports it, and only to draw the game window. It is downloaded by `go` on the first run.
- Ebitengine needs, to open the window:
  - **Windows:** nothing extra (no cgo).
  - **macOS:** nothing extra with the Xcode command line tools installed.
  - **Linux:** cgo plus the X11/OpenGL development packages. On Debian/Ubuntu: `sudo apt install gcc libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config`.
- If your machine cannot build Ebitengine, use the `nogame` tag (see below): the game is compiled out and everything the enunciado asks for still runs.
- *Optional, only for the race detector on Windows:* a C compiler (gcc) in the `PATH`. See [Race detector](#race-detector).

## How to run

```
go run .                  # scenario + simulation in the console, then the game window opens
go run . -web             # scenario + simulation + the web dashboard (no game window)
go run -tags nogame .     # scenario + simulation only; builds WITHOUT Ebitengine
```

All three start the same way:

1. **Mandatory scenario (Section 6).** It prints steps 1-4 and the result of the four required queries.
2. **Bonus: concurrent simulation.** It runs one goroutine per patient for a few rounds and prints a summary.

Then they differ:

| Command | Third step |
|---|---|
| `go run .` | Opens the game window (480x270 pixel art scaled x2) on the main menu: New shift, Continue, Demo (the enunciado scenario, step by step) and more. Closing the window ends the program. There is no web server. |
| `go run . -web` | **Bonus: web interface.** It starts a server and prints `Open http://localhost:8080`. Open that URL in a browser. Press `Ctrl+C` to stop. No game window. |
| `go run -tags nogame .` | Prints a short message saying the game was compiled out and exits. The `nogame` build never imports Ebitengine, so it works on a machine without the graphics libraries. Use it to run the mandatory scenario anywhere. |

The web page has three tabs:

| Tab | What it shows |
|---|---|
| **Mandatory Scenario** | The exact console output of the scenario, plus the four queries as cards. It uses its own hospital, so nothing done in the other tabs changes it. |
| **Hospital Management** | A live hospital. You can admit patients, hire doctors and orderlies, add rooms, trigger sleep attacks, wake patients up and assign rooms. It also shows statistics, room cards, staff and the episode history. |
| **Concurrent Simulation** | Starts the goroutine simulation on the live hospital. The page polls a safe snapshot twice per second, so you can watch beds fill and empty in real time. |

## How to test

```
go test ./...
```

There are 8 tests in `hospital/hospital_test.go`:

| Test | What it checks |
|---|---|
| `TestAssignRoomWithFreeRoom` | Room assignment when a room is free *(required)* |
| `TestAssignRoomNoRoomAvailable` | Failure when no room is free: error, patient stays in the hallway *(required)* |
| `TestPatientsInHallway` | Query 5.1 *(required: one of the queries)* |
| `TestSevereReport` | Query 5.4, including Severe patients with 0 episodes |
| `TestDoctorMyEpisodes` | Query 5.2 |
| `TestStaffDispatchIsPolymorphic` | The `[]Attender` dispatches a Doctor and then an Orderly |
| `TestWakeUpReleasesRoom` | Waking up frees the bed for the patient left in the hallway |
| `TestSimulationIsRaceFree` | Goroutines and concurrent snapshots keep the hospital consistent |

`game/game_test.go` tests the game rules on top of the model: the chosen staff member is dispatched, a patient with no bed stays where they fell asleep, discharge frees the room, cooldowns, one-time overwork penalty, the clock and pause, scoring, five full seeded shifts, and that the scripted Demo ends with the same four queries as the console scenario.

### Race detector

```
go test -race ./...
go run -race .
```

On Windows `-race` needs cgo, so `gcc` must be installed. For example:

```
winget install BrechtSanders.WinLibs.POSIX.UCRT
```

After installing, open a new terminal and check that `gcc --version` works.

### Quality checks

```
gofmt -l .      # prints nothing when every file is formatted
go vet ./...    # prints nothing when there are no warnings
```

---

## Project layout

```
go.mod                 module costenos-narcolepsia
main.go                builds the scenario, prints it, then opens the game or (-web) the web server (no business logic)
game_on.go             `!nogame`: opens the game window (the only file in main that reaches Ebitengine)
game_off.go            `nogame`: prints that the game was compiled out
art/                   sprite grids (sprites/*.txt), generated PNGs (png/) and art.go, which embeds the PNGs into the executable
hospital/              THE MODEL: pure logic, no printing, no display code
  person.go            Person (embedded in Patient, Doctor and Orderly)
  patient.go           Patient + NarcolepsyLevel + PatientState
  doctor.go            Doctor (implements Attender)
  orderly.go           Orderly (second type implementing Attender)
  attender.go          Attender interface
  room.go              Room + RoomState
  episode.go           EpisodeRecord
  errors.go            sentinel errors (ErrNoRoomAvailable, ...)
  hospital.go          Hospital + the four required queries
  snapshot.go          race-free copy of the state, used by readers
  simulation.go        bonus: one goroutine per patient
  hospital_test.go     tests
game/                  GAME RULES (standard library only): clock, patients, staff, score, days, report, demo
gameview/              GAME WINDOW: menu, camera screen, Demo and the -agent test backdoor; the only package that imports Ebitengine
  assets/fonts/        Pixelify Sans and VT323 (.ttf) with their OFL licences
web/                   THE VIEW: REST API + static HTML/CSS/JS
  server.go            routes and handlers
  dto.go               JSON structs and conversion from hospital.Snapshot
  static/              index.html, styles.css, app.js (embedded with go:embed)
```

## Design decisions

- **Encapsulation.** Every entity field is lowercase (unexported). Code outside the `hospital` package can only use constructors (`NewPatient`, `NewRoom`, …), getters and methods. Constructors validate their input and return an `error`. Getters that return slices return copies (`slices.Clone`), so callers cannot change internal lists.
- **Composition.** `Person` (id, name, age) is embedded **by value** in `Patient`, `Doctor` and `Orderly`. Its methods `ID()`, `Name()` and `Age()` are promoted. That is also why `Doctor` and `Orderly` already have two of the methods `Attender` requires.
- **Polymorphism.** `Hospital` keeps all staff in a `[]Attender` and dispatches them round-robin, skipping anyone whose `IsAvailable()` is false. It never checks the concrete type to attend a patient. A new type (for example a `Nurse`) only needs the five methods and a call to `HireStaff`. `EpisodeRecord` also stores its attender as an `Attender`.
- **Typed states.** `NarcolepsyLevel`, `PatientState` and `RoomState` are `int`-based types with `iota` constants and a `String()` method. They are never compared as loose strings.
- **Errors, not panics.**
  - Foreseeable problems return sentinel errors wrapped with `%w`, for example "no room available" or "patient is already awake". Callers check them with `errors.Is`.
  - `RegisterEpisode` always records the episode, because the patient did fall asleep. If there was no room it returns `ErrNoRoomAvailable`. It uses `errors.Join` when both the room and the staff failed.
- **Relationships live inside the entities.** A `Doctor` stores its own episodes, so `MyEpisodes()` needs no hospital parameter. A `Patient` knows its room and doctor, and a `Room` knows its occupants. `Room.Occupy` and `Room.Release` update both sides, so they never disagree.
- **Slices vs. map.**
  - Patients, rooms and staff are slices because order matters: the first free room, admission order, and round-robin turns.
  - `SevereReport` returns a `map[*Patient]int` as the enunciado requires. It is a natural "patient → count" relation. Maps have no order, so `main` sorts the keys by ID before printing.
- **`EpisodeRecord` is an immutable value.** It stores the room *number*, not a `*Room`, so it remains a faithful record of what happened even after the room changes.
- **Concurrency.**
  - `Hospital` owns a `sync.Mutex`. Every public method locks it. Private `...Locked` helpers assume it is already held, because `sync.Mutex` is not reentrant.
  - `Snapshot()` copies the whole state under the lock, so readers like the web server never race with writers.
  - Episode IDs come from an `atomic.Int64`.
- **Clean separation of logic and view.**
  - `hospital` has no JSON tags, no HTTP and no printing (only `EpisodeRecord.Summary()`, which the enunciado requires).
  - `web` converts snapshots to its own DTOs. The `hospital` package did not change when the interface was added.
- **`runScenario(out io.Writer)`.** The same function writes the scenario to the console and to a buffer, through `io.MultiWriter`. That buffer is what the first web tab shows.

## REST API

| Method | Route | Body | Purpose |
|---|---|---|---|
| GET | `/api/scenario` | – | Console log and snapshot of the mandatory scenario |
| GET | `/api/state` | – | Snapshot of the live hospital and simulation status |
| POST | `/api/patients` | `{id, name, age, level}` | Admit a patient |
| POST | `/api/doctors` | `{id, name, age, specialty}` | Hire a doctor |
| POST | `/api/orderlies` | `{id, name, age}` | Hire an orderly |
| POST | `/api/rooms` | `{number, capacity}` | Add a room |
| POST | `/api/patients/{id}/sleep` | `{location}` | Register a sleep episode |
| POST | `/api/patients/{id}/wake` | – | Wake a patient up (frees the bed) |
| POST | `/api/patients/{id}/assign-room` | – | Assign a room to a hallway patient |
| POST | `/api/simulation` | `{rounds, maxDelayMs}` | Start the concurrent simulation (409 if one is running) |
| POST | `/api/reset` | – | Reset the live hospital to its initial data |

Errors are returned as `{"error": "..."}` with status 400, 404 or 409, chosen with `errors.Is` on the model's sentinel errors.

## AI usage

See [AI_USAGE.md](AI_USAGE.md).

## Credits

- **Fonts** (both under the [SIL Open Font License 1.1](https://openfontlicense.org), embedded in `gameview/assets/fonts/` together with their `OFL-*.txt` files):
  - **Pixelify Sans** by Stefie Justprince, from [Google Fonts](https://fonts.google.com/specimen/Pixelify+Sans). Used for the menu and every interface text.
  - **VT323** by Peter Hull, from [Google Fonts](https://fonts.google.com/specimen/VT323). Used only for the camera stamp, the clock and the map buttons.
- **Game engine:** [Ebitengine](https://ebitengine.org) (Apache 2.0).
- **Pixel art:** drawn for this project from text grids (`art/sprites/`).

## Test backdoor (`-agent`)

For checking the game without a person looking at the window. Full table of flags and an example script in `docs/HospitalCosteños/Arquitectura.md`.

```
go run . -agent -scene menu -input script.txt -quit
go run . -agent -scene demo -step 3 -cam 7B -input script.txt -quit
```

# AI Usage Report

This file describes honestly how AI tools were used in this project, as required by Section 8.1 of the assignment.

## Tools used and for what

| Tool | Used for |
|---|---|
| **Claude Code** (Anthropic CLI agent, model Claude Opus 5.5) | Planning the design together with me before writing any code. It also wrote most of the code (model, tests, `main`, REST API and HTML/CSS/JS), ran `gofmt`, `go vet` and the tests, made the commits I asked for, and drafted this README, this report and a Spanish study guide for the defense. |

How we worked:

1. Claude wrote a full plan first. I reviewed it, answered its design questions and approved it.
2. The code was written in phases, with one commit per phase so the history shows how the project grew.
3. All code identifiers are in English and all comments are in Spanish, so I can study every line before the oral defense.

## Decisive prompts

1. *"Vamos a hacer el proyecto que se encuentra en el md anclado, vamos a hacer primero el plan completo (…) Sé muy apegado a las consideraciones del enunciado, el código no tiene que ser tan complejo, la idea es que tenga la esencia básica de objetos y cumpla todas las rúbricas. El código se hará en inglés pero los comentarios en español (…)"*
   This set the scope. The design had to follow the rubric closely and stay simple enough for me to defend.
2. *"Para la interfaz gráfica me gusta la idea que se haga con algo de HTML y JS (…) no se deben usar muchas librerías extra, hagamos el bono de concurrencia pero debe ser muy explícito ya que no conozco bien del tema. Para el arranque que sea todo junto, podríamos implementar una pequeña interfaz para separar el escenario obligatorio del resto."*
   This led to the standard-library REST API and the three web tabs. It also made the concurrency code heavily commented, with `Snapshot()` as an explicit race-free way to read state.
3. The answers to the planning questions: English for user-facing text, tabs instead of a console menu, and installing gcc to verify `-race`. These fixed the final structure before any code was written.

## Cases where the generated code was wrong or not idiomatic, and how it was fixed

1. **Broken accents after an automated edit (real bug, committed and then fixed).**
   - *What went wrong:* The AI changed one number in `main.go` with a PowerShell `Get-Content`/`Set-Content` command. Windows PowerShell 5.1 read the UTF-8 file as ANSI and wrote it back as UTF-8. Every accented character was double-encoded: `Costeños` became `CosteÃ±os`, including the hospital name printed by the program. The same command also added a byte-order mark, which `gofmt` flagged.
   - *How it was fixed:* The BOM was removed right away. The double encoding was only noticed one phase later, so it reached the repository in the concurrency commit. It was fixed in its own commit (`fix: restore UTF-8 accents…`) by reversing the encoding, and the result was checked with a diff against the previous version.
   - *Lesson:* Do not edit source files with PowerShell 5.1 text cmdlets.
2. **Non-idiomatic ID generation in a test.**
   - *What went wrong:* The first version of the concurrency test built patient IDs with `"P-" + string(rune('0'+i))`. That only works for one-digit numbers and is hard to read.
   - *How it was fixed:* It was replaced with `fmt.Sprintf("P-%03d", i)`.
3. **`gofmt` rejected the first test file.** Trailing comments on consecutive lines were not aligned as `gofmt` expects. The fix was running `gofmt -w`. Since then every phase runs `gofmt -l .` and `go vet ./...` before committing.
4. **Uninformative simulation output.** With an even number of rounds, every patient alternated sleep/wake and always finished *awake*. The console summary then showed every room empty, which demonstrated nothing. The number of rounds was changed to 7, so the final state shows occupied rooms and patients in the hallway.
5. **Layout bug in the web page.** In the scenario tab the wide console text pushed the right-hand column off the screen, because the grid used `1fr 1fr`. This was found by taking screenshots with headless Edge. The fix was `minmax(0, 1fr)` columns, so the console scrolls instead.
6. **Possible lost clicks.** The page redrew every table every 500 ms. A click could be lost if the button was replaced between mouse-down and mouse-up. Now the page only redraws when the received data actually changed.
7. **Windows line endings broke `gofmt`.** After a `git checkout` of a file, Git for Windows (`core.autocrlf=true`) rewrote it with CRLF endings, and `gofmt -l` flagged it. Anyone cloning the repository on Windows would have seen every file flagged. The fix was a `.gitattributes` with `* text=auto eol=lf`, then checking a fresh clone from GitHub: `gofmt`, `go vet` and `go test` were all clean.
8. **Design adjustment: circular dependency.** `Patient` needs `*Doctor` and `Doctor` needs `*Patient`, so the planned "Patient/Room/Episode first, Doctor later" phases did not compile on their own. The `assignedDoctor` field was added to `Patient` in the same phase as `Doctor`.

## Part 2: the camera game (graphical interface bonus)

> *Draft written by Claude on 2026-10-18 from the project's design log and mistakes log (`docs/HospitalCosteños/`). To be reviewed by me before submitting.*

After the base project was finished, I designed a small pixel-art game on top of the `hospital` model: you watch security cameras of the hospital, assign a doctor or orderly to each patient who falls asleep, and discharge them when they wake up. The `hospital` package was **not changed at all** (`git diff main -- hospital/` is empty), as Section 7 requires.

### How AI was used

| Tool | Used for |
|---|---|
| **Claude Code** (main session, Claude Opus 5.5) | Turning my idea into a written design (Obsidian notes in `docs/HospitalCosteños/`), asking me design questions with a recommended answer each time, logging every decision with my own words, coordinating the agents below, reviewing their work (checks, screenshots) and making the commits. |
| **Art agent** (Claude Opus 5.5, subagent) | Palette, pixel-art sprites (written as text grids in `art/sprites/` and converted to PNG by `tools/sprites`), and every screen as a mock-up in Figma before any code. |
| **Builder agent** (Claude Sonnet 5.5, subagent) | The `game/` rules package, the `gameview/` Ebitengine view, the `-agent` test mode and the tests. |
| **Tester agent** (Claude Sonnet 5.5, subagent) | Driving the game through the `-agent` test mode, comparing screenshots with the approved designs and reporting bugs (it does not fix them). |
| **Figma + Figwright** | Every screen was approved by me as a picture (a Figma frame, or a review image when the Figma plugin was disconnected) before it was programmed. |

How we worked: idea → design notes and decision log → I approve each screen as a picture → the builder programs it → screenshots are checked against the design → I play it and report what feels wrong.

The art agents generated the large sprite grids (backgrounds, characters) with **throwaway Python scripts that are not in the repository**. The `.txt` grids are the real source, and each generated one says so in a comment.

### Decisive prompts (game)

1. *"La idea es que sea de tipo revisión de cámaras de seguridad, donde cada cámara es una habitación del hospital, a un lado, habrá un panel con la información reducida de camilleros y doctores (…) el rol del jugador es asignarle a esos pacientes que se duerman un doctor/camillero disponible (…) la parte visual quiero que sea de ambiente cálido (…) el juego busca ser algo más de ambiente relajado."*
   This defined the whole game and its warm, non-horror look.
2. *"Los cambios que dices de la parte lógica hay que documentarlos muy bien, podrías específicarmelos en un md aparte dentro de obsidian?"*
   This produced `Lógica del juego.md`, the spec of the `game/` layer, including how the player chooses the staff member without changing `hospital/`.
3. *"si un doctor llena su cupo diario, baje un poco los puntos del hospital por 'exceso de trabajo'"*
   This turned a limit of the model (`maxPatientsPerDoctor = 4`) into a game mechanic.

### Cases where the generated code was wrong, and how it was fixed

The full list (about 20 cases) is in `docs/HospitalCosteños/Errores de la IA.md`. The most instructive ones:

1. **The tests passed but the real game did not work (found by me, playing).** After changing how the game is drawn (to fix blurry text), mouse hover and clicks landed in the wrong place: pointing at "Nuevo turno" highlighted "Créditos", and the map, pause and Demo buttons did not respond. The cause: `ebiten.CursorPosition()` now returned coordinates in a 2× space, but hit-testing still used 480×270 game pixels. The automated test mode did not catch it because it injected clicks directly in game coordinates, **bypassing the real mouse conversion**. Fix: one pure function `windowToGame` with a table test for scales 1–4, and the test mode now clicks through that same function.
2. **A character drawn as a placeholder rectangle (found by me, playing).** The view looked up each patient's sprite by the first word of the name, so `Abelardo "Mañe" Julio` searched for `pac_abelardo_*` while the files were `pac_mane_*`. Fix: sprites are looked up by ID, and a test fails if any cast member has no sprite.
3. **Blurry text.** Everything was drawn into a 480×270 image and then enlarged, so the vector fonts were smeared. Found by the main session reviewing 4× screenshots. Fix: draw at the real window resolution (sprites scaled without smoothing, text at the final size).
4. **A silent dependency change.** `go get` of Ebitengine raised the `go` directive in `go.mod` from 1.22 to 1.25 without warning. Fix: pin Ebitengine v2.8.8, which keeps "Go 1.22 or later".
5. **An invented API.** The builder called `ebiten.SetRunnableOnUnfocusedWindow`, which does not exist (the real one is `SetRunnableOnUnfocused`). The compiler caught it.
6. **Agents working in parallel mixed their commits.** Two builder agents shared Git's staging area, so one agent's commit included the other's files. History was not rewritten; the lesson is to commit with explicit paths or separate worktrees.

## What I learned

> *This section must be written by me (Maria José) in my own words before submitting.*

- …
- …
- …

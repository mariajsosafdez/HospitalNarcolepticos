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

## What I learned

> *This section must be written by me (Maria José) in my own words before submitting.*

- …
- …
- …

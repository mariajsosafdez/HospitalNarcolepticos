package gameview

import (
	"io"
	"math"
)

// Funciones PURAS del sonido (sin Ebitengine): volumen constante, fundidos,
// silencio y huecos entre ronquidos. Así se prueban con tablas. El reproductor
// que sí usa Ebitengine está en sound.go.

// Cue es un efecto de sonido suelto.
type Cue string

const (
	CueClick     Cue = "click"     // clic de menú, panel o botón
	CueDischarge Cue = "alta"      // "ding" de alta
	CueWake      Cue = "despierta" // pop del "!" de quien despierta
	CueNoBed     Cue = "sin_cama"  // aviso suave de "no hay cama"
	CueEndShift  Cue = "fin_turno" // jingle del reporte de fin de turno
	CueSnore     Cue = "ronquido"  // ronquido (muy bajito)
	CueRoll      Cue = "camilla"   // ruedas de la camilla (en bucle, bajito)
)

// Track es la música que debería sonar.
type Track string

const (
	TrackNone  Track = ""
	TrackMenu  Track = "menu"  // menú, créditos y ajustes
	TrackTurno Track = "turno" // durante un turno y la Demo
)

// Los archivos de assets/sound llevan el nombre del cue: sfx_<cue>.ogg y
// music_<pista>.ogg.
func cueFile(c Cue) string     { return "sfx_" + string(c) + ".ogg" }
func trackFile(t Track) string { return "music_" + string(t) + ".ogg" }

// Volumen constante ------------------------------------------------------------

// Cuánto debe sonar cada archivo, medido en dBFS de RMS (0 dBFS es el máximo
// posible). La música queda 2 dB por debajo de los efectos; el ronquido y la
// camilla, mucho más bajitos porque suenan seguido de fondo.
const (
	targetMusicDB = -20.0
	targetSFXDB   = -18.0
	targetSnoreDB = -30.0
	targetRollDB  = -32.0
	peakLimitDB   = -1.0 // ningún pico pasa de aquí, para que nada sature
)

// soundTargetDB dice a cuántos dB de RMS debe quedar el archivo.
func soundTargetDB(file string) float64 {
	switch file {
	case cueFile(CueSnore):
		return targetSnoreDB
	case cueFile(CueRoll):
		return targetRollDB
	case trackFile(TrackMenu), trackFile(TrackTurno):
		return targetMusicDB
	}
	return targetSFXDB
}

// toDB pasa una amplitud (1,0 = máximo) a dBFS. El silencio da -Inf.
func toDB(amplitude float64) float64 {
	if amplitude <= 0 {
		return math.Inf(-1)
	}
	return 20 * math.Log10(amplitude)
}

// fromDB es lo contrario: cuántas veces se multiplica la señal.
func fromDB(db float64) float64 { return math.Pow(10, db/20) }

// gateWindow es el tamaño (en cuadros estéreo) de las ventanas con que se mide
// la parte fuerte: 10 ms a 44 100 Hz.
const gateWindow = 441

// measureLoudness mide un audio PCM de 16 bits little endian, estéreo: el pico
// y el RMS "de la parte fuerte". Para el RMS corta el audio en ventanas de
// 10 ms y solo cuenta las que están a menos de 20 dB de la más fuerte, así
// los silencios del principio o del final no hacen que parezca más bajito.
func measureLoudness(pcm []byte) (rmsDB, peakDB float64) {
	n := len(pcm) / 2 // muestras
	if n == 0 {
		return math.Inf(-1), math.Inf(-1)
	}
	sample := func(i int) float64 {
		return float64(int16(uint16(pcm[2*i])|uint16(pcm[2*i+1])<<8)) / 32768
	}
	peak := 0.0
	for i := 0; i < n; i++ {
		peak = math.Max(peak, math.Abs(sample(i)))
	}

	// Energía media de cada ventana.
	win := gateWindow * 2 // muestras por ventana (2 canales)
	var energies []float64
	var counts []int
	loudest := 0.0
	for start := 0; start < n; start += win {
		end := min(start+win, n)
		sum := 0.0
		for i := start; i < end; i++ {
			s := sample(i)
			sum += s * s
		}
		energies = append(energies, sum)
		counts = append(counts, end-start)
		loudest = math.Max(loudest, sum/float64(end-start))
	}
	// Se juntan las ventanas fuertes (energía media >= la más fuerte - 20 dB).
	total, samples := 0.0, 0
	for i, e := range energies {
		if e/float64(counts[i]) >= loudest/100 {
			total += e
			samples += counts[i]
		}
	}
	return toDB(math.Sqrt(total / float64(samples))), toDB(peak)
}

// gainDB es cuántos dB hay que subir (o bajar) un archivo para que su RMS
// llegue a target sin que el pico pase de peakLimitDB. Si el pico manda, el
// archivo queda un poco más bajito que su objetivo.
func gainDB(rmsDB, peakDB, target float64) float64 {
	return math.Min(target-rmsDB, peakLimitDB-peakDB)
}

// scalePCM multiplica por gain cada muestra de 16 bits (little endian) de buf,
// y la deja en el rango de 16 bits si se pasara.
func scalePCM(buf []byte, gain float64) {
	for i := 0; i+1 < len(buf); i += 2 {
		v := float64(int16(uint16(buf[i])|uint16(buf[i+1])<<8)) * gain
		v = math.Max(-32768, math.Min(32767, math.Round(v)))
		u := uint16(int16(v))
		buf[i], buf[i+1] = byte(u), byte(u>>8)
	}
}

// gainReader le aplica una ganancia fija a lo que lee de otro lector de PCM de
// 16 bits. Así la música se normaliza mientras suena, sin cargarla entera.
type gainReader struct {
	src  io.ReadSeeker
	gain float64
}

func (g *gainReader) Seek(offset int64, whence int) (int64, error) {
	return g.src.Seek(offset, whence)
}

func (g *gainReader) Read(p []byte) (int, error) {
	n, err := g.src.Read(p)
	// Una muestra son 2 bytes: si llegó un byte suelto, se pide el que falta.
	if n%2 == 1 && n < len(p) {
		m, err2 := io.ReadFull(g.src, p[n:n+1])
		n += m
		if err == nil {
			err = err2
		}
	}
	scalePCM(p[:n], g.gain)
	return n, err
}

// Fundidos y silencio ------------------------------------------------------------

const (
	fadeTicks   = 15  // un fundido dura 0,25 s (salir) y otros 0,25 s (entrar)
	duckVolume  = 0.4 // con la pausa la música baja al 40 %
	duckTicks   = 18  // y baja o sube en 0,3 s
	musicVolume = 1.0 // la música ya viene normalizada: no se baja más
)

// approach acerca level a target, a lo sumo step por tick.
func approach(level, target, step float64) float64 {
	// El 1e-9 evita que los errores de redondeo dejen un resto minúsculo.
	switch {
	case math.Abs(target-level) <= step+1e-9:
		return target
	case level < target:
		return level + step
	}
	return level - step
}

// fadeVolume convierte el nivel lineal (0 a 1) del fundido en volumen. Al
// cuadrado se oye más parejo que en línea recta.
func fadeVolume(level float64) float64 { return level * level }

// musicStep es un tick del cambio de pista: si lo que suena (playing) no es lo
// que se quiere (want), baja el nivel hasta 0; cuando llega a 0 cambia de pista
// (start = true: hay que abrir la nueva) y la sube desde 0. Si ya suena lo
// que se quiere, solo termina de subir.
func musicStep(playing, want Track, level float64) (next Track, newLevel float64, start bool) {
	step := 1.0 / fadeTicks
	if playing == want {
		return playing, approach(level, 1, step), false
	}
	if level > 0 {
		return playing, approach(level, 0, step), false
	}
	return want, 0, want != TrackNone
}

// volumeGain convierte un volumen de Ajustes (0 a 100) en el número por el que
// se multiplica la señal (0 a 1). El oído no oye en línea recta: la mitad de la
// amplitud (50 %) apenas suena "un poco más bajito", y para que la barra se
// sienta pareja se usa una curva cuadrática, gain = (v/100)². Así 50 % es 0,25
// (-12 dB) y 10 % es 0,01 (-40 dB), y cada paso de la barra baja unos dB
// parecidos. 0 % da exactamente 0 (silencio) y 100 % da 1 (el nivel ya
// normalizado de cada archivo). Fuera de 0 a 100 se corta.
func volumeGain(percent int) float64 {
	v := float64(min(max(percent, 0), 100)) / 100
	return v * v
}

// musicVolumeNow es el volumen real de la música: el fundido, el "agachado" de
// la pausa (duck, de 1 a duckVolume), el volumen de Ajustes (ya pasado por
// volumeGain, de 0 a 1) y el interruptor de Sonido.
func musicVolumeNow(level, duck, gain float64, enabled bool) float64 {
	if !enabled {
		return 0
	}
	return fadeVolume(level) * duck * gain * musicVolume
}

// duckTarget dice a qué nivel debe estar el "agachado": en pausa baja.
func duckTarget(paused bool) float64 {
	if paused {
		return duckVolume
	}
	return 1
}

// Ronquidos --------------------------------------------------------------------

const (
	snoreMinGap = 6 * 60  // ticks (6 s) como mínimo entre ronquidos
	snoreMaxGap = 14 * 60 // y 14 s como máximo
)

// snoreGap da cuántos ticks se espera hasta el próximo ronquido, a partir de
// un número r entre 0 y 1 (de la semilla del turno, así es repetible).
func snoreGap(r float64) int {
	r = math.Max(0, math.Min(r, 0.999999))
	return snoreMinGap + int(r*float64(snoreMaxGap-snoreMinGap))
}

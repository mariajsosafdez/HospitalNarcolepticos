package gameview

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"
)

// decodeAll decodifica un OGG embebido a PCM de 16 bits estéreo.
func decodeAll(t *testing.T, file string) []byte {
	t.Helper()
	stream, err := loadOGG(file)
	if err != nil {
		t.Fatalf("%s: %v", file, err)
	}
	pcm, err := io.ReadAll(stream)
	if err != nil {
		t.Fatalf("%s: %v", file, err)
	}
	return pcm
}

// soundFiles lista los OGG embebidos.
func soundFiles(t *testing.T) []string {
	t.Helper()
	entries, err := fs.ReadDir(soundFS, "assets/sound")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names
}

// TestMeasureSounds imprime la tabla de ganancias. Solo corre con SOUND_MEASURE=1:
//
//	SOUND_MEASURE=1 go test ./gameview -run TestMeasureSounds -v
//
// Lo impreso al final es lo que va en sound_gain.go.
func TestMeasureSounds(t *testing.T) {
	if os.Getenv("SOUND_MEASURE") == "" {
		t.Skip("SOUND_MEASURE no está puesto")
	}
	var b strings.Builder
	for _, f := range soundFiles(t) {
		rms, peak := measureLoudness(decodeAll(t, f))
		target := soundTargetDB(f)
		g := gainDB(rms, peak, target)
		fmt.Fprintf(&b, "\t// antes: RMS %.1f dBFS, pico %.1f dBFS; objetivo RMS %.0f dBFS\n\t%q: %.1f,\n", rms, peak, target, f, g)
		fmt.Printf("%-18s before RMS %6.1f peak %6.1f | target %4.0f gain %+5.1f | after RMS %6.1f peak %6.1f\n",
			f, rms, peak, target, g, rms+g, peak+g)
	}
	fmt.Print(b.String())
}

// TestEverySoundHasGain: cada OGG embebido tiene su ganancia en la tabla, y la
// tabla no tiene archivos que ya no existen.
func TestEverySoundHasGain(t *testing.T) {
	files := soundFiles(t)
	for _, f := range files {
		if _, ok := soundGainDB[f]; !ok {
			t.Errorf("%s has no entry in soundGainDB", f)
		}
	}
	for f := range soundGainDB {
		if !slices.Contains(files, f) {
			t.Errorf("soundGainDB has %s but the file is not embedded", f)
		}
	}
	// Y todos los cues y pistas del juego tienen archivo.
	for _, c := range allCues {
		if !slices.Contains(files, cueFile(c)) {
			t.Errorf("cue %s has no file %s", c, cueFile(c))
		}
	}
	for _, tr := range []Track{TrackMenu, TrackTurno} {
		if !slices.Contains(files, trackFile(tr)) {
			t.Errorf("track %s has no file %s", tr, trackFile(tr))
		}
	}
}

// TestSoundsAreNormalized: con la ganancia aplicada, cada archivo queda en su
// objetivo de RMS (a menos de 0,3 dB) o, si el pico lo limita, bajo el objetivo;
// y ningún pico pasa de -1 dBFS.
func TestSoundsAreNormalized(t *testing.T) {
	for _, f := range soundFiles(t) {
		pcm := decodeAll(t, f)
		scalePCM(pcm, fromDB(soundGainDB[f]))
		rms, peak := measureLoudness(pcm)
		target := soundTargetDB(f)
		if peak > peakLimitDB+0.1 {
			t.Errorf("%s: peak %.2f dBFS above the %.0f limit", f, peak, peakLimitDB)
		}
		if rms > target+0.3 || (rms < target-0.3 && math.Abs(peak-peakLimitDB) > 0.3) {
			t.Errorf("%s: rms %.2f dBFS, want %.0f (peak %.2f)", f, rms, target, peak)
		}
	}
}

func TestMeasureLoudnessSine(t *testing.T) {
	// Seno de amplitud 0,5 en los dos canales: RMS = 0,5/raíz(2) = -9,03 dB, pico = -6,02 dB.
	frames := 44100
	pcm := make([]byte, frames*4)
	for i := 0; i < frames; i++ {
		v := int16(0.5 * 32767 * math.Sin(2*math.Pi*440*float64(i)/44100))
		for ch := 0; ch < 2; ch++ {
			pcm[4*i+2*ch], pcm[4*i+2*ch+1] = byte(uint16(v)), byte(uint16(v)>>8)
		}
	}
	rms, peak := measureLoudness(pcm)
	if math.Abs(rms-(-9.03)) > 0.05 || math.Abs(peak-(-6.02)) > 0.05 {
		t.Fatalf("rms %.2f peak %.2f, want -9.03 and -6.02", rms, peak)
	}
	// Con 1 s de silencio delante, el RMS "de la parte fuerte" es el mismo.
	padded := append(make([]byte, frames*4), pcm...)
	if rms2, _ := measureLoudness(padded); math.Abs(rms2-rms) > 0.05 {
		t.Fatalf("silence changed the rms: %.2f vs %.2f", rms2, rms)
	}
	if r, p := measureLoudness(nil); !math.IsInf(r, -1) || !math.IsInf(p, -1) {
		t.Fatalf("empty audio should be -Inf, got %v %v", r, p)
	}
}

func TestGainDB(t *testing.T) {
	tests := []struct {
		name              string
		rms, peak, target float64
		want              float64
	}{
		{"sube", -26, -12, -20, 6},
		{"baja", -10, -3, -18, -8},
		{"el pico limita", -22, -3, -18, 2}, // subir 4 dB dejaría el pico en +1: solo caben 2
	}
	for _, tc := range tests {
		if got := gainDB(tc.rms, tc.peak, tc.target); math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("%s: gain %.2f, want %.2f", tc.name, got, tc.want)
		}
	}
}

func TestSoundTargets(t *testing.T) {
	tests := []struct {
		file string
		want float64
	}{
		{"music_menu.ogg", -20}, {"music_turno.ogg", -20},
		{"sfx_alta.ogg", -18}, {"sfx_click.ogg", -18},
		{"sfx_ronquido.ogg", -30}, {"sfx_camilla.ogg", -32},
	}
	for _, tc := range tests {
		if got := soundTargetDB(tc.file); got != tc.want {
			t.Errorf("%s: target %v, want %v", tc.file, got, tc.want)
		}
	}
}

func TestScalePCM(t *testing.T) {
	// Muestras: 1000, -1000, 30000, -30000. Con ganancia 2: 2000, -2000, y las dos últimas se recortan.
	vals := []int16{1000, -1000, 30000, -30000}
	buf := make([]byte, 0, 8)
	for _, v := range vals {
		buf = append(buf, byte(uint16(v)), byte(uint16(v)>>8))
	}
	scalePCM(buf, 2)
	want := []int16{2000, -2000, 32767, -32768}
	for i, w := range want {
		got := int16(uint16(buf[2*i]) | uint16(buf[2*i+1])<<8)
		if got != w {
			t.Errorf("sample %d: got %d, want %d", i, got, w)
		}
	}
}

// oddReader entrega los bytes de a 3 para probar que gainReader no parte muestras.
type oddReader struct{ r *bytes.Reader }

func (o oddReader) Read(p []byte) (int, error) { return o.r.Read(p[:min(3, len(p))]) }
func (o oddReader) Seek(off int64, wh int) (int64, error) {
	return o.r.Seek(off, wh)
}

func TestGainReaderKeepsSamplesWhole(t *testing.T) {
	var src []byte
	for i := 0; i < 50; i++ {
		v := int16(100 * (i + 1))
		src = append(src, byte(uint16(v)), byte(uint16(v)>>8))
	}
	g := &gainReader{oddReader{bytes.NewReader(src)}, 0.5}
	got, err := io.ReadAll(g)
	if err != nil || len(got) != len(src) {
		t.Fatalf("read %d bytes, err %v; want %d", len(got), err, len(src))
	}
	for i := 0; i < 50; i++ {
		v := int16(uint16(got[2*i]) | uint16(got[2*i+1])<<8)
		if want := int16(50 * (i + 1)); v != want {
			t.Fatalf("sample %d: %d, want %d", i, v, want)
		}
	}
}

func TestApproach(t *testing.T) {
	tests := []struct{ level, target, step, want float64 }{
		{0, 1, 0.25, 0.25}, {0.9, 1, 0.25, 1}, {1, 0.4, 0.25, 0.75},
		{0.5, 0.4, 0.25, 0.4}, {0.4, 0.4, 0.25, 0.4},
	}
	for _, tc := range tests {
		if got := approach(tc.level, tc.target, tc.step); math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("approach(%v,%v,%v) = %v, want %v", tc.level, tc.target, tc.step, got, tc.want)
		}
	}
}

func TestFadeVolume(t *testing.T) {
	tests := []struct{ level, want float64 }{{0, 0}, {0.5, 0.25}, {1, 1}}
	for _, tc := range tests {
		if got := fadeVolume(tc.level); math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("fadeVolume(%v) = %v, want %v", tc.level, got, tc.want)
		}
	}
}

func TestMusicStep(t *testing.T) {
	tests := []struct {
		name          string
		playing, want Track
		level         float64
		next          Track
		wantLevel     float64
		start         bool
	}{
		{"arranca de cero", TrackNone, TrackMenu, 0, TrackMenu, 0, true},
		{"sube", TrackMenu, TrackMenu, 0, TrackMenu, 1.0 / fadeTicks, false},
		{"ya al tope", TrackMenu, TrackMenu, 1, TrackMenu, 1, false},
		{"baja antes de cambiar", TrackMenu, TrackTurno, 1, TrackMenu, 1 - 1.0/fadeTicks, false},
		{"cambia al llegar a cero", TrackMenu, TrackTurno, 0, TrackTurno, 0, true},
		{"se quiere silencio", TrackMenu, TrackNone, 0, TrackNone, 0, false},
	}
	for _, tc := range tests {
		next, level, start := musicStep(tc.playing, tc.want, tc.level)
		if next != tc.next || math.Abs(level-tc.wantLevel) > 1e-9 || start != tc.start {
			t.Errorf("%s: got (%q, %.3f, %v), want (%q, %.3f, %v)", tc.name, next, level, start, tc.next, tc.wantLevel, tc.start)
		}
	}
}

// Un cambio de menú a turno completo: baja en fadeTicks, cambia, sube en fadeTicks.
func TestMusicStepFullCrossfade(t *testing.T) {
	playing, level, ticks, switched := TrackMenu, 1.0, 0, -1
	for ; ticks < 100 && !(playing == TrackTurno && level == 1); ticks++ {
		var start bool
		playing, level, start = musicStep(playing, TrackTurno, level)
		if start {
			switched = ticks
		}
	}
	if switched != fadeTicks || ticks > 2*fadeTicks+2 {
		t.Fatalf("switched at tick %d, finished at %d; want %d and about %d", switched, ticks, fadeTicks, 2*fadeTicks)
	}
}

func TestMusicVolumeNow(t *testing.T) {
	tests := []struct {
		name        string
		level, duck float64
		gain        float64
		enabled     bool
		want        float64
	}{
		{"normal", 1, 1, 1, true, 1},
		{"en pausa", 1, duckVolume, 1, true, duckVolume},
		{"a medio fundido", 0.5, 1, 1, true, 0.25},
		{"apagado", 1, 1, 1, false, 0},
		{"apagado en pausa", 0.5, duckVolume, 1, false, 0},
		{"volumen 50 %", 1, 1, volumeGain(50), true, 0.25},
		{"volumen 0 %", 1, 1, volumeGain(0), true, 0},
		{"todo junto", 0.5, duckVolume, volumeGain(70), true, 0.25 * duckVolume * 0.49},
	}
	for _, tc := range tests {
		if got := musicVolumeNow(tc.level, tc.duck, tc.gain, tc.enabled); math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("%s: %v, want %v", tc.name, got, tc.want)
		}
	}
	if duckTarget(true) != duckVolume || duckTarget(false) != 1 {
		t.Error("duckTarget wrong")
	}
}

func TestSnoreGap(t *testing.T) {
	tests := []struct {
		r    float64
		want int
	}{
		{0, snoreMinGap}, {0.5, (snoreMinGap + snoreMaxGap) / 2}, {-3, snoreMinGap}, {1, snoreMaxGap - 1}, {7, snoreMaxGap - 1},
	}
	for _, tc := range tests {
		if got := snoreGap(tc.r); got != tc.want {
			t.Errorf("snoreGap(%v) = %d, want %d", tc.r, got, tc.want)
		}
	}
}

// El Sound en modo pruebas anota los cues, y apagado no anota nada.
func TestSoundLogAndMute(t *testing.T) {
	s := NewSound(true)
	s.Play(CueClick)
	s.Update(TrackMenu, false, true)
	s.Roll()
	s.Update(TrackMenu, false, true)
	s.Update(TrackMenu, false, true) // sin Roll: se detiene
	s.Roll()
	s.Update(TrackMenu, false, true)
	if got := strings.Join(s.Log(), ","); got != "click,camilla,camilla" {
		t.Fatalf("log %q", got)
	}
	s.Update(TrackMenu, false, false) // se apaga
	s.Play(CueDischarge)
	s.Roll()
	s.Update(TrackMenu, false, false)
	if len(s.Log()) != 3 {
		t.Fatalf("muted sound still logged: %v", s.Log())
	}
	var nilSound *Sound // las pantallas de las pruebas no tienen Sound
	nilSound.Play(CueClick)
	nilSound.Roll()
	nilSound.Update(TrackMenu, false, true)
}

func TestVolumeGain(t *testing.T) {
	tests := []struct {
		percent int
		want    float64
	}{
		{0, 0}, {10, 0.01}, {50, 0.25}, {70, 0.49}, {80, 0.64}, {100, 1}, {-20, 0}, {250, 1},
	}
	for _, tt := range tests {
		if got := volumeGain(tt.percent); math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("volumeGain(%d) = %v, want %v", tt.percent, got, tt.want)
		}
	}
	// Siempre sube con la barra y 50 % son unos -12 dB.
	for p := 0; p < 100; p += volumeStep {
		if volumeGain(p+volumeStep) <= volumeGain(p) {
			t.Errorf("volumeGain must grow: %d -> %d", p, p+volumeStep)
		}
	}
	if db := toDB(volumeGain(50)); math.Abs(db+12.04) > 0.01 {
		t.Errorf("50 %% = %.2f dB, want about -12", db)
	}
}

// Con 0 % cada grupo calla, y el otro sigue igual; con el sonido apagado
// callan los dos.
func TestVolumesMute(t *testing.T) {
	s := NewSound(true)
	s.SetVolumes(0, 100)
	if s.MusicGain() != 0 || s.SfxGain() != 1 {
		t.Errorf("music 0 %%: music %v, sfx %v", s.MusicGain(), s.SfxGain())
	}
	s.SetVolumes(100, 0)
	if s.MusicGain() != 1 || s.SfxGain() != 0 {
		t.Errorf("sfx 0 %%: music %v, sfx %v", s.MusicGain(), s.SfxGain())
	}
	s.SetVolumes(70, 80)
	s.SetEnabled(false)
	if s.MusicGain() != 0 || s.SfxGain() != 0 {
		t.Error("with the sound off both groups must be silent")
	}
	if musicVolumeNow(1, 1, volumeGain(0), true) != 0 {
		t.Error("music at 0 % must give volume 0")
	}
}

package gameview

// soundGainDB es la ganancia, en decibelios, que se le aplica a cada OGG al
// cargarlo para que todos suenen parejo. Cada una sale de MEDIR el archivo
// (ver measureLoudness en sound_mix.go) y es:
//
//	ganancia = objetivo de RMS - RMS medido   (sin pasar de -1 dBFS de pico)
//
// Objetivos: música -20 dBFS, efectos -18, ronquido -30 y camilla -32 (estos
// dos suenan seguido y de fondo). Para volver a medir (por ejemplo si se
// cambia un sonido):
//
//	SOUND_MEASURE=1 go test ./gameview -run TestMeasureSounds -v
//
// y se pega lo que imprime. TestEverySoundHasGain y TestSoundsAreNormalized
// fallan si falta un archivo o si la tabla ya no deja el volumen en su objetivo.
var soundGainDB = map[string]float64{
	// antes: RMS -16.4 dBFS, pico -0.3 dBFS; objetivo RMS -20 dBFS
	"music_menu.ogg": -3.6,
	// antes: RMS -17.7 dBFS, pico -1.6 dBFS; objetivo RMS -20 dBFS
	"music_turno.ogg": -2.3,
	// antes: RMS -15.7 dBFS, pico -1.0 dBFS; objetivo RMS -18 dBFS
	"sfx_alta.ogg": -2.3,
	// antes: RMS -26.6 dBFS, pico -13.4 dBFS; objetivo RMS -32 dBFS
	"sfx_camilla.ogg": -5.4,
	// antes: RMS -17.5 dBFS, pico -1.5 dBFS; objetivo RMS -18 dBFS
	"sfx_click.ogg": -0.5,
	// antes: RMS -16.5 dBFS, pico -0.0 dBFS; objetivo RMS -18 dBFS
	"sfx_despierta.ogg": -1.5,
	// antes: RMS -16.0 dBFS, pico -2.6 dBFS; objetivo RMS -18 dBFS
	"sfx_fin_turno.ogg": -2.0,
	// antes: RMS -23.5 dBFS, pico -5.7 dBFS; objetivo RMS -30 dBFS
	"sfx_ronquido.ogg": -6.5,
	// antes: RMS -14.8 dBFS, pico -0.9 dBFS; objetivo RMS -18 dBFS
	"sfx_sin_cama.ogg": -3.2,
}

package gameview

import (
	"bytes"
	"embed"
	"fmt"
	"io"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

// Reproductor de sonido. Es el único archivo que toca el audio de Ebitengine.
// Todas las decisiones (qué suena, cuándo, cuánto) están en sound_mix.go como
// funciones puras. Los sonidos son los candidatos "por defecto": la autora
// confirma de oído (ver Sonido.md).

//go:embed assets/sound/*.ogg
var soundFS embed.FS

// sampleRate es el de la salida de audio; todo se remuestrea a él.
const sampleRate = 44100

// allCues son todos los efectos, en el orden en que se cargan.
var allCues = []Cue{CueClick, CueDischarge, CueWake, CueNoBed, CueEndShift, CueSnore, CueRoll}

// loadOGG abre un OGG embebido como PCM de 16 bits, estéreo, a sampleRate.
func loadOGG(file string) (*vorbis.Stream, error) {
	raw, err := soundFS.ReadFile("assets/sound/" + file)
	if err != nil {
		return nil, err
	}
	return vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(raw))
}

// Sound es la mesa de sonido del juego. Con out == nil no suena nada (modo
// pruebas o sin tarjeta de sonido): solo anota los cues para el volcado.
type Sound struct {
	out *soundOut

	log     []string // cues que se pidieron (para el volcado de pruebas)
	enabled bool     // el interruptor "Sonido" de Ajustes
	musicG  float64  // volumen de la música (0 a 1, ya con la curva de volumeGain)
	sfxG    float64  // volumen de los efectos (0 a 1, ya con la curva)

	playing Track   // pista que suena
	level   float64 // nivel del fundido de esa pista (0 a 1)
	duck    float64 // 1 normal, duckVolume en pausa

	rolling    bool // la camilla suena ahora
	rollWanted bool // alguien la pidió en este tick
}

// soundOut es lo que de verdad suena: el contexto, los efectos ya cargados y
// los reproductores en uso.
type soundOut struct {
	ctx    *audio.Context
	sfx    map[Cue][]byte // PCM ya normalizado de cada efecto
	music  *audio.Player
	roll   *audio.Player
	active []*audio.Player // efectos sonando (para soltarlos al terminar)
}

// NewSound arma la mesa de sonido. Con silent (modo pruebas) no abre audio. Si
// la tarjeta de sonido falla, el juego sigue en silencio.
func NewSound(silent bool) *Sound {
	s := &Sound{enabled: true, duck: 1, musicG: volumeGain(70), sfxG: volumeGain(80)}
	if silent {
		return s
	}
	out, err := newSoundOut()
	if err != nil {
		return s
	}
	s.out = out
	return s
}

func newSoundOut() (*soundOut, error) {
	o := &soundOut{ctx: audio.NewContext(sampleRate), sfx: make(map[Cue][]byte)}
	for _, c := range allCues {
		stream, err := loadOGG(cueFile(c))
		if err != nil {
			return nil, err
		}
		pcm, err := io.ReadAll(&gainReader{stream, fromDB(soundGainDB[cueFile(c)])})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", cueFile(c), err)
		}
		o.sfx[c] = pcm
	}
	return o, nil
}

// Log devuelve los cues pedidos hasta ahora (solo para el volcado de pruebas).
func (s *Sound) Log() []string {
	if s == nil {
		return nil
	}
	return s.log
}

// Music dice qué pista suena (para el volcado de pruebas).
func (s *Sound) Music() Track {
	if s == nil {
		return TrackNone
	}
	return s.playing
}

// Play hace sonar un efecto. Con el sonido apagado no hace nada.
func (s *Sound) Play(c Cue) {
	if s == nil || !s.enabled {
		return
	}
	s.log = append(s.log, string(c))
	if s.out == nil {
		return
	}
	if s.sfxG == 0 {
		return // volumen de efectos en 0: nada que reproducir
	}
	p := s.out.ctx.NewPlayerFromBytes(s.out.sfx[c])
	p.SetVolume(s.sfxG)
	p.Play()
	s.out.active = append(s.out.active, p)
}

// SetEnabled prende o apaga el sonido (el interruptor de Ajustes). Apagado no
// suena ningún efecto ni la música.
func (s *Sound) SetEnabled(on bool) {
	if s != nil {
		s.enabled = on
	}
}

// SetVolumes pone los volúmenes de Ajustes (0 a 100) de la música y de los
// efectos. Se llama en cada tick con lo guardado, así lo que se arrastra en la
// barra se oye de inmediato: la música lo toma en el próximo update y los
// efectos que ya suenan (ronquido, camilla) también.
func (s *Sound) SetVolumes(music, sfx int) {
	if s != nil {
		s.musicG, s.sfxG = volumeGain(music), volumeGain(sfx)
	}
}

// MusicGain y SfxGain dan el volumen efectivo (0 a 1) para el volcado de
// pruebas: 0 si el sonido está apagado.
func (s *Sound) MusicGain() float64 {
	if s == nil || !s.enabled {
		return 0
	}
	return s.musicG
}

func (s *Sound) SfxGain() float64 {
	if s == nil || !s.enabled {
		return 0
	}
	return s.sfxG
}

// Roll pide que suenen las ruedas de la camilla en este tick. Hay que llamarlo
// en cada tick mientras se vea una camilla; si deja de llamarse, se paran.
func (s *Sound) Roll() {
	if s != nil {
		s.rollWanted = true
	}
}

// Update lo llama la aplicación una vez por tick, después de actualizar la
// pantalla: want es la pista que toca, paused dice si el juego está en pausa
// (la música baja) y enabled es el interruptor de Ajustes.
func (s *Sound) Update(want Track, paused, enabled bool) {
	if s == nil {
		return
	}
	s.enabled = enabled
	s.duck = approach(s.duck, duckTarget(paused), (1-duckVolume)/duckTicks)

	var start bool
	s.playing, s.level, start = musicStep(s.playing, want, s.level)
	roll := s.rollWanted && enabled
	s.rollWanted = false
	if roll && !s.rolling {
		s.log = append(s.log, string(CueRoll))
	}
	s.rolling = roll
	if s.out != nil {
		s.out.update(s, start, roll)
	}
}

// update mueve los reproductores reales según el estado ya calculado.
func (o *soundOut) update(s *Sound, start, roll bool) {
	if start {
		o.openMusic(s.playing)
	}
	if o.music != nil {
		o.music.SetVolume(musicVolumeNow(s.level, s.duck, s.musicG, s.enabled))
		switch {
		case !s.enabled && o.music.IsPlaying():
			o.music.Pause()
		case s.enabled && !o.music.IsPlaying():
			o.music.Play()
		}
	}
	o.updateRoll(roll, s.sfxG)
	// Se sueltan los efectos que ya terminaron; los que siguen sonando toman
	// el volumen de efectos de ahora.
	kept := o.active[:0]
	for _, p := range o.active {
		if p.IsPlaying() {
			p.SetVolume(s.sfxG)
			kept = append(kept, p)
		} else {
			p.Close()
		}
	}
	o.active = kept
}

// openMusic cierra la pista anterior y abre la nueva en bucle. Si algo falla,
// se queda sin música (el juego sigue).
func (o *soundOut) openMusic(t Track) {
	if o.music != nil {
		o.music.Close()
		o.music = nil
	}
	if t == TrackNone {
		return
	}
	file := trackFile(t)
	stream, err := loadOGG(file)
	if err != nil {
		return
	}
	loop := audio.NewInfiniteLoop(&gainReader{stream, fromDB(soundGainDB[file])}, stream.Length())
	p, err := o.ctx.NewPlayer(loop)
	if err != nil {
		return
	}
	p.SetVolume(0)
	o.music = p
}

// updateRoll arranca o frena el bucle de las ruedas.
func (o *soundOut) updateRoll(on bool, gain float64) {
	if on && o.roll == nil {
		loop := audio.NewInfiniteLoop(bytes.NewReader(o.sfx[CueRoll]), int64(len(o.sfx[CueRoll])))
		p, err := o.ctx.NewPlayer(loop)
		if err != nil {
			return
		}
		p.Play()
		o.roll = p
	}
	if o.roll != nil {
		o.roll.SetVolume(gain)
	}
	if !on && o.roll != nil {
		o.roll.Close()
		o.roll = nil
	}
}

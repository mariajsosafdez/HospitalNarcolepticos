package game

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Settings son los ajustes del jugador (sección 10). El sonido tiene un
// interruptor general (SoundOff) y dos volúmenes. Un guardado viejo sin
// "musicVolume" o "sfxVolume" queda con los de fábrica: Load parte de
// DefaultSettings y json.Unmarshal solo pisa los campos que sí vienen.
type Settings struct {
	SoundOff     bool `json:"soundOff"`     // sonido apagado (se guarda al revés para que, por defecto, esté prendido)
	MusicVolume  int  `json:"musicVolume"`  // 0-100
	SfxVolume    int  `json:"sfxVolume"`    // 0-100
	BlinkOnSleep bool `json:"blinkOnSleep"` // parpadeo de cámara cuando alguien se duerme
	BlinkOnWake  bool `json:"blinkOnWake"`  // ... y cuando alguien despierta
	Fullscreen   bool `json:"fullscreen"`
}

// DefaultSettings devuelve los ajustes de fábrica: parpadeos apagados,
// música 70 %, efectos 80 %, sonido prendido y ventana normal.
func DefaultSettings() Settings {
	return Settings{MusicVolume: 70, SfxVolume: 80}
}

// SaveData es lo que se autoguarda: el último día desbloqueado, las mejores
// estrellas por día (índice 0 = día 1) y los ajustes.
type SaveData struct {
	LastUnlockedDay int         `json:"lastUnlockedDay"`
	BestStars       [MaxDay]int `json:"bestStars"`
	Settings        Settings    `json:"settings"`
	TutorialSeen    bool        `json:"tutorialSeen"` // el tutorial del Día 1 ya se vio completo
}

// NewSaveData devuelve una partida nueva: solo el día 1 desbloqueado.
func NewSaveData() SaveData {
	return SaveData{LastUnlockedDay: 1, Settings: DefaultSettings()}
}

// RecordShift anota el resultado de un turno: mejora las estrellas del día
// si hace falta y desbloquea el siguiente día.
func (d *SaveData) RecordShift(day, stars int) {
	if day < 1 || day > MaxDay {
		return
	}
	d.BestStars[day-1] = max(d.BestStars[day-1], stars)
	if day < MaxDay {
		d.LastUnlockedDay = max(d.LastUnlockedDay, day+1)
	}
}

// HasProgress dice si hay algo que continuar: algún día jugado hasta el
// final (estrellas guardadas) o un día siguiente desbloqueado. Guardar solo
// los ajustes no cuenta, para que "Continuar" no se prenda sin haber jugado.
func (d SaveData) HasProgress() bool {
	if d.LastUnlockedDay > 1 {
		return true
	}
	for _, stars := range d.BestStars {
		if stars > 0 {
			return true
		}
	}
	return false
}

// SaveStore lee y escribe el JSON de guardado. La ruta se inyecta al crearlo:
// así los tests (y la puerta -agent) usan una ruta temporal y nunca tocan el
// guardado real de la autora.
type SaveStore struct {
	path string
}

// NewSaveStore crea un almacén que guarda en path.
func NewSaveStore(path string) *SaveStore { return &SaveStore{path: path} }

// DefaultSavePath es la ruta real del guardado: dentro de la carpeta de
// configuración del usuario (en Windows, %AppData%).
func DefaultSavePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "HospitalCostenos", "save.json"), nil
}

// Load lee el guardado. Si el archivo no existe devuelve una partida nueva
// (no es un error: es la primera vez que se juega).
func (st *SaveStore) Load() (SaveData, error) {
	raw, err := os.ReadFile(st.path)
	if errors.Is(err, os.ErrNotExist) {
		return NewSaveData(), nil
	}
	if err != nil {
		return NewSaveData(), err
	}
	data := NewSaveData()
	if err := json.Unmarshal(raw, &data); err != nil {
		return NewSaveData(), err
	}
	return data, nil
}

// Save escribe el guardado, creando la carpeta si hace falta.
func (st *SaveStore) Save(data SaveData) error {
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(st.path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(st.path, raw, 0o644)
}

// Exists dice si ya hay un archivo de guardado. El menú lo usa para dejar
// "Continuar" gris cuando es la primera vez que se juega.
func (st *SaveStore) Exists() bool {
	_, err := os.Stat(st.path)
	return err == nil
}

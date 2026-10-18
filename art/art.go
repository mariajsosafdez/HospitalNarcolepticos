// Package art expone los PNG y la paleta de esta carpeta como un sistema de
// archivos embebido en el ejecutable. go:embed solo puede leer archivos que
// estén en la carpeta del paquete (o debajo), y gameview vive en otra
// carpeta, así que el paquete que embebe tiene que ser este.
//
// Solo se leen los PNG a 1x (los de @4x son para revisar en Figma). Los
// sprites se regeneran con `go run ./tools/sprites`; para que el juego vea
// un PNG nuevo hay que volver a compilar.
package art

import "embed"

// Files contiene png/*.png y palette.hex.
//
//go:embed png/*.png palette.hex
var Files embed.FS

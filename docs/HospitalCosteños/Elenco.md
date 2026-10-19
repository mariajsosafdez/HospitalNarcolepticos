# Elenco fijo

> ✅ **Aprobado** el 2026-10-17: "revisé el elenco y los textos y me gustan bastante, aprobados :D"
>
> Los 5 pacientes, los 2 doctores y el camillero del **escenario obligatorio** (`main.go`) están incluidos con el mismo ID, nombre, edad y nivel, para que el modo Demo use exactamente esos datos.

## Pacientes (10)

| ID | Nombre | Edad | Nivel | Personalidad / cómo se ve | Demo |
|---|---|---|---|---|---|
| P-001 | Yesid Pacheco | 34 | Moderado | Vendedor de raspao. Camiseta del Junior; no para de hablar del partido | ✅ |
| P-002 | Yeimy Padilla | 28 | Severo | Bailarina de champeta. Se duerme a mitad de un paso, con los brazos arriba | ✅ |
| P-003 | Rafa Cantillo | 52 | Leve | Taxista con sombrero vueltiao. Discute de béisbol con cualquiera | ✅ |
| P-004 | Wilfrido Berrio | 45 | Severo | Pescador de Ciénaga con su nevera de icopor. Ronca durísimo | ✅ |
| P-005 | Marelvis Ortega | 61 | Severo | La señora del sancocho: llega con su olla y le ofrece a todo el mundo | ✅ |
| P-006 | Nayeli Mercado | 19 | Leve | Estudiante con audífonos; se duerme encima del celular | |
| P-007 | Abelardo "Mañe" Julio | 73 | Moderado | Abuelito con abanico de palma y guayabera; cuenta historias de "cuando yo era pelao" | |
| P-008 | Kelly Johana Ruiz | 37 | Moderado | Peluquera, siempre con rulos puestos y una revista | |
| P-009 | Eder Salcedo | 41 | Leve | Vallenatero con su acordeón; se queda dormido tocando | |
| P-010 | Yuranis Polo | 25 | Moderado | Vendedora de bollos ("¡bollo limpio, bollo 'e yuca!") con su ponchera en la cabeza | |

Niveles: 3 Leves, 4 Moderados, 3 Severos (al menos 2 severos, como pide el enunciado).

## Doctores (3)

| ID | Nombre | Especialidad | Personalidad / cómo se ve | Demo |
|---|---|---|---|---|
| D-001 | Dra. Karen Ospina | Neurología | Tranquila y ordenada, gafas y moño | ✅ |
| D-002 | Dr. Álvaro Mendoza | Medicina del sueño | Veterano, bata sobre guayabera, bigote | ✅ |
| D-003 | Dra. Luz Dary Fontalvo | Medicina general | Acelerada, siempre con un tinto en la mano. Aparece desde el Día 3 | |

## Camilleros (2)

| ID | Nombre | Personalidad / cómo se ve | Demo |
|---|---|---|---|
| O-001 | Chepe Barrios | Silba vallenatos mientras empuja la camilla; chancletas | ✅ |
| O-002 | Jhon Jairo "El Mono" Arrieta | Grandote y callado, gorra al revés. Aparece desde el Día 4 | |

## Director

**Don Ernesto Cassiani**, director del hospital. Guayabera blanca, sombrero vueltiao y un ventilador de mano. Habla con refranes y mucho cariño, pero cuando el turno sale mal, regaña con drama. Sale en el tutorial y en los mensajes de fin de turno (ver [[Textos del Director]]).

## Easter eggs ✅ (2026-10-19)

Pacientes extra, raros, inspirados en fotos de la autora (las fotos no están en el repo; solo el pixel art). Sprites `pac_egg<N>_*`.

| N | Nombre | Nivel | Cómo se ve |
|---|---|---|---|
| 1 | **Misifu** (gato) | Severo | Persa naranja y crema, cara aplastada de bravo, melena blanca. Collar rojo |
| 2 | **Alejo Urrego** | Leve | Crespo oscuro, gafas rectangulares, camiseta azul oscuro, pulgar arriba |
| 3 | **Samu Giraldo** | Leve | Mullet crespo, suéter gris |
| 4 | **Cote** | Severo | Mullet largo y ondulado, sin gafas ("quitale las gafas al cote y hazle mas mullet porfi"), camiseta crema, shorts de jean, medias con margaritas, reloj grande |
| 5 | **Sebs** | Moderado | Crespo corto, camiseta blanca de maratón, shorts verdes, mano de rock |

**Reglas** (palabras de la autora: "que sea raro que salga, por cierto, los dos que tienes severo se duermen mucho mas seguido que los demás, evitando el cooldown para poderse volver a dormir"):
- En cada turno hay **15 % de probabilidad** de que llegue **un** easter egg al azar, como paciente extra.
- **Misifu y Cote** (los severos) se duermen **mucho más seguido** que los demás y **no tienen la espera de 10 min** después del alta para volver a dormirse.
- Solo para el gato, Don Ernesto dice al llegar: "Erda, ¿y ese gato quién lo dejó entrar?" ("sisiisisi solo para el gato").

## Lugar de la Demo ✅

En el escenario, P-002 se duerme en el "champeta dance hall". Por eso se agregó la cámara **CAM 8 · Salón de champeta** (ver [[Mapa de Cámaras]]). Aprobado: "Sipi".

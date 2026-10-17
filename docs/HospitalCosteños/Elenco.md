# Elenco fijo

> Estado: **propuesta, esperando aprobación**. El elenco fijo se aprobó el 2026-10-17 ("el elenco hagamoslo fijo"); los nombres y personalidades de abajo son la propuesta de Claude.
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

## Easter eggs (si sobra tiempo)

Personajes extra que aparecen muy de vez en cuando. Se definen después.

## Por decidir

- La Demo dice que la paciente P-002 se durmió en el **"champeta dance hall"**, pero ese lugar no está entre las cámaras. Propuesta: **agregar la cámara "Salón de champeta"** (CAM 8, la sala de terapia de baile). Así la Demo coincide con el escenario sin cambiar `main.go`.

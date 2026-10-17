# Game Design

> Las secciones marcadas **(propuesta)** son ideas de Claude para llenar huecos. Nada está decidido hasta que pase al [[#Registro de decisiones]].

## La idea en palabras de Maria José

> "La idea es que sea de tipo revisión de cámaras de seguridad, donde cada cámara es una habitación del hospital, a un lado, habrá un panel con la información reducida de camilleros y doctores, en las cámaras, iran apareciendo personajes que pueden o no quedarse dormidos en la sala, el rol del jugador es asignarle a esos pacientes que se duerman un doctor/camillero disponible y este lo mandará a una habitación, el turno constará de 6horas (en más o menos 3 min de la vida real), la distribución de las cámaras estará basada en las cámaras de la saga de FNAF, el juego tendrá un sistema de rating dependiendo de la demora en atender personajes. los camilleros/doctores tienen tiempos de reutilización después de desplazar pacientes, además de que hay que estar verificando las habitaciones para ver si los pacientes ya se despertaron y asi desocupar la habitación."

> "la parte visual quiero que sea de ambiente cálido, aunque el sistema de cámaras esté basado en FNAF, no hace falta que las paletas de colores y la ambientación lo estén, el juego busca ser algo más de ambiente relajado."

## Resumen en una frase

Eres quien vigila las cámaras del hospital durante un turno de 6 horas: cuando un costeño se duerme en algún lugar, le mandas un doctor o un camillero libre para que lo lleve a una cama, y vas revisando las habitaciones para liberar las camas de los que ya despertaron.

## El bucle de juego

1. **Mirar.** Cambias entre cámaras con el mapa (como en FNAF). Los pacientes entran por recepción y caminan por las zonas.
2. **Detectar.** Un paciente se duerme (cae dormido, aparecen "zzz"). Empieza a correr su **tiempo de espera**.
3. **Asignar.** Seleccionas al paciente dormido y eliges a alguien libre del panel de personal. El personal lo lleva a la primera cama libre.
4. **Esperar cooldown.** Ese doctor/camillero queda ocupado un rato (traslado + descanso).
5. **Revisar habitaciones.** Los pacientes se despiertan solos en su cama. Tienes que verlo en la cámara de la habitación y pulsar **"Dar de alta"** para liberar la cama.
6. **Repetir** hasta las 6 horas. Al final: reporte del turno y estrellas.

Si no hay cama libre, el paciente se queda dormido en el pasillo (igual que en el enunciado) y su espera sigue contando hasta que liberes una cama.

## Tiempo ✅

| Juego | Real |
|---|---|
| Turno completo: 6 h | 3 min |
| 1 hora | 30 s |
| 1 minuto | 0,5 s |

**Turno de la tarde, 12:00 → 18:00.** Encaja con el enunciado (se duermen "a las 2:15 PM en medio de una frase" después del sancocho) y con el ambiente cálido: la luz de las cámaras pasa de mediodía blanco a atardecer naranja. Es el "contrario relajado" de las noches de FNAF.

## Personal ✅

Panel lateral, una tarjeta por persona: nombre, rol, estado y barra de cooldown.

| | Doctor | Camillero |
|---|---|---|
| Estados | Disponible / En traslado / Descansando / Sin cupo | Disponible / En traslado / Descansando |
| Traslado + descanso | 5 s + 15 s | 3 s + 5 s |
| Cupo | **máx. 4 pacientes a cargo por turno** (regla real del modelo: `maxPatientsPerDoctor = 4`, y no se liberan al despertar) | ilimitado |
| Ventaja | Diagnostica: atender a un paciente **Severo** con doctor da bonus de rating | Rápido y siempre disponible después del cooldown |

Esto hace que la decisión sea interesante: el doctor es "caro" (cooldown largo y cupo limitado), el camillero es "barato".

✅ **Exceso de trabajo** (idea de la autora): si un doctor llena su cupo del día, el hospital pierde puntos. Y en cada partida los pacientes no deben pasar mucho del cupo total de los doctores. Detalle en [[Lógica del juego#4.5 Cupo del doctor (regla del modelo)]].

## Pacientes ✅

El nivel de narcolepsia del modelo (Mild / Moderate / Severe) cambia el comportamiento. Se muestra con el color de la manilla del paciente.

| Nivel | Prob. de dormirse | Duerme en cama |
|---|---|---|
| Leve | baja | 30–60 min de juego |
| Moderado | media | 1–2 h |
| Severo | alta | 2–3 h |

Personajes con identidad costeña: sombrero vueltiao, abanico, el que baila champeta, el que discute de béisbol, la señora con el plato de sancocho. Cada uno tiene nombre (como "Wilfrido Berrio" del enunciado).

## Rating ✅

Por cada paciente se mide **tiempo de espera** = desde que se duerme hasta que llega a una cama (por confirmar, pregunta 11). Puntos exactos en [[Lógica del juego#7. Puntuación (`score.go`)]].

| Espera (juego) | Espera (real) | Nota |
|---|---|---|
| ≤ 20 min | ≤ 10 s | Excelente |
| ≤ 45 min | ≤ 22 s | Bien |
| ≤ 90 min | ≤ 45 s | Regular |
| > 90 min | > 45 s | Mal |

Además:

- **Cama ocupada por alguien despierto:** si un paciente despierta y no le das de alta en ~15 min de juego, empieza a restar.
- **Bonus:** Severo atendido por doctor.
- Al final: **1 a 5 estrellas** según el promedio.

Como el juego es relajado, **no hay game over**: el peor resultado es 1 estrella y un mensaje divertido del Director sobre la mala administración. Las penalizaciones pesan más cada día.

## Reporte de fin de turno = las 4 consultas del enunciado

Idea clave para la defensa: la pantalla final del turno **es** el resultado de las consultas de la sección 5:

| Consulta | Cómo se ve en el reporte |
|---|---|
| 5.1 `PatientsInHallway` | "Se quedaron en el pasillo" |
| 5.2 `MyEpisodes` | Ficha de cada doctor con sus episodios |
| 5.3 `AssignRoom` | Ocupación de camas / veces que no hubo cama |
| 5.4 `SevereReport` | Pacientes severos y cuántos episodios tuvieron |

## Menú principal ✅

| Opción | Qué hace |
|---|---|
| **Nuevo turno** | Empieza en el Día 1 |
| **Continuar** | Sigue en el último día desbloqueado |
| **Demo (Escenario del enunciado)** | Reproduce la sección 6 paso a paso en las cámaras (ver abajo) |
| **Cómo jugar** | Tutorial corto con imágenes |
| **Ajustes** | Volumen música/efectos, parpadeo de cámaras (apagado por defecto), pantalla completa. Autoguardado |
| **Créditos** | Autora, música (licencias CC), curso |

### Progresión por días ✅

Como las noches de FNAF: Día 1 tranquilo (pocos pacientes, 3 camas) y cada día sube un poco (más pacientes, más cámaras, otra cama, menos personal). 5 días.

### Modo Demo

Turno guionado, sin azar, que reproduce exactamente el escenario obligatorio:

1. 2 doctores, 1 camillero, 3 habitaciones de capacidad 1, 5 pacientes (2 severos).
2. 4 pacientes se duermen en lugares distintos. 3 consiguen cama; el 4.º se queda en el pasillo y aparece el mensaje de error legible.
3. Uno despierta, se libera la cama y el del pasillo pasa a esa cama.
4. Se muestra el reporte con las 4 consultas.

Con botones **Siguiente paso / Auto**, para usarlo en la defensa.

## Plan de trabajo (mínimo viable)

✅ Orden aprobado, con la Demo adelante:

1. Menú + **Demo** (escenario del enunciado en las cámaras)
2. Día 1 + reporte de fin de turno
3. Días 2–5
4. Sonido
5. Tutorial (mensajes del Director)
6. *Si sobra tiempo:* personajes easter egg

Todo se trabaja en la rama **`game`**. Entrega del enlace del repo: 19/10, 23:59. Sustentación: 20/10.

## Preguntas abiertas

Rondas 1 y 2 respondidas (ver el registro). Ronda 3, con mi recomendación en **negrita**:

1. **Al cerrar la ventana del juego**, ¿termina todo el programa (incluido el dashboard web)? **Sí**: un solo programa, una sola forma de cerrarlo. El dashboard se usa mientras la ventana está abierta.
2. **Camillero 3 s**: ¿aplica al traslado (5 s → 3 s), al descanso (5 s → 3 s) o a ambos? Lo dejé en el **traslado**.
3. **Probabilidad de irse tras el alta**: **Leve 60 %, Moderado 40 %, Severo 20 %**.
4. **Movimiento aleatorio**: **cada paciente se mueve cada 5–25 min de juego (2,5–12,5 s), sorteado cada vez**.
5. **Multiplicador de penalizaciones por día**: **×1,0 / ×1,25 / ×1,5 / ×1,75 / ×2,0**.
6. **Título en el menú**: **"Hospital de los Costeños con Narcolepsia"** con un subtítulo tipo *"Turno de la tarde"*. ¿O un nombre más corto de juego?
7. **Herramientas de arte**: ¿tienes **Figma** (con el plugin Figwright abierto) y **Aseprite** instalados? Si no hay Aseprite, los sprites se pueden hacer como PNG por código. Es más lento de ajustar, pero funciona.
8. **Fuente pixel**: **una fuente pixel libre (licencia OFL) que tenga tildes y ñ**.

## Registro de decisiones

| Fecha | Decisión | Por qué | Palabras de la autora |
|---|---|---|---|
| 2026-10-17 | Juego de vigilancia de cámaras sobre el modelo del hospital | Idea original | "La idea es que sea de tipo revisión de cámaras de seguridad…" |
| 2026-10-17 | Cámaras al estilo FNAF, pero ambiente cálido y relajado | Idea original | "el juego busca ser algo más de ambiente relajado" |
| 2026-10-17 | Turno de 6 h ≈ 3 min reales | Idea original | "el turno constará de 6horas (en más o menos 3 min de la vida real)" |
| 2026-10-17 | Primero bóveda de ideas, luego Figma, luego código | Orden de trabajo | "aun no vamos a escribir código, primero definiremos muy bien la idea" |
| 2026-10-17 | Se apunta a tenerlo para la entrega | La autora cree que da el tiempo | "la entrega es el 20 realmente, yo creo que nos da para hacerlo para el 20 :D" |
| 2026-10-17 | Pixel art | Se ve mejor | "El juego será en pixelart, asi se pueden ver mejor las cosas" |
| 2026-10-17 | Fuera del navegador (ventana aparte; por confirmar qué significa) | Preferencia | "me gustaría que el sitio de ejecución sea una pestaña extra, no una en el navegador como teniamos en la interfaz" |
| 2026-10-17 | El personal lleva al paciente a la primera habitación libre | Igual que `AssignRoom` | "el jugador no elije la habitación, cuando asigna un personal médico y un paciente, este va automaticamente a la primera libre" |
| 2026-10-17 | Parpadeo de cámara al dormirse **y al despertarse**, opción apagada por defecto | Ayuda opcional | "me gustaría que la cámara parpadee si se duerme alguien, asi como si se despierta, pero que sea una opción que por default esté desactivada" |
| 2026-10-17 | Los pacientes se mueven entre cámaras | Que no sean estáticos | "Los pacientes se van a mover por las cámaras, la idea es que no sean estáticos." |
| 2026-10-17 | Cámaras aprobadas como en [[Mapa de Cámaras]] | — | "Me gusta la propuesta de cámaras que me dices." |
| 2026-10-17 | Dormido sin atender y despierto sin alta viven solo en `game` | Huecos 2 y 3 | "los huecos a llenar relacionados al modelo, punto 2 y 3 son como dices" |
| 2026-10-17 | Pacientes por partida cerca del cupo de los doctores + penalización por exceso de trabajo | Que se usen los doctores | "si un doctor llena su cupo diario, baje un poco los puntos del hospital por 'exceso de trabajo'" |
| 2026-10-17 | Los cambios de lógica se documentan en una nota aparte | Para entenderlos y defenderlos | "Los cambios que dices de la parte lógica hay que documentarlos muy bien" → [[Lógica del juego]] |
| 2026-10-17 | Motor Ebitengine; el juego se abre con `go run .` | Ventana propia, fuera del navegador | "Si, vamos a usar Ebltengine y que se abra con go run ." |
| 2026-10-17 | Trabajar en la rama `game`. Repo el 19/10, sustentación el 20/10 | — | "trabajemos en la rama game" |
| 2026-10-17 | Mínimo viable con la Demo primero, luego Día 1 + reporte | — | "podríamos cambiar de orden el día 1 + reporte por la demo" |
| 2026-10-17 | 480×270 escalado; sin oficina, siempre desde las cámaras | — | "la resolución me parece bien, no hay 'oficina', siempre se trabaja desde las cámaras" |
| 2026-10-17 | Paciente sin cama se queda donde se durmió (cualquier zona) y se manda a cama manualmente | — | "tengo que asignarlo manualmente igual" / "se puede dormir en cualquier sala, no solo en el pasillo" |
| 2026-10-17 | Español con sabor costeño, elenco fijo, easter eggs si sobra tiempo | — | "dejemos la parte visual con español con sabor costeño, el elenco hagamoslo fijo (igual aparte se hace lo del easter egg)" |
| 2026-10-17 | Tutorial con mensajes del Director | — | "me gusta la idea del tutorial" |
| 2026-10-17 | Con 1★, mensaje divertido del Director | — | "si el hospital llega a 1 estrella, puede salir un mensaje divertido del director diciendo algo relacionado a la mala administración" |
| 2026-10-17 | La espera cuenta hasta llegar a cama; tras el alta el paciente se queda o se va | Preguntas 11 y 12 | "los pacientes pueden o quedarse en el hospital o irse de él" |
| 2026-10-17 | Penalizaciones más fuertes cada día | Progresión | "mientras suba el día, las acciones que restan puntos pueden valer un poco más" |
| 2026-10-17 | Movimiento de pacientes por ticks aleatorios | Más difícil seguirlos | "Me gustaría que los pacientes se rijan por ticks aleatorios para su movimiento" |
| 2026-10-17 | Volver a dormirse: mínimo 10 min después del alta | — | "un paciente dado de alta puede dormirse a los 10 min mínimo (si no se va del hospital)" |
| 2026-10-17 | Tiempos de personal aprobados; camillero 3 s | — | "solo cambia que el camillero sea de 3s, no de 5s" |
| 2026-10-17 | Aprobados: tiempo, personal, pacientes, rating, puntos, menú, días, pausa, fin de turno, demo, ajustes, autoguardado, ambiente (16–21) | — | "me gusta el fin del turno, me gusta tu plantemiento del demo. me gusta la parte de ajustes y autoguardado" |

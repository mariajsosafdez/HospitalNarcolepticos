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

1. 2 doctores, 1 camillero, 3 habitaciones de capacidad 1, 5 pacientes (3 severos, como en `main.go`).
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

Rondas 1, 2 y 3 respondidas (ver el registro). Las pendientes de personajes y textos están en [[Elenco]] y [[Textos del Director]].

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
| 2026-10-17 | El descanso del camillero se queda en 5 s (solo el traslado baja a 3 s) | — | "nah, no le cambies el descanso al camillero" |
| 2026-10-17 | Aprobados: irse tras el alta 60/40/20 %, movimiento cada 5–25 min de juego, penalizaciones ×1,0–×2,0, fuente pixel libre con tildes y ñ | Ronda 3 | "3, 4, 5, 6 y 8 aprobados con tu recomendación" |
| 2026-10-17 | Título: **Hospital de los Costeños con Narcolepsia**, subtítulo **"El hueco más profundo de Montería"** | — | "para el título, cambia turno de la tarde por 'El hueco más profundo de Montería'" |
| 2026-10-17 | No hay Aseprite: sprites como PNG generados por código (ver [[Lógica del juego]] → arte); Figma + Figwright para las pantallas | Herramientas disponibles | "Aseprite no lo tengo, pero si figma con figwright abierto" |
| 2026-10-17 | No se usará el dashboard web; separarlo del juego | — | "Hay manera de tener el dashboard web por aparte del juego externo? no creo que vayamos a usar el dashboard la verdad." |
| 2026-10-17 | Aprobados: dirección de arte, manillas verde/amarillo/rojo, elenco fijo, textos del Director, lista de pantallas de Figma | — | "me gusta la dirección de arte, la manilla de cada nivel como la recomendaste. me gusta el elenco fijo, los textos de recomendación y las pantallas de figma" |
| 2026-10-17 | `go run .` abre el juego sin dashboard; `go run . -web` abre el dashboard sin juego; `web/` se queda en el repo | Respaldo para la defensa | "Sip" |
| 2026-10-17 | Equipo de agentes (arte, builder, tester) en la rama `game` | Flujo de trabajo | "Sí, me parece bien :D" |
| 2026-10-17 | Nueva cámara CAM 8 · Salón de champeta (la Demo coincide con `main.go`) | Lugar del escenario que faltaba | "Sipi" |
| 2026-10-17 | Elenco fijo y textos del Director aprobados | — | "revisé el elenco y los textos y me gustan bastante, aprobados :D" |
| 2026-10-17 | Menú aprobado con cambios: sin sombrero, sin esquinas de cámara ni REC, ¡ a tamaño 10, atardecer fijo, descripciones del agente aprobadas | Revisión de la vista | "yo creo que es mejor quitarlo" / "nop" / "ok" / "me gusta tu recomendacion" / "me gustan como estan" |
| 2026-10-17 | Tag `entrega-base` en `main` como respaldo; builder arranca con `game/`, `-web` y `nogame`; arte sigue con el elenco y la cámara de la Demo | Plan de trabajo | "Sip" |
| 2026-10-17 | La ventana del juego arranca más pequeña: ×2 (960×540) en vez de ×3 | Se veía muy grande | "tal vez la pantalla del juego puede iniciar mas pequeña? la veo muy grande" |
| 2026-10-17 | Días 3–5 con 10 pacientes (el elenco tiene 10); subir a 12 con personajes extra si sobra tiempo | Elenco fijo | "okeyyy" |
| 2026-10-17 | Aprobadas las decisiones del builder: nivel según el elenco, CAM 2 conectada al Pasillo Este, `MoveToBed` = traslado + medio descanso, errores internos en inglés | Revisión | "sip" / "ok" |
| 2026-10-17 | Dormido sin atender al final del turno resta 10 (× multiplicador del día) en vez de sumar | No fueron atendidos | "cambiemoslo mejor por -10, porque realmente no se atendieron" |
| 2026-10-17 | Pantalla de cámaras aprobada, con retoques (ver Art log) | Revisión de la vista | "si" |
| 2026-10-17 | La cámara se dibuja a ×2; panel y mapa a ×1 | Que los personajes no queden diminutos | "mantenlo así" |
| 2026-10-17 | Una sola cama por habitación (capacidad 1, como el modelo), con más decoración | — | "ok, una sola cama por habitación, pon también más decoración en la habitación" |
| 2026-10-17 | Cámara bloqueada: candado rojo al pasar el mouse | — | "Hay manera de que cuando el usuario haga hover en la cámara aparezca un candado rojo?" |
| 2026-10-17 | Textos nuevos del agente de arte aprobados (pantalla de cámaras y pasos de la Demo) | — | "me gustan" |
| 2026-10-17 | Se dibujan todas las salas que faltan | — | "dibuja también las salas faltantes" |
| 2026-10-17 | La línea de ayuda queda "Enter: dale" sin ¡! | La ¡ de Pixelify parece una i | "quita los ¡!" |
| 2026-10-17 | En pantalla se quita la ¡ inicial y se deja el ! final (las notas conservan el español correcto) | La ¡ de Pixelify Sans parece una i | "Me gusta" |
| 2026-10-17 | Salas nuevas y resto del elenco aprobados, con un cambio: la mayoría se duerme en el piso; solo Mañe y Marelvis en silla (Yeimy de pie) | Que no todos se duerman en sillas | "Me gustaría cambiar que no todos se duerman en sillas, cambialo para que algunos (la gran mayoría) se duerma en el piso como los demás" |
| 2026-10-17 | Mensajes del builder aprobados: "Un momentico", "Cupo lleno!", "Turno terminado", "Ya no se puede" | — | "me gusta" |
| 2026-10-17 | Panel de personal con tarjetas compactas para que quepan 5 sin páginas | Días 3–5 | "okeyyy" |
| 2026-10-18 | La ayuda "Elige a alguien del panel…" lleva borde oscuro, como los nombres de las cámaras | Se leía mal sobre la pared clara | "ponle borde negro, asi como hicimos para los nombres de las cámaras" |
| 2026-10-18 | Dormidos en el piso aprobados | — | "okey" |
| 2026-10-18 | Pantallas del turno aprobadas (reporte, Director, pausa, ajustes, panel de 5) y sus textos; solo "Menú" al final de la Demo y del Día 5; salir desde la pausa abandona el turno | Revisión | "sip" |
| 2026-10-18 | Yeimy: brazos unidos (3 px abajo, no 1) y 2 px más arriba cuando los alza | Que se vea que baila | "subele un poco más los brazos cuando ella los alce, tipo unos 2 pixeles más" |
| 2026-10-18 | Zoom de cámaras confirmado normal; controles de teclado aprobados; retoque de Yeimy aprobado; plan del día aprobado (sonido ahora, tester en la noche, unir a `main` el 19 temprano) | Prueba de la autora | "ya está normal" / "funcionan bien :D" / "sip" |
| 2026-10-18 | Camas al centro de la habitación y cada habitación con distinta organización y objetos | El mensaje del Director tapaba las camas | "podríamos mover las camas de las habitaciones al centro? y buscar que entre las habitaciones cambie un poco el diseño?" |
| 2026-10-18 | Seis habitaciones distintas (101–106) con la cama al centro, aprobadas | — | "sip" |
| 2026-10-18 | Créditos aprobados con cambios: la clase es Compiladores y Lenguajes; Don Ernesto: "Gracias por la ayuda, vale!"; se queda la línea de IA | Revisión | "Es clase de Compiladores y lenguajes :P" / "pon \"gracias por la ayuda vale!\"" / "Sip" |
| 2026-10-18 | Dígitos dibujados en pixel art (el 5 de Pixelify parece una S) | Legibilidad | "okey" |
| 2026-10-18 | El recuadro de la consulta 5.1 se titula "Atendidos sin cama" | Que no parezca contradictorio | "mejor" |
| 2026-10-19 | El recuadro de 5.1 pasa a "Sin cama al cerrar el turno" con dos grupos: "Atendidos, sin cama (5.1)" y "Sin atender" | El reporte no contaba a los que nunca se tocaron (el modelo los ve despiertos) | "no los añade a la cantidad de pacientes que se quedaron sin cama o dormidos en el pasillo" |
| 2026-10-18 | Obstáculos en todas las salas para que nadie se duerma sobre los muebles | — | "me gusta" |
| 2026-10-18 | Dígitos en pixel art aprobados (0 pequeño sin punto, medianos con trazo de 1 px) | El 5 de Pixelify parecía una S | "si" / "ok" / "ok" |
| 2026-10-18 | Nadie se duerme en los últimos 10 s reales del turno (17:40–18:00) | Si no, la penalización sería segura | "los pacientes no deberían dormirse faltando menos de 10 segundos para que se acabe el juego, porque sino seria garantizar una penalización" |
| 2026-10-18 | Pantalla completa nítida (escala entera con bordes) | Se veía borrosa | "el modo de pantalla completa hace que se vea algo borroso el texto y el juego en general" |
| 2026-10-18 | Si no hay cama, el personal no queda "En traslado": atiende y pasa directo a descansar | En la Demo parecía que Karen y Chepe trasladaban a la vez | "el traslado lo hacen la doctora karen y el camillero a la vez, no sé si sea error" |
| 2026-10-18 | Rework de la Demo: reloj quieto entre partes, cámara que sigue cada evento, 8 partes (texto en Textos del Director) | Se veía solo el que quedaba sin cama y el turno se acababa solo | "sipiriri" |
| 2026-10-18 | Las restas al terminar el turno se quedan; un tester comprueba cada caso | Revisión | "ok, haz prubas replicando terminar el turno con esos casos y mira que la resta de puntos si sea correcta" |
| 2026-10-18 | Pantalla completa que llene la pantalla a cualquier resolución (escala fraccionaria nítida) | Con escala entera quedaba pequeña | "bro la verdad está un poco popó, no hay manera de que se adapte a la resolución de cada pantalla?" |
| 2026-10-18 | Textos de la Demo guiada aprobados | — | "me gustan" |
| 2026-10-18 | 5★ desde el 80 % (antes 85 %) | El turno perfecto sin severos no llegaba a 5★ | "okeyy" |
| 2026-10-18 | Las restas se suman y se redondean una vez por tipo con el multiplicador del día | El redondeo por cobro castigaba de más en los días 2 y 4 | "okey" |
| 2026-10-19 | Cinco personajes easter egg a partir de fotos de la autora: gato (Severo), foto 2 (Leve), foto 3 (Leve, sin el Shrek), foto 4 (Severo), foto 5 (Moderado). Diseño primero, luego código. Las fotos no se suben al repo | Easter eggs (si sobra tiempo) | "Hola, vamos a hacer los personajes easter eggs (...) antes de programarlos hagamos los diseños en figma" |
| 2026-10-19 | Nombres de los easter eggs: Misifu (gato), Alejo Urrego, Samu Giraldo, Cote, Sebs | — | "1. Misifu 2. Alejo Urrego 3. Samu Giraldo 4. Cote 5. Sebs" |
| 2026-10-19 | 15 % por turno de que llegue un easter egg; los severos (Misifu y Cote) se duermen mucho más y sin la espera tras el alta; mensaje del Director solo para el gato | Que sea raro y especial | "sisi, que sea raro que salga (...) evitando el cooldown para poderse volver a dormir" / "sisiisisi solo para el gato" |
| 2026-10-19 | `main` adelantado a la rama `game` (fast-forward) y verificado con un clon limpio | Asegurar la entrega | "okey" |
| 2026-10-19 | El desfile de Créditos incluye a todo el elenco (10 pacientes, 3 doctores, 2 camilleros y los 5 easter eggs) | Que salgan todos | "en los creditos incluye a todo el elenco, incluido los del easter egg" |
| 2026-10-19 | El juego se entrega **sin sonido**: se quitan el bloque de música de Créditos y los volúmenes de Ajustes | Prioridad: preparar la sustentación | "Yo creo que mejor deberiamos dejar el juego sin música, no creo que nos dé el tiempo, tengo que preparar la sustentación" |
| 2026-10-19 | Sonido aprobado y unido a la entrega (`main`): bossa en el menú, "Green Salon" en el turno, efectos CC0, barras de volumen | Le gustó al escucharlo | "unela, me gustó bastante" |
| 2026-10-19 | El parpadeo de cámara sigue hasta que el jugador mira esa cámara (o el motivo desaparece) | Solo parpadeaba un momento | "las cámaras no están parpadeando para siempre, solo lo hacen un momento" |

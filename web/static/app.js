// JavaScript de la interfaz. No usa librerías: solo fetch() para hablar con
// la API REST de Go y template strings para armar el HTML.
// Toda la lógica del hospital está en Go; aquí solo se muestran datos y se
// envían acciones.
'use strict';

const $ = (selector) => document.querySelector(selector);

// esc evita que un texto (por ejemplo, un nombre) se interprete como HTML.
const esc = (value) => String(value ?? '').replace(/[&<>"']/g, (c) => (
  { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]
));

// ---------------------------------------------------------------------------
// Comunicación con el servidor
// ---------------------------------------------------------------------------

// api hace la petición y devuelve el JSON. Si el servidor responde con error
// (4xx/5xx), lanza una excepción con el mensaje que mandó Go.
async function api(method, url, body) {
  const options = { method, headers: {} };
  if (body !== undefined) {
    options.headers['Content-Type'] = 'application/json';
    options.body = JSON.stringify(body);
  }
  const response = await fetch(url, options);
  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data.error || `HTTP ${response.status}`);
  }
  return data;
}

let toastTimer;
function toast(text, kind = 'ok') {
  const box = $('#toast');
  box.textContent = text;
  box.className = `toast show ${kind}`;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => { box.className = 'toast'; }, 3500);
}

// notify muestra la respuesta de una acción: mensaje normal o advertencia.
function notify(data) {
  if (data.warning) toast(data.warning, 'warn');
  else toast(data.message || 'Done');
}

// ---------------------------------------------------------------------------
// Pestañas
// ---------------------------------------------------------------------------

let activeTab = 'scenario';

function showTab(name) {
  activeTab = name;
  document.querySelectorAll('.tab').forEach((b) => b.classList.toggle('active', b.dataset.tab === name));
  document.querySelectorAll('.panel').forEach((panel) => {
    panel.classList.toggle('active', panel.id === `tab-${name}`);
  });
  history.replaceState(null, '', `#${name}`); // la URL recuerda la pestaña (sirve al recargar)
  refresh();
}

document.querySelectorAll('.tab').forEach((button) => {
  button.addEventListener('click', () => showTab(button.dataset.tab));
});

// ---------------------------------------------------------------------------
// Piezas de HTML reutilizables
// ---------------------------------------------------------------------------

const STATE_ICON = { awake: '☀️', hallway: '💤', bed: '🛏️' };

function badge(patient) {
  return `<span class="badge ${patient.stateCode}">${STATE_ICON[patient.stateCode]} ${esc(patient.state)}</span>`;
}

function levelTag(level) {
  return `<span class="level ${esc(level.toLowerCase())}">${esc(level)}</span>`;
}

function roomsHTML(rooms) {
  if (rooms.length === 0) return '<p class="empty">No rooms registered.</p>';
  return rooms.map((room) => {
    const who = room.occupants.length === 0
      ? '<span class="mono">empty</span>'
      : room.occupants.map((p) => `🛏️ ${esc(p.id)} ${esc(p.name)}`).join('<br>');
    return `
      <div class="room ${room.available ? 'free' : 'full'}">
        <div class="num">Room ${room.number}</div>
        <div class="meta">${room.occupants.length}/${room.capacity} beds · ${esc(room.state)}</div>
        <div class="who">${who}</div>
      </div>`;
  }).join('');
}

function hallwayHTML(patients) {
  if (patients.length === 0) return '<p class="empty">Nobody is asleep in a hallway. 🎉</p>';
  return `<table><tr><th>ID</th><th>Patient</th><th>Where</th></tr>${
    patients.map((p) => `<tr><td class="mono">${esc(p.id)}</td><td>${esc(p.name)}</td><td>💤 ${esc(p.location)}</td></tr>`).join('')
  }</table>`;
}

function destination(episode) {
  return episode.room ? `room ${episode.room}` : '<b style="color:var(--hallway)">stayed in hallway</b>';
}

function doctorsHTML(groups) {
  return groups.map((group) => {
    const rows = group.episodes.length === 0
      ? '<p class="empty">No episodes.</p>'
      : `<table><tr><th>Time</th><th>Patient</th><th>Location</th><th>Result</th></tr>${
        group.episodes.map((e) => `
          <tr>
            <td class="mono">${esc(e.time.slice(0, 16))}</td>
            <td>${esc(e.patientId)} ${esc(e.patientName)}</td>
            <td>${esc(e.location)}</td>
            <td>${destination(e)}</td>
          </tr>`).join('')
      }</table>`;
    return `<h3>${esc(group.doctorName)} · ${esc(group.specialty)}</h3>${rows}`;
  }).join('');
}

function severeHTML(entries) {
  if (entries.length === 0) return '<p class="empty">No Severe patients.</p>';
  return `<table><tr><th>ID</th><th>Patient</th><th>Episodes today</th></tr>${
    entries.map((e) => `<tr><td class="mono">${esc(e.patient.id)}</td><td>${esc(e.patient.name)}</td><td><b>${e.episodes}</b></td></tr>`).join('')
  }</table>`;
}

function historyHTML(history, limit) {
  const latest = history.slice(-limit).reverse(); // los más recientes primero
  if (latest.length === 0) return '<p class="empty">No episodes yet.</p>';
  return `<table><tr><th>Episode</th><th>Time</th><th>Patient</th><th>Location</th><th>Result</th><th>Attended by</th></tr>${
    latest.map((e) => `
      <tr>
        <td class="mono">${esc(e.id)}</td>
        <td class="mono">${esc(e.time.slice(11))}</td>
        <td>${esc(e.patientId)} ${esc(e.patientName)}</td>
        <td>${esc(e.location)}</td>
        <td>${destination(e)}</td>
        <td>${esc(e.attendedBy)} <span class="mono">${esc(e.role)}</span></td>
      </tr>`).join('')
  }</table>`;
}

// ---------------------------------------------------------------------------
// Pestaña 1: escenario obligatorio (se carga una sola vez, nunca cambia)
// ---------------------------------------------------------------------------

async function loadScenario() {
  try {
    const data = await api('GET', '/api/scenario');
    $('#scenario-log').textContent = data.log;
    $('#sc-hallway').innerHTML = hallwayHTML(data.state.hallway);
    $('#sc-doctors').innerHTML = doctorsHTML(data.state.doctorEpisodes);
    $('#sc-rooms').innerHTML = roomsHTML(data.state.rooms);
    $('#sc-severe').innerHTML = severeHTML(data.state.severeReport);
    $('#clock').textContent = `scenario run at ${data.state.takenAt}`;
  } catch (error) {
    toast(`Could not load the scenario: ${error.message}`, 'error');
  }
}

// ---------------------------------------------------------------------------
// Pestaña 2: gestión del hospital
// ---------------------------------------------------------------------------

function renderManagement(state) {
  const freeRooms = state.rooms.filter((r) => r.available).length;
  const asleep = state.patients.filter((p) => p.stateCode !== 'awake').length;
  $('#stats').innerHTML = `
    <div class="stat"><div class="value">${state.patients.length}</div><div class="label">Patients admitted (${asleep} asleep)</div></div>
    <div class="stat ${state.hallway.length ? 'alert' : ''}"><div class="value">${state.hallway.length}</div><div class="label">Asleep in a hallway</div></div>
    <div class="stat"><div class="value">${freeRooms}/${state.rooms.length}</div><div class="label">Rooms available</div></div>
    <div class="stat"><div class="value">${state.history.length}</div><div class="label">Episodes registered</div></div>`;

  // Los botones llevan data-action y data-id; un solo listener en la tabla
  // los atiende a todos (delegación de eventos).
  $('#patients-table').innerHTML = `
    <tr><th>ID</th><th>Name</th><th>Age</th><th>Level</th><th>State</th><th>Location</th><th>Doctor</th><th>Actions</th></tr>
    ${state.patients.map((p) => `
      <tr>
        <td class="mono">${esc(p.id)}</td>
        <td>${esc(p.name)}</td>
        <td>${p.age}</td>
        <td>${levelTag(p.level)}</td>
        <td>${badge(p)}</td>
        <td>${esc(p.location)}</td>
        <td>${esc(p.doctor || '—')}</td>
        <td>
          ${p.stateCode === 'awake' ? `<button class="btn small" data-action="sleep" data-id="${esc(p.id)}">💤 Sleep</button>` : ''}
          ${p.stateCode !== 'awake' ? `<button class="btn small" data-action="wake" data-id="${esc(p.id)}">☀️ Wake</button>` : ''}
          ${p.stateCode === 'hallway' ? `<button class="btn small" data-action="assign-room" data-id="${esc(p.id)}">🛏️ Assign room</button>` : ''}
        </td>
      </tr>`).join('')}`;

  $('#mg-rooms').innerHTML = roomsHTML(state.rooms);
  $('#mg-history').innerHTML = historyHTML(state.history, 15);
  $('#mg-staff').innerHTML = `<table><tr><th>Staff</th><th>Role</th><th>Episodes</th></tr>${
    state.staff.map((s) => `
      <tr>
        <td>${esc(s.name)}<br><span class="mono">${esc(s.id)}${s.available ? '' : ' · full'}</span></td>
        <td>${s.role === 'Doctor' ? '🩺' : '🛏️'} ${esc(s.role)}</td>
        <td>${s.episodes}</td>
      </tr>`).join('')
  }</table>`;
}

$('#patients-table').addEventListener('click', async (event) => {
  const button = event.target.closest('button[data-action]');
  if (!button) return;
  const { action, id } = button.dataset;
  const url = `/api/patients/${encodeURIComponent(id)}/${action}`;
  try {
    const body = action === 'sleep' ? { location: $('#location-select').value } : {};
    notify(await api('POST', url, body));
  } catch (error) {
    toast(error.message, 'error');
  }
  refresh();
});

// formData convierte un formulario en objeto; los campos numéricos se pasan a número.
function formData(form, numberFields) {
  const data = Object.fromEntries(new FormData(form));
  numberFields.forEach((field) => { data[field] = Number(data[field]); });
  return data;
}

async function submitForm(event, url, numberFields) {
  event.preventDefault();
  const form = event.target;
  try {
    notify(await api('POST', typeof url === 'function' ? url(form) : url, formData(form, numberFields)));
    form.reset();
    updateSpecialtyField();
  } catch (error) {
    toast(error.message, 'error');
  }
  refresh();
}

$('#form-patient').addEventListener('submit', (e) => submitForm(e, '/api/patients', ['age']));
$('#form-room').addEventListener('submit', (e) => submitForm(e, '/api/rooms', ['number', 'capacity']));
$('#form-staff').addEventListener('submit', (e) => submitForm(
  e,
  (form) => (form.role.value === 'doctor' ? '/api/doctors' : '/api/orderlies'),
  ['age'],
));

// El camillero no tiene especialidad: ocultamos ese campo cuando se elige Orderly.
function updateSpecialtyField() {
  const isDoctor = $('#staff-role').value === 'doctor';
  $('#staff-specialty').style.display = isDoctor ? '' : 'none';
  $('#staff-specialty').required = isDoctor;
}
$('#staff-role').addEventListener('change', updateSpecialtyField);
updateSpecialtyField();

$('#reset-btn').addEventListener('click', async () => {
  if (!confirm('Reset the live hospital to its initial data?')) return;
  try {
    notify(await api('POST', '/api/reset'));
  } catch (error) {
    toast(error.message, 'error');
  }
  refresh();
});

// ---------------------------------------------------------------------------
// Pestaña 3: simulación concurrente
// ---------------------------------------------------------------------------

function renderSimulation(state) {
  const running = state.simulation.running;
  const status = $('#sim-status');
  status.textContent = running ? 'Running… goroutines at work' : 'Idle';
  status.classList.toggle('running', running);
  $('#sim-start').disabled = running;

  $('#sim-rooms').innerHTML = roomsHTML(state.rooms);
  $('#sim-hallway').innerHTML = hallwayHTML(state.hallway);

  const latest = state.history.slice(-10).reverse();
  $('#sim-feed').innerHTML = latest.length === 0
    ? '<p class="empty">No episodes yet. Press Start!</p>'
    : `<ul class="feed">${latest.map((e) => `
        <li><span class="mono">${esc(e.time.slice(11))}</span> ·
          <b>${esc(e.patientName)}</b> fell asleep at ${esc(e.location)} → ${destination(e)}
          <span class="mono">(${esc(e.attendedBy)})</span></li>`).join('')}</ul>`;

  const last = state.simulation.last;
  $('#sim-result').innerHTML = last
    ? `<h2>Last simulation result</h2>
       <div class="result-grid">
         <div class="stat"><div class="value">${last.goroutines}</div><div class="label">Goroutines</div></div>
         <div class="stat"><div class="value">${last.episodes}</div><div class="label">Sleep attacks</div></div>
         <div class="stat alert"><div class="value">${last.noRoom}</div><div class="label">No room available</div></div>
         <div class="stat"><div class="value">${last.wakeUps}</div><div class="label">Wake ups</div></div>
         <div class="stat"><div class="value">${last.relocations}</div><div class="label">Hallway → bed</div></div>
       </div>`
    : '<p class="empty">The result appears here when a simulation finishes.</p>';
}

$('#sim-start').addEventListener('click', async () => {
  try {
    notify(await api('POST', '/api/simulation', {
      rounds: Number($('#sim-rounds').value),
      maxDelayMs: Number($('#sim-speed').value),
    }));
  } catch (error) {
    toast(error.message, 'error');
  }
  refresh();
});

// ---------------------------------------------------------------------------
// Actualización periódica (polling): cada 500 ms se pide /api/state.
// ---------------------------------------------------------------------------

let refreshing = false;
let lastRendered = '';

async function refresh() {
  if (activeTab === 'scenario' || refreshing) return;
  refreshing = true; // evita pedir de nuevo si la petición anterior no ha terminado
  try {
    const state = await api('GET', '/api/state');
    // Solo se redibuja si algo cambió. Redibujar sin cambios cada 500 ms
    // podría "comerse" un clic justo cuando se reemplaza un botón.
    const key = JSON.stringify({ ...state, takenAt: '' });
    if (key !== lastRendered) {
      lastRendered = key;
      renderManagement(state);
      renderSimulation(state);
    }
    $('#clock').textContent = `live data · updated ${state.takenAt}`;
  } catch (error) {
    $('#clock').textContent = 'server not reachable';
  } finally {
    refreshing = false;
  }
}

loadScenario();
const initialTab = location.hash.slice(1);
if (['management', 'simulation'].includes(initialTab)) showTab(initialTab);
setInterval(refresh, 500);

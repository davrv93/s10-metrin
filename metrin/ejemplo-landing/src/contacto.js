// Formulario de contacto de la landing de TurnoClaro.
//
// Configuración del formulario:
//   - Envía por fetch un POST JSON a CONFIG.endpoint (/api/contacto, tomado del
//     atributo data-endpoint del <form>), sin recargar la página.
//   - Valida en el navegador antes de enviar (reglas en REGLAS).
//   - Tiempo máximo de espera de la petición: CONFIG.timeoutMs (10 s).
//   - Anti-spam: campo trampa «sitio_web» (debe ir vacío) y un envío como
//     máximo cada CONFIG.esperaEntreEnviosMs (30 s) por navegador.
//   - El destino final de los mensajes (ventas@turnoclaro.example) lo decide
//     el backend; aquí solo se manda el campo «plan» para enrutarlo.

const CONFIG = {
  endpoint: '/api/contacto',
  timeoutMs: 10000,
  esperaEntreEnviosMs: 30000,
  claveUltimoEnvio: 'turnoclaro:ultimo-envio',
};

const REGLAS = {
  nombre: { requerido: true, min: 2, max: 80 },
  email: { requerido: true, patron: /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/ },
  negocio: { requerido: false, max: 120 },
  mensaje: { requerido: true, min: 20, max: 1000 },
  acepto: { casilla: true },
};

const MENSAJES = {
  requerido: 'Este campo es obligatorio.',
  min: (n) => `Escribe al menos ${n} caracteres.`,
  max: (n) => `Máximo ${n} caracteres.`,
  patron: 'Revisa el formato del correo.',
  casilla: 'Debes aceptar la política de privacidad.',
  espera: 'Acabas de enviar un mensaje. Espera unos segundos antes de enviar otro.',
  ok: '¡Gracias! Te responderemos en menos de un día laborable.',
  error: 'No pudimos enviar el mensaje. Inténtalo de nuevo o escríbenos a hola@turnoclaro.example.',
};

function validarCampo(nombre, valor, regla) {
  if (regla.casilla) return valor ? null : MENSAJES.casilla;
  const v = (valor || '').trim();
  if (!v) return regla.requerido ? MENSAJES.requerido : null;
  if (regla.min && v.length < regla.min) return MENSAJES.min(regla.min);
  if (regla.max && v.length > regla.max) return MENSAJES.max(regla.max);
  if (regla.patron && !regla.patron.test(v)) return MENSAJES.patron;
  return null;
}

function validarFormulario(form) {
  const errores = {};
  for (const [nombre, regla] of Object.entries(REGLAS)) {
    const campo = form.elements[nombre];
    const valor = regla.casilla ? campo.checked : campo.value;
    const error = validarCampo(nombre, valor, regla);
    campo.setAttribute('aria-invalid', error ? 'true' : 'false');
    if (error) errores[nombre] = error;
  }
  return errores;
}

function puedeEnviar() {
  try {
    const ultimo = Number(localStorage.getItem(CONFIG.claveUltimoEnvio) || 0);
    return Date.now() - ultimo > CONFIG.esperaEntreEnviosMs;
  } catch {
    return true;
  }
}

async function enviar(form, estado) {
  const datos = Object.fromEntries(new FormData(form));
  if (datos.sitio_web) return; // bot: se descarta en silencio
  delete datos.sitio_web;
  datos.acepto = form.elements.acepto.checked;

  const controlador = new AbortController();
  const temporizador = setTimeout(() => controlador.abort(), CONFIG.timeoutMs);
  try {
    const r = await fetch(form.dataset.endpoint || CONFIG.endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(datos),
      signal: controlador.signal,
    });
    if (!r.ok) throw new Error(`HTTP ${r.status}`);
    try { localStorage.setItem(CONFIG.claveUltimoEnvio, String(Date.now())); } catch {}
    form.reset();
    estado.textContent = MENSAJES.ok;
    estado.dataset.tipo = 'ok';
  } catch {
    estado.textContent = MENSAJES.error;
    estado.dataset.tipo = 'error';
  } finally {
    clearTimeout(temporizador);
  }
}

document.addEventListener('DOMContentLoaded', () => {
  const form = document.getElementById('form-contacto');
  const estado = document.getElementById('form-estado');
  if (!form) return;
  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const errores = validarFormulario(form);
    const primero = Object.keys(errores)[0];
    if (primero) {
      estado.textContent = errores[primero];
      estado.dataset.tipo = 'error';
      form.elements[primero].focus();
      return;
    }
    if (!puedeEnviar()) {
      estado.textContent = MENSAJES.espera;
      estado.dataset.tipo = 'error';
      return;
    }
    const boton = form.querySelector('button[type=submit]');
    boton.disabled = true;
    estado.textContent = 'Enviando…';
    await enviar(form, estado);
    boton.disabled = false;
  });
});

if (typeof module !== 'undefined') module.exports = { validarCampo, REGLAS, CONFIG };

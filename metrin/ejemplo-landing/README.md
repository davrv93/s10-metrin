# Landing de TurnoClaro

Landing estática de **TurnoClaro**, un SaaS ficticio de turnos para cafeterías,
panaderías y restaurantes pequeños. Sirve de repo de ejemplo para el RAG de
`rag-go`.

## Estructura

- `index.html` — hero, funciones, planes (sin precios: los precios viven en la
  hoja de tarifas del bucket de documentación), FAQ y formulario de contacto.
- `styles.css` — estilos; paleta en variables CSS de `:root`.
- `src/contacto.js` — validación y envío del formulario de contacto.
- `package.json` — scripts de desarrollo.

## Formulario de contacto

El formulario `#form-contacto` no recarga la página: `src/contacto.js`
intercepta el `submit`, valida y envía un `POST` JSON a `/api/contacto`
(atributo `data-endpoint`). Campos: `nombre` (2–80), `email` (formato),
`negocio` (opcional, máx. 120), `plan` (básico/pro/empresa, Pro por defecto),
`mensaje` (20–1000) y la casilla `acepto` (obligatoria). Lleva un campo trampa
`sitio_web` contra bots y limita a un envío cada 30 s por navegador.
La petición se aborta a los 10 s.

## Desarrollo

```bash
npm install
npm run dev      # servidor estático en http://localhost:5173
npm run lint
```

-- Esquema del visor de trazas JEV (Fase 3/5).
-- Una traza = una petición a POST /v1/systemone de jeva.cpp:
--   traces: la petición (estado compartido + metadatos)
--   jev:    sus preguntas (choice/score/noul) y respuestas
-- El visor (trazas/trazas.py · RepositorioPostgres) reconstruye
-- el JSON anidado con un JOIN; el API no cambia con respecto
-- al mock. Aplica con:  psql $TRAZAS_PG_DSN -f trazas/esquema.sql

CREATE TABLE IF NOT EXISTS traces (
    id             TEXT PRIMARY KEY,             -- ULID del emisor
    ts             TIMESTAMPTZ NOT NULL DEFAULT now(),
    modelo         TEXT NOT NULL DEFAULT '',
    origen         TEXT NOT NULL DEFAULT '',     -- p. ej. 'metrin-chat', 'reportes'
    plantilla      TEXT,
    ms             INTEGER NOT NULL DEFAULT 0,
    tokens_entrada INTEGER NOT NULL DEFAULT 0,
    estado         JSONB NOT NULL,               -- string | objeto | array
    correcta       BOOLEAN                        -- veredicto de muestra etiquetada; NULL = sin evaluar
);

CREATE TABLE IF NOT EXISTS jev (
    trace_id      TEXT NOT NULL REFERENCES traces (id) ON DELETE CASCADE,
    pregunta_id   TEXT NOT NULL,
    tipo          TEXT NOT NULL CHECK (tipo IN ('choice', 'score', 'noul')),
    instrucciones JSONB,
    criterios     JSONB,
    respuesta     JSONB NOT NULL,                -- choice/score/noul según el tipo
    PRIMARY KEY (trace_id, pregunta_id)
);

CREATE INDEX IF NOT EXISTS traces_ts_idx      ON traces (ts DESC);
CREATE INDEX IF NOT EXISTS traces_modelo_idx  ON traces (modelo);
CREATE INDEX IF NOT EXISTS jev_tipo_idx       ON jev (tipo);

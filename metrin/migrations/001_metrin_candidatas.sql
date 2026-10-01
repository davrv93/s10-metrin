-- Migración 001: Tablas de preguntas candidatas y configuración de rúbricas.
-- Aplica con: psql "$METRIN_PG_DSN" -f migrations/001_metrin_candidatas.sql

CREATE TABLE IF NOT EXISTS metrin_preguntas_candidatas (
    id                  BIGSERIAL PRIMARY KEY,
    nodo_id             TEXT NOT NULL,
    modulo              TEXT NOT NULL,
    seccion             TEXT NOT NULL,
    contenido           TEXT NOT NULL,
    nodo                TEXT NOT NULL,
    url                 TEXT,
    pregunta            TEXT NOT NULL,
    tipo                TEXT,
    respuesta_esperada  TEXT,
    modelo_generador    TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ix_candidatas_nodo
    ON metrin_preguntas_candidatas (nodo_id);
CREATE INDEX IF NOT EXISTS ix_candidatas_modulo_seccion
    ON metrin_preguntas_candidatas (modulo, seccion);

CREATE TABLE IF NOT EXISTS metrin_rubricas_config (
    id              BIGSERIAL PRIMARY KEY,
    nombre          TEXT NOT NULL UNIQUE,
    descripcion     TEXT NOT NULL,
    peso            REAL NOT NULL,
    umbral_minimo   REAL NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO metrin_rubricas_config (nombre, descripcion, peso, umbral_minimo) VALUES
    ('claridad',        'La pregunta es clara, sin ambigüedades, sin dobles negaciones.',               0.25, 0.80),
    ('especificidad',   'Menciona términos concretos del S10. No es genérica.',                           0.20, 0.75),
    ('respondibilidad', 'Es respondible SOLO con el texto del nodo de origen.',                           0.25, 0.85),
    ('utilidad',        'Responde a una necesidad real de un usuario del ERP.',                           0.20, 0.75),
    ('tono',            'Tono profesional, español neutro, sin jergas regionales.',                       0.10, 0.85)
ON CONFLICT (nombre) DO NOTHING;

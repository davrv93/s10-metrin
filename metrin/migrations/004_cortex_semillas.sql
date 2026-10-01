-- Migración 004: Tabla de semillas publicadas en Cortex con VB.
-- Aplica con: psql "$METRIN_PG_DSN" -f migrations/004_cortex_semillas.sql

CREATE TABLE IF NOT EXISTS cortex_semillas (
    id                  BIGSERIAL PRIMARY KEY,
    nodo_id             TEXT NOT NULL REFERENCES cortex_nodos(id),
    pregunta            TEXT NOT NULL,
    respuesta_esperada  TEXT,
    tipo                TEXT,
    origen              TEXT DEFAULT 'generacion_automatica',
    vb                  BOOLEAN DEFAULT FALSE,
    vb_por              TEXT,
    vb_at               TIMESTAMPTZ,
    jev_claridad        REAL,
    jev_dificultad      INT,
    jev_tipo_conf       REAL,
    rubricas            JSONB,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(nodo_id, pregunta)
);

CREATE INDEX IF NOT EXISTS ix_semillas_nodo
    ON cortex_semillas (nodo_id);
CREATE INDEX IF NOT EXISTS ix_semillas_vb
    ON cortex_semillas (vb) WHERE vb = TRUE;

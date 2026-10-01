-- Migración 002: Tabla de preguntas juzgadas con puntajes y rúbricas.
-- Aplica con: psql "$METRIN_PG_DSN" -f migrations/002_metrin_juzgadas.sql

CREATE TABLE IF NOT EXISTS metrin_preguntas_juzgadas (
    id                  BIGSERIAL PRIMARY KEY,
    candidata_id        BIGINT NOT NULL REFERENCES metrin_preguntas_candidatas(id) ON DELETE CASCADE,
    nodo_id             TEXT NOT NULL,
    -- Puntajes individuales por rúbrica
    rubrica_claridad        REAL,
    rubrica_especificidad   REAL,
    rubrica_respondibilidad REAL,
    rubrica_utilidad        REAL,
    rubrica_tono            REAL,
    rubrica_dificultad      INT,
    -- Tipo elegido por el juez
    jev_tipo            TEXT,
    jev_tipo_conf       REAL,
    -- Rúbricas crudas para auditoría
    rubricas_raw        JSONB,
    -- Puntaje global y veredicto
    puntaje_global      REAL,
    veredicto           TEXT NOT NULL DEFAULT 'revision'
                         CHECK (veredicto IN ('aceptada','revision','rechazada')),
    -- Visto bueno
    vb                  BOOLEAN DEFAULT FALSE,
    vb_por              TEXT,
    vb_at               TIMESTAMPTZ,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(candidata_id)
);

CREATE INDEX IF NOT EXISTS ix_juzgadas_nodo
    ON metrin_preguntas_juzgadas (nodo_id);
CREATE INDEX IF NOT EXISTS ix_juzgadas_vb
    ON metrin_preguntas_juzgadas (vb) WHERE vb = TRUE;
CREATE INDEX IF NOT EXISTS ix_juzgadas_veredicto
    ON metrin_preguntas_juzgadas (veredicto);

-- Migración 003: Schema de Cortex (jerarquía Módulo → Sección → Contenido → Nodo).
-- Aplica con: psql "$METRIN_PG_DSN" -f migrations/003_cortex_schema.sql

CREATE TABLE IF NOT EXISTS cortex_modulos (
    id      TEXT PRIMARY KEY,
    nombre  TEXT NOT NULL,
    orden   INT
);

CREATE TABLE IF NOT EXISTS cortex_secciones (
    id          TEXT PRIMARY KEY,
    modulo_id   TEXT NOT NULL REFERENCES cortex_modulos(id),
    nombre      TEXT NOT NULL,
    orden       INT
);

CREATE TABLE IF NOT EXISTS cortex_contenidos (
    id          TEXT PRIMARY KEY,
    seccion_id  TEXT NOT NULL REFERENCES cortex_secciones(id),
    nombre      TEXT NOT NULL,
    orden       INT
);

CREATE TABLE IF NOT EXISTS cortex_nodos (
    id            TEXT PRIMARY KEY,
    contenido_id  TEXT NOT NULL REFERENCES cortex_contenidos(id),
    nombre        TEXT NOT NULL,
    texto         TEXT NOT NULL,
    url           TEXT,
    version       TEXT,
    orden         INT,
    metadata      JSONB
);

CREATE INDEX IF NOT EXISTS ix_cortex_modulo
    ON cortex_secciones (modulo_id);
CREATE INDEX IF NOT EXISTS ix_cortex_seccion
    ON cortex_contenidos (seccion_id);
CREATE INDEX IF NOT EXISTS ix_cortex_contenido
    ON cortex_nodos (contenido_id);

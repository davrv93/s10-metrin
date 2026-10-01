# Estado de Jev

## Servicio activo
- **Host:** `127.0.0.1:8765`
- **Modelo:** `jev-style-0.8b-decision-v3`
- **Backend:** MLX del Mac (launchd `com.jevstyle.serve`)
- **Endpoint:** `POST /v1/systemone` requiere `state`; responde `noul`, `choice`, `score`
- **Pipeline conectada:** `metrin/pipeline/preguntas.py` con `JEV_URL=http://127.0.0.1:8765`

## Configuración en `.env`
```bash
JEV_URL=http://host.docker.internal:8765
JEV_MODEL=jev-style-0.8b-decision-v3
```

## Uso actual
- **Modo:** ya NO es mock
- **Fase en curso:** juzgado real de 2,553 candidatas con Jev (background)
- **Comando:** `JEV_URL=http://127.0.0.1:8765 JEV_MODEL=jev-style-0.8b-decision-v3 .venv/bin/python3 metrin/pipeline/preguntas.py juzgar`
- **Siguiente:** `vb` → `sync_cortex_semillas.py` → `reindex`

## Orquestación
Ver `ORCHESTRATION.md` para el flujo completo del pipeline.

## Nota
- Puerto 8080 sigue ocupado por `mlx_lm.server` (conversacional Metrín)
- Jev y Metrín conversacional coexisten sin conflicto en puertos distintos

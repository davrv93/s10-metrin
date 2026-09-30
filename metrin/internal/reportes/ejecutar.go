package reportes

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Fila es un registro resultado: columnas en orden y valores.
type Fila struct {
	Columnas []string
	Valores  []any
}

// Resultado de una consulta ejecutada.
type Resultado struct {
	Columnas []string
	Filas    []Fila
	MS       int64
	Truncada bool // se alcanzó el tope de filas
}

// prepara convierte :nombrado a $n (PostgreSQL) o ? (SQLite) y devuelve los
// valores en orden.
func prepara(sql string, params map[string]any, postgres bool) (string, []any, error) {
	var valores []any
	var sb strings.Builder
	i := 0
	for {
		j := strings.IndexByte(sql[i:], ':')
		if j < 0 {
			sb.WriteString(sql[i:])
			break
		}
		sb.WriteString(sql[i : i+j])
		k := i + j + 1
		fin := k
		for fin < len(sql) && (sql[fin] == '_' || sql[fin] >= 'a' && sql[fin] <= 'z' ||
			sql[fin] >= 'A' && sql[fin] <= 'Z' || sql[fin] >= '0' && sql[fin] <= '9') {
			fin++
		}
		if fin == k { // no era un parámetro
			sb.WriteByte(':')
			i = k
			continue
		}
		nombre := sql[k:fin]
		v, ok := params[nombre]
		if !ok {
			return "", nil, fmt.Errorf("parámetro sin valor: :%s", nombre)
		}
		if postgres {
			if v == nil {
				sb.WriteString("NULL")
			} else {
				sb.WriteString(fmt.Sprintf("$%d", len(valores)+1))
				valores = append(valores, v)
			}
		} else {
			sb.WriteString("?")
			valores = append(valores, v)
		}
		i = fin
	}
	return sb.String(), valores, nil
}

// Ejecutar corre la consulta validada contra SQLite (ruta) o PostgreSQL (URL).
// La sesión es read-only por conexión; el tope de filas es cortesía de red.
func Ejecutar(ctx context.Context, consultaSQL string, params map[string]any, bd string) (*Resultado, error) {
	cctx, cancel := context.WithTimeout(ctx, TimeoutConsulta*time.Second)
	defer cancel()

	postgres := strings.HasPrefix(bd, "postgresql://") || strings.HasPrefix(bd, "postgres://")
	comp, valores, err := prepara(consultaSQL, params, postgres)
	if err != nil {
		return nil, err
	}

	var conn *sql.DB
	if postgres {
		if !strings.Contains(bd, "sslmode") {
			bd += "sslmode=disable&"
		}
		conn, err = sql.Open("pgx", bd)
		if err == nil {
			conn.SetMaxOpenConns(1)
			if err := conn.PingContext(cctx); err != nil {
				conn.Close()
				return nil, fmt.Errorf("postgres: %w", err)
			}
			// read-only por sesión y timeout por servidor
			if _, err := conn.ExecContext(cctx, "SET default_transaction_read_only = on; SET statement_timeout = " +
				fmt.Sprint(TimeoutConsulta*1000) + ";"); err != nil {
				conn.Close()
				return nil, fmt.Errorf("postgres readonly: %w", err)
			}
		}
	} else {
		conn, err = sql.Open("sqlite", bd+"?mode=ro&_timeusec=0")
		if err == nil {
			conn.SetMaxOpenConns(1)
		}
	}
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	t0 := time.Now()
	rows, err := conn.QueryContext(cctx, comp, valores...)
	if err != nil {
		return nil, fmt.Errorf("consulta: %w", err)
	}
	defer rows.Close()

	res := &Resultado{}
	res.Columnas, err = rows.Columns()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		if len(res.Filas) >= MaxFilas {
			res.Truncada = true
			break
		}
		vals := make([]any, len(res.Columnas))
		ptrs := make([]any, len(res.Columnas))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		limpia := make([]any, len(vals))
		for i, v := range vals {
			switch t := v.(type) {
			case []byte:
				limpia[i] = string(t)
			default:
				limpia[i] = t
			}
		}
		res.Filas = append(res.Filas, Fila{Columnas: res.Columnas, Valores: limpia})
	}
	res.MS = time.Since(t0).Milliseconds()
	return res, rows.Err()
}

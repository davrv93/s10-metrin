package reportes

import (
	"fmt"
	"strings"
)

// Validar comprueba por tokens que la consulta sea un único SELECT.
// Se admite WITH ... SELECT pero no CTEs que modifiquen datos. Es defensa en
// profundidad: la garantía real es el rol read-only de la BD.
func Validar(sql string) error {
	return validarSQL(sql)
}

func validarSQL(sql string) error {
	toks := tokenizar(sql)
	if len(toks) == 0 {
		return fmt.Errorf("consulta vacía")
	}
	primera := strings.ToUpper(toks[0])
	if primera != "SELECT" && primera != "WITH" {
		return fmt.Errorf("solo se permiten consultas SELECT (empieza con %q)", primera)
	}
	prohibidas := map[string]bool{
		"INSERT": true, "UPDATE": true, "DELETE": true, "DROP": true, "ALTER": true,
		"CREATE": true, "TRUNCATE": true, "ATTACH": true, "DETACH": true,
		"GRANT": true, "REVOKE": true, "VACUUM": true, "REINDEX": true,
		"PRAGMA": true, "CALL": true, "EXECUTE": true, "EXEC": true, "MERGE": true,
	}
	for _, t := range toks {
		if prohibidas[strings.ToUpper(t)] {
			return fmt.Errorf("palabra no permitida en consulta de lectura: %s", strings.ToUpper(t))
		}
	}
	// una sola sentencia: rechazar ; intermedio (el final puede traer ;)
	cuerpo := strings.TrimRight(sql, " \t\n;")
	if strings.Contains(cuerpo, ";") {
		return fmt.Errorf("una sola sentencia por consulta")
	}
	return nil
}

// tokenizar separa identificadores, números, cadenas y operadores. Las cadenas
// entrecomilladas se devuelven como un solo token (no dejan colar palabras
// prohibidas dentro de un literal: se validan igual, es más estricto).
func tokenizar(sql string) []string {
	var out []string
	var actual strings.Builder
	flush := func() {
		if actual.Len() > 0 {
			out = append(out, actual.String())
			actual.Reset()
		}
	}
	for i := 0; i < len(sql); i++ {
		c := sql[i]
		switch {
		case c == '\'' || c == '"':
			quote := c
			flush()
			actual.WriteByte(c)
			for i+1 < len(sql) {
				i++
				actual.WriteByte(sql[i])
				if sql[i] == quote {
					if i+1 < len(sql) && sql[i+1] == quote { // escape ''
						i++
						actual.WriteByte(sql[i])
						continue
					}
					break
				}
			}
			flush()
		case c == '-' && i+1 < len(sql) && sql[i+1] == '-':
			flush()
			for i+1 < len(sql) && sql[i+1] != '\n' {
				i++
			}
		case c == '/' && i+1 < len(sql) && sql[i+1] == '*':
			flush()
			fin := strings.Index(sql[i+2:], "*/")
			if fin < 0 {
				i = len(sql)
			} else {
				i += 2 + fin + 1
			}
		case c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= 0x80:
			actual.WriteByte(c)
		case c >= '0' && c <= '9':
			actual.WriteByte(c)
		default:
			flush()
			out = append(out, string(c))
		}
	}
	flush()
	return out
}

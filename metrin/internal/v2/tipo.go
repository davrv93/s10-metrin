package v2

// Tipo de respuesta: CONCEPT, PROCEDURE, NAVIGATION, TROUBLESHOOTING, CONFIGURATION, COMPARISON, UNKNOWN o SOCIAL.
//
// Orden (el código controla; los modelos solo proponen):
//  1. Charla pura (todas las palabras son de cortesía, internal/clasificar.PuntajeSocial) → SOCIAL.
//  2. Reglas léxicas claras (una sola familia, o una que domina a las demás) → confianza 0,90.
//  3. kNN de tipo de respuesta (V2_TIPO_MODELO) si está cargado: decide su top-1 si supera V2_TIPO_UMBRAL y saca
//     al top-2 al menos V2_TIPO_MARGEN (si chocaron reglas, solo si es una de ellas).
//  4. El clasificador de intención de V1 (defecto.json, solo lectura) detecta «social»/«limite» en mensajes cortos.
//  5. Nada claro → UNKNOWN con confianza 0.
//
// Si nada lo deja decidido, el orquestador pregunta al motor de decisión (response_type) con las opciones que
// chocaron (o las mejores del kNN, o todas) más UNKNOWN; si el motor falla, UNKNOWN y se pide aclaración
// (recomendación de kb/catalogos/EVALUACION.md: con el embebedor estático el kNN es una SEÑAL, no un decisor).

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"rag-go/internal/clasificar"
	"rag-go/internal/embed"
	"rag-go/internal/v2/tipos"
)

// Confianzas de las reglas léxicas (no son probabilidades calibradas).
const (
	confReglaClara    = 0.90
	confReglaConflict = 0.45
	confSocialLexica  = 0.95
)

// TiposRespuesta: las siete clases de trabajo, en el orden de prioridad de las reglas (SOCIAL va aparte).
var TiposRespuesta = []tipos.TipoRespuesta{
	tipos.Problema, tipos.Comparacion, tipos.Concepto, tipos.Navegacion, tipos.Configuracion, tipos.Procedimiento,
	tipos.Desconocido,
}

// patronesTipo: frases (sobre el texto normalizado, con espacios a los lados) que delatan cada clase.
var patronesTipo = map[tipos.TipoRespuesta][]string{
	tipos.Problema: {" no puedo ", " no me deja ", " no me sale ", " no sale ", " no se puede ", " no funciona ",
		" no guarda ", " no se guarda ", " no graba ", " no aparece ", " no aparecen ", " no carga ", " no abre ",
		" no permite ", " no cuadra ", " no cuadran ", " no calcula ", " error ", " errores ", " falla ", " fallo ",
		" problema ", " se cierra ", " se cuelga ", " se traba ", " se bloquea ", " bloqueado ", " rechazado ",
		" rechaza ", " me sale un ", " me salio un ", " me aparece un ", " lento "},
	tipos.Comparacion: {" diferencia ", " diferencias ", " comparar ", " compara ", " comparacion ", " versus ",
		" vs ", " frente a ", " cual es mejor ", " cual conviene ", " es lo mismo ", " son lo mismo ", " o mejor "},
	tipos.Concepto: {" que es ", " que son ", " q es ", " que significa ", " q significa ", " que significan ",
		" que quiere decir ", " para que sirve ", " para que sirven ", " en que consiste ", " definicion ",
		" define ", " concepto de ", " a que se refiere ", " a que se refieren "},
	tipos.Navegacion: {" donde ", " dnd ", " en que menu ", " en que parte ", " en que pantalla ", " en que opcion ",
		" en que pestana ", " en que pestaña ", " en que modulo ", " ubico ", " ubicar ", " ubicacion ", " encuentro ",
		" encontrar ", " como llego ", " como accedo ", " como entro ", " acceder a ", " ruta del menu ",
		" no encuentro ", " no ubico "},
	tipos.Configuracion: {" configur", " parametr", " ajustes ", " personaliz", " preferencias ", " por defecto ",
		" predeterminad", " habilit", " deshabilit", " activar ", " desactivar ", " instal", " actualizar s10 ",
		" tipo de cambio ", " permisos ", " usuarios y permisos "},
	tipos.Procedimiento: {" como ", " cmo ", " pasos ", " paso a paso ", " procedimiento ", " registr", " crear ",
		" creo ", " generar ", " genero ", " agregar ", " agrego ", " anadir ", " añadir ", " elaborar ", " hacer ",
		" hago ", " ingresar ", " ingreso ", " importar ", " exportar ", " calcular ", " calculo ", " imprimir ",
		" eliminar ", " modificar ", " editar ", " emitir ", " anular ", " copiar ", " aprobar ", " ensename ",
		" enseñame ", " quiero ", " necesito ", " ayudame a ", " tutorial ", " guia ", " guiame "},
}

// ReglasTipo aplica las reglas léxicas al texto normalizado. Devuelve la clase elegida, las que coincidieron (en
// orden de prioridad) y la confianza: 0,90 si una domina, 0,45 si chocan, 0 si ninguna.
//
// Dominancias (no son conflicto): PROCEDURE cede ante CONFIGURATION, NAVIGATION y TROUBLESHOOTING (son formas de
// procedimiento: «¿cómo configuro…?»); TROUBLESHOOTING domina a CONFIGURATION y NAVIGATION («error al configurar»,
// «no encuentro el botón»); COMPARISON domina a CONCEPT.
func ReglasTipo(norm string) (tipos.TipoRespuesta, []tipos.TipoRespuesta, float64) {
	t := " " + norm + " "
	hay := map[tipos.TipoRespuesta]bool{}
	for tipo, frases := range patronesTipo {
		for _, f := range frases {
			if strings.Contains(t, f) {
				hay[tipo] = true
				break
			}
		}
	}
	if hay[tipos.Problema] {
		delete(hay, tipos.Configuracion)
		delete(hay, tipos.Navegacion)
		delete(hay, tipos.Procedimiento)
	}
	if hay[tipos.Configuracion] || hay[tipos.Navegacion] {
		delete(hay, tipos.Procedimiento)
	}
	if hay[tipos.Comparacion] {
		delete(hay, tipos.Concepto)
	}
	var coinciden []tipos.TipoRespuesta
	for _, tipo := range TiposRespuesta {
		if hay[tipo] {
			coinciden = append(coinciden, tipo)
		}
	}
	switch len(coinciden) {
	case 0:
		return "", nil, 0
	case 1:
		return coinciden[0], coinciden, confReglaClara
	}
	return coinciden[0], coinciden, confReglaConflict
}

// Clasificador del tipo de respuesta. Todos sus campos son opcionales: sin ninguno quedan las reglas léxicas.
type Clasificador struct {
	Emb       embed.Embebedor    // embebedor del índice; nil = sin kNN
	Intencion *clasificar.Modelo // modelo de intención de V1 (solo lectura): social / limite
	Tipo      *clasificar.Modelo // modelo kNN de tipo de respuesta (V2_TIPO_MODELO); nil = sin kNN de tipo
	Umbral    float64            // V2_TIPO_UMBRAL: similitud mínima del top-1 del kNN (provisional)
	Margen    float64            // V2_TIPO_MARGEN: ventaja mínima del top-1 sobre el top-2 (provisional)
}

// ResultadoTipo es lo que propone el clasificador (todavía sin motor de decisión).
type ResultadoTipo struct {
	Tipo        tipos.TipoRespuesta
	Confianza   float64
	Metodo      string                // reglas | knn | intencion_v1 | ninguno
	Decidido    bool                  // una regla clara, el kNN con margen o la intención de V1 lo dejaron decidido
	Candidatos  []tipos.TipoRespuesta // opciones para el motor de decisión si no quedó decidido
	IntencionV1 string                // «social» / «limite» / … si se consultó el modelo de V1
	SimilitudV1 float64
	TipoKNN     tipos.TipoRespuesta // top-1 del kNN de tipo (señal, aunque no decida)
	SimKNN      float64
	MargenKNN   float64
}

// Dudoso: nada lo dejó decidido y decide el motor de decisión.
func (r ResultadoTipo) Dudoso() bool { return !r.Decidido }

// Clasificar propone el tipo de respuesta de una pregunta.
func (c *Clasificador) Clasificar(ctx context.Context, pregunta string) ResultadoTipo {
	// Charla pura primero: «¿cómo estás?» lleva «cómo» pero todas sus palabras son de cortesía.
	if clasificar.PuntajeSocial(pregunta) == 1 {
		return ResultadoTipo{Tipo: tipos.Social, Confianza: confSocialLexica, Metodo: "reglas", Decidido: true,
			IntencionV1: clasificar.Social}
	}
	elegido, coinciden, conf := ReglasTipo(Normalizar(pregunta))
	if conf >= confReglaClara {
		return ResultadoTipo{Tipo: elegido, Confianza: conf, Metodo: "reglas", Decidido: true, Candidatos: coinciden}
	}
	res := ResultadoTipo{Tipo: elegido, Confianza: conf, Metodo: "reglas", Candidatos: coinciden}
	if elegido == "" {
		res = ResultadoTipo{Tipo: tipos.Desconocido, Metodo: "ninguno"}
	}
	if c == nil || c.Emb == nil {
		return res
	}
	umbral, margen := c.Umbral, c.Margen
	if umbral <= 0 {
		umbral = UmbralTipoDefecto
	}
	if margen <= 0 {
		margen = MargenTipoDefecto
	}
	// kNN de tipo: decide con top-1 por encima del umbral y con margen sobre el top-2 (si chocaron reglas, solo
	// entre las que chocaron). Si no, queda como señal: sus mejores clases son las opciones del motor.
	if c.Tipo != nil {
		if ins, sims, err := c.Tipo.Puntajes(ctx, c.Emb, pregunta); err == nil && len(ins) > 0 && !math.IsNaN(sims[0]) {
			knn := tipos.TipoRespuesta(ins[0].Nombre)
			segunda := 0.0
			if len(sims) > 1 {
				segunda = sims[1]
			}
			res.TipoKNN, res.SimKNN, res.MargenKNN = knn, sims[0], sims[0]-segunda
			if valido(knn) && sims[0] >= umbral && res.MargenKNN >= margen && (elegido == "" || contieneTipo(coinciden, knn)) {
				res.Tipo, res.Confianza, res.Metodo, res.Decidido = knn, sims[0], "knn", true
				return res
			}
			if elegido == "" {
				for i := 0; i < len(ins) && i < 3; i++ {
					if t := tipos.TipoRespuesta(ins[i].Nombre); valido(t) && t != tipos.Desconocido {
						res.Candidatos = append(res.Candidatos, t)
					}
				}
			}
		}
	}
	// Intención de V1: solo charla y fuera de alcance en mensajes cortos sin señal de trabajo. El modelo de V1 ya
	// aplica su umbral calibrado (0,40): si dice «social», queda decidido.
	if elegido == "" && c.Intencion != nil {
		if in, sim, err := c.Intencion.Clasificar(ctx, c.Emb, pregunta); err == nil {
			res.IntencionV1 = in
			if !math.IsInf(sim, 0) && !math.IsNaN(sim) {
				res.SimilitudV1 = sim
			}
			if (in == clasificar.Social || in == "limite") && !preguntaLarga(pregunta) {
				res.Tipo, res.Confianza, res.Metodo, res.Decidido = tipos.Social, res.SimilitudV1, "intencion_v1", true
				res.Candidatos = nil
			}
		}
	}
	return res
}

func valido(t tipos.TipoRespuesta) bool {
	return t == tipos.Social || contieneTipo(TiposRespuesta, t)
}

func contieneTipo(ts []tipos.TipoRespuesta, t tipos.TipoRespuesta) bool {
	for _, x := range ts {
		if x == t {
			return true
		}
	}
	return false
}

// preguntaLarga: «?» con más de 3 palabras (misma excepción de seguridad que V1).
func preguntaLarga(s string) bool {
	return strings.Contains(s, "?") && len(strings.Fields(s)) > 3
}

// CargarModeloTipo carga el modelo kNN de tipo de respuesta:
//   - «.json»: un modelo de internal/clasificar (clasificar.Cargar), con las clases como intenciones;
//   - «.yml»/«.yaml»: el catálogo kb/catalogos/tipo_respuesta.yml; se entrena al arrancar con emb (centroides por
//     clase, como el candidato modelos/clasificador-tipo-respuesta-candidato.json; tarda milisegundos).
//
// La huella del modelo debe coincidir con el embebedor (si no, se rechaza: los vectores no serían comparables).
func CargarModeloTipo(ctx context.Context, ruta string, emb embed.Embebedor, umbral float64) (*clasificar.Modelo, error) {
	if emb == nil {
		return nil, fmt.Errorf("modelo de tipo: sin embebedor")
	}
	switch strings.ToLower(filepath.Ext(ruta)) {
	case ".json":
		m, err := clasificar.Cargar(ruta)
		if err != nil {
			return nil, err
		}
		if m.HuellaEmb != "" && m.HuellaEmb != emb.Nombre() {
			return nil, fmt.Errorf("modelo de tipo %s: huella %s ≠ embebedor %s", ruta, m.HuellaEmb, emb.Nombre())
		}
		return m, nil
	case ".yml", ".yaml":
		ejemplos, err := EjemplosCatalogoTipo(ruta)
		if err != nil {
			return nil, err
		}
		m, err := clasificar.Entrenar(ctx, emb, emb.Nombre(), umbral, ejemplos)
		if err != nil {
			return nil, err
		}
		// Modo centroide (el que eligió el leave-one-out de kb/catalogos/EVALUACION.md): sin ejemplos sueltos.
		m.Ejemplos = nil
		return m, nil
	}
	return nil, fmt.Errorf("modelo de tipo %s: extensión no admitida (.json o .yml)", ruta)
}

// EjemplosCatalogoTipo lee solo «clases → <CLASE> → ejemplos» del catálogo de tipo de respuesta. Es un lector de
// líneas a propósito (el catálogo usa bloques «>-» que el YAML mínimo del repo no admite y aquí no hacen falta).
// Solo acepta como clase las de tipos.TipoRespuesta.
func EjemplosCatalogoTipo(ruta string) (map[string][]string, error) {
	f, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string][]string{}
	enClases, clase, enEjemplos := false, "", false
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	for sc.Scan() {
		linea := sc.Text()
		recortada := strings.TrimSpace(linea)
		if recortada == "" || strings.HasPrefix(recortada, "#") {
			continue
		}
		sangria := len(linea) - len(strings.TrimLeft(linea, " "))
		switch {
		case sangria == 0:
			enClases = recortada == "clases:"
			clase, enEjemplos = "", false
		case !enClases:
		case sangria == 2 && strings.HasSuffix(recortada, ":"):
			nombre := strings.TrimSuffix(recortada, ":")
			clase, enEjemplos = "", false
			if valido(tipos.TipoRespuesta(nombre)) {
				clase = nombre
			}
		case clase != "" && sangria == 4:
			enEjemplos = recortada == "ejemplos:"
		case clase != "" && enEjemplos && strings.HasPrefix(recortada, "- "):
			if t := escalarYAML(strings.TrimSpace(recortada[2:])); t != "" {
				out[clase] = append(out[clase], t)
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("catálogo de tipo %s: sin ejemplos", ruta)
	}
	return out, nil
}

// escalarYAML quita las comillas de un escalar de una línea y un comentario final fuera de comillas.
func escalarYAML(s string) string {
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') {
		q := s[0]
		for i := 1; i < len(s); i++ {
			if s[i] == '\\' && q == '"' {
				i++
				continue
			}
			if s[i] == q {
				v := s[1:i]
				if q == '"' {
					v = strings.NewReplacer(`\"`, `"`, `\\`, `\`).Replace(v)
				} else {
					v = strings.ReplaceAll(v, "''", "'")
				}
				return strings.TrimSpace(v)
			}
		}
		return strings.TrimSpace(s[1:])
	}
	if i := strings.Index(s, " #"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

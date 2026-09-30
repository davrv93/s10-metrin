package embed

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"math"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Copiado y adaptado de pjgfarma-analista/analista/interno/preguntas/modelo.go
// (no se importa el paquete interno). Formato PJGE v1 (little endian). Lo escribe
// analista/herramientas/destilar/destilar.py y lo lee Cargar:
//
//	"PJGE"                      magia
//	u32 versión                 = 1
//	u32 dim                     dimensión del embedding
//	u32 n                       número de piezas (= filas de la matriz)
//	u32 unk                     índice de la pieza desconocida ([UNK])
//	u8  cuantización            1 = int8 con escala por fila, 2 = float16
//	u8  banderas                bit0: aislar la puntuación ASCII con espacios
//	u16 reservado
//	u32 nNorm                   entradas de la tabla de normalización
//	nNorm × { u32 runa, u8 largo, largo bytes UTF-8 }   runa → reemplazo
//	n × { u8 largo, largo bytes UTF-8, f32 puntaje }    piezas Unigram
//	matriz int8:   n·dim int8, luego n f32 (escalas)
//	matriz float16: n·dim u16
//	u32 CRC-32 IEEE de todo lo anterior
//
// Tokenización (la del tokenizador de Hugging Face del modelo, recortado):
// normalización por tabla (la «Precompiled» de SentencePiece, runa a runa),
// aislamiento opcional de la puntuación ASCII, partición por espacios, «▁»
// delante de cada palabra y Viterbi Unigram sobre las piezas. Las piezas
// desconocidas se descartan y el embedding es la media de las filas,
// normalizada a norma 1 (como StaticModel.encode de model2vec).

const (
	cuantInt8 = 1
	cuantF16  = 2

	banderaPuntuacion = 1

	// penalización de HF tokenizers (Unigram::K_UNK_PENALTY)
	penalizacionUnk = 10.0
	// model2vec corta en 512 tokens
	maxTokens = 512
)

// Modelo es un modelo de embeddings estáticos cargado en memoria. Es de
// solo lectura: se puede usar desde varias goroutines a la vez.
type Modelo struct {
	dim        int
	piezas     map[string]int32
	puntajes   []float32
	unk        int32
	puntajeUnk float64
	maxBytes   int
	norm       map[rune]string
	banderas   uint8
	cuant      uint8
	q          []int8
	escalas    []float32
	h          []uint16
	huella     string
}

// Huella es el SHA-256 del fichero del modelo, en hexadecimal: la misma que
// escribe destilar.py en el .json de al lado. Entra en analista.modelo.version
// (CONTRATO §12): si cambia, los embeddings guardados se recalculan.
func (m *Modelo) Huella() string { return m.huella }

// Dim es la dimensión de los embeddings.
func (m *Modelo) Dim() int { return m.dim }

// Piezas es el tamaño del vocabulario.
func (m *Modelo) Piezas() int { return len(m.puntajes) }

// CargarArchivo abre y carga un fichero .pjge.
func CargarArchivo(ruta string) (*Modelo, error) {
	f, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Cargar(bufio.NewReaderSize(f, 1<<16))
}

type lectorCRC struct {
	r   io.Reader
	crc uint32
	sha hash.Hash
}

func (l *lectorCRC) Read(p []byte) (int, error) {
	n, err := l.r.Read(p)
	l.crc = crc32.Update(l.crc, crc32.IEEETable, p[:n])
	l.sha.Write(p[:n])
	return n, err
}

// Cargar lee un modelo PJGE y comprueba su CRC.
func Cargar(r io.Reader) (*Modelo, error) {
	lc := &lectorCRC{r: r, sha: sha256.New()}
	var cab struct {
		Magia                [4]byte
		Version, Dim, N, Unk uint32
		Cuant, Banderas      uint8
		Reservado            uint16
		NNorm                uint32
	}
	if err := binary.Read(lc, binary.LittleEndian, &cab); err != nil {
		return nil, fmt.Errorf("cabecera: %w", err)
	}
	if string(cab.Magia[:]) != "PJGE" || cab.Version != 1 {
		return nil, errors.New("no es un modelo PJGE v1")
	}
	if cab.Dim == 0 || cab.Dim > 4096 || cab.N == 0 || cab.N > 4_000_000 || cab.Unk >= cab.N {
		return nil, errors.New("cabecera PJGE incoherente")
	}
	m := &Modelo{
		dim: int(cab.Dim), unk: int32(cab.Unk), banderas: cab.Banderas, cuant: cab.Cuant,
		norm:     make(map[rune]string, cab.NNorm),
		piezas:   make(map[string]int32, cab.N),
		puntajes: make([]float32, cab.N),
	}
	var buf [256]byte
	for i := uint32(0); i < cab.NNorm; i++ {
		var ent struct {
			Runa  uint32
			Largo uint8
		}
		if err := binary.Read(lc, binary.LittleEndian, &ent); err != nil {
			return nil, fmt.Errorf("tabla de normalización: %w", err)
		}
		if _, err := io.ReadFull(lc, buf[:ent.Largo]); err != nil {
			return nil, err
		}
		m.norm[rune(ent.Runa)] = string(buf[:ent.Largo])
	}
	minPuntaje := math.Inf(1)
	for i := uint32(0); i < cab.N; i++ {
		if _, err := io.ReadFull(lc, buf[:1]); err != nil {
			return nil, fmt.Errorf("piezas: %w", err)
		}
		l := int(buf[0])
		if _, err := io.ReadFull(lc, buf[:l+4]); err != nil {
			return nil, err
		}
		p := string(buf[:l])
		s := math.Float32frombits(binary.LittleEndian.Uint32(buf[l : l+4]))
		m.piezas[p] = int32(i)
		m.puntajes[i] = s
		minPuntaje = math.Min(minPuntaje, float64(s))
		m.maxBytes = max(m.maxBytes, l)
	}
	m.puntajeUnk = minPuntaje - penalizacionUnk
	total := int(cab.N) * m.dim
	switch cab.Cuant {
	case cuantInt8:
		m.q = make([]int8, total)
		if err := binary.Read(lc, binary.LittleEndian, m.q); err != nil {
			return nil, fmt.Errorf("matriz: %w", err)
		}
		m.escalas = make([]float32, cab.N)
		if err := binary.Read(lc, binary.LittleEndian, m.escalas); err != nil {
			return nil, fmt.Errorf("escalas: %w", err)
		}
	case cuantF16:
		m.h = make([]uint16, total)
		if err := binary.Read(lc, binary.LittleEndian, m.h); err != nil {
			return nil, fmt.Errorf("matriz: %w", err)
		}
	default:
		return nil, fmt.Errorf("cuantización %d desconocida", cab.Cuant)
	}
	calculado := lc.crc
	var guardado uint32
	if err := binary.Read(r, binary.LittleEndian, &guardado); err != nil {
		return nil, fmt.Errorf("crc: %w", err)
	}
	if calculado != guardado {
		return nil, errors.New("CRC del modelo no coincide: fichero dañado")
	}
	binary.Write(lc.sha, binary.LittleEndian, guardado)
	m.huella = hex.EncodeToString(lc.sha.Sum(nil))
	return m, nil
}

// esPuntuacionASCII: los mismos 32 signos que el normalizador del modelo
// rodea de espacios.
func esPuntuacionASCII(r rune) bool {
	return r < 128 && strings.ContainsRune("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", r)
}

// normalizar aplica la tabla del modelo y, si toca, aísla la puntuación
// (también la que sale de la tabla: «！» pasa a «!» y luego se aísla).
func (m *Modelo) normalizar(texto string) string {
	var b strings.Builder
	b.Grow(len(texto) + 8)
	escribir := func(r rune) {
		if m.banderas&banderaPuntuacion != 0 && esPuntuacionASCII(r) {
			b.WriteByte(' ')
			b.WriteRune(r)
			b.WriteByte(' ')
			return
		}
		b.WriteRune(r)
	}
	for _, r := range texto {
		if s, ok := m.norm[r]; ok {
			for _, r2 := range s {
				escribir(r2)
			}
			continue
		}
		escribir(r)
	}
	return b.String()
}

// Tokenizar devuelve los ids de las piezas del texto, [UNK] incluido, igual
// que el tokenizador de Hugging Face con add_special_tokens=False.
func (m *Modelo) Tokenizar(texto string) []int32 {
	var ids []int32
	for _, palabra := range strings.FieldsFunc(m.normalizar(texto), unicode.IsSpace) {
		ids = m.viterbi("▁"+palabra, ids)
	}
	return ids
}

type nodoViterbi struct {
	puntaje float64
	desde   int
	id      int32
	listo   bool
}

// viterbi segmenta s con el algoritmo de Unigram::encode_optimized de HF:
// recorre las posiciones de izquierda a derecha, prueba las piezas de largo
// creciente y solo reemplaza un camino si el nuevo es estrictamente mejor.
// Si ninguna pieza cubre el carácter, avanza un carácter como [UNK]. Los
// [UNK] consecutivos se funden en uno.
func (m *Modelo) viterbi(s string, ids []int32) []int32 {
	n := len(s)
	nodos := make([]nodoViterbi, n+1)
	nodos[0].listo = true
	for pos := 0; pos < n; {
		if !nodos[pos].listo {
			_, t := utf8.DecodeRuneInString(s[pos:])
			pos += t
			continue
		}
		base := nodos[pos].puntaje
		_, largoCar := utf8.DecodeRuneInString(s[pos:])
		unCaracter := false
		for fin := pos + 1; fin <= n && fin-pos <= m.maxBytes; fin++ {
			if fin < n && !utf8.RuneStart(s[fin]) {
				continue
			}
			id, ok := m.piezas[s[pos:fin]]
			if !ok {
				continue
			}
			c := base + float64(m.puntajes[id])
			if nd := &nodos[fin]; !nd.listo || c > nd.puntaje {
				*nd = nodoViterbi{puntaje: c, desde: pos, id: id, listo: true}
			}
			if fin-pos == largoCar {
				unCaracter = true
			}
		}
		if !unCaracter {
			fin := pos + largoCar
			c := base + m.puntajeUnk
			if nd := &nodos[fin]; !nd.listo || c > nd.puntaje {
				*nd = nodoViterbi{puntaje: c, desde: pos, id: m.unk, listo: true}
			}
		}
		pos += largoCar
	}
	// reconstruir hacia atrás
	var rev []int32
	for fin := n; fin > 0; fin = nodos[fin].desde {
		id := nodos[fin].id
		if id == m.unk && len(rev) > 0 && rev[len(rev)-1] == m.unk {
			continue
		}
		rev = append(rev, id)
	}
	for i := len(rev) - 1; i >= 0; i-- {
		ids = append(ids, rev[i])
	}
	return ids
}

// fila suma la fila id (decuantizada) en acc.
func (m *Modelo) sumarFila(acc []float64, id int32) {
	ini := int(id) * m.dim
	if m.cuant == cuantInt8 {
		e := float64(m.escalas[id])
		for j, v := range m.q[ini : ini+m.dim] {
			acc[j] += float64(float32(v) * float32(e))
		}
		return
	}
	for j, v := range m.h[ini : ini+m.dim] {
		acc[j] += float64(f16a32(v))
	}
}

// Embeber devuelve el embedding del texto (norma 1), o un vector de ceros si
// ninguna pieza es conocida.
func (m *Modelo) Embeber(texto string) []float32 {
	acc := make([]float64, m.dim)
	n := 0
	for _, id := range m.Tokenizar(texto) {
		if id == m.unk {
			continue
		}
		if n == maxTokens {
			break
		}
		m.sumarFila(acc, id)
		n++
	}
	out := make([]float32, m.dim)
	if n == 0 {
		return out
	}
	var norma float64
	for j := range acc {
		acc[j] /= float64(n)
		norma += acc[j] * acc[j]
	}
	norma = math.Sqrt(norma) + 1e-32
	for j := range acc {
		out[j] = float32(acc[j] / norma)
	}
	return out
}

// f16a32 convierte un float16 IEEE 754 a float32.
func f16a32(h uint16) float32 {
	signo := float32(1)
	if h&0x8000 != 0 {
		signo = -1
	}
	exp := int(h>>10) & 0x1f
	man := float32(h & 0x3ff)
	switch exp {
	case 0: // cero o subnormal
		return signo * man * float32(math.Pow(2, -24))
	case 0x1f:
		if man == 0 {
			return signo * float32(math.Inf(1))
		}
		return float32(math.NaN())
	}
	return signo * (1 + man/1024) * float32(math.Pow(2, float64(exp-15)))
}

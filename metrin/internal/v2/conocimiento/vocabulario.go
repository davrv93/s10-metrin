package conocimiento

import "unicode"

// vocabulario: las palabras de los manuales (fragmentos) y de los YAML, por su forma fonética. Sirve para
// corregir faltas de ortografía de la consulta ANTES de buscar, con dos reglas generales:
//
//  1. Una palabra que no está en el vocabulario pero suena igual que una que sí («nuebo» → «nuevo»,
//     «kiero» → «quiero», «aser» → «hacer») se cambia por esa; si la corregida es una palabra vacía, se
//     quita.
//  2. Si no, y tiene 6 letras o más, se cambia por la palabra del vocabulario a una edición de distancia
//     (fonética) más frecuente, si aparece al menos 3 veces («prosupuesto» → «presupuesto»).
//
// Una palabra desconocida sin corrección se queda tal cual: sigue pesando (con IDF máximo) en la
// cobertura, que es lo que hace que «criptomonedas» o «huella de carbono» no encuentren un procedimiento.
type vocabulario struct {
	cuenta   map[string]int          // palabra → veces (mientras se arma; cerrar() lo vacía)
	fon      map[string]*entradaVoc  // fonética → palabra más vista con esa fonética
	porLetra map[claveLetra][]string // fonéticas vistas ≥ 3 veces, por primera letra y largo
}

type claveLetra struct {
	letra byte
	largo int
}

type entradaVoc struct {
	palabra string
	n       int // veces de todas las grafías con esta fonética
	mejor   int // veces de la grafía elegida (la más vista)
}

func nuevoVocabulario() *vocabulario {
	v := &vocabulario{cuenta: map[string]int{}, fon: map[string]*entradaVoc{}, porLetra: map[claveLetra][]string{}}
	for w := range vacias {
		v.fon[fonetica(w)] = &entradaVoc{palabra: w, n: 1 << 20}
	}
	return v
}

func soloLetras(w string) bool {
	for _, r := range w {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func (v *vocabulario) agregar(texto string) { v.agregarPalabras(palabras(texto)) }

func (v *vocabulario) agregarPalabras(ws []string) {
	for _, w := range ws {
		if len(w) >= 3 && len(w) <= 24 {
			v.cuenta[w]++
		}
	}
}

// cerrar arma el índice por letra. Después, el vocabulario es de solo lectura.
func (v *vocabulario) cerrar() {
	for w, n := range v.cuenta {
		if !soloLetras(w) {
			continue
		}
		f := fonetica(w)
		e, ok := v.fon[f]
		if !ok {
			v.fon[f] = &entradaVoc{palabra: w, n: n, mejor: n}
			continue
		}
		if e.n >= 1<<20 { // una palabra vacía con esa fonética manda
			continue
		}
		if n > e.mejor || (n == e.mejor && w < e.palabra) {
			e.palabra, e.mejor = w, n
		}
		e.n += n
	}
	v.cuenta = nil
	for f, e := range v.fon {
		if len(f) >= 5 && e.n >= 3 {
			k := claveLetra{f[0], len(f)}
			v.porLetra[k] = append(v.porLetra[k], f)
		}
	}
}

// corregir devuelve la palabra corregida (o la misma) y si quedó vacía.
func (v *vocabulario) corregir(w string) (string, bool) {
	if v == nil || vacias[w] {
		return w, vacias[w]
	}
	f := fonetica(w)
	if e, ok := v.fon[f]; ok {
		return e.palabra, vacias[e.palabra]
	}
	if len(w) < 6 || !soloLetras(w) {
		return w, false
	}
	var mejor *entradaVoc
	for l := len(f) - 1; l <= len(f)+1; l++ {
		for _, g := range v.porLetra[claveLetra{f[0], l}] {
			if e := v.fon[g]; distancia1(f, g) && (mejor == nil || e.n > mejor.n || (e.n == mejor.n && e.palabra < mejor.palabra)) {
				mejor = e
			}
		}
	}
	if mejor == nil {
		return w, false
	}
	return mejor.palabra, vacias[mejor.palabra]
}

// terminosCorregidos: como terminos(), con cada palabra corregida.
func (v *vocabulario) terminosCorregidos(s string) []string {
	var out []string
	for _, w := range palabras(s) {
		if len(w) < 2 || vacias[w] {
			continue
		}
		c, vacia := v.corregir(w)
		if vacia {
			continue
		}
		out = append(out, raiz(c))
	}
	return out
}

package clasificar

import (
	"regexp"
	"strings"
)

// Social es la intención de charla (saludos, cortesías, respuestas cortas).
const Social = "social"

// maxPalabrasSocial: más largo que esto ya no es charla suelta.
const maxPalabrasSocial = 8

// vocabSocial: palabras (normalizadas, sin tildes) que por sí solas no piden
// nada de obra ni de S10. Un mensaje es charla solo si TODAS sus palabras
// están aquí: «bien y tu», «ok gracias», «hola metrin». Basta una palabra de
// trabajo («hola, dime los metrados») para que decida el embedding/RAG.
// Los embeddings estáticos dan ~0.3 a mensajes de 2-3 palabras aunque sean
// saludos obvios; esta capa no necesita reentrenar ni el modelo .pjge.
var vocabSocial = conjunto(`
hola holi hey ey oye alo buenas buenos buen dia dias tardes noches saludos saludo
que tal como estas esta estan estamos andas anda va vas te tu ti usted ustedes vos
y e yo me mi nos igual igualmente tambien bien muy mal regular mas menos todo toda
tranquilo tranquila genial excelente perfecto perfecta super chevere bacan bravo buenisimo
ok okay oki okey vale dale listo lista entendido entendida entiendo claro si sip no nop
gracias muchas muchisimas mil agradezco agradecido agradecida amable
gusto mucho un una el la lo los las de del al a en con para por pues bueno ya ah oh uy
wow asi placer encantado encantada alegro adios chau chao hasta luego pronto mañana manana vemos
cuidate cuidese haces hace cuentas bendiciones feliz contento contenta cansado cansada triste aburrido aburrida
sueño sueno estoy ando aqui ahi sigues eres quien bot robot amigo amiga crack metrin
tengo jaja jeje jiji lol xd
`)

var (
	reNoLetra = regexp.MustCompile(`[^a-zñ]+`)
	reRisa    = regexp.MustCompile(`^(a|e|i)?((j|h)(a|e|i|o)){2,}(j|h)?$`)
)

func conjunto(s string) map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(s) {
		m[w] = true
	}
	return m
}

// colapsar quita letras repetidas seguidas: «holaaa» → «hola», «okk» → «ok».
func colapsar(w string) string {
	var b strings.Builder
	var prev rune
	for _, r := range w {
		if r != prev {
			b.WriteRune(r)
		}
		prev = r
	}
	return b.String()
}

func esSocial(w string) bool {
	if vocabSocial[w] || reRisa.MatchString(w) {
		return true
	}
	c := colapsar(w)
	return vocabSocial[c] || reRisa.MatchString(c)
}

// PuntajeSocial es la fracción de palabras del mensaje que son de charla
// (0 si está vacío o es largo). 1 significa charla pura.
func PuntajeSocial(texto string) float64 {
	palabras := strings.Fields(reNoLetra.ReplaceAllString(normalizarTexto(texto), " "))
	if len(palabras) == 0 || len(palabras) > maxPalabrasSocial {
		return 0
	}
	n := 0
	for _, w := range palabras {
		if esSocial(w) {
			n++
		}
	}
	return float64(n) / float64(len(palabras))
}

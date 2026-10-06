// Package config lee la configuración del entorno y, si existe, de un .env
// (KEY=VALOR, sin exportar ni comillas obligatorias). El entorno manda sobre
// el .env.
package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"rag-go/internal/traza"
	"rag-go/internal/v2"
)

type Config struct {
	DirDatos string // índice chromem + estado + sin_respuesta.jsonl

	EmbedProvider  string // estatico | ollama
	ModeloEstatico string
	OllamaURL      string
	EmbedModelo    string // modelo de embeddings en Ollama
	LLMModelo      string
	LLMProvider    string // ollama | mlx (mlx_lm.server con adaptador LoRA)
	MLXURL         string
	Temperatura    float64
	TimeoutSeg     int
	LLMHilos       int     // Ollama num_thread (0 = lo decide Ollama)
	LLMMaxTokens   int     // tope de tokens por respuesta (0 = sin tope)
	MaxDistancia   float64 // distancia coseno (1-sim) por encima de la cual no hay contexto

	S3Endpoint, S3Region, S3AccessKey, S3SecretKey, S3Bucket string

	// Decisiones JEV (jeva.cpp, POST /v1/systemone). Vacío =
	// sin decisiones JEV: el flujo actual no cambia.
	JEVURL    string
	JEVModelo string
	JEVTrazas string // JSONL de trazas para el visor (vacío = no emitir)
	JEVOrigen string // «origen» de la traza (p. ej. "reportes")

	Puerto string

	// Modo traza del chat (METRIN_TRAZA=1|true). Apagado por defecto: /ask
	// responde igual aunque la petición traiga "traza": true.
	Traza bool

	// V2 de Metrín (AGENT_VERSION, AGENT_V2_ENABLED, AGENT_V2_PERCENTAGE,
	// límites, DECISION_*, V2_TIPO_*): apagada por defecto. Ver internal/v2.
	V2 v2.Config
}

// CargarEnv mete en el entorno las claves del fichero que no estén ya
// definidas. Si el fichero no existe no hace nada.
func CargarEnv(ruta string) error {
	f, err := os.Open(ruta)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		k, v, ok := strings.Cut(l, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(strings.TrimPrefix(k, "export "))
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		if _, ya := os.LookupEnv(k); !ya {
			os.Setenv(k, v)
		}
	}
	return sc.Err()
}

func Leer() Config {
	return Config{
		DirDatos:       def("RAG_DATOS", "./datos"),
		EmbedProvider:  def("EMBED_PROVIDER", "estatico"),
		ModeloEstatico: def("EMBED_MODELO_ESTATICO", "./modelos/potion-es-int8.pjge"),
		OllamaURL:      def("OLLAMA_URL", "http://localhost:11434"),
		EmbedModelo:    def("OLLAMA_EMBED_MODEL", "nomic-embed-text"),
		LLMModelo:      def("OLLAMA_MODEL", "qwen2.5-coder:7b"),
		LLMProvider:    def("LLM_PROVIDER", "ollama"),
		MLXURL:         def("MLX_URL", "http://localhost:8080"),
		Temperatura:    num("RAG_TEMPERATURA", 0.2),
		TimeoutSeg:     int(num("RAG_TIMEOUT", 300)),
		LLMHilos:       int(num("OLLAMA_NUM_THREAD", 0)),
		LLMMaxTokens:   int(num("LLM_MAX_TOKENS", 0)),
		MaxDistancia:   num("RAG_MAX_DISTANCIA", 0.80),
		S3Endpoint:     def("S3_ENDPOINT", "http://localhost:4790"),
		S3Region:       def("S3_REGION", "garage"),
		S3AccessKey:    os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:    os.Getenv("S3_SECRET_KEY"),
		S3Bucket:       def("S3_BUCKET", "rag-demo"),
		JEVURL:         def("JEV_URL", ""),
		JEVModelo:      def("JEV_MODEL", ""),
		JEVTrazas:      def("JEV_TRAZAS", ""),
		JEVOrigen:      def("JEV_ORIGEN", "metrin"),
		Puerto:         def("RAG_PUERTO", "4760"),
		Traza:          traza.Habilitada(os.Getenv("METRIN_TRAZA")),
		V2:             v2.LeerConfig(),
	}
}

func def(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func num(k string, d float64) float64 {
	if v, err := strconv.ParseFloat(os.Getenv(k), 64); err == nil {
		return v
	}
	return d
}

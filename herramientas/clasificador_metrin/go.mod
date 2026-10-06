// Arnés de evaluación del clasificador de Metrín. Es un módulo aparte para no
// tocar metrin/: importa rag-go/internal/clasificar y rag-go/internal/embed
// (solo biblioteca estándar) a través de go.work. La ruta «rag-go/...» es la
// que permite importar los paquetes internal/ de rag-go.
module rag-go/herramientas/clasificadormetrin

go 1.27.1

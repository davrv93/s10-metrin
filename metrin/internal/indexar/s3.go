package indexar

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"rag-go/internal/almacen"
	"rag-go/internal/trocear"
)

// ConfigS3 describe un endpoint compatible con S3 (Garage).
type ConfigS3 struct {
	Endpoint, Region, AccessKey, SecretKey, Bucket string
}

// NuevoClienteS3 crea un cliente con endpoint propio y path-style (Garage no
// sirve virtual-host sin configurar un dominio raíz).
func NuevoClienteS3(c ConfigS3) (*s3.Client, error) {
	if c.AccessKey == "" || c.SecretKey == "" {
		return nil, errors.New("faltan S3_ACCESS_KEY / S3_SECRET_KEY (ver .env.example)")
	}
	return s3.New(s3.Options{
		BaseEndpoint: aws.String(c.Endpoint),
		Region:       c.Region,
		UsePathStyle: true,
		Credentials:  credentials.NewStaticCredentialsProvider(c.AccessKey, c.SecretKey, ""),
		// Garage no entiende los checksums CRC nuevos del SDK en todas las
		// versiones: solo cuando la operación lo exige.
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	}), nil
}

// S3 es un bucket (o un prefijo de él).
type S3 struct {
	Cliente *s3.Client
	Bucket  string
	Prefijo string
	// MaxBytes limita el tamaño de lo que se descarga (0 = 5 MiB).
	MaxBytes int64
}

func (o S3) Fuente() string { return almacen.FuenteS3 }

func (o S3) Cita(d Documento, n int) string {
	return "s3:" + o.Bucket + "/" + d.Ruta + "#" + itoa(n)
}

// Abarca dice si una clave cae dentro del prefijo listado: lo de fuera no se
// borra aunque no aparezca en el listado.
func (o S3) Abarca(clave string) bool { return strings.HasPrefix(clave, "s3:"+o.Prefijo) }

// Subir sube un objeto (para sembrar el bucket de ejemplo).
func (o S3) Subir(ctx context.Context, key string, cuerpo []byte, tipo string) error {
	_, err := o.Cliente.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(o.Bucket), Key: aws.String(key),
		Body: bytes.NewReader(cuerpo), ContentType: aws.String(tipo),
	})
	return err
}

// Listar pagina ListObjectsV2; la versión es el ETag (sin comillas), así que
// no se descarga nada que no haya cambiado.
func (o S3) Listar(ctx context.Context) ([]Documento, error) {
	max := o.MaxBytes
	if max == 0 {
		max = 5 << 20
	}
	var out []Documento
	p := s3.NewListObjectsV2Paginator(o.Cliente, &s3.ListObjectsV2Input{
		Bucket: aws.String(o.Bucket),
		Prefix: aws.String(o.Prefijo),
	})
	for p.HasMorePages() {
		pag, err := p.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listar s3://%s/%s: %w", o.Bucket, o.Prefijo, err)
		}
		for _, obj := range pag.Contents {
			key := aws.ToString(obj.Key)
			if strings.HasSuffix(key, "/") || !trocear.EsTexto(key) || aws.ToInt64(obj.Size) > max {
				continue
			}
			out = append(out, Documento{
				Clave:   "s3:" + key,
				Ruta:    key,
				Version: strings.Trim(aws.ToString(obj.ETag), `"`),
				Cargar:  func(ctx context.Context) (string, error) { return o.leer(ctx, key, max) },
			})
		}
	}
	return out, nil
}

func (o S3) leer(ctx context.Context, key string, max int64) (string, error) {
	r, err := o.Cliente.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(o.Bucket), Key: aws.String(key)})
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	b, err := io.ReadAll(io.LimitReader(r.Body, max))
	return string(b), err
}

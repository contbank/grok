package grok

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// NewRedisConnection ...
//
// Aceita dois formatos em connectionString, sem exigir mudança de código em quem chama:
//   - "host:porta" (formato histórico, sem autenticação) — comportamento idêntico ao de antes.
//   - "redis://[:senha@]host:porta[/db]" ou "rediss://..." (URL completa, com suporte a senha, DB e TLS) —
//     necessário para Redis self-hosted com autenticação habilitada.
//
// A detecção é automática: tenta interpretar como URL primeiro; se não for uma URL válida (ex.: o formato
// histórico "host:porta", que não tem esquema), cai no comportamento antigo. Nenhum chamador precisa mudar.
func NewRedisConnection(connectionString string) *redis.Client {
	options, err := redis.ParseURL(connectionString)
	if err != nil {
		options = &redis.Options{
			Addr: connectionString,
		}
	}

	client := redis.NewClient(options)

	_, err = client.Ping(context.Background()).Result()

	if err != nil {
		logrus.WithError(err).Panic("Error pinging Redis")
	}

	return client
}

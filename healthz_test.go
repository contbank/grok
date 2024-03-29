package grok_test

import (
	"testing"

	"github.com/contbank/grok"
	"github.com/stretchr/testify/assert"
)

func TestHealthz(t *testing.T) {
	settings := &grok.Settings{}
	grok.FromYAML("tests/config.yaml", settings)

	t.Run("Mongo Success", func(t *testing.T) {
		healthz := grok.NewHealthz(
			grok.WithMongo(),
			grok.WithRedis(),
			grok.WithHealthzSettings(settings))

		err := healthz.Healthz()

		assert.NoError(t, err)
	})
}

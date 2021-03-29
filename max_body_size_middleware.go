package grok

import (
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func SetMaxBodyBytesMiddleware(maxMegaBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		maxMBytes := maxMegaBytes
		if maxMBytes == 0 {
			maxMBytes = 1
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxMegaBytes<<20)

		_, err := ioutil.ReadAll(c.Request.Body)

		if err != nil {
			logrus.WithField("error", err).Error("error on max body middleware")
			entityTooLarge := NewError(http.StatusRequestEntityTooLarge, "payload too large")
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, entityTooLarge)
			return
		}
	}
}

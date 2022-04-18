package grok

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

//LogMiddleware ...
func LogMiddleware(restricteds []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer recovery(c)
		defer c.Request.Body.Close()

		requestID := uuid.New()

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		blw.Header().Set("Request-Id", requestID.String())
		c.Writer = blw

		c.Set("Request-Id", requestID.String())

		now := time.Now()
		req := request(c, restricteds)

		c.Next()

		elapsed := time.Since(now)
		fields := make(map[string]interface{})

		fields["request"] = req
		fields["claims"] = c.Keys
		fields["errors"] = c.Errors
		fields["ip"] = c.ClientIP()
		fields["latency"] = elapsed.Seconds()
		fields["request_id"] = requestID.String()
		fields["response"] = response(blw, restricteds)

		logrus.WithFields(fields).Infof(
			"Request incoming from %s elapsed %s completed with %d",
			c.ClientIP(),
			elapsed.String(),
			c.Writer.Status(),
		)
	}
}

func request(context *gin.Context, restricteds []string) interface{} {
	r := make(map[string]interface{})

	bodyCopy := new(bytes.Buffer)
	io.Copy(bodyCopy, context.Request.Body)
	bodyData := bodyCopy.Bytes()

	var body map[string]interface{}
	json.Unmarshal(bodyData, &body)

	r["body"] = restricted(body, restricteds)
	r["host"] = context.Request.Host
	r["form"] = context.Request.Form
	r["path"] = context.Request.URL.Path
	r["full_path"] = context.FullPath()
	r["method"] = context.Request.Method
	r["headers"] = restricted(context.Request.Header, restricteds)
	r["url"] = context.Request.URL.String()
	r["post_form"] = context.Request.PostForm
	r["remote_addr"] = context.Request.RemoteAddr
	r["query_string"] = context.Request.URL.Query()

	context.Request.Body = ioutil.NopCloser(bytes.NewReader(bodyData))

	return r
}

func response(writer *bodyLogWriter, restricteds []string) interface{} {
	r := make(map[string]interface{})

	var body map[string]interface{}
	json.Unmarshal(writer.body.Bytes(), &body)

	r["body"] = restricted(body, restricteds)
	r["status"] = writer.Status()
	r["headers"] = writer.Header()

	return r
}

func recovery(c *gin.Context) {
	if err := recover(); err != nil {
		logrus.WithField("error", err).
			WithField("stack", string(debug.Stack())).
			Error("Error on logging middleware")
		internalServerError := NewError(http.StatusInternalServerError, "internal server error")
		c.AbortWithStatusJSON(http.StatusInternalServerError, internalServerError)
	}
}

func restricted(v interface{}, restricteds []string) interface{} {
	str := marshal(v)

	for _, restricted := range restricteds {
		result := gjson.Get(str, restricted)

		if result.Index <= 0 {
			continue
		}

		str, _ = sjson.Set(str, restricted, "RESTRICTED")
	}
	return unmarshal(str)
}

func marshal(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func unmarshal(str string) interface{} {
	v := make(map[string]interface{})

	json.Unmarshal([]byte(str), &v)

	return v
}

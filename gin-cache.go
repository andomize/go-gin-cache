package gogincache

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/andomize/go-gin-cache/redis"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GinCache struct {
	triggers map[string][]Trigger
	pool     redis.Pool
}

func New(pool redis.Pool) *GinCache {
	return &GinCache{
		triggers: make(map[string][]Trigger),
		pool:     pool,
	}
}

func (gc *GinCache) CacheRouter() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// Просматриваем все зарегистрированные тригеры и если
		// какой-либо из триггеров совпадает по условиям, то
		// очищаем кеш для данного метода.

		for handlerId, triggers := range gc.triggers {
			for _, trigger := range triggers {
				if trigger != nil && trigger.Comparable(ctx.Request) {

					// Если запрос попадает под условие зарегистрированного триггера,
					// то используем идентификатор обработчика, что бы сбросить кеш.
					conn := gc.pool.Get(ctx)

					if err := conn.Del(genKey(handlerId, "*")); err != nil {
						log.Printf("[gin-cache] error occured on conn.Del: %v", err)
						return
					}
				}
			}
		}
	}
}

func (gc *GinCache) Cache(td time.Duration, triggers ...Trigger) gin.HandlerFunc {

	// У каждого обработчика будет свой уникальный идентификатор.
	// Идентификатор генерируется единожды при регистрации обработчика.
	handlerId := uuid.New().String()

	return func(ctx *gin.Context) {
		uri := ctx.Request.RequestURI
		method := ctx.Request.Method

		// Регистрация триггеров. При вызове стандартной 'CacheRouter' функции,
		// будет выполнена проверка на соответствие всем зарегистрированным
		// триггерам и очищены данные в кеше.
		if gc.triggers[handlerId] == nil {
			gc.triggers[handlerId] = make([]Trigger, len(triggers))
			gc.triggers[handlerId] = append(gc.triggers[handlerId], triggers...)
		}

		if method != http.MethodGet {
			log.Printf("[gin-cache] The cache only supports the GET method")
			return
		}

		conn := gc.pool.Get(ctx)
		content, ok, err := conn.Get(genKey(handlerId, uri))
		if err != nil {
			log.Printf("[gin-cache] error occured on conn.Get: %v", err)
			return
		}
		if !ok {

			blw := &bodyLogWriter{
				body:           bytes.NewBufferString(""),
				ResponseWriter: ctx.Writer,
			}
			ctx.Writer = blw
			ctx.Next()

			c := redis.Content{
				StatusCode:  ctx.Writer.Status(),
				ContentType: ctx.Writer.Header().Get("Content-Type"),
				Data:        blw.body.Bytes(),
			}
			err = conn.Set(genKey(handlerId, uri), &c, td)
			if err != nil {
				log.Printf("[gin-cache] error occured on conn.Set: %v", err)
				return
			}
			return
		}

		ctx.Data(content.StatusCode, content.ContentType, content.Data)
		ctx.Abort()
	}
}

func genKey(handlerId, uri string) string {
	return handlerId + ":" + uri
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

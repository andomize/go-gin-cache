package gogincache

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/andomize/go-gin-cache/clients"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GinCache struct {
	triggersStats   map[string][]*TriggerStats
	triggersStatsMx sync.Mutex
	pool            clients.Pool
}

type TriggerStats struct {
	trigger Trigger
	urls    map[string]struct{}
}

func New(pool clients.Pool) *GinCache {
	return &GinCache{
		triggersStats: make(map[string][]*TriggerStats, 0),
		pool:          pool,
	}
}

func (gc *GinCache) CacheRouter() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		gc.triggersStatsMx.Lock()
		defer gc.triggersStatsMx.Unlock()

		// Просматриваем все зарегистрированные тригеры и если
		// какой-либо из триггеров совпадает по условиям, то
		// очищаем все записи в кеше для данного триггера.

		for _, triggerStats := range gc.triggersStats {
			for _, triggerStat := range triggerStats {
				if triggerStat.trigger.Comparable(ctx.Request) {
					// url := ctx.Request.URL.Path
					// if _, has := triggerStat.urls[url]; !has {
					// 	continue
					// }

					for url := range triggerStat.urls {
						go func(key string) {
							// Если запрос попадает под условие зарегистрированного триггера,
							// то используем идентификатор обработчика, что бы сбросить кеш.
							conn := gc.pool.Get(context.Background())

							if err := conn.Del(key); err != nil {
								log.Printf("[gin-cache] error occured on conn.Del: %v", err)
								return
							}
						}(url)
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
		method := ctx.Request.Method
		uri := ctx.Request.URL.RequestURI()
		url := ctx.Request.URL.Path

		// Регистрация триггеров. При вызове стандартной 'CacheRouter' функции,
		// будет выполнена проверка на соответствие всем зарегистрированным
		// триггерам и очищены данные в кеше.
		gc.triggerStatsInit(handlerId, triggers...)
		gc.triggerStatsSetURL(handlerId, url)

		if method != http.MethodGet {
			log.Printf("[gin-cache] The cache only supports the GET method")
			return
		}

		conn := gc.pool.Get(ctx)
		content, ok, err := conn.Get(uri)
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

			rc := &clients.Content{
				StatusCode:  ctx.Writer.Status(),
				ContentType: ctx.Writer.Header().Get("Content-Type"),
				Data:        blw.body.Bytes(),
			}
			go func(rc *clients.Content) {
				conn := gc.pool.Get(context.Background())
				err = conn.Set(uri, rc, td)
				if err != nil {
					log.Printf("[gin-cache] error occured on conn.Set: %v", err)
					return
				}
			}(rc)

			return
		}

		ctx.Data(content.StatusCode, content.ContentType, content.Data)
		ctx.Abort()
	}
}

func (gc *GinCache) triggerStatsInit(handlerId string, triggers ...Trigger) {
	gc.triggersStatsMx.Lock()
	defer gc.triggersStatsMx.Unlock()

	if _, has := gc.triggersStats[handlerId]; has {
		return
	}

	gc.triggersStats[handlerId] = make([]*TriggerStats, 0)
	for _, trigger := range triggers {
		if trigger == nil {
			continue
		}
		gc.triggersStats[handlerId] = append(
			gc.triggersStats[handlerId], &TriggerStats{
				trigger: trigger,
				urls:    make(map[string]struct{}),
			})
	}
}

func (gc *GinCache) triggerStatsSetURL(handlerId string, url string) {
	gc.triggersStatsMx.Lock()
	defer gc.triggersStatsMx.Unlock()

	if _, has := gc.triggersStats[handlerId]; !has {
		return
	}

	for _, triggerStat := range gc.triggersStats[handlerId] {
		triggerStat.urls[url] = struct{}{}
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

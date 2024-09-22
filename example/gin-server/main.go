package main

import (
	"context"
	"log"
	"net/http"
	"time"

	gogincache "github.com/andomize/go-gin-cache"
	"github.com/andomize/go-gin-cache/clients/goredis/v9"
	"github.com/gin-gonic/gin"
	lib "github.com/redis/go-redis/v9"
)

var (
	// Время кеширования - 3 минуты
	cacheTime = 3 * time.Minute
)

func main() {

	// Create a pool with go-redis which is the pool will use
	// while communicating with Redis. This can also be any pool
	// that implements the `redis.Pool` interface.
	client := lib.NewUniversalClient(&lib.UniversalOptions{
		Addrs: []string{"localhost:6379"},
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("redis ping error: %v\n", err)
	}

	router := gin.New()

	// Инициализируем библиотеку и передаём клиент
	// для взаимодействия с сервисом кеширования.
	cache := gogincache.New(goredis.NewPool(client))

	// Устанавливаем глобальный Middleware <CacheRouter>.
	// Он проверяет все запросы на соответствие триггерам,
	// если запрос совместим с каким-либо триггерром, то
	// выполняется сброс кеша.
	router.Use(cache.CacheRouter())

	// Получение списка пользователей с использованием кеша.
	// Для сброса кеша используется триггер, срабатывающий
	// при вызове методов создания или изменения пользователя.
	router.GET("/users", cache.Cache(cacheTime, &gogincache.TriggerURI{
		Methods: gogincache.DefaultUpdateMethods,
		URI:     ".*/users.*",
	}), handlerUserList())

	// Получение одиночного пользователя с использованием кеша.
	// Для сброса кеша используется триггер, срабатывающий
	// при вызове методов создания или изменения пользователя.
	router.GET("/users/:USER", cache.Cache(cacheTime, &gogincache.TriggerURI{
		Methods: gogincache.DefaultUpdateMethods,
		URI:     ".*/users.*",
	}), handlerUserGet())

	// При вызове этих методов кеш списка пользователей будет сброшен.
	router.POST("/users", handlerUserCreate())
	router.PATCH("/users/:USER", handlerUserPatch())

	go router.Run() // listen and serve on 0.0.0.0:8080
	TesterRun()
	time.Sleep(1 * time.Hour)
}

func handlerUserList() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		time.Sleep(5 * time.Second) // slow method
		ctx.JSON(http.StatusOK, []gin.H{
			{"username": "a"},
			{"username": "b"},
		})
	}
}

func handlerUserGet() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		time.Sleep(5 * time.Second) // slow method
		ctx.JSON(http.StatusOK, gin.H{
			"username": ctx.Param("USER"),
		})
	}
}

func handlerUserCreate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusCreated, gin.H{
			"status": "ok",
		})
	}
}

func handlerUserPatch() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusCreated, gin.H{
			"status": "ok",
		})
	}
}

package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func TesterRun() {
	go func() {
		time.Sleep(5 * time.Second)

		// t := time.Now()
		go TesterGet(100)
		// fmt.Println(time.Since(t).Milliseconds())
		go TesterPost(100)

		// t = time.Now()
		// TesterGet(5)
		// fmt.Println(time.Since(t).Milliseconds())
	}()
}

func TesterGet(count int) {
	wg := sync.WaitGroup{}
	wg.Add(count)

	for i := 0; i < count; i++ {

		// random sleep from 1 ms to 5 sec
		time.Sleep(time.Duration(1000+rand.Intn(5000)) * time.Millisecond)

		go func(i int) {
			url := fmt.Sprintf("http://localhost:8080/users?search=%d", i)
			fmt.Printf("[%d] GET: %s\n", i, url)

			req, _ := http.NewRequest("GET", url, nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			// b, _ := io.ReadAll(resp.Body)
			// fmt.Printf("[%d] resp: %s\n", n, string(b))
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TesterPost(count int) {
	wg := sync.WaitGroup{}
	wg.Add(count)

	for i := 0; i < count; i++ {

		// random sleep from 1 ms to 5 sec
		time.Sleep(time.Duration(1000+rand.Intn(5000)) * time.Millisecond)

		go func(i int) {
			url := "http://localhost:8080/users"
			fmt.Printf("[%d] POST: %s\n", i, url)

			req, _ := http.NewRequest("POST", url, nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			// b, _ := io.ReadAll(resp.Body)
			// fmt.Printf("[%d] resp: %s\n", n, string(b))
			wg.Done()
		}(i)
	}

	wg.Wait()
}

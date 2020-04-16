package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

func main() {
	wg := &sync.WaitGroup{}
	worker := NewWorker(10)
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go worker.Start(ctx, wg)

	api := NewApi(worker)

	router := fasthttprouter.New()
	router.POST("/messages", api.HandleCreateMessage)

	server := fasthttp.Server{
		Handler: router.Handler,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		fmt.Println("server is running on port 8080")
		log.Println(server.ListenAndServe("localhost:8080"))
	}()

	<-stop
	_ = server.Shutdown()
	cancel()
	wg.Wait()
	log.Println("Finished")
}

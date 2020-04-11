package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/valyala/fasthttp"
)

func Post(requestURI string, payload []byte) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(requestURI)
	req.Header.SetContentType("application/json; charset=utf-8")
	req.Header.SetMethod(fasthttp.MethodPost)
	req.SetBody(payload)

	var timeOut = 3 * time.Second
	var err = fasthttp.DoTimeout(req, resp, timeOut)
	if err != nil {
		return nil, err
	}

	var out = fasthttp.AcquireResponse()
	resp.CopyTo(out)

	return out, nil
}

func sendMessage(i int64) int64 {
	clientID := rand.Intn(10) + 1
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	body := fmt.Sprintf(`{"text":"hello world","content_id":%v,"client_id":%v,"timestamp":%v}`, i, clientID, timestamp)
	_, err := Post("http://localhost:8080/messages", []byte(body))
	if err != nil {
		fmt.Println(err)
		return -1
	}

	return i
}

func worker(jobs <-chan int64, results chan<- int64) {
	for i := range jobs {
		results <- sendMessage(i)
	}
}

func main() {
	const requestsCount int64 = 1 * 1000 * 1000

	start := time.Now()

	jobs := make(chan int64, requestsCount)
	results := make(chan int64, requestsCount)

	maxWorkers := runtime.NumCPU() * 2
	for i := 0; i < maxWorkers; i++ {
		go worker(jobs, results)
	}

	for i := int64(1); i <= requestsCount; i++ {
		jobs <- i
	}
	close(jobs)

	for i := int64(1); i <= requestsCount; i++ {
		<-results
	}

	duration := time.Now().Sub(start)
	fmt.Printf(`%v sec`, duration.Seconds())
}

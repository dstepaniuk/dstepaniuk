package main

import (
	"encoding/json"
	"log"

	"github.com/valyala/fasthttp"
)

type Api struct {
	worker *Worker
}

func NewApi(w *Worker) *Api {
	return &Api{worker: w}
}

func (api *Api) HandleCreateMessage(ctx *fasthttp.RequestCtx) {
	var m Message
	body := ctx.PostBody()
	if err := json.Unmarshal(body, &m); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		log.Println(err)
		return
	}
	m.Payload = make([]byte, len(body))
	copy(m.Payload, body)

	api.worker.Messages <- m

	ctx.SetStatusCode(fasthttp.StatusNoContent)
}

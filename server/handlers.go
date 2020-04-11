package main

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
)

func HandleCreateMessage(ctx *fasthttp.RequestCtx) {
	payload := ctx.PostBody()

	message := Message{ Payload: payload }

	err := json.Unmarshal(payload, &message)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
	}

	CreateMessage(message)
	ctx.SetStatusCode(fasthttp.StatusNoContent)
}

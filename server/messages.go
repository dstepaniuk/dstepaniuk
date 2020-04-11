package main

import (
	"fmt"
	"sync"
)

var messages []Message
var mu *sync.RWMutex

func init() {
	mu = &sync.RWMutex{}
}

func CreateMessage(message Message) {
	mu.Lock()
	messages = append(messages, message)
	mu.Unlock()
	fmt.Println(string(message.Payload))
}

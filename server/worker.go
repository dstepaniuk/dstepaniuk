package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Message struct {
	ContentID int64  `json:"content_id"`
	ClientID  int    `json:"client_id"`
	Timestamp int64  `json:"timestamp"`
	Payload   []byte
}

func (m Message) Key() string {
	date := time.Unix(m.Timestamp / 1000, 0).Format("2006-01-02")
	return fmt.Sprintf(`/chat/%v/content_logs_%v_%v`, date, date, m.ClientID)
}

type Worker struct {
	Messages chan Message

	uploaders map[string]*S3Uploader
}

func NewWorker(buffSize int) *Worker {
	w := &Worker{
		Messages:  make(chan Message, buffSize),
		uploaders: make(map[string]*S3Uploader),
	}

	return w
}

func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case m := <-w.Messages:
			if err := w.processMessage(ctx, wg, m); err != nil {
				log.Printf("Could not process message: %s", err)
			}
		}
	}
}

func (w *Worker) processMessage(ctx context.Context, wg *sync.WaitGroup, m Message) error {
	key := m.Key()

	if _, ok := w.uploaders[key]; !ok {
		up, err := NewS3Uploader(key)
		if err != nil {
			return err
		}
		w.uploaders[key] = up
		wg.Add(1)
		go up.Start(ctx, wg)
	}

	w.uploaders[key].Enqueue(m)

	return nil
}

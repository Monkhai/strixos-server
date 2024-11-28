package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Monkhai/strixos-server.git/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	var wg sync.WaitGroup
	var s = server.NewServer(&ctx, &wg)

	wg.Add(1)
	go s.QueueLoop(ctx, &wg)
	http.HandleFunc("/ws", s.WebSocketHandler)

	go func() {
		fmt.Println("Server started on :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	<-signalChan

	fmt.Println("\nShutting down gracefully...")
	cancel()

	fmt.Println("\nCancel Called")
	wg.Wait()
	fmt.Println("\nAll waitgroups are done")
}

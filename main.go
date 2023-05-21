package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	// Sets us some randomization between runs.
	// rand.Seed(time.Now().UnixNano())

	// Create a new server listening on port 80. This will listen on all available IP addresses.
	serv, err := newServer(80)
	if err != nil {
		panic(err)
	}
	// Start our server. This blocks, so we have it do it in its own goroutine.
	go serv.start()

	// Sleep long enough for the server to start.
	time.Sleep(500 * time.Millisecond)

	// Create a client that is pointed at our localhost address on port 80.
	client, err := New("http://127.0.0.1:80")
	if err != nil {
		panic(err)
	}

	// We are going to fetch several responses concurrently and put them in this channel.
	results := make(chan string, 2)

	ctx := context.Background()
	wg := sync.WaitGroup{}

	// Get a quote from Mark Twain. He has the best quotes.
	wg.Add(1)
	go func() {
		defer wg.Done()
		quote, err := client.Get(ctx, "Mark Twain")
		if err != nil {
			panic(err)
		}
		results <- quote
	}()

	// Get a random quote from another person.
	wg.Add(1)
	go func() {
		defer wg.Done()
		quote, err := client.Get(ctx, "")
		if err != nil {
			panic(err)
		}
		results <- quote
	}()

	// When we have finished getting quotes, close our results channel.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Read the returned quotes until the results channel is closed.
	for result := range results {
		fmt.Println(result)
	}
}

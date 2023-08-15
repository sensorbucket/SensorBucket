package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	WORKER_COUNT = 5
	WORKER_DELAY = 10 * time.Millisecond
	START_IX     = 0
)

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func Run() error {
	if len(os.Args) < 3 {
		return errors.New("./test http://localhost:3000/process/abcdef ./data.json\ndata.json should be an array")
	}
	if len(os.Args) > 3 {
		ix, _ := strconv.Atoi(os.Args[3])
		START_IX = ix
	}
	url := os.Args[1]
	jsonFile, err := os.Open(os.Args[2])
	if err != nil {
		fmt.Println(err)
	}
	var data []json.RawMessage
	if err := json.NewDecoder(jsonFile).Decode(&data); err != nil {
		return errors.New("Failed to read input file")
	}
	jsonFile.Close()

	wg := sync.WaitGroup{}
	fanout := make(chan json.RawMessage, WORKER_COUNT)

	// Start workers
	for i := 0; i < WORKER_COUNT; i++ {
		id := i
		wg.Add(1)
		go func() {
			fmt.Printf("Worker(%d): Online\n", id)
			client := &http.Client{}
			for msg := range fanout {
				req, _ := http.NewRequest("POST", url, bytes.NewBuffer(msg))
				req.Header.Add("Content-Type", "application/json")
				req.Header.Add("Connection", "close")
				r, err := client.Do(req)
				if err != nil {
					fmt.Printf("Worker(%d): error: %v\n", id, err)
				} else if r.StatusCode < 200 || r.StatusCode > 299 {
					fmt.Printf("Worker(%d): sus status: %v\n", id, r.StatusCode)
				}
				time.Sleep(WORKER_DELAY)
			}
			fmt.Printf("Worker(%d): Finished\n", id)
			wg.Done()
		}()
	}

	// Start producer
	for i := START_IX; i < len(data); i++ {
		fanout <- data[i]
		if i%25 == 0 {
			fmt.Printf("Producer: %d of %d\n", i, len(data))
		}
	}

	return nil
}

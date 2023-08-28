package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func Run() error {
	if len(os.Args) < 4 {
		return errors.New("./ttn-simulator http-ingress pipeline-id data-set.json")
	}

	url := os.Args[1]
	pipelineId := os.Args[2]
	dataSet := os.Args[3]

	jsonFile, err := os.Open(dataSet)
	if err != nil {
		return err
	}
	log.Println("caching data")
	var data []json.RawMessage
	if err := json.NewDecoder(jsonFile).Decode(&data); err != nil {
		return errors.New("failed to read input file")
	}
	jsonFile.Close()
	log.Println("done caching")

	for {
		randomIndex := rand.Intn(len(data))
		el := data[randomIndex]
		fmt.Println("sending", string(el))
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", url, pipelineId), bytes.NewBuffer(el))
		if err != nil {
			fmt.Println("Error creating request:", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			continue
		}
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error sending request:", err)
			continue
		}
		fmt.Println("Response Status:", resp.Status)
		fmt.Println("Response Body:", string(respBody))
		time.Sleep(time.Millisecond * 33)
	}
}

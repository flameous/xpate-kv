package main

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/flameous/xpate-kv/kv/service"
)

// just for testing
func main() {

	time.Sleep(time.Second)

	cli, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
		return
	}

	i := service.InputAction{
		Action: "DELETE",
		Key:    "test_key",
		Value:  "test_value",
	}

	b, _ := json.Marshal(i)

	_, err = cli.Write(b)
	if err != nil {
		log.Fatal(err)
		return
	}
}

package service

import (
	"encoding/json"
	"log"
	"net"
	"time"
)

// just for testing
func DoClientStuff() {

	time.Sleep(time.Second)

	cli, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
		return
	}

	i := inputAction{
		Action: putAction,
		Key:    "foo",
		Value:  "bar",
	}

	b, _ := json.Marshal(i)

	_, err = cli.Write(b)
	if err != nil {
		log.Fatal(err)
		return
	}
}

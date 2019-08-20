package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/flameous/xpate-kv/kv"
	"io"
	"log"
	"net"
)

const (
	putAction    = "PUT"
	readAction   = "READ"
	deleteAction = "DELETE"
)

func NewListener(cache kv.Cacher) *Service {
	return &Service{
		cache: cache,
	}
}

type Service struct {
	cache kv.Cacher
}

type inputAction struct {
	Action string `json:"action"`
	Key    string `json:"key"`

	Value string `json:"value,omitempty"`
	TTL   *int64 `json:"ttl,omitempty"`
}

func (s *Service) Start(port string) error {
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	for {
		// accept connection
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go s.work(conn)
	}
}

func (s *Service) work(conn net.Conn) {
	defer conn.Close()
	log.Println("work")

	// read incoming data
	buf := bytes.Buffer{}
	_, err := io.Copy(&buf, conn)
	if err != nil {
		log.Printf("failed to copy data from connection: %v\n", err)
		return
	}

	// deserialize it to inner struct
	var ia inputAction
	b := buf.Bytes()
	err = json.Unmarshal(b, &ia)
	if err != nil {
		log.Printf("failed to unmarshal input action data. err: %v, raw data: %s\n", err, b)
		return
	}

	switch ia.Action {
	case readAction:
		value, ok := s.cache.Read(ia.Key)
		var msg string
		if ok {
			msg = value
		} else {
			msg = "ERROR: NOT_FOUND"
		}
		_, err = conn.Write([]byte(msg))

	case putAction:
		s.cache.Set(ia.Key, ia.Value, ia.TTL)
		_, err = conn.Write([]byte("ok"))

	case deleteAction:
		s.cache.Delete(ia.Key)
		_, err = conn.Write([]byte("ok"))
	default:
		err = errors.New("unexpected action: " + ia.Action)
	}
	if err != nil {
		log.Printf("action error: %v\n", err)
	}
	log.Println(s.cache)
}

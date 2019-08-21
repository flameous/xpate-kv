package kv

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

const defaultTTL = 60 * 1000 * 1000 * 1000 // 60 seconds

type value struct {
	Val         string `json:"val"`
	CreatedTime int64  `json:"created_time"`
	TTL         int64  `json:"ttl"`
}

type Cacher interface {
	Set(key, value string, ttl *int64)
	Read(key string) (string, bool)
	Delete(key string)
}

func NewCacher() Cacher {
	var (
		m   map[string]value
		err error
	)
	m, err = getDataFromFile()
	if err != nil {
		m = make(map[string]value)
	}
	c := &cache{
		container: m,
		mu:        &sync.RWMutex{},
	}
	go func() {
		for {
			time.Sleep(10 * time.Second)
			go c.dumpToTheFile()
		}
	}()

	go c.clearCacheFromOldData()
	return c
}

type cache struct {
	container map[string]value
	mu        *sync.RWMutex
	action    chan struct{}
}

func (c *cache) Set(k, v string, ttl *int64) {
	c.mu.Lock()

	innerValue := value{
		Val:         v,
		CreatedTime: time.Now().UnixNano(),
		TTL:         defaultTTL,
	}
	if ttl != nil {
		innerValue.TTL = *ttl
	}

	c.container[k] = innerValue
	c.mu.Unlock()
}

func (c *cache) Read(key string) (string, bool) {
	c.mu.RLock()
	v, ok := c.container[key]
	c.mu.RUnlock()

	if !ok {
		return "", false
	}

	// data was expired
	if time.Now().UnixNano() > v.CreatedTime+v.TTL {
		c.Delete(key)
		return "", false
	}
	return v.Val, true
}

func (c *cache) Delete(key string) {
	c.mu.Lock()
	delete(c.container, key)
	c.mu.Unlock()
}

func (c *cache) clearCacheFromOldData() {
	for {
		time.Sleep(10 * time.Second)

		m := make(map[string]value)
		tnNano := time.Now().UnixNano()
		c.mu.RLock()
		for k, v := range c.container {
			if v.CreatedTime+v.TTL > tnNano {
				m[k] = v
			}
		}
		c.mu.RUnlock()

		c.mu.Lock()
		c.container = m
		c.mu.Unlock()
	}
}

func getDataFromFile() (map[string]value, error) {
	file, err := os.Open("./dump")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := bytes.Buffer{}
	_, err = io.Copy(&buf, file)
	if err != nil {
		return nil, err
	}

	var m map[string]value
	err = json.Unmarshal(buf.Bytes(), &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (c *cache) dumpToTheFile() {
	file, err := os.OpenFile("./dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Println("failed to open file to dump data", err)
		return
	}
	defer file.Close()

	c.mu.Lock()
	b, err := json.Marshal(c.container)
	c.mu.Unlock()
	if err != nil {
		log.Println("failed to serialize map container", err)
		return
	}
	_, err = file.Write(b)
	if err != nil {
		log.Println("failed to write serialized data to file", err)
	}
}

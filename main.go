package main

import (
	"github.com/flameous/xpate-kv/kv"
	"github.com/flameous/xpate-kv/kv/service"
	"log"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	c := kv.NewCacher()
	l := service.NewListener(c)
	log.Fatal(l.Start("8080"))
}

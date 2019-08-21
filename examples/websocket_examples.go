package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8000", "http service address")

func main() {
	flag.Parse()
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	fmt.Println(u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial: ", err)
	}
	defer c.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {

			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)

		}
	}()

	op := OpCommand{
		Op: "subscribe",
		Args: []string{
			// "blockinfo",
			"slash",
			// "depth:sdu1/cet",
			// "bancor:sdu1/cet",
			// "bancor-trade:coinex18c3hryxtjdtjjvnjm63r8k3p8tlhm0l6k96l9v",
			// "unlock:coinex18c3hryxtjdtjjvnjm63r8k3p8tlhm0l6k96l9v",
			// "unlock:coinex1kc2nguz9xfttfpav4drldh2w96xyzrnqss9scw",
			// "ticker:sdu1/cet",
			// "deal:sdu1/cet",
			// "comment:cet",
			// "order:coinex16ur229a4xkj9e0xu06nqge9c23y70g7sl5vj98",
			// "kline:sdu1/cet:16",
			// "txs:coinex16ur229a4xkj9e0xu06nqge9c23y70g7sl5vj98",
			// "income:coinex16ur229a4xkj9e0xu06nqge9c23y70g7sl5vj98",
			"redelegation:coinex19yl5cehvef8aghtpv0xf9wrmgn3s244deafs57",
			"unbonding:coinex19yl5cehvef8aghtpv0xf9wrmgn3s244deafs57",
		},
	}

	bz, err := json.Marshal(op)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = c.WriteMessage(websocket.TextMessage, bz)
	if err != nil {
		log.Fatal(err)
		return
	}
	<-done

}

type OpCommand struct {
	Op   string
	Args []string
}

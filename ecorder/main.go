package main

import (
	"flag"
	"fmt"
	"github.com/evecentral/eccore"
	"github.com/evecentral/ecorder"
	"github.com/gorilla/mux"
	"github.com/theatrus/crestmarket"
	"github.com/theatrus/crestmarket/helper"
	"github.com/theatrus/gomemcache/memcache"
	"log"
	"net/http"
)

var ecorderPort int
var mcHost string

func init() {
	flag.IntVar(&ecorderPort, "ecorder.port", 1933, "Port for HTTP server")
	flag.StringVar(&mcHost, "ecorder.memcache.host", "localhost:11211", "Host:port for memcache, only one supported")
}

func main() {
	flag.Parse()

	log.Println("Connecting to memcache")
	mc := memcache.New(mcHost)
	// Dummy write
	mc.Set(&memcache.Item{Key: "test", Value: []byte("ok")})

	log.Println("Connecting to the database")
	db, err := eccore.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Building a static cache of the universe")
	static, err := eccore.NewStaticItemsFromDatabase(db)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Loading CREST settings")
	settings, err := crestmarket.LoadSettings("settings.json")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("CREST authentication")
	transport, err := helper.InteractiveStartup("token.json", settings)
	if err != nil {
		log.Fatal(err)
	}

	requestor, err := crestmarket.NewCrestRequestor(transport)
	if err != nil {
		log.Fatal(err)
	}

	if requestor == nil {
		log.Fatal("wut")
	}

	log.Println("Building hydrator")
	hydrator, err := ecorder.NewCrestHydrator(requestor, static)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Building DB facade hydrator")
	dbOrderStore, err := eccore.NewOrderStore(db, static)
	if err != nil {
		log.Fatal(err)
	}
	dbHydrator, err := ecorder.NewDBStoreOrderHydrator(hydrator, dbOrderStore)
	if err != nil {
		log.Fatal(err)
	}
	
	log.Println("Building cache")
	orderCache := &ecorder.OrderCache{Hydrator: dbHydrator,
		Mc: mc}

	_, err = orderCache.OrdersForType(34, 10000002)
	if err != nil {
		log.Fatal(err)
	}

	rtr := mux.NewRouter()

	http.Handle("/", rtr)

	bind := fmt.Sprintf("localhost:%d", ecorderPort)
	log.Printf("Binding to %s", bind)
	http.ListenAndServe(bind, nil)
}

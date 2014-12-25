package main

import (
	"flag"
	"fmt"
	"github.com/evecentral/eccore"
	"github.com/gorilla/mux"
	"github.com/theatrus/crestmarket"
	"github.com/theatrus/crestmarket/helper"
	"log"
	"net/http"
)

var ecorderPort int

func init() {
	flag.IntVar(&ecorderPort, "ecorder.port", 1933, "Port for HTTP server")
}

func main() {
	flag.Parse()

	db, err := eccore.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	_, err = eccore.NewStaticItemsFromDatabase(db)
	if err != nil {
		log.Fatal(err)
	}

	settings, err := crestmarket.LoadSettings("settings.json")
	if err != nil {
		log.Fatal(err)
	}

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

	rtr := mux.NewRouter()

	http.Handle("/", rtr)

	bind := fmt.Sprintf(":%d", ecorderPort)
	log.Printf("Binding to %s", bind)
	http.ListenAndServe(bind, nil)
}

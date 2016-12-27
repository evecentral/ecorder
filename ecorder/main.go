package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/evecentral/sdetools"

	"cloud.google.com/go/datastore"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
	"github.com/theatrus/mediate"

	"github.com/evecentral/esiapi/client"
	"github.com/evecentral/esiapi/helper"

	"context"

	"gopkg.in/redis.v5"

	"log"
	"net/http"
)

var ecorderPort int
var redisHost string
var sdepath string
var settingsFile string
var tokenName string
var project string

func init() {
	flag.IntVar(&ecorderPort, "ecorder.port", 1933, "Port for HTTP server")
	flag.StringVar(&redisHost, "ecorder.redis.host", "localhost:11211", "Host:port for redis")
	flag.StringVar(&sdepath, "sde", "sde", "Path to the SDE root")
	flag.StringVar(&settingsFile, "esi.settings", "settings.json", "Default settings file")
	flag.StringVar(&project, "esi.ds.project", "", "Google Cloud Datastore Project")
	flag.StringVar(&tokenName, "esi.token.name", "default-token", "Name of the token")

}

func main() {
	flag.Parse()

	log.Println("Connecting to redis")
	rd := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "",
		DB:       0,
	})
	rd.Set("a", "a", 0)

	log.Println("Loading static SDE")
	sde := sdetools.SDE{
		BaseDir: sdepath,
	}
	sde.Init()

	log.Println("Loading tokens")
	settings, err := helper.LoadSettings(settingsFile)
	if err != nil {
		log.Fatal(err)
	}

	store, err := helper.NewDatastoreTokenStore(project, tokenName, 120*time.Second)
	if err != nil {
		fmt.Printf("Can't load datastore %v\n", err)
		return
	}

	transport, err := helper.InteractiveStartupWithTokenStore(store, settings)
	if err != nil {
		fmt.Printf("Can't do startup %v\n", err)
		return
	}

	cliTransport := httptransport.New("esi.tech.ccp.is", "/latest", []string{"https"})
	cliTransport.Transport = mediate.RateLimit(10, 1*time.Second, transport)

	// Connect to cloud datastore
	ctx := context.Background()

	_, err = datastore.NewClient(ctx, project)
	if err != nil {
		log.Fatal(err)
	}

	_ = client.New(cliTransport, strfmt.Default)

	rtr := mux.NewRouter()

	http.Handle("/", rtr)

	bind := fmt.Sprintf("localhost:%d", ecorderPort)
	log.Printf("Binding to %s", bind)
	http.ListenAndServe(bind, nil)
}

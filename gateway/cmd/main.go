package main

import (
	"APIGateway/gateway/config"
	"APIGateway/gateway/pkg/api"
	"APIGateway/gateway/pkg/middl"
	"flag"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

type server struct {
	api *api.API
}

func init() {
	if err := godotenv.Load("gateway/.env"); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	var srv server
	cfg := config.New()
	port := cfg.Gateway.AdrPort
	newsPort := cfg.News.AdrPort
	censorPort := cfg.Censor.AdrPort
	comment := cfg.Comments.AdrPort

	portFlag := flag.String("gateway-port", port, "Gateway port")
	portFlagNews := flag.String("news-port", newsPort, "News service port")
	portFlagCensor := flag.String("censor-port", censorPort, "Censor service port")
	portFlagComment := flag.String("comments-port", comment, "Comments service port")

	flag.Parse()
	srv.api = api.New(cfg, *portFlagNews, *portFlagCensor, *portFlagComment)
	srv.api.Router().Use(middl.Middle)

	log.Println("Gateway running at http://127.0.0.1" + *portFlag)
	log.Fatal(http.ListenAndServe(*portFlag, srv.api.Router()))
}

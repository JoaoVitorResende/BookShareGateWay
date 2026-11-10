package main

import (
	"BookShareGateWay/internal/httpapi"
	"BookShareGateWay/internal/messaging"
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitURL := os.Getenv("RABBIT_URL")
	pubKeyPath := os.Getenv("JWT_PUBLIC_KEY_PATH")
	if rabbitURL == "" || pubKeyPath == "" {
		log.Fatal("RABBIT_URL e JWT_PUBLIC_KEY_PATH são obrigatórios")
	}

	pem, err := os.ReadFile(pubKeyPath)
	if err != nil {
		log.Fatalf("lendo chave pública: %v", err)
	}
	pub, err := jwt.ParseRSAPublicKeyFromPEM(pem)
	if err != nil {
		log.Fatalf("chave pública inválida: %v", err)
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("conectando ao rabbit: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("abrindo canal: %v", err)
	}
	defer ch.Close()

	rpc, err := messaging.NewRPCClient(ch)
	if err != nil {
		log.Fatalf("criando rpc client: %v", err)
	}

	router := httpapi.NewRouter(
		httpapi.NewAuthHandler(rpc),
		httpapi.NewBookHandler(rpc),
		httpapi.NewAuthMiddleware(pub),
	)

	log.Fatal(http.ListenAndServe(":8080", router))
}

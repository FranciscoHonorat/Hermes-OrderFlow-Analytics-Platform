package main

import (
	"log"

	"github.com/FranciscoHonorat/ordemflow/services/order-service/config"
)

func main() {
	cfg := config.Get()

	log.Printf("iniciando order-service no ambiente: %s", cfg.Env)

	db, err := postegres.NewPostgresDB(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Fail to connect to PostgreSQL DB: %v", err)
	}
}

package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"dota_market_notifier/internal/currency/monobank"
	"dota_market_notifier/internal/market"
	"dota_market_notifier/internal/repository/db"
)

func main() {
	apiKey := mustGetEnvVar("API_KEY")
	repo := db.NewRepo(db.NewDB(getDBCredentials()))
	currency := monobank.New()
	marketParser := market.New(repo, currency, apiKey)
	tradeLots, err := marketParser.GetLastPrices(context.Background())
	if err != nil {
		panic(err)
	}
	for _, tradeLot := range tradeLots {
		fmt.Println(tradeLot)
	}
	if err := marketParser.UpdateHistory(context.Background(), tradeLots); err != nil {
		panic(err)
	}
}

func getDBCredentials() (host string, port int, username, password, dbName string) {
	host = mustGetEnvVar("DB_HOST")
	port, err := strconv.Atoi(mustGetEnvVar("DB_PORT"))
	if err != nil {
		panic(err)
	}
	username = mustGetEnvVar("DB_USER")
	password = mustGetEnvVar("DB_PASS")
	dbName = mustGetEnvVar("DB_NAME")
	return
}

func mustGetEnvVar(key string) (value string) {
	value, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("%s is not set", key))
	}
	return
}

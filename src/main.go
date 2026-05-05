package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func car_csv_handler() {

	file, err := os.Open("./data/cars.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		vin := record[6]
		year := record[0]
		make := record[1]
		model := record[2]
		trim := record[3]
		body := record[4]
		transmission := record[5]
		condition, _ := strconv.ParseFloat(record[8], 64)
		odometer, _ := strconv.Atoi(record[9])
		color := record[10]
		interior := record[11]
		seller := record[12]
		mmr, _ := strconv.Atoi(record[13])
		sellingPrice, _ := strconv.Atoi(record[14])
	}
}

func main() {

	// TODO: Vytvoř connection string pro PostgreSQL
	// Hint: user, password, localhost, port, database jméno
	connStr := "host=172.25.254.161 port=5432 user=user dbname=benchmark_db sslmode=disable"

	// TODO: Otevři spojení
	// Hint: sql.Open("pgx", connStr)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		// TODO: Co s tímto errorem? (log.Fatal je dobrá volba)
		log.Fatal(err)
	}

	// TODO: Zavři spojení na konci
	// Hint: defer db.Close()
	defer db.Close()

	// TODO: Ověří, že všechno je OK
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Připojeno k PostgreSQL!")
}

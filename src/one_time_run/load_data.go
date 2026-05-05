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

func fillPostgress(db *sql.DB) {
	file, err := os.Open("./data/cars.csv")
	if err != nil {
		log.Fatal("Nelze otevřít CSV:", err)
	}
	defer file.Close()

	yearMap := make(map[string]int)
	makeMap := make(map[string]int)
	modelMap := make(map[string]int)
	trimMap := make(map[string]int)
	bodyMap := make(map[string]int)
	transmissionMap := make(map[string]int)
	colorMap := make(map[string]int)
	interiorMap := make(map[string]int)

	reader := csv.NewReader(file)
	_, _ = reader.Read() // přeskočíme hlavičku

	// START TRANSAKCE - tohle je ten "turbo" režim
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	// Pokud program spadne, Rollback zruší rozdělanou práci
	defer tx.Rollback()

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// PARSOVÁNÍ S OŠETŘENÍM (v CSV bývají prázdné stringy)
		vin := record[6]
		rawYear := record[0]
		rawMake := record[1]
		rawModel := record[2]
		rawTrim := record[3]
		rawBody := record[4]
		rawTransmission := record[5]
		rawColor := record[10]
		rawInterior := record[11]

		sellingPrice, _ := strconv.ParseInt(record[14], 10, 64)
		mmr, _ := strconv.ParseInt(record[13], 10, 64)
		seller := record[12]
		odometer, _ := strconv.ParseInt(record[9], 10, 64)
		condition, _ := strconv.ParseFloat(record[8], 64)

		getOrCreateID := func(table string, column string, value string, cache map[string]int) int {
			if value == "" {
				value = "UNKNOWN"
			} // ošetření prázdných stringů
			if id, exists := cache[value]; exists {
				return id
			}
			var id int
			//query := fmt.Sprintf("INSERT INTO %s(%s) VALUES($1) ON CONFLICT (%s) DO UPDATE SET %s=EXCLUDED.%s RETURNING %s_id", table, column, column, column, column, table)
			query := fmt.Sprintf("INSERT INTO %s(%s) VALUES($1) ON CONFLICT (%s) DO UPDATE SET %s=EXCLUDED.%s RETURNING %s_id", table, column, column, column, column, table[4:])
			err := tx.QueryRow(query, value).Scan(&id)
			if err != nil {
				log.Fatalf("Chyba u %s (%s): %v", table, value, err)
			}
			cache[value] = id
			return id
		}

		makeID := getOrCreateID("dim_make", "make_name", rawMake, makeMap)
		modelID := getOrCreateID("dim_model", "model_name", rawModel, modelMap)
		trimID := getOrCreateID("dim_trim", "trim_name", rawTrim, trimMap)
		bodyID := getOrCreateID("dim_body", "body_name", rawBody, bodyMap)
		transmissionID := getOrCreateID("dim_transmission", "transmission_name", rawTransmission, transmissionMap)
		colorID := getOrCreateID("dim_color", "color_name", rawColor, colorMap)
		interiorID := getOrCreateID("dim_interior", "interior_name", rawInterior, interiorMap)

		var yearID int
		if id, exists := yearMap[rawYear]; exists {
			yearID = id
		} else {
			yVal, _ := strconv.ParseInt(rawYear, 10, 64)
			err := tx.QueryRow("INSERT INTO dim_year(year) VALUES($1) ON CONFLICT (year) DO UPDATE SET year=EXCLUDED.year RETURNING year_id", yVal).Scan(&yearID)
			if err != nil {
				log.Fatal(err)
			}
			yearMap[rawYear] = yearID
		}

		// INSERT FAKTU (opět přes tx)
		_, err = tx.Exec(`INSERT INTO fact_sales(
			vin, year_id, make_id, model_id, trim_id, body_id, transmission_id, 
			color_id, interior_id, odometer, condition, mmr, seller, selling_price
		) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) ON CONFLICT DO NOTHING`,
			vin, yearID, makeID, modelID, trimID, bodyID, transmissionID,
			colorID, interiorID, odometer, condition, mmr, seller, sellingPrice)
		if err != nil {
			log.Printf("Chyba u VIN %s: %v", vin, err)
		}

		count++
		if count%10000 == 0 {
			fmt.Printf("Načteno %d řádků...\n", count)
		}
	}

	// POTVRZENÍ VŠEHO NA DISK
	err = tx.Commit()
	if err != nil {
		log.Fatal("Nepodařilo se commitnout transakci:", err)
	}
	fmt.Printf("✅ Hotovo! Celkem importováno %d řádků.\n", count)
}

func createPostgresConnection() *sql.DB {
	// Vytvoř connection string pro PostgreSQL
	connStr := "host=172.25.254.161 port=5432 user=user dbname=benchmark_db sslmode=disable"

	// Otevři spojení
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Ověří, že všechno je OK
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Připojeno k PostgreSQL!")
	return db
}

func main() {
	db := createPostgresConnection()
	defer db.Close()

	fillPostgress(db)
}

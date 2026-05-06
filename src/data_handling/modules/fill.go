package data_handling

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/marcboeker/go-duckdb"
)

func fillDuckDBWithCars(db *sql.DB) {
	/*fill duckdb with cars data*/
	// Otevři DuckDB (nebo vytvoř, pokud neexistuje)

	// COPY FROM CSV (nejrychlejší pro DuckDB)
	query := `COPY raw_sales FROM './data/cars_clean.csv' (FORMAT CSV, HEADER true)`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Chyba při vkládání dat do DuckDB:", err)
	}
	fmt.Println("DuckDB filled!")
}

func fillPostgresWithCars(db *sql.DB) {
	/*fill postgres with cars data*/
	file, err := os.Open("./data/cars_clean.csv")
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
		vin := record[0]
		rawYear := record[1]
		rawMake := record[2]
		rawModel := record[3]
		rawTrim := record[4]
		rawBody := record[5]
		rawTransmission := record[6]
		rawColor := record[7]
		rawInterior := record[8]

		sellingPrice, _ := strconv.ParseInt(record[9], 10, 64)
		mmr, _ := strconv.ParseInt(record[10], 10, 64)
		seller := record[11]
		odometer, _ := strconv.ParseInt(record[12], 10, 64)
		condition, _ := strconv.ParseFloat(record[13], 64)

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

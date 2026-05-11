package data_handling

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	. "golap-benchmark/backend"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TableAsStringToCSV(content string, dbType SQLOverhead, newFilePathAndName string, headers []string) error {
	outFile, err := os.Create(newFilePathAndName)
	if err != nil {
		return fmt.Errorf("nelze vytvořit soubor %s: %v", newFilePathAndName, err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// Pokud máme headers, zapiš je jako první řádek
	if len(headers) > 0 {
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("chyba při zápisu headeru: %v", err)
		}
	}

	// Určit separator podle typu databáze
	var separator rune
	switch dbType.(type) {
	case *PostgresOverhead:
		// PostgreSQL vrací tab-separated values
		separator = '\t'
	case *DuckDBOverhead:
		// DuckDB vrací pipe-separated values
		separator = '|'
	default:
		return fmt.Errorf("neznámý typ databáze: %T", dbType)
	}

	// Zparsuj obsah podle separatoru
	reader := csv.NewReader(strings.NewReader(content))
	reader.Comma = separator

	// Čti řádky a piš do CSV
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("chyba při čtení řádku: %v", err)
		}

		// Trimuj whitespace z každého pole (pro tab-separated z Postgresu)
		for i := range record {
			record[i] = strings.TrimSpace(record[i])
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("chyba při zápisu do CSV: %v", err)
		}
	}

	return nil
}

func PreprocessCSV() {
	/*data cleaning*/
	// Otevři original CSV
	file, err := os.Open("./data/cars.csv")
	if err != nil {
		log.Fatal("Nelze otevřít CSV:", err)
	}
	defer file.Close()

	// Vytvoř nový CSV file
	outFile, err := os.Create("./data/cars_clean.csv")
	if err != nil {
		log.Fatal("Nelze vytvořit clean CSV:", err)
	}
	defer outFile.Close()

	reader := csv.NewReader(file)
	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// načteme hlavičku a z ní vybereme jen potřebné sloupce
	header, err := reader.Read()
	if err != nil {
		log.Fatal("Nelze přečíst hlavičku CSV:", err)
	}

	headerClean := []string{header[6], header[0], header[1], header[2], header[3], header[4], header[5], header[10], header[11], header[14], header[13], header[12], header[9], header[8]}
	err = writer.Write(headerClean)
	if err != nil {
		log.Fatal("Nelze zapsat hlavičku do clean CSV:", err)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("Chyba při čtení CSV:", err)
		}
		recordClean := []string{record[6], record[0], record[1], record[2], record[3], record[4], record[5], record[10], record[11], record[14], record[13], record[12], record[9], record[8]}

		// NAHRAĎ prázdné stringy na "UNKNOWN" (ale ne pro čísla!)
		// Index 12 = odometer, index 13 = condition (obojí čísla - nechej prázdné)
		for i, val := range recordClean {
			if val == "" && i != 12 && i != 13 {
				recordClean[i] = "UNKNOWN"
			}
		}

		err = writer.Write(recordClean)
		if err != nil {
			log.Fatal("Nelze zapsat řádek do clean CSV:", err)
		}
	}
}

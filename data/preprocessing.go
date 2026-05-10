package csv_handling

import (
	"encoding/csv"
	"io"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/marcboeker/go-duckdb"
)

func preprocessCSV() {
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
		err = writer.Write(recordClean)
		if err != nil {
			log.Fatal("Nelze zapsat řádek do clean CSV:", err)
		}
	}
}

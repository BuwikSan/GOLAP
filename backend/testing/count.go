package testing

import (
	"fmt"
	"log"
	"strings"

	. "golap-benchmark/backend"
)

func PostgresGetFactSales(pg SQLOverhead) {
	query := fmt.Sprintf("SELECT Count(*) FROM %s", "fact_sales")

	// -noheader a -list zajistí, že dostaneme jen to číslo
	output, duration, err := pg.QueryRow(query)
	if err != nil {
		log.Fatal(err)
	}

	cleanOut := strings.TrimSpace(output)

	// Převedeme string výstup na číslo
	var count int
	fmt.Sscanf(cleanOut, "%d", &count)

	fmt.Printf("Postgres: Načteno %d řádků za %v\n", count, duration)
}

// Přidej tuhle metodu do svého DuckDB wrapperu
func DuckGetRawSales(d SQLOverhead) {
	query := fmt.Sprintf("SELECT Count(*) FROM %s", "raw_sales")

	// -noheader a -list zajistí, že dostaneme jen to číslo
	output, duration, err := d.QueryRow(query)
	if err != nil {
		log.Fatal(err)
	}

	cleanOut := strings.TrimSpace(output)

	// Převedeme string výstup na číslo
	var count int
	fmt.Sscanf(cleanOut, "%d", &count)

	fmt.Printf("DuckDB: Načteno %d řádků za %v\n", count, duration)
}

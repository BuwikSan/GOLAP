// CompareDataBetweenDatabases porovná počty záznamů a checksumem identifikuje rozdíly
package testing

import (
	"fmt"
	. "golap-benchmark/src"
	"strconv"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func CompareDataBetweenDatabases(postgresDB SQLOverhead, duckdbDB SQLOverhead) {
	fmt.Println("===================================")
	fmt.Println("🔍 POROVNÁNÍ DAT MEZI DATABÁZEMI")
	fmt.Println("===================================")

	// 1. Porovnání počtu záznamů
	fmt.Println("\n📊 Počet záznamů:")

	postgresCountStr, _, err1 := postgresDB.QueryRow("SELECT COUNT(*) FROM fact_sales")
	postgresCountStr = strings.TrimSpace(postgresCountStr)
	fmt.Printf("[DEBUG] PostgreSQL response: '%s', error: %v\n", postgresCountStr, err1)
	postgresCount, _ := strconv.ParseInt(postgresCountStr, 10, 64)

	duckdbCountStr, _, err2 := duckdbDB.QueryRow("SELECT COUNT(*) FROM raw_sales")
	duckdbCountStr = strings.TrimSpace(duckdbCountStr)
	fmt.Printf("[DEBUG] DuckDB response: '%s', error: %v\n", duckdbCountStr, err2)
	duckdbCount, _ := strconv.ParseInt(duckdbCountStr, 10, 64)

	fmt.Printf("PostgreSQL (fact_sales): %d záznamů\n", postgresCount)
	fmt.Printf("DuckDB (raw_sales):      %d záznamů\n", duckdbCount)

	if postgresCount == duckdbCount {
		fmt.Printf("✅ Počty se shodují!\n")
	} else {
		fmt.Printf("❌ ROZDÍL: %d záznamů\n", postgresCount-duckdbCount)
	}

	// 2. Porovnání podle VINu (který by měl být unikátní)
	fmt.Println("\n📊 Porovnání unikátních VINů:")

	postgresVINStr, _, _ := postgresDB.QueryRow("SELECT COUNT(DISTINCT vin) FROM fact_sales WHERE vin IS NOT NULL AND vin != ''")
	postgresVINStr = strings.TrimSpace(postgresVINStr)
	postgresVINCount, _ := strconv.ParseInt(postgresVINStr, 10, 64)

	duckdbVINStr, _, _ := duckdbDB.QueryRow("SELECT COUNT(DISTINCT vin) FROM raw_sales WHERE vin IS NOT NULL AND vin != 'UNKNOWN'")
	duckdbVINStr = strings.TrimSpace(duckdbVINStr)
	duckdbVINCount, _ := strconv.ParseInt(duckdbVINStr, 10, 64)

	fmt.Printf("PostgreSQL (unique VINs): %d\n", postgresVINCount)
	fmt.Printf("DuckDB (unique VINs):      %d\n", duckdbVINCount)

	if postgresVINCount == duckdbVINCount {
		fmt.Printf("✅ VINy se shodují!\n")
	} else {
		fmt.Printf("❌ ROZDÍL: %d VINů\n", postgresVINCount-duckdbVINCount)
	}

	// 3. Porovnání agregací (průměrná cena, suma, atd.)
	fmt.Println("\n💰 Porovnání agregací:")

	postgresAvgStr, _, _ := postgresDB.QueryRow("SELECT ROUND(AVG(CAST(selling_price AS NUMERIC)), 2) FROM fact_sales")
	postgresAvgStr = strings.TrimSpace(postgresAvgStr)
	postgresAvgPrice, _ := strconv.ParseFloat(postgresAvgStr, 64)

	duckdbAvgStr, _, _ := duckdbDB.QueryRow("SELECT ROUND(AVG(selling_price), 2) FROM raw_sales")
	duckdbAvgStr = strings.TrimSpace(duckdbAvgStr)
	duckdbAvgPrice, _ := strconv.ParseFloat(duckdbAvgStr, 64)

	fmt.Printf("PostgreSQL (avg price): %.2f\n", postgresAvgPrice)
	fmt.Printf("DuckDB (avg price):     %.2f\n", duckdbAvgPrice)

	if postgresAvgPrice == duckdbAvgPrice {
		fmt.Printf("✅ Průměrné ceny se shodují!\n")
	} else {
		fmt.Printf("⚠️  ROZDÍL: %.2f\n", postgresAvgPrice-duckdbAvgPrice)
	}

	// 4. Porovnání počtů podle kategorie (např. UNKNOWN)
	fmt.Println("\n📋 Porovnání null/unknown záznamů:")

	postgresUnknownStr, _, _ := postgresDB.QueryRow("SELECT COUNT(*) FROM fact_sales fs JOIN dim_make dm ON fs.make_id = dm.make_id WHERE dm.make_name = 'UNKNOWN'")
	postgresUnknownStr = strings.TrimSpace(postgresUnknownStr)
	postgresUnknownMake, _ := strconv.ParseInt(postgresUnknownStr, 10, 64)

	duckdbUnknownStr, _, _ := duckdbDB.QueryRow("SELECT COUNT(*) FROM raw_sales WHERE make = 'UNKNOWN'")
	duckdbUnknownStr = strings.TrimSpace(duckdbUnknownStr)
	duckdbUnknownMake, _ := strconv.ParseInt(duckdbUnknownStr, 10, 64)

	fmt.Printf("PostgreSQL (UNKNOWN makes): %d\n", postgresUnknownMake)
	fmt.Printf("DuckDB (UNKNOWN makes):     %d\n", duckdbUnknownMake)

	if postgresUnknownMake == duckdbUnknownMake {
		fmt.Printf("✅ UNKNOWN záznamy se shodují!\n")
	} else {
		fmt.Printf("❌ ROZDÍL: %d záznamů\n", postgresUnknownMake-duckdbUnknownMake)
	}

	fmt.Println("\n" + "===================================")
	if postgresCount == duckdbCount && postgresVINCount == duckdbVINCount && postgresUnknownMake == duckdbUnknownMake {
		fmt.Println("✅ VŠECHNA DATA SE SHODUJÍ! Databáze jsou identické.")
	} else {
		fmt.Println("⚠️  POZOR: Existují rozdíly mezi databázemi!")
	}
	fmt.Println("===================================")

	// 5. Detailní kontrola všech sloupců
	fmt.Println("\n🔬 DETAILNÍ ANALÝZA SLOUPCŮ:")

	// Pro PostgreSQL musíme dělat JOINy na dimension tabulky!
	columns := []struct {
		name           string
		postgresTable  string
		postgresColumn string
		duckdbTable    string
		duckdbColumn   string
	}{
		{"make", "dim_make", "make_name", "raw_sales", "make"},
		{"model", "dim_model", "model_name", "raw_sales", "model"},
		{"body", "dim_body", "body_name", "raw_sales", "body"},
		{"transmission", "dim_transmission", "transmission_name", "raw_sales", "transmission"},
		{"color", "dim_color", "color_name", "raw_sales", "color"},
		{"interior", "dim_interior", "interior_name", "raw_sales", "interior"},
	}

	for _, col := range columns {
		fmt.Printf("\n📋 Sloupec '%s':\n", col.name)

		// PostgreSQL - počet unikátních hodnot (JOIN na dimension tabulku)
		postgresDistinctQuery := fmt.Sprintf("SELECT COUNT(DISTINCT %s) FROM %s", col.postgresColumn, col.postgresTable)
		postgresDistinctStr, _, _ := postgresDB.QueryRow(postgresDistinctQuery)
		postgresDistinct, _ := strconv.ParseInt(strings.TrimSpace(postgresDistinctStr), 10, 64)

		// DuckDB - počet unikátních hodnot
		duckdbDistinctQuery := fmt.Sprintf("SELECT COUNT(DISTINCT %s) FROM %s", col.duckdbColumn, col.duckdbTable)
		duckdbDistinctStr, _, _ := duckdbDB.QueryRow(duckdbDistinctQuery)
		duckdbDistinct, _ := strconv.ParseInt(strings.TrimSpace(duckdbDistinctStr), 10, 64)

		fmt.Printf("  PostgreSQL unique values: %d\n", postgresDistinct)
		fmt.Printf("  DuckDB unique values:     %d\n", duckdbDistinct)

		if postgresDistinct == duckdbDistinct {
			fmt.Printf("  ✅ Shodují se!\n")
		} else {
			fmt.Printf("  ❌ ROZDÍL: %d\n", postgresDistinct-duckdbDistinct)
		}

		// TOP 5 nejčastěji se vyskytující hodnoty (PostgreSQL - JOIN)
		fmt.Printf("  Top 5 hodnot (PostgreSQL):\n")
		postgresTopQuery := fmt.Sprintf("SELECT %s, COUNT(*) as cnt FROM %s GROUP BY %s ORDER BY cnt DESC LIMIT 5", col.postgresColumn, col.postgresTable, col.postgresColumn)
		postgresTopStr, _, _ := postgresDB.QueryRow(postgresTopQuery)
		fmt.Printf("    %s\n", strings.TrimSpace(postgresTopStr))

		fmt.Printf("  Top 5 hodnot (DuckDB):\n")
		duckdbTopQuery := fmt.Sprintf("SELECT %s, COUNT(*) as cnt FROM %s GROUP BY %s ORDER BY cnt DESC LIMIT 5", col.duckdbColumn, col.duckdbTable, col.duckdbColumn)
		duckdbTopStr, _, _ := duckdbDB.QueryRow(duckdbTopQuery)
		fmt.Printf("    %s\n", strings.TrimSpace(duckdbTopStr))
	}

	// Numerické sloupce
	fmt.Println("\n💰 NUMERICKÉ SLOUPCE:")

	numericCols := []struct {
		name  string
		table string
	}{
		{"selling_price", "raw_sales"},
		{"mmr", "raw_sales"},
		{"odometer", "raw_sales"},
		{"condition", "raw_sales"},
	}

	for _, col := range numericCols {
		fmt.Printf("\n📊 Sloupec '%s':\n", col.name)

		// PostgreSQL stats - filtruj NULL hodnoty!
		postgresStatsQuery := fmt.Sprintf("SELECT MIN(%s), MAX(%s), AVG(%s), STDDEV(%s) FROM fact_sales WHERE %s IS NOT NULL AND %s > 0", col.name, col.name, col.name, col.name, col.name, col.name)
		postgresStatsStr, _, _ := postgresDB.QueryRow(postgresStatsQuery)
		fmt.Printf("  PostgreSQL: %s\n", strings.TrimSpace(postgresStatsStr))

		// DuckDB stats - stejně tak
		duckdbStatsQuery := fmt.Sprintf("SELECT MIN(%s), MAX(%s), AVG(%s), STDDEV(%s) FROM raw_sales WHERE %s IS NOT NULL", col.name, col.name, col.name, col.name, col.name)
		duckdbStatsStr, _, _ := duckdbDB.QueryRow(duckdbStatsQuery)
		fmt.Printf("  DuckDB:     %s\n", strings.TrimSpace(duckdbStatsStr))

		// Kontrola nul/prázdných hodnot
		postgresNullQuery := fmt.Sprintf("SELECT COUNT(*) FROM fact_sales WHERE %s IS NULL OR %s = 0", col.name, col.name)
		postgresNullStr, _, _ := postgresDB.QueryRow(postgresNullQuery)
		postgresNull, _ := strconv.ParseInt(strings.TrimSpace(postgresNullStr), 10, 64)

		duckdbNullQuery := fmt.Sprintf("SELECT COUNT(*) FROM raw_sales WHERE %s IS NULL", col.name)
		duckdbNullStr, _, _ := duckdbDB.QueryRow(duckdbNullQuery)
		duckdbNull, _ := strconv.ParseInt(strings.TrimSpace(duckdbNullStr), 10, 64)

		fmt.Printf("  NULL/prázdné (PostgreSQL): %d\n", postgresNull)
		fmt.Printf("  NULL/prázdné (DuckDB):     %d\n", duckdbNull)

		if postgresNull != duckdbNull {
			fmt.Printf("  ⚠️  ROZDÍL V NULL HANDLINGU: %d\n", postgresNull-duckdbNull)
		}
	}
}

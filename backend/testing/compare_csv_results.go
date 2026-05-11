package testing

import (
"fmt"
)

// PrintDetailedQueryComparison vytiskne detailní porovnání jedné query (analytika sloupců)
func PrintDetailedQueryComparison(pgFilePath, duckFilePath, queryName string) {
fmt.Printf("\nDETAILNÍ ANALYTIKA: %s\n", queryName)
fmt.Printf("PostgreSQL: %s\n", pgFilePath)
fmt.Printf("DuckDB:     %s\n", duckFilePath)

// Sloupce analytika
fmt.Printf("\n ANALYTIKA SLOUPCŮ (Porovnání hodnot a jejich výskytů):\n")

// Načti headers
rows1, _ := readCSVFile(pgFilePath)
var headers []string
if len(rows1) > 0 {
headers = rows1[0]
}

columnAnalytics, err := AnalyzeAllColumns(pgFilePath, duckFilePath, headers)
if err != nil {
fmt.Printf(" Chyba při analytice sloupců: %v\n", err)
return
}

// Vytiskni analytiku (jen problémy - sloupce s rozdíly)
PrintColumnAnalytics(columnAnalytics, true)
}

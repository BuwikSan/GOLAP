package testing

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// ColumnAnalytics obsahuje statistiky jednoho sloupce
type ColumnAnalytics struct {
	ColumnName string
	ColumnIdx  int
	IsNumeric  bool // Detekce, zda je sloupec numeric

	// PostgreSQL
	PG_Distinct     int             // Počet unikátních hodnot
	PG_TopValues    []ValueCount    // Top N hodnoty a jejich počty
	PG_NullCount    int             // Počet NULL/prázdných
	PG_ValueSet     map[string]bool // Všechny unikátní hodnoty
	PG_SampleValues []string        // Vzorek hodnot

	// DuckDB
	Duck_Distinct     int
	Duck_TopValues    []ValueCount
	Duck_NullCount    int
	Duck_ValueSet     map[string]bool
	Duck_SampleValues []string

	// Srovnání
	HasDifferences  bool
	CommonValues    int // Kolik hodnot mají oba
	OnlyInPG        []string
	OnlyInDuck      []string
	DifferenceRatio float64 // 0-1, kde 1 = zcela rozdílné
}

// ValueCount reprezentuje počet výskytů hodnoty
type ValueCount struct {
	Value string
	Count int
}

// Helper functions

func isNumericValue(val string) bool {
	if val == "" || val == "NULL" {
		return true // Treat empty and NULL as acceptable in numeric columns
	}
	_, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
	return err == nil
}

func detectIsNumericColumn(values []string) bool {
	if len(values) == 0 {
		return false
	}

	// Check if at least 80% of values are numeric or NULL
	numericCount := 0
	nonEmptyCount := 0

	for _, val := range values {
		if val != "" && val != "NULL" {
			nonEmptyCount++
			if isNumericValue(val) {
				numericCount++
			}
		}
	}

	if nonEmptyCount == 0 {
		return false
	}

	return float64(numericCount)/float64(nonEmptyCount) > 0.8
}

func normalizeNumericValue(val string) string {
	if val == "" || val == "NULL" {
		return "NULL"
	}

	trimmed := strings.TrimSpace(val)
	if isNumericValue(trimmed) {
		floatVal, _ := strconv.ParseFloat(trimmed, 64)
		// If it's an integer displayed as float, normalize it
		if floatVal == float64(int64(floatVal)) {
			return strconv.FormatInt(int64(floatVal), 10)
		}
		// Otherwise keep 2 decimal places
		return fmt.Sprintf("%.2f", floatVal)
	}

	return trimmed
}

// AnalyzeColumn provádí analytiku jednoho sloupce
func AnalyzeColumn(pgValues, duckValues []string, columnName string, columnIdx int) *ColumnAnalytics {
	analytics := &ColumnAnalytics{
		ColumnName:    columnName,
		ColumnIdx:     columnIdx,
		PG_ValueSet:   make(map[string]bool),
		Duck_ValueSet: make(map[string]bool),
		IsNumeric:     false,
	}

	// Detekuj zda je sloupec numeric
	combinedValues := append(pgValues, duckValues...)
	analytics.IsNumeric = detectIsNumericColumn(combinedValues)

	// Analýza PostgreSQL
	pgCounts := make(map[string]int)
	for _, val := range pgValues {
		normalizedVal := val
		if analytics.IsNumeric {
			normalizedVal = normalizeNumericValue(val)
		}

		analytics.PG_ValueSet[normalizedVal] = true
		pgCounts[normalizedVal]++

		if normalizedVal == "" || normalizedVal == "NULL" {
			analytics.PG_NullCount++
		}
	}
	analytics.PG_Distinct = len(analytics.PG_ValueSet)
	analytics.PG_TopValues = getTopValues(pgCounts, 5)
	analytics.PG_SampleValues = getSampleValues(pgValues, 3)

	// Analýza DuckDB
	duckCounts := make(map[string]int)
	for _, val := range duckValues {
		normalizedVal := val
		if analytics.IsNumeric {
			normalizedVal = normalizeNumericValue(val)
		}

		analytics.Duck_ValueSet[normalizedVal] = true
		duckCounts[normalizedVal]++

		if normalizedVal == "" || normalizedVal == "NULL" {
			analytics.Duck_NullCount++
		}
	}
	analytics.Duck_Distinct = len(analytics.Duck_ValueSet)
	analytics.Duck_TopValues = getTopValues(duckCounts, 5)
	analytics.Duck_SampleValues = getSampleValues(duckValues, 3)

	// Porovnání
	analytics.compareValueSets()

	return analytics
}

// compareValueSets porovnává value sets obou verzí
func (ca *ColumnAnalytics) compareValueSets() {
	// Najdi commonValues
	for val := range ca.PG_ValueSet {
		if ca.Duck_ValueSet[val] {
			ca.CommonValues++
		} else {
			ca.OnlyInPG = append(ca.OnlyInPG, val)
		}
	}

	// Najdi values pouze v DuckDB
	for val := range ca.Duck_ValueSet {
		if !ca.PG_ValueSet[val] {
			ca.OnlyInDuck = append(ca.OnlyInDuck, val)
		}
	}

	// Vypočítej ratio rozdílnosti
	totalUnique := len(ca.PG_ValueSet) + len(ca.Duck_ValueSet) - ca.CommonValues
	if totalUnique > 0 {
		ca.DifferenceRatio = float64(len(ca.OnlyInPG)+len(ca.OnlyInDuck)) / float64(totalUnique)
	}

	ca.HasDifferences = len(ca.OnlyInPG) > 0 || len(ca.OnlyInDuck) > 0
}

// AnalyzeAllColumns provádí analytiku všech sloupců
func AnalyzeAllColumns(file1Path, file2Path string, columnNames []string) (map[string]*ColumnAnalytics, error) {
	// Načti data
	rows1, err := readCSVFile(file1Path)
	if err != nil {
		return nil, err
	}
	rows2, err := readCSVFile(file2Path)
	if err != nil {
		return nil, err
	}

	// Určí zda má soubor header (pokud columnNames není prázdný, máme header)
	var pgData [][]string
	var duckData [][]string
	hasHeader := len(columnNames) > 0 && columnNames[0] != ""

	// Pokud máme header, přeskoč první řádek
	if hasHeader && len(rows1) > 1 {
		pgData = rows1[1:]
	} else if len(rows1) > 0 {
		pgData = rows1
	}

	if hasHeader && len(rows2) > 1 {
		duckData = rows2[1:]
	} else if len(rows2) > 0 {
		duckData = rows2
	}

	// Určí počet sloupců
	numColumns := 0
	if len(rows1) > 0 {
		numColumns = len(rows1[0])
	}
	if len(rows2) > 0 && len(rows2[0]) > numColumns {
		numColumns = len(rows2[0])
	}

	// Pokud nemáme columnNames, vytvoř default jména
	if len(columnNames) == 0 || !hasHeader {
		columnNames = make([]string, numColumns)
		for i := 0; i < numColumns; i++ {
			columnNames[i] = fmt.Sprintf("Column_%d", i)
		}
	}

	// Sběr hodnot pro každý sloupec
	analytics := make(map[string]*ColumnAnalytics)

	for colIdx := 0; colIdx < numColumns; colIdx++ {
		colName := ""
		if colIdx < len(columnNames) {
			colName = columnNames[colIdx]
		} else {
			colName = fmt.Sprintf("Column_%d", colIdx)
		}

		// Sbírej hodnoty z PostgreSQL
		var pgValues []string
		for _, row := range pgData {
			if colIdx < len(row) {
				pgValues = append(pgValues, row[colIdx])
			}
		}

		// Sbírej hodnoty z DuckDB
		var duckValues []string
		for _, row := range duckData {
			if colIdx < len(row) {
				duckValues = append(duckValues, row[colIdx])
			}
		}

		// Analyzuj sloupec
		analytics[colName] = AnalyzeColumn(pgValues, duckValues, colName, colIdx)
	}

	return analytics, nil
}

// PrintColumnAnalytics vytiskne analytiku sloupců
func PrintColumnAnalytics(analytics map[string]*ColumnAnalytics, problemOnly bool) {
	fmt.Println("\n" + strings.Repeat("=", 120))
	fmt.Printf("ANALYTIKA SLOUPCŮ - Porovnání PostgreSQL vs DuckDB\n")
	fmt.Println(strings.Repeat("=", 120))

	// Seřaď sloupce
	var colNames []string
	for name := range analytics {
		colNames = append(colNames, name)
	}
	sort.Strings(colNames)

	problemCount := 0

	for _, colName := range colNames {
		ca := analytics[colName]

		// Pokud problemOnly, přeskoč sloupce bez problémů
		if problemOnly && !ca.HasDifferences {
			continue
		}

		if ca.HasDifferences {
			problemCount++
			fmt.Printf("\n🔴 PROBLÉM: %s (Col %d) [Type: %s]\n", colName, ca.ColumnIdx,
				map[bool]string{true: "NUMERIC", false: "STRING"}[ca.IsNumeric])
		} else {
			fmt.Printf("\n✅ OK: %s (Col %d) [Type: %s]\n", colName, ca.ColumnIdx,
				map[bool]string{true: "NUMERIC", false: "STRING"}[ca.IsNumeric])
		}

		fmt.Printf("   PostgreSQL: %d distinct values (NULL: %d)\n", ca.PG_Distinct, ca.PG_NullCount)
		fmt.Printf("   DuckDB:     %d distinct values (NULL: %d)\n", ca.Duck_Distinct, ca.Duck_NullCount)
		fmt.Printf("   Common values: %d | Difference ratio: %.2f%%\n", ca.CommonValues, ca.DifferenceRatio*100)

		// Zobraz top hodnoty
		if len(ca.PG_TopValues) > 0 {
			fmt.Printf("\n   PostgreSQL Top Values:\n")
			for _, vc := range ca.PG_TopValues {
				fmt.Printf("      '%s': %d times\n", vc.Value, vc.Count)
			}
		}

		if len(ca.Duck_TopValues) > 0 {
			fmt.Printf("\n   DuckDB Top Values:\n")
			for _, vc := range ca.Duck_TopValues {
				fmt.Printf("      '%s': %d times\n", vc.Value, vc.Count)
			}
		}

		// Zobraz rozdíly
		if len(ca.OnlyInPG) > 0 {
			fmt.Printf("\n   ⬅️  Jen v PostgreSQL (%d):\n", len(ca.OnlyInPG))
			// Zobraz max 5
			for i := 0; i < len(ca.OnlyInPG) && i < 5; i++ {
				fmt.Printf("      '%s'\n", ca.OnlyInPG[i])
			}
			if len(ca.OnlyInPG) > 5 {
				fmt.Printf("      ... a %d dalších\n", len(ca.OnlyInPG)-5)
			}
		}

		if len(ca.OnlyInDuck) > 0 {
			fmt.Printf("\n   ➡️  Jen v DuckDB (%d):\n", len(ca.OnlyInDuck))
			// Zobraz max 5
			for i := 0; i < len(ca.OnlyInDuck) && i < 5; i++ {
				fmt.Printf("      '%s'\n", ca.OnlyInDuck[i])
			}
			if len(ca.OnlyInDuck) > 5 {
				fmt.Printf("      ... a %d dalších\n", len(ca.OnlyInDuck)-5)
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 120))
	fmt.Printf("Celkem problematických sloupců: %d/%d\n", problemCount, len(analytics))
	fmt.Println(strings.Repeat("=", 120))
}

// Helper functions

func getTopValues(counts map[string]int, topN int) []ValueCount {
	var values []ValueCount
	for val, count := range counts {
		values = append(values, ValueCount{val, count})
	}

	// Seřaď descending by count
	sort.Slice(values, func(i, j int) bool {
		return values[i].Count > values[j].Count
	})

	if len(values) > topN {
		values = values[:topN]
	}

	return values
}

func getSampleValues(values []string, sampleSize int) []string {
	var sample []string
	seen := make(map[string]bool)

	for _, val := range values {
		if !seen[val] && len(sample) < sampleSize {
			sample = append(sample, val)
			seen[val] = true
		}
	}

	return sample
}

// readCSVFile načte CSV soubor a vrátí všechny řádky (včetně headeru)
func readCSVFile(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

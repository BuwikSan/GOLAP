package testing

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// CSVComparison obsahuje výsledky porovnání dvou CSV souborů
type CSVComparison struct {
	File1Path       string
	File2Path       string
	RowsFile1       int
	RowsFile2       int
	ColumnsMatch    bool
	RowCountMatch   bool
	MatchingRows    int
	DifferentRows   []RowDifference
	MissingInFile2  []int // řádky chybějící v souboru 2
	MissingInFile1  []int // řádky chybějící v souboru 1
	ExtraInFile1    []int // extra řádky v souboru 1
	ExtraInFile2    []int // extra řádky v souboru 2
	MatchPercentage float64
}

// RowDifference obsahuje informace o rozdílech v řádku
type RowDifference struct {
	RowNumber  int
	Column     int
	ColumnName string
	File1Value string
	File2Value string
}

// QueryResultComparison porovnává výsledky jednoho query na obou DB
type QueryResultComparison struct {
	QueryName   string
	QueryNumber int
	PostgreSQL  *CSVComparison
	DuckDB      *CSVComparison
	AreSame     bool
	Summary     string
}

// AllQueriesComparison obsahuje porovnání všech queries
type AllQueriesComparison struct {
	Queries          []QueryResultComparison
	DataPath         string
	SavedQueriesPath string
	TotalQueries     int
	MatchingQueries  int
	DifferentQueries int
}

// CompareTwoCSVFiles porovnává dva CSV soubory a vrací detailní výsledky
func CompareTwoCSVFiles(file1Path, file2Path string) (*CSVComparison, error) {
	comparison := &CSVComparison{
		File1Path:      file1Path,
		File2Path:      file2Path,
		DifferentRows:  []RowDifference{},
		MissingInFile2: []int{},
		MissingInFile1: []int{},
		ExtraInFile1:   []int{},
		ExtraInFile2:   []int{},
	}

	// Načti oba soubory
	rows1, err := readCSVFile(file1Path)
	if err != nil {
		return nil, fmt.Errorf("chyba při čtení souboru 1 (%s): %v", file1Path, err)
	}

	rows2, err := readCSVFile(file2Path)
	if err != nil {
		return nil, fmt.Errorf("chyba při čtení souboru 2 (%s): %v", file2Path, err)
	}

	comparison.RowsFile1 = len(rows1)
	comparison.RowsFile2 = len(rows2)

	// Kontrola počtu sloupců (v prvním řádku - header)
	if len(rows1) > 0 && len(rows2) > 0 {
		comparison.ColumnsMatch = len(rows1[0]) == len(rows2[0])
	}

	// Kontrola počtu řádků (bez headeru)
	comparison.RowCountMatch = comparison.RowsFile1 == comparison.RowsFile2

	// Porovnání řádků
	var headers []string
	minRows := comparison.RowsFile1
	if comparison.RowsFile2 < minRows {
		minRows = comparison.RowsFile2
	}

	// Vezmi headers z prvního souboru
	if len(rows1) > 0 {
		headers = rows1[0]
	}

	// Porovnání dat (od řádku 1, header přeskočit)
	for i := 1; i < minRows; i++ {
		row1 := rows1[i]
		row2 := rows2[i]

		minCols := len(row1)
		if len(row2) < minCols {
			minCols = len(row2)
		}

		rowDiffers := false
		for j := 0; j < minCols; j++ {
			if row1[j] != row2[j] {
				rowDiffers = true
				columnName := ""
				if j < len(headers) {
					columnName = headers[j]
				}
				comparison.DifferentRows = append(comparison.DifferentRows, RowDifference{
					RowNumber:  i,
					Column:     j,
					ColumnName: columnName,
					File1Value: row1[j],
					File2Value: row2[j],
				})
			}
		}

		if !rowDiffers {
			comparison.MatchingRows++
		}
	}

	// Setkej chybějící/extra řádky
	if comparison.RowsFile1 > comparison.RowsFile2 {
		for i := comparison.RowsFile2; i < comparison.RowsFile1; i++ {
			comparison.ExtraInFile1 = append(comparison.ExtraInFile1, i)
		}
	} else if comparison.RowsFile2 > comparison.RowsFile1 {
		for i := comparison.RowsFile1; i < comparison.RowsFile2; i++ {
			comparison.ExtraInFile2 = append(comparison.ExtraInFile2, i)
		}
	}

	// Vypočítej procentuální shodu
	if comparison.RowsFile1 > 0 {
		comparison.MatchPercentage = (float64(comparison.MatchingRows) / float64(comparison.RowsFile1-1)) * 100
	}

	return comparison, nil
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

// QueryPair je pomocná struktura pro hledání párů files
type QueryPair struct {
	QueryName      string
	PostgreSQLFile string
	DuckDBFile     string
}

// CompareSavedQueries je wrapper funkce, která porovnává všechny uložené výsledky queries
// Vyhledá všechny saved_queries soubory v data/saved_queries a porovnává je
func CompareSavedQueries(dataPath string) (*AllQueriesComparison, error) {
	allComparison := &AllQueriesComparison{
		DataPath:         dataPath,
		SavedQueriesPath: filepath.Join(dataPath, "saved_queries"),
		Queries:          []QueryResultComparison{},
	}

	// Zkontroluj, zda existuje saved_queries složka
	if _, err := os.Stat(allComparison.SavedQueriesPath); os.IsNotExist(err) {
		fmt.Printf("⚠️  Složka saved_queries neexistuje: %s\n", allComparison.SavedQueriesPath)
		return allComparison, nil // Vrať prázdnou strukturu místo nil
	}

	// Vyhledej všechny soubory v saved_queries
	files, err := ioutil.ReadDir(allComparison.SavedQueriesPath)
	if err != nil {
		return nil, fmt.Errorf("chyba při čtení saved_queries složky: %v", err)
	}

	// Setkej query páry (PostgreSQL a DuckDB verze)
	queryPairs := make(map[string]*QueryPair)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()

		// Očekávaný formát: query_1_postgres.csv, query_1_duckdb.csv
		if strings.HasSuffix(fileName, ".csv") {
			parts := strings.Split(strings.TrimSuffix(fileName, ".csv"), "_")
			if len(parts) >= 3 {
				dbType := parts[len(parts)-1] // "postgres" nebo "duckdb"
				baseQueryName := strings.Join(parts[:len(parts)-1], "_")

				if _, exists := queryPairs[baseQueryName]; !exists {
					queryPairs[baseQueryName] = &QueryPair{
						QueryName: baseQueryName,
					}
				}

				filePath := filepath.Join(allComparison.SavedQueriesPath, fileName)

				if dbType == "postgres" {
					queryPairs[baseQueryName].PostgreSQLFile = filePath
				} else if dbType == "duckdb" {
					queryPairs[baseQueryName].DuckDBFile = filePath
				}
			}
		}
	}

	// Porovnaj jednotlivé páry
	queryNum := 1
	for _, pair := range queryPairs {
		if pair.PostgreSQLFile != "" && pair.DuckDBFile != "" {
			pgComparison, err := CompareTwoCSVFiles(pair.PostgreSQLFile, pair.DuckDBFile)
			if err != nil {
				fmt.Printf("Chyba při porovnávání %s: %v\n", pair.QueryName, err)
				continue
			}

			queryComp := QueryResultComparison{
				QueryName:   pair.QueryName,
				QueryNumber: queryNum,
				PostgreSQL:  pgComparison,
				DuckDB:      pgComparison, // stejné porovnání
				AreSame:     len(pgComparison.DifferentRows) == 0 && pgComparison.RowCountMatch,
			}

			if queryComp.AreSame {
				queryComp.Summary = fmt.Sprintf("✅ %s: IDENTICKÉ (řádky: %d, shoda: %.2f%%)",
					pair.QueryName, pgComparison.RowsFile1, pgComparison.MatchPercentage)
				allComparison.MatchingQueries++
			} else {
				queryComp.Summary = fmt.Sprintf("❌ %s: ROZDÍLNÉ (PostgreSQL: %d řádků, DuckDB: %d řádků, rozdílů: %d)",
					pair.QueryName, pgComparison.RowsFile1, pgComparison.RowsFile2, len(pgComparison.DifferentRows))
				allComparison.DifferentQueries++
			}

			allComparison.Queries = append(allComparison.Queries, queryComp)
			queryNum++
		}
	}

	allComparison.TotalQueries = len(allComparison.Queries)

	return allComparison, nil
}

// PrintComparison vytiskne výsledky porovnání v čitelné formě
func PrintComparison(comparison *CSVComparison) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("Porovnání CSV souborů\n")
	fmt.Printf("Soubor 1: %s (%d řádků)\n", comparison.File1Path, comparison.RowsFile1)
	fmt.Printf("Soubor 2: %s (%d řádků)\n", comparison.File2Path, comparison.RowsFile2)
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("Sloupce se shodují: %v\n", comparison.ColumnsMatch)
	fmt.Printf("Počet řádků se shoduje: %v\n", comparison.RowCountMatch)
	fmt.Printf("Shodujících se řádků: %d\n", comparison.MatchingRows)
	fmt.Printf("Shoda: %.2f%%\n", comparison.MatchPercentage)

	if len(comparison.DifferentRows) > 0 {
		fmt.Printf("\n❌ Rozdílů: %d\n", len(comparison.DifferentRows))
		// Zobraz prvních 10 rozdílů
		displayCount := 10
		if len(comparison.DifferentRows) < displayCount {
			displayCount = len(comparison.DifferentRows)
		}
		for i := 0; i < displayCount; i++ {
			diff := comparison.DifferentRows[i]
			fmt.Printf("  Řádek %d, Sloupec %d (%s): '%s' vs '%s'\n",
				diff.RowNumber, diff.Column, diff.ColumnName, diff.File1Value, diff.File2Value)
		}
		if len(comparison.DifferentRows) > displayCount {
			fmt.Printf("  ... a %d dalších rozdílů\n", len(comparison.DifferentRows)-displayCount)
		}
	} else {
		fmt.Println("✅ Všechny řádky jsou identické!")
	}

	if len(comparison.ExtraInFile1) > 0 {
		fmt.Printf("\n⚠️  Extra řádků v souboru 1: %d\n", len(comparison.ExtraInFile1))
	}
	if len(comparison.ExtraInFile2) > 0 {
		fmt.Printf("\n⚠️  Extra řádků v souboru 2: %d\n", len(comparison.ExtraInFile2))
	}
}

// PrintAllQueriesComparison vytiskne výsledky všech query porovnání
func PrintAllQueriesComparison(allComparison *AllQueriesComparison) {
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Printf("POROVNÁNÍ VŠECH ULOŽENÝCH QUERIES\n")
	fmt.Printf("Složka: %s\n", allComparison.SavedQueriesPath)
	fmt.Printf("Celkem queries: %d\n", allComparison.TotalQueries)
	fmt.Printf("Shodující se: %d | Rozdílné: %d\n", allComparison.MatchingQueries, allComparison.DifferentQueries)
	fmt.Println(strings.Repeat("=", 100))

	for _, query := range allComparison.Queries {
		fmt.Println(query.Summary)
		if !query.AreSame {
			fmt.Printf("  PostgreSQL: %d řádků, DuckDB: %d řádků\n",
				query.PostgreSQL.RowsFile1, query.PostgreSQL.RowsFile2)
			if len(query.PostgreSQL.DifferentRows) > 0 {
				fmt.Printf("  Rozdílů v datech: %d\n", len(query.PostgreSQL.DifferentRows))
			}
		}
	}

	fmt.Println(strings.Repeat("=", 100))
	if allComparison.DifferentQueries == 0 {
		fmt.Println("✅ VŠECHNY QUERIES JSOU IDENTICKÉ!")
	} else {
		fmt.Printf("⚠️  %d queries má rozdílné výsledky\n", allComparison.DifferentQueries)
	}
}

// ExportComparisonToCSV exportuje výsledky porovnání do CSV souboru
func ExportComparisonToCSV(comparison *CSVComparison, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	header := []string{"RowNumber", "Column", "ColumnName", "File1Value", "File2Value"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Differences
	for _, diff := range comparison.DifferentRows {
		row := []string{
			strconv.Itoa(diff.RowNumber),
			strconv.Itoa(diff.Column),
			diff.ColumnName,
			diff.File1Value,
			diff.File2Value,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

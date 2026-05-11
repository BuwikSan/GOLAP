package backend

import (
	"database/sql"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type DuckDBOverhead struct {
	Path    string
	BinPath string // cesta k duckdb.exe
}

func (d *DuckDBOverhead) RunRaw(query string) error {
	cmd := exec.Command(d.BinPath, d.Path, "-c", query)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("DuckDB error: %v, output: %s", err, string(output))
	}
	return nil
}

// Exec spustí SQL příkaz, který nevrací řádky (INSERT, CREATE, atd.)
func (db *DuckDBOverhead) Exec(query string) (string, error) {
	cmd := exec.Command(db.BinPath, db.Path, "-c", query)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("DuckDB error: %v, output: %s", err, string(output))
	}
	return "", nil
}

// QueryRow spustí dotaz a vrátí výsledek jako string (dobré pro benchmarky)
func (db *DuckDBOverhead) QueryRow(query string) (string, string, error) {
	// -noheader a -list zajistí, že dostaneme jen čistý výsledek
	cmd := exec.Command(db.BinPath, db.Path, "-noheader", "-list", "-c", query)
	start := time.Now()
	output, err := cmd.Output()
	duration := time.Since(start).String()
	if err != nil {
		return "", "", err
	}
	return string(output), duration, nil
}

func (db *DuckDBOverhead) GetSource() (*sql.DB, error) {
	return nil, fmt.Errorf("DuckDB uses CLI only, can't return SQL connection")
}

func (db *DuckDBOverhead) String() string {
	return "duckdb"
}

type PostgresOverhead struct {
	*sql.DB
}

func (p *PostgresOverhead) RunRaw(query string) error {
	_, err := p.DB.Exec(query) // Tady se "vstřebají" ty ...any a sql.Result
	return err
}

func (p *PostgresOverhead) Exec(query string) (string, error) {
	_, err := p.DB.Exec(query)
	if err != nil {
		return "", err
	}
	return "", nil
}

func (p *PostgresOverhead) QueryRow(query string) (string, string, error) {
	start := time.Now()
	rows, err := p.DB.Query(query)
	duration := time.Since(start).String()
	if err != nil {
		return "", duration, err
	}
	defer rows.Close()

	// Get column names
	cols, err := rows.Columns()
	if err != nil {
		return "", duration, err
	}

	// Pre-allocate slices once outside the loop
	numCols := len(cols)
	values := make([]interface{}, numCols)
	valuePtrs := make([]interface{}, numCols)
	for i := range cols {
		valuePtrs[i] = &values[i]
	}

	// Use strings.Builder for efficient string concatenation
	var sb strings.Builder
	sb.Grow(1024) // Pre-allocate reasonable buffer

	for rows.Next() {
		// Scan the row
		if err := rows.Scan(valuePtrs...); err != nil {
			return "", duration, err
		}

		// Format row as tab-separated values
		for i, val := range values {
			if i > 0 {
				sb.WriteByte('\t')
			}
			if val == nil {
				sb.WriteString("NULL")
			} else {
				fmt.Fprintf(&sb, "%v", val)
			}
		}
		sb.WriteByte('\n')
	}

	return sb.String(), duration, rows.Err()
}

func (p *PostgresOverhead) GetSource() (*sql.DB, error) {
	return p.DB, nil
}

func (p *PostgresOverhead) String() string {
	return "postgres"
}

type SQLOverhead interface {
	// Upravíme to tak, aby to vracelo jen error pro jednoduchost

	RunRaw(query string) error

	Exec(query string) (string, error) // output by měl být něco jako (result, error) kde result je nějakej shluk info o querry, výsledku a novém stavu db

	QueryRow(query string) (string, string, error)

	GetSource() (*sql.DB, error)

	String() string
}

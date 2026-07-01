package spreadsheet

import (
	"fmt"
	"strconv"

	"github.com/xuri/excelize/v2"
)

// RowRecord represents a uniform data line parsed from an excel ledger example
type RowRecord struct {
	AccountID   string
	ContactName string
	BalanceDue  float64
}

// ParseLedgerSheet opens an Excel file and extracts target accounting data rows
func ParseLedgerSheet(filePath, sheetName string) ([]RowRecord, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet rows: %w", err)
	}

	var records []RowRecord

	// Skip row 0 assuming it is the header row (Account, Name, Balance)
	for i, row := range rows {
		if i == 0 || len(row) < 3 {
			continue
		}

		balance, _ := strconv.ParseFloat(row[2], 64)

		records = append(records, RowRecord{
			AccountID:   row[0],
			ContactName: row[1],
			BalanceDue:  balance,
		})
	}

	return records, nil
}

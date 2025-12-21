package compressState

import (
	"fmt"
	"log"

	"github.com/xuri/excelize/v2"
)

func Excel() {
	rows := [][]string{
		{"Name", "Role", "City"},
		{"Jeremy", "Dev", "St. John's"},
		{"Alice", "Analyst", "Toronto"},
	}

	f := excelize.NewFile()
	sheet := "Sheet1"

	// Write rows to cells
	for r, row := range rows {
		for c, val := range row {
			cell, err := excelize.CoordinatesToCellName(c+1, r+1) // 1-based
			if err != nil {
				log.Fatal(err)
			}
			if err := f.SetCellValue(sheet, cell, val); err != nil {
				log.Fatal(err)
			}
		}
	}

	// Optional: make columns a bit wider
	if err := f.SetColWidth(sheet, "A", "C", 18); err != nil {
		log.Fatal(err)
	}

	if err := f.SaveAs("output.xlsx"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Wrote output.xlsx")
}

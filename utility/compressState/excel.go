package compressState

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

func IntsToString(nums []int) string {
	strs := make([]string, len(nums))
	for i, n := range nums {
		strs[i] = strconv.Itoa(n)
	}
	return strings.Join(strs, ",")
}

func Excel(lineups [][2][]int, answers map[string]string) {

	// get unique teams in a map
	maxLength := 0
	teamMap := make(map[string]bool)
	for i := 0; i < len(lineups); i++ {
		if len(lineups[i][0]) > maxLength {
			maxLength = len(lineups[i][0])
		}
		if len(lineups[i][1]) > maxLength {
			maxLength = len(lineups[i][1])
		}

		stringA := IntsToString(lineups[i][0])
		stringB := IntsToString(lineups[i][1])
		if _, ok := teamMap[stringA]; !ok {
			teamMap[stringA] = true
		}
		if _, ok := teamMap[stringB]; !ok {
			teamMap[stringB] = true
		}
	}

	// put teams into an array of array strings, 0th is 1 size, 1st is 2 size ...
	teamSizes := make([][]string, maxLength)
	for i := 0; i < maxLength; i++ {
		teamSizes[i] = []string{}
	}
	for key := range teamMap {
		// split key string by , to get count
		parts := strings.Split(key, ",")
		count := len(parts)

		// push key to index count - 1
		teamSizes[count-1] = append(teamSizes[count-1], key)
	}

	// sort each array of array ints
	for i := 0; i < len(teamSizes); i++ {
		sort.Strings(teamSizes[i])
	}

	f := excelize.NewFile()
	sheet := "Sheet1"

	y := 1

	// convert to alphabet notation and write the table headers
	for i, j := 0, len(teamSizes)-1; i <= j; i, j = i+1, j-1 {

		// draw column headers
		for k := 0; k < len(teamSizes[i]); k++ {
			alphanumeric := CompressPattern(teamSizes[i][k])
			cell, _ := excelize.CoordinatesToCellName(1, k+1+y) // column A, row k+1+y
			f.SetCellValue(sheet, cell, alphanumeric)
		}

		// draw row headers
		for k := 0; k < len(teamSizes[j]); k++ {
			alphanumeric := CompressPattern(teamSizes[j][k])
			cell, _ := excelize.CoordinatesToCellName(k+2, y) // column k+2, row y
			f.SetCellValue(sheet, cell, alphanumeric)
		}

		// draw cells
		for a := 0; a < len(teamSizes[i]); a++ {
			for b := 0; b < len(teamSizes[j]); b++ {
				alphanumeric1 := CompressPattern(teamSizes[i][a])
				alphanumeric2 := CompressPattern(teamSizes[j][b])
				cell, _ := excelize.CoordinatesToCellName(b+2, a+1+y)
				key := alphanumeric1 + ";" + alphanumeric2
				val, ok := answers[key]
				if ok {
					f.SetCellValue(sheet, cell, val)
				} else {
					fmt.Println("key not found", key)
				}
			}
		}

		y += len(teamSizes[i]) + 2

	}

	f.SetColWidth(sheet, "A", "Z", 18)

	if err := f.SaveAs("output.xlsx"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Wrote output.xlsx")

}

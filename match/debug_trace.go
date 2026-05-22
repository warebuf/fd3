package match

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/warebuf/fd3/engine/state"
)

type TraceEntry struct {
	Log   any              `json:"log"`
	State *state.GameState `json:"state"`
}

func (match *Match) WriteTrace() {
	if err := os.MkdirAll("traces", 0755); err != nil {
		fmt.Println("WriteTrace: mkdir error:", err)
		return
	}
	data, err := json.MarshalIndent(match.DebugTrace, "", "  ")
	if err != nil {
		fmt.Println("WriteTrace: marshal error:", err)
		return
	}
	filename := fmt.Sprintf("traces/%s.json", match.ID)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Println("WriteTrace: write error:", err)
	}
}

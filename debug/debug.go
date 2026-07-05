package debug

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

func WriteTrace(id string, trace []TraceEntry) {
	if err := os.MkdirAll("traces", 0755); err != nil {
		fmt.Println("WriteTrace: mkdir error:", err)
		return
	}
	data, err := json.MarshalIndent(trace, "", "  ")
	if err != nil {
		fmt.Println("WriteTrace: marshal error:", err)
		return
	}
	filename := fmt.Sprintf("traces/%s.json", id)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Println("WriteTrace: write error:", err)
	}
}

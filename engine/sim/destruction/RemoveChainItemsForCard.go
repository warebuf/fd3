package destruction

import (
	"github.com/warebuf/fd3/engine/logs"
	"github.com/warebuf/fd3/engine/state"
)

// removes any chain items sourced from cardIndex, e.g. after it's destroyed
func RemoveChainItemsForCard(gs *state.GameState, logList *[]any, UpdateID *int, cardIndex int, executeAtWallTime int64) {
	filtered := gs.Chain[:0]
	for _, item := range gs.Chain {
		if item.CardIndex != cardIndex {
			filtered = append(filtered, item)
			continue
		}
		*logList = append(*logList, logs.PurgeChainLog{
			UpdateID:      *UpdateID,
			WallTimestamp: executeAtWallTime,
			Type:          logs.PurgeChain,
			CardIndex:     cardIndex,
		})
		*UpdateID++
	}
	gs.Chain = filtered
}

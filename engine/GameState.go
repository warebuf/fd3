package engine

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/warebuf/fd3/utility/shallowCopy"
)

type GameState struct {
	// time the state is set for
	GameStateID int
	Timestamp   time.Time
	Round       int

	// team sets: index (unitID) ->  team index
	Team []int
	Zone []int // 0: Bench, 1: Field, 2: Dead

	// model set
	ModelNo []int

	// core unit stats: index (unitID) ->  HP
	HP []int

	// abilities
	Abilities       []int   // abilities are flattened for each unit
	AbilityStartIdx []int   // the starting index in Abilities[] for this unit
	AbilityCount    []int   // how many abilities belong to each unit
	AbilityCharge   []int64 // flattened, saved in millisconds
	AbilitySPD      []int64 // flattened, in milliseconds

	SummonCondition []SummonType
}

type Telemetry struct {
	TimesAttacked  []int64 // index (unitID) -> val
	TimesDefended  []int64
	DamageDealt    []int64
	DamageReceived []int64
}

// CreateGameState initializes an empty game state with outer slices ready
func CreateFirstGameState(t time.Time, GSID int) *GameState {
	gm := &GameState{
		GameStateID:     GSID,
		Timestamp:       t,
		Round:           0,
		Team:            make([]int, 0),
		Zone:            make([]int, 0),
		ModelNo:         make([]int, 0),
		HP:              make([]int, 0),
		Abilities:       make([]int, 0),
		AbilityStartIdx: make([]int, 0),
		AbilityCount:    make([]int, 0),
		AbilityCharge:   make([]int64, 0),
		AbilitySPD:      make([]int64, 0),
		SummonCondition: make([]SummonType, 0),
	}
	return gm
}

func CopyGameState(timestamp time.Time, prevState *GameState) *GameState {
	if prevState == nil {
		return nil
	}

	newState := &GameState{
		GameStateID: prevState.GameStateID + 1, // increment ID for new state
		Timestamp:   timestamp,
		Round:       prevState.Round,
	}

	newState.Team = shallowCopy.IntSlice(prevState.Team)
	newState.Zone = shallowCopy.IntSlice(prevState.Zone)
	newState.ModelNo = shallowCopy.IntSlice(prevState.ModelNo)
	newState.HP = shallowCopy.IntSlice(prevState.HP)
	newState.Abilities = shallowCopy.IntSlice(prevState.Abilities)
	newState.AbilityStartIdx = shallowCopy.IntSlice(prevState.AbilityStartIdx)
	newState.AbilityCount = shallowCopy.IntSlice(prevState.AbilityCount)
	newState.AbilityCharge = shallowCopy.Int64Slice(prevState.AbilityCharge)
	newState.AbilitySPD = shallowCopy.Int64Slice(prevState.AbilitySPD)
	newState.SummonCondition = CopySummonTypeSlice(prevState.SummonCondition)

	return newState
}

func PrintState(state *GameState) {
	for i := 0; i < len(state.HP); i++ {
		if state.Team[i] == 0 {
			fmt.Printf(strconv.Itoa(state.ModelNo[i]) + ":" + strconv.Itoa(state.HP[i]) + " ")
		}
	}
	fmt.Println("")
	for i := 0; i < len(state.HP); i++ {
		if state.Team[i] == 1 {
			fmt.Printf(strconv.Itoa(state.ModelNo[i]) + ":" + strconv.Itoa(state.HP[i]) + " ")
		}
	}
	fmt.Println("")
}

// each unit retains order, simply save HP
func ConvertStateToHPKey(state *GameState) string {

	teamStrings := []string{}

	i := 0
	for {
		currentString := ""
		for j := 0; j < len(state.HP); j++ {
			if state.Team[j] == i {
				currentString += strconv.Itoa(state.ModelNo[j])
				currentString += ":"
				currentString += strconv.Itoa(state.HP[j])
				currentString += " "
			}
		}

		if len(currentString) == 0 {
			break
		}

		teamStrings = append(teamStrings, currentString)
		i += 1

	}

	return strings.Join(teamStrings, ";")
}

// sort by ID
func ConvertStateToModelNoKey(state *GameState) string {

	teamStrings := []string{}

	i := 0
	found := false
	for {
		units := []int{}
		for j := 0; j < len(state.ModelNo); j++ {
			if state.Team[j] == i {
				found = true
				if state.HP[j] > 0 {
					units = append(units, state.ModelNo[j])
				}
			}
		}

		if !found {
			break
		}

		// sort
		sort.Ints(units)

		// convert to strings
		unitStrings := make([]string, len(units))
		for a, b := range units {
			unitStrings[a] = strconv.Itoa(b)
		}

		teamStrings = append(teamStrings, strings.Join(unitStrings, ","))
		i += 1
		found = false
	}

	return strings.Join(teamStrings, ";")
}

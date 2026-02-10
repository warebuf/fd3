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

	// additional info on the state
	Round int

	// structure data
	StructureHP   []int
	StructureTeam []int

	// team sets: index (unitID) ->  team index
	Team []int
	Zone []int // 0: Bench, 1: Field, 2: Dead, 3: QueuedSummon

	// model set
	ModelNo []int

	// core unit stats: index (unitID) ->  HP
	HP  []int
	EXP []int

	// functions
	Functions        []int // abilities are flattened for each unit
	FunctionType     []int // 0 is auto, 1 is manual, 2 is passive
	FunctionStartIdx []int // the starting index in Functions[] for this unit
	FunctionCount    []int // how many functions belong to each unit
	FunctionCharge   []int // flattened, saved in millisconds
	FunctionSPD      []int // flattened, in milliseconds
	FunctionTarget   []int // on init set to -1, each function gets 2 fixed elements, can be 1-6

}

// CreateGameState initializes an empty game state with outer slices ready
func CreateFirstGameState(t time.Time, GSID int) *GameState {
	gm := &GameState{
		GameStateID:      GSID,
		Timestamp:        t,
		Round:            0,
		StructureHP:      make([]int, 0),
		StructureTeam:    make([]int, 0),
		Team:             make([]int, 0),
		Zone:             make([]int, 0),
		ModelNo:          make([]int, 0),
		HP:               make([]int, 0),
		EXP:              make([]int, 0),
		Functions:        make([]int, 0),
		FunctionType:     make([]int, 0),
		FunctionStartIdx: make([]int, 0),
		FunctionCount:    make([]int, 0),
		FunctionCharge:   make([]int, 0),
		FunctionSPD:      make([]int, 0),
		FunctionTarget:   make([]int, 0),
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

	newState.StructureHP = shallowCopy.IntSlice(prevState.StructureHP)
	newState.StructureTeam = shallowCopy.IntSlice(prevState.StructureTeam)

	newState.Team = shallowCopy.IntSlice(prevState.Team)
	newState.Zone = shallowCopy.IntSlice(prevState.Zone)
	newState.ModelNo = shallowCopy.IntSlice(prevState.ModelNo)
	newState.HP = shallowCopy.IntSlice(prevState.HP)
	newState.EXP = shallowCopy.IntSlice(prevState.EXP)
	newState.Functions = shallowCopy.IntSlice(prevState.Functions)
	newState.FunctionType = shallowCopy.IntSlice(prevState.FunctionType)
	newState.FunctionStartIdx = shallowCopy.IntSlice(prevState.FunctionStartIdx)
	newState.FunctionCount = shallowCopy.IntSlice(prevState.FunctionCount)
	newState.FunctionCharge = shallowCopy.IntSlice(prevState.FunctionCharge)
	newState.FunctionSPD = shallowCopy.IntSlice(prevState.FunctionSPD)
	newState.FunctionTarget = shallowCopy.IntSlice(prevState.FunctionTarget)

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

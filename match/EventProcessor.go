package match

// so we have to balance between client inputs and autoevents

// every client input allows us to predict the next n+1 autoevent

// 200ms before every autoevent, we are guaranteed for that autoevent/state to happen
// so 200ms before every autoevent, we are valid to calculate the next autoevent

// we can work out all cases to write the EventSequencer / EventScheduler

// we can have an Event data structure, with inputs, autoevents, and autoschedulers
// we can also have a GameState data structure

// maybe we put the InputEventChan and the AutoScheduler into a priority queue sorted by timestamp
// chatgpt is telling me two goroutines doesnt really work because they block each other because of the mutex on the gamestate
// which makes sense i think

// should have one loop that is EventScheduler - a priority queue that has a timer, pushes into a channel
// another loop that is EventProcessor - reads the channel and does the behavior

func (match *Match) EventProcessor() {

	for {
	}

}

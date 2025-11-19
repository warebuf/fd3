package match

// so we have to balance between client inputs and autoevents

// every client input can calculate the next n+1 autoevent

// 200ms before every autoevent, we are guaranteed for that autoevent/state to happen
// so 200ms before every autoevent, we are valid to calculate the next autoevent

// we can work out all cases to write the EventSequencer / EventScheduler


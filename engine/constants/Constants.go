package constants

const LEAD = 200             // milliseconds, we schedule all inputs 200 milliseconds into the future from when the server receives it
const REACTION_WINDOW = 5000 // milliseconds, how long the reaction window is
const SIM_SPEED = 5          // multiplier applied to SimulationPQ-scheduled durations (OnTick, mana regen); does not affect ReactionStack windows
const PAUSE_BANK = 300000    // milliseconds, how much pause time per player
const START_MANA = 300
const BOARD_SETUP_DURATION = 120000 // milliseconds, how long players have to stage

package constants

const LEAD = 200           // milliseconds, we schedule all inputs 200 milliseconds into the future from when the server receives it
const CHAIN_WINDOW = 30000 // milliseconds, how long the chain window is
const SIM_SPEED = 5        // multiplier applied to simpq durations
const PAUSE_BANK = 300000  // milliseconds, how much pause time per player
const START_MANA = 300
const DRAWS_COMPLETE_AT = 3000 // ms after draw phase starts; matches the last staggered ScheduleSystemDraw call
const STAGING_WINDOW = 120000  // ms after draws finish during which players may stage cards before Play begins
const TRAP_SET_TIME = 20000    // ms a trap spends setting before it's ready; scaled by SIM_SPEED like OnTick durations

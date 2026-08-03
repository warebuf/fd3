package constants

const LEAD = 200             // milliseconds, we schedule all inputs 200 milliseconds into the future from when the server receives it
const REACTION_WINDOW = 5000 // milliseconds, how long the reaction window is
const SIM_SPEED = 5          // multiplier applied to simpq durations
const PAUSE_BANK = 300000    // milliseconds, how much pause time per player
const START_MANA = 300
const DRAWS_COMPLETE_AT = 3000 // ms after draw phase starts; matches the last staggered ScheduleSystemDraw call
const STAGING_WINDOW = 120000  // ms after draws finish during which players may stage cards before Play begins

package defs

import (
	"math"

	"github.com/warebuf/fd3/engine/state"
)

const (
	TargetSelf         = 1 << iota // the card itself
	TargetOwnField                 // friendly field cards (excluding self)
	TargetOppField                 // opponent field cards
	TargetOwnHand                  // friendly hand cards
	TargetOppHand                  // opponent hand cards
	TargetOwnCont                  // friendly continuous zone
	TargetOppCont                  // opponent continuous zone
	TargetOwnTrap                  // friendly set traps
	TargetOppTrap                  // opponent set traps
	TargetOppStructure             // opponent structures, targeted by slot rather than by card index
)

// which zone(s) the owning card must be in for OnTrigger
const (
	TriggerZoneField = 1 << iota
	TriggerZoneContinuous
	TriggerZoneGraveyard
	TriggerZoneHand
)

type Ability struct {
	Effects      [3]state.Effect
	Cons         [3]state.ConditionType
	Cost         int
	Unlimited    bool              // if true is always available, and ignores Charges/Pending
	Milliseconds int64             // only used by OnTick
	TargetFlags  int               // bitmask of Target flags
	Trigger      state.TriggerType // only used by OnTrigger, condition we do a trigger check
	TriggerZones int               // only used by OnTrigger, bitmask of TriggerZone flags
	CanAutoFire  bool              // only used by Active
	ChainEntry   state.ChainEntry  // used by Active, OnTick and OnTrigger, the slots that push chain items
	SkipChain    bool              // only used by OnTick
}

// change from dynamic arrs to fixed size
type CardDef struct {
	Type        int
	Name        string
	Description string

	Archetype int
	Taunt     bool
	Bodyguard bool

	ChainEntry state.ChainEntry
	IsToken    bool // if set true, does not go to grave
	IsAuction  bool // if set true, this trap's Active[0] opens a Standoff-style blind mana auction instead of resolving normally; see match.handleBid

	OnPlay    [3]Ability
	OnTick    [3]Ability
	Active    [3]Ability
	OnTrigger [3]Ability

	EquipTargetFlags     int // for equipment cards, bitmask of Target flags
	AttackBonusPerSummon int // for equipment cards: added to the host's attack damage per unit its own controller has ever summoned (still on the field, or already in the graveyard)

	Charges    int // starting charges; shared pool any ability slot on this card may gain or spend from
	MaxCharges int

	BaseHP int
	Cost   int
}

var CardDefs = []CardDef{
	0: {}, // empty, so DefNo lines up with index
	1: {
		Type:        state.CardTypeUnit,
		Name:        "Fallen Soldier",
		Description: "Deal 1 random damage every 30 ticks to an enemy",
		BaseHP:      2,
		Cost:        60,
		ChainEntry:  state.ChainAction,
		OnTick: [3]Ability{{
			ChainEntry:   state.ChainAction,
			Milliseconds: 30000,
			Effects: [3]state.Effect{
				{
					Type: state.EffectAttack,
					A:    0, // targetType, 0: random, 1: user-targetted, 2: prioritize archetype
					B:    1, // amount
				},
			},
		}},
	},
	2: {
		Type:        state.CardTypeTrap,
		Name:        "Year Zero",
		Description: "Destroy all units on the field",
		Cost:        100,
		Active: [3]Ability{{
			ChainEntry: state.ChainDual,
			Effects: [3]state.Effect{
				{
					Type: state.EffectDestroy,
					A:    0, // Destroy all units
				},
			},
		}},
	},
	3: {
		Type:        state.CardTypeUnit,
		Archetype:   state.ArchetypeTree,
		Name:        "Money Tree",
		Description: "Gain 10 mana every 10 ticks",
		BaseHP:      1,
		Cost:        60,
		ChainEntry:  state.ChainAction,
		OnTick: [3]Ability{{
			SkipChain:    true,
			Milliseconds: 10000,
			Effects: [3]state.Effect{
				{
					Type: state.EffectMana, // fill A, which is team
					B:    10,               // amount
				},
			},
		}},
	},
	4: {
		Type:        state.CardTypeTrap,
		Name:        "Downforce",
		Description: "Gain 30 mana",
		Cost:        0,
		Active: [3]Ability{{
			ChainEntry: state.ChainDual,
			Effects: [3]state.Effect{
				{
					Type: state.EffectMana, // fill A, which is team
					B:    30,               // amount
				},
			},
		}},
	},
	5: {
		Type:        state.CardTypeUnit,
		Name:        "Barkbreaker",
		Description: "Prioritizes and double-damages 'Tree' archetype. Deal 1 damage every 20 ticks to a random",
		BaseHP:      2,
		Cost:        100,
		ChainEntry:  state.ChainAction,
		OnTick: [3]Ability{{
			ChainEntry:   state.ChainAction,
			Milliseconds: 20000,
			Effects: [3]state.Effect{
				{
					Type: state.EffectAttack,
					A:    2,                   // targetType, 0: random, 1: user-targetted, 2: prioritize archetype
					B:    1,                   // base damage
					C:    state.ArchetypeTree, // priority archetype
				},
			},
		}},
	},
	6: {
		Type:        state.CardTypeUnit,
		Name:        "Metal Heart",
		Description: "Taunt",
		BaseHP:      6,
		Cost:        80,
		ChainEntry:  state.ChainAction,
		Taunt:       true,
	},
	7: {
		Type:        state.CardTypeUnit,
		Name:        "Tomoe",
		Description: "Heal 1 HP every 30 ticks to a random friendly unit",
		BaseHP:      2,
		Cost:        60,
		ChainEntry:  state.ChainAction,
		OnTick: [3]Ability{{
			ChainEntry:   state.ChainAction,
			Milliseconds: 30000,
			Effects: [3]state.Effect{
				{
					Type: state.EffectHeal,
					B:    1, // amount
				},
			},
		}},
	},
	8: {
		Type:        state.CardTypeTrap,
		Name:        "Countertrap",
		Description: "Negate a summon or ability and destroy the source card.",
		Cost:        30,
		Active: [3]Ability{{
			ChainEntry: state.ChainReaction,
			Effects: [3]state.Effect{
				{Type: state.EffectCounter},
			},
		}},
	},
	9: {
		Type:        state.CardTypeContinuous,
		Name:        "Rolling in the Deep",
		Description: "Use (50): Draw a card. 3 uses.",
		Cost:        80,
		ChainEntry:  state.ChainDual,
		Charges:     3,
		MaxCharges:  math.MaxInt,
		Active: [3]Ability{{
			Cost:        50,
			Cons:        [3]state.ConditionType{state.ConditionManaAtLeast},
			CanAutoFire: true,
			ChainEntry:  state.ChainDual,
			Effects:     [3]state.Effect{{Type: state.EffectDraw, B: 1}},
		}},
	},
	10: {
		Type:        state.CardTypeUnit,
		Name:        "Archangel",
		Description: "Maximum Charges: 3. Gain 1 charge every 30 ticks. Spend a charge to deal 1 damage to a target.",
		BaseHP:      2,
		Cost:        70,
		ChainEntry:  state.ChainAction,
		MaxCharges:  3,
		OnTick: [3]Ability{{
			SkipChain:    true,
			Milliseconds: 30000,
			Effects: [3]state.Effect{
				{Type: state.EffectChargeSelf},
			},
		}},
		Active: [3]Ability{{
			TargetFlags: TargetSelf | TargetOwnField | TargetOppField,
			Cons:        [3]state.ConditionType{state.ConditionChargeAtLeastOne},
			CanAutoFire: true,
			ChainEntry:  state.ChainDual,
			Effects: [3]state.Effect{
				{
					Type: state.EffectAttack,
					A:    1, // targetType, 0: random, 1: user-targetted, 2: prioritize archetype, 3: always struct
					B:    1, // amount
				},
			},
		}},
	},
	11: {
		Type:        state.CardTypeUnit,
		Name:        "Forsaken",
		Description: "Use: Sacrifice a friendly unit you choose. Deal damage directly to the enemy structure equal to 1/10th its mana cost.",
		BaseHP:      3,
		Cost:        60,
		ChainEntry:  state.ChainAction,
		Active: [3]Ability{{
			Unlimited:   true,
			TargetFlags: TargetSelf | TargetOwnField,
			ChainEntry:  state.ChainDual,
			Effects: [3]state.Effect{
				{Type: state.EffectSacrifice},
				{Type: state.EffectAttack, A: 3}, // A: 3, always hits enemy structure; B is filled in by the sacrifice at runtime
			},
		}},
	},
	12: {
		Type:        state.CardTypeContinuous,
		Name:        "Charge Reservoir",
		Description: "Gain 1 charge whenever one of your traps resolves.",
		Cost:        60,
		ChainEntry:  state.ChainDual,
		MaxCharges:  math.MaxInt,
		OnTrigger: [3]Ability{{
			Trigger:      state.TriggerAllyTrapResolve,
			TriggerZones: TriggerZoneContinuous,
			ChainEntry:   state.ChainDual,
			Effects:      [3]state.Effect{{Type: state.EffectChargeSelf}},
		}},
	},
	13: {
		Type:        state.CardTypeUnit,
		Name:        "Charge Consumer",
		Description: "Summon: Destroy a friendly card you choose; gain its charges as HP. Use: Deal 1 damage to an enemy, plus those charges, then this unit is destroyed.",
		BaseHP:      1,
		Cost:        60,
		ChainEntry:  state.ChainAction,
		Charges:     1,
		MaxCharges:  1,
		OnPlay: [3]Ability{{
			TargetFlags: TargetOwnField | TargetOwnCont,
			Effects:     [3]state.Effect{{Type: state.EffectConsume, C: 0}}, // C: 0, feeds Active slot 0's Bonus
		}},
		Active: [3]Ability{{
			TargetFlags: TargetOppField,
			Cons:        [3]state.ConditionType{state.ConditionChargeAtLeastOne},
			CanAutoFire: true,
			ChainEntry:  state.ChainDual,
			Effects: [3]state.Effect{
				{Type: state.EffectAttack, A: 1, B: 1}, // A: 1, user-targeted; base 1 damage, plus Bonus
				{Type: state.EffectSelfDestruct},
			},
		}},
	},
	14: {
		Type:        state.CardTypeUnit,
		Name:        "Charge Token",
		Description: "Destroyed after 10 ticks.",
		BaseHP:      1,
		IsToken:     true,
		ChainEntry:  state.ChainAction,
		OnTick: [3]Ability{{
			SkipChain:    true,
			Milliseconds: 10000,
			Effects:      [3]state.Effect{{Type: state.EffectSelfDestruct}},
		}},
	},
	15: {
		Type:        state.CardTypeTrap,
		Name:        "Temp Tokens",
		Description: "Destroy a friendly Charge Reservoir you choose; for every 2 charges it had, create a token with 1 HP that lasts 10 ticks.",
		Cost:        60,
		Active: [3]Ability{{
			ChainEntry:  state.ChainDual,
			TargetFlags: TargetOwnField | TargetOwnCont,
			Effects:     [3]state.Effect{{Type: state.EffectConvertToTokens, B: 14}}, // B: 14, Charge Token's DefNo
		}},
	},
	16: {
		Type:        state.CardTypeTrap,
		Name:        "Kindling",
		Description: "If none of your cards have 2+ charges, give a card 2 charges.",
		Cost:        50,
		Active: [3]Ability{{
			ChainEntry:  state.ChainDual,
			TargetFlags: TargetOwnField | TargetOppField | TargetOwnCont | TargetOppCont,
			Effects:     [3]state.Effect{{Type: state.EffectChargeTarget, B: 2}},
		}},
	},
	17: {
		Type:        state.CardTypeTrap,
		Name:        "Rummage",
		Description: "Draw a card.",
		Cost:        40,
		Active: [3]Ability{{
			ChainEntry: state.ChainDual,
			Effects:    [3]state.Effect{{Type: state.EffectDraw, B: 1}},
		}},
	},
	18: {
		Type:             state.CardTypeEquip,
		Name:             "Deathless",
		Description:      "If the equipped unit would be destroyed, negate the damage and destroy this instead.",
		Cost:             40,
		ChainEntry:       state.ChainDual,
		EquipTargetFlags: TargetOwnField | TargetOppField,
		OnTrigger: [3]Ability{{
			Trigger: state.TriggerHostWouldBeDestroyed,
			Effects: [3]state.Effect{{Type: state.EffectAbsorbDestruction}},
		}},
	},
	19: {
		Type:        state.CardTypeTrap,
		Name:        "Donation",
		Description: "Give a friendly unit you choose to the opponent.",
		Cost:        0,
		Active: [3]Ability{{
			ChainEntry:  state.ChainDual,
			TargetFlags: TargetOwnField,
			Effects:     [3]state.Effect{{Type: state.EffectGiveControl}},
		}},
	},
	20: {
		Type:        state.CardTypeTrap,
		Name:        "Sanctuary",
		Description: "Give 3 HP to every unit on the field that has no damaging ability.",
		Cost:        50,
		Active: [3]Ability{{
			ChainEntry: state.ChainDual,
			Effects:    [3]state.Effect{{Type: state.EffectHealNonDamaging, B: 3}},
		}},
	},
	21: {
		Type:        state.CardTypeUnit,
		Name:        "Extortionist",
		Description: "While on the field, whenever the opponent activates a card, ability, trap, or structure draw, they pay you 10 mana if they can; otherwise you gain 10 mana anyway.",
		BaseHP:      2,
		Cost:        80,
		ChainEntry:  state.ChainAction,
		OnTrigger: [3]Ability{{
			Trigger:      state.TriggerEnemyPaidMana,
			TriggerZones: TriggerZoneField,
			ChainEntry:   state.ChainDual,
			Effects:      [3]state.Effect{{Type: state.EffectExtort, B: 10}},
		}},
	},
	22: {
		Type:        state.CardTypeTrap,
		Name:        "Repossession",
		Description: "Take control of a unit you originally owned. It immediately activates one of its abilities for free.",
		Cost:        60,
		Active: [3]Ability{{
			ChainEntry:  state.ChainDual,
			TargetFlags: TargetOppField,
			Cons:        [3]state.ConditionType{state.ConditionTargetOwnedBySelf},
			Effects:     [3]state.Effect{{Type: state.EffectRepossess}},
		}},
	},
	23: {
		Type:        state.CardTypeUnit,
		Name:        "Volatile Core",
		Description: "When this unit is destroyed, deal 5 damage to all allied units.",
		BaseHP:      3,
		Cost:        70,
		ChainEntry:  state.ChainAction,
		OnTrigger: [3]Ability{{
			Trigger:      state.TriggerSelfDestroyed,
			TriggerZones: TriggerZoneGraveyard,
			Effects:      [3]state.Effect{{Type: state.EffectDamageAllAllies, B: 5}},
		}},
	},
	24: {
		Type:        state.CardTypeUnit,
		Name:        "Mana Leech",
		Description: "Drains 50 mana from the opponent and gains 1 HP every 25 ticks.",
		BaseHP:      2,
		Cost:        80,
		ChainEntry:  state.ChainAction,
		OnTick: [3]Ability{{
			ChainEntry:   state.ChainAction,
			Milliseconds: 25000,
			Effects: [3]state.Effect{
				{Type: state.EffectDrainMana, B: 50},
				{Type: state.EffectHealSelf, B: 1},
			},
		}},
	},
	25: {
		Type:        state.CardTypeTrap,
		Name:        "Liquidator",
		Description: "Destroy one of your units. Convert its health, charges, and attachments into mana.",
		Cost:        0,
		Active: [3]Ability{{
			ChainEntry:  state.ChainDual,
			TargetFlags: TargetOwnField,
			Effects:     [3]state.Effect{{Type: state.EffectLiquidate}},
		}},
	},
	26: {
		Type:        state.CardTypeTrap,
		Name:        "Mutual Destruction",
		Description: "Each player chooses one of their own units. Exchange them, then both immediately attack.",
		Cost:        40,
		Active: [3]Ability{{
			ChainEntry:  state.ChainDual,
			TargetFlags: TargetOwnField,
			Effects:     [3]state.Effect{{Type: state.EffectMutualDestruction, B: 3}},
		}},
	},
	27: {
		Type:        state.CardTypeUnit,
		Name:        "Bodyguard",
		Description: "When any effect selects a target, this may become its target instead.",
		BaseHP:      4,
		Cost:        60,
		ChainEntry:  state.ChainAction,
		Bodyguard:   true,
	},
	28: {
		Type:        state.CardTypeTrap,
		Name:        "Standoff",
		Description: "Both players secretly bid mana. The higher bidder chooses which of two effects happens: gain 40 mana, or deal 8 damage to a random enemy structure. Both bids are spent either way; ties favor whoever set this trap.",
		Cost:        20,
		IsAuction:   true,
		Active: [3]Ability{{
			ChainEntry: state.ChainDual,
		}},
	},
	29: {
		Type:        state.CardTypeUnit,
		Name:        "Gravedigger",
		Description: "Use: For each card in your graveyard, deal 1 damage to a random enemy. Hits that land on a structure also drain that much mana.",
		BaseHP:      2,
		Cost:        70,
		ChainEntry:  state.ChainAction,
		Active: [3]Ability{{
			Unlimited:   true,
			Cost:        30,
			Cons:        [3]state.ConditionType{state.ConditionManaAtLeast},
			CanAutoFire: true,
			ChainEntry:  state.ChainDual,
			Effects:     [3]state.Effect{{Type: state.EffectGraveyardBarrage, B: 1}},
		}},
	},
	30: {
		Type:        state.CardTypeUnit,
		Name:        "Leech",
		Description: "Use: Steal 40 mana from the opponent and deal 3 damage to a random enemy.",
		BaseHP:      2,
		Cost:        70,
		ChainEntry:  state.ChainAction,
		Active: [3]Ability{{
			Unlimited:   true,
			Cost:        30,
			Cons:        [3]state.ConditionType{state.ConditionManaAtLeast},
			CanAutoFire: true,
			ChainEntry:  state.ChainDual,
			Effects: [3]state.Effect{
				{Type: state.EffectDrainMana, B: 40},
				{Type: state.EffectAttack, A: 0, B: 3},
			},
		}},
	},
	31: {
		Type:        state.CardTypeUnit,
		Name:        "Overwhelm",
		Description: "Use: Deal 3 damage to a random enemy for each unit the opponent has. Any excess from overkilling spills onto another enemy unit or structure.",
		BaseHP:      2,
		Cost:        70,
		ChainEntry:  state.ChainAction,
		Active: [3]Ability{{
			Unlimited:   true,
			Cost:        30,
			Cons:        [3]state.ConditionType{state.ConditionManaAtLeast},
			CanAutoFire: true,
			ChainEntry:  state.ChainDual,
			Effects:     [3]state.Effect{{Type: state.EffectScalingSpillover, B: 3}},
		}},
	},
	32: {
		Type:        state.CardTypeTrap,
		Name:        "Reinforcements",
		Description: "Summon two 1 HP tokens. They last 10 ticks.",
		Cost:        30,
		Active: [3]Ability{{
			ChainEntry: state.ChainDual,
			Effects:    [3]state.Effect{{Type: state.EffectSpawnTokens, B: 14, C: 2}}, // B: 14, Charge Token's DefNo
		}},
	},
	33: {
		Type:                 state.CardTypeEquip,
		Name:                 "Veteran's Standard",
		Description:          "Equipped unit deals 1 extra damage for each unit its controller has summoned.",
		Cost:                 40,
		ChainEntry:           state.ChainDual,
		EquipTargetFlags:     TargetOwnField | TargetOppField,
		AttackBonusPerSummon: 1,
	},
}

// IsDamaging reports whether this CardDef has an EffectAttack in any of its ability slots
func IsDamaging(defNo int) bool {
	def := CardDefs[defNo]
	for _, slots := range [][3]Ability{def.OnPlay, def.OnTick, def.Active, def.OnTrigger} {
		for _, ability := range slots {
			for _, eff := range ability.Effects {
				if eff.Type == state.EffectAttack {
					return true
				}
			}
		}
	}
	return false
}

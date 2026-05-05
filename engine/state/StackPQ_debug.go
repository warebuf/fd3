package state

import (
	"fmt"
	"strings"
	"time"
)

func (et EffectType) String() string {
	switch et {
	case EffectAttack:
		return "Attack"
	case EffectActivate:
		return "Activate"
	case EffectHeal:
		return "Heal"
	case EffectDraw:
		return "Draw"
	case EffectMana:
		return "Mana"
	case EffectSilence:
		return "Silence"
	case EffectGrantAbility:
		return "GrantAbility"
	case SummonResolution:
		return "SummonResolution"
	case EffectDestroy:
		return "Destroy"
	default:
		return fmt.Sprintf("Unknown(%d)", int(et))
	}
}

func chainStepTypeName(t int) string {
	switch t {
	case ChainStepCard:
		return "card"
	case ChainStepCardEff:
		return "card_eff"
	case ChainStepAbility:
		return "ability"
	case ChainStepOther:
		return "other"
	case ChainStepSummon:
		return "summon"
	case ChainStepSpell:
		return "spell"
	case ChainStepRecurring:
		return "recurring_ability"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

func fmtMs(ms int64) string {
	return fmt.Sprintf("%d", ms%100_000)
}

func DebugReactionPQ(pq *PriorityQueue) string {
	if pq == nil || len(pq.Items) == 0 {
		return "ReactionPQ: (empty)"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "ReactionPQ t=%d:\n", time.Now().UnixMilli()%100_000)

	for ci, chain := range pq.Items {
		fmt.Fprintf(&b, "  [%d] Chain %d: effectIndex=%d popTime=%s\n",
			ci, chain.ChainID, chain.EffectIndex, fmtMs(chain.PopTime))

		for si, step := range chain.ChainSteps {
			recurring := ""
			if step.RecurringPeriod > 0 {
				recurring = fmt.Sprintf(" recurring=%dms", step.RecurringPeriod)
			}
			fmt.Fprintf(&b, "    [%d] Step type=%s cardID=%d startTime=%s popTime=%s occurTime=%s%s\n",
				si, chainStepTypeName(step.Type), step.CardID, fmtMs(step.StartTime), fmtMs(step.PopTime), fmtMs(step.OccurTime), recurring)

			for fi, eff := range step.Effects {
				fmt.Fprintf(&b, "      [%d] Effect type=%s A=%d B=%d C=%d\n",
					fi, eff.Type, eff.A, eff.B, eff.C)
			}
		}
	}

	return b.String()
}

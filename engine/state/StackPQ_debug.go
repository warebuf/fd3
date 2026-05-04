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

func entryTypeName(t int) string {
	switch t {
	case 0:
		return "card"
	case 1:
		return "card_eff"
	case 2:
		return "ability"
	case 3:
		return "other"
	case 4:
		return "summon"
	case 5:
		return "spell"
	case 6:
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

	for si, stack := range pq.Items {
		fmt.Fprintf(&b, "  [%d] Stack %d: effectIndex=%d popTime=%s\n",
			si, stack.StackID, stack.EffectIndex, fmtMs(stack.PopTime))

		for ei, entry := range stack.StackEntries {
			recurring := ""
			if entry.RecurringPeriod > 0 {
				recurring = fmt.Sprintf(" recurring=%dms", entry.RecurringPeriod)
			}
			fmt.Fprintf(&b, "    [%d] Entry type=%s cardID=%d startTime=%s popTime=%s occurTime=%s%s\n",
				ei, entryTypeName(entry.Type), entry.CardID, fmtMs(entry.StartTime), fmtMs(entry.PopTime), fmtMs(entry.OccurTime), recurring)

			for fi, eff := range entry.Effects {
				fmt.Fprintf(&b, "      [%d] Effect type=%s A=%d B=%d C=%d\n",
					fi, eff.Type, eff.A, eff.B, eff.C)
			}
		}
	}

	return b.String()
}

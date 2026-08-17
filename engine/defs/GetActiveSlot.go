package defs

func GetActiveSlot(defNo int, requestedSlot int) int {
	actives := CardDefs[defNo].Active

	onlySlot := -1
	realSlots := 0
	for i, ability := range actives {
		if ability.Effects[0].Type == 0 {
			continue // unused ability slot
		}
		realSlots++
		onlySlot = i
	}

	if realSlots == 0 {
		return -1
	}
	if realSlots == 1 {
		return onlySlot
	}

	if requestedSlot < 0 || requestedSlot >= len(actives) {
		return -1
	}
	if actives[requestedSlot].Effects[0].Type == 0 {
		return -1
	}
	return requestedSlot
}

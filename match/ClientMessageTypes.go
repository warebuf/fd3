package match

type handMessage struct {
	Type      string `json:"Type"`
	CardIndex int    `json:"CardIndex"`
}

type stageCardMessage struct {
	Type      string `json:"type"`
	CardIndex int    `json:"CardIndex"`
}

type readyMessage struct {
	Type string `json:"type"`
}

package engine

type FuncType string

const (
	FuncHeal   FuncType = "Heal"
	FuncAOE    FuncType = "AOE"
	FuncAttack FuncType = "Attack"
)

type TargetType string

const (
	TgtSingleAny    TargetType = "SingleAny"    // any unit (ally or enemy); here we pick lowest hp ally by example
	TgtEnemiesAll   TargetType = "EnemiesAll"   // all enemies
	TgtAttackSingle TargetType = "AttackSingle" // one enemy (we pick lowest hp enemy)
)

type FunctionDef struct {
	Type        FuncType   `json:"Type"`
	Target      TargetType `json:"Target"`
	ATK         int        `json:"ATK"`
	SPD         int        `json:"SPD"` // higher = faster; cooldown = 1000 / SPD ms
	Description string     `json:"Description"`
}

type UnitDef struct {
	No        int           `json:"No"`
	Name      string        `json:"Name"`
	HP        int           `json:"HP"`
	Functions []FunctionDef `json:"Functions"`
}

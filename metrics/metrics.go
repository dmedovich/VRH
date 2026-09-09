package metrics

// Result contains the main quality metrics for an algorithm.
type Result struct {
	Balance      float64
	Collisions   int
	Migration    float64
	MigrationPct float64
}

// Calculate builds a metrics result from owners and scores.
func Calculate(
	nodes []uint64,
	owners []uint64,
	scores []uint64,
	oldOwners []uint64,
) Result {
	migration := 0.0

	if oldOwners != nil {
		migration = Migration(oldOwners, owners)
	}

	return Result{
		Balance:      Balance(nodes, owners),
		Collisions:   CollisionCount(scores),
		Migration:    migration,
		MigrationPct: migration * 100,
	}
}

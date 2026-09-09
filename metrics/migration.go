package metrics

// Migration calculates the fraction of keys whose owner changed.
//
// oldOwners and newOwners must describe the same keys in the same order.
// The result is in [0, 1].
func Migration(oldOwners, newOwners []uint64) float64 {
	if len(oldOwners) == 0 {
		return 0
	}

	if len(oldOwners) != len(newOwners) {
		panic("metrics: owner slices have different lengths")
	}

	changed := 0

	for i := range oldOwners {
		if oldOwners[i] != newOwners[i] {
			changed++
		}
	}

	return float64(changed) / float64(len(oldOwners))
}

// MigrationPercent returns migration as a percentage.
func MigrationPercent(oldOwners, newOwners []uint64) float64 {
	return Migration(oldOwners, newOwners) * 100
}

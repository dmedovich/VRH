package metrics

import (
	"math"
	"testing"
)

func TestMigration(t *testing.T) {
	oldOwners := []uint64{1, 1, 2, 2, 3, 3, 4, 4}
	newOwners := []uint64{1, 2, 2, 3, 3, 3, 4, 1}

	got := Migration(oldOwners, newOwners)

	want := 3.0 / 8.0

	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("expected %.12f, got %.12f", want, got)
	}
}

func TestMigrationNoChanges(t *testing.T) {
	owners := []uint64{1, 2, 3, 4}

	if got := Migration(owners, owners); got != 0 {
		t.Fatalf("expected 0 migration, got %v", got)
	}
}

func TestMigrationAllChanged(t *testing.T) {
	oldOwners := []uint64{1, 1, 2, 2}
	newOwners := []uint64{3, 3, 4, 4}

	if got := Migration(oldOwners, newOwners); got != 1 {
		t.Fatalf("expected 1.0 migration, got %v", got)
	}
}

func TestMigrationPercent(t *testing.T) {
	oldOwners := []uint64{1, 1, 2, 2}
	newOwners := []uint64{1, 2, 2, 3}

	got := MigrationPercent(oldOwners, newOwners)

	if math.Abs(got-50) > 1e-12 {
		t.Fatalf("expected 50%%, got %v", got)
	}
}

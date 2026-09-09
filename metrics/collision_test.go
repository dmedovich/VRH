package metrics

import "testing"

func TestCollisionCount(t *testing.T) {
	scores := []uint64{
		10, 20, 30, 40,
		20, 50, 60, 40,
	}

	got := CollisionCount(scores)

	if got != 2 {
		t.Fatalf("expected 2 collisions, got %d", got)
	}
}

func TestCollisionCountWithoutCollisions(t *testing.T) {
	scores := []uint64{
		10, 20, 30, 40,
	}

	if got := CollisionCount(scores); got != 0 {
		t.Fatalf("expected 0 collisions, got %d", got)
	}

	if HasCollision(scores) {
		t.Fatal("expected no collision")
	}
}

func TestHasCollision(t *testing.T) {
	if !HasCollision([]uint64{1, 2, 3, 1}) {
		t.Fatal("expected collision")
	}
}

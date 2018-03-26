package firedb

import (
	"context"
	"testing"
)

type DummyMonsterStore struct {
	UpdatePositionCount int
	MonsterPosition     *MonsterPosition
}

func (s *DummyMonsterStore) UpdatePosition(ctx context.Context, p *MonsterPosition) error {
	s.UpdatePositionCount++
	s.MonsterPosition = p

	return nil
}

func TestMonsterStore_UpdatePosition(t *testing.T) {
	dummy := &DummyMonsterStore{}
	SetMonsterStore(dummy)

	s := NewMonsterStore()
	p := &MonsterPosition{
		ID:     "dummy",
		Angle:  180,
		IsMove: false,
		X:      1000,
		Y:      1000,
	}
	ctx := context.Background()
	if err := s.UpdatePosition(ctx, p); err != nil {
		t.Fatalf("failed MonsterStore.UpdatePosition. err=%+v", err)
	}
	if e, g := 1, dummy.UpdatePositionCount; e != g {
		t.Fatalf("expected UpdatePositionCount is %d; got %d", e, g)
	}
	if e, g := p.ID, dummy.MonsterPosition.ID; e != g {
		t.Fatalf("expected MonsterPosition.ID is %s; got %s", e, g)
	}
	if e, g := p.Angle, dummy.MonsterPosition.Angle; e != g {
		t.Fatalf("expected MonsterPosition.Angle is %f; got %f", e, g)
	}
	if e, g := p.IsMove, dummy.MonsterPosition.IsMove; e != g {
		t.Fatalf("expected MonsterPosition.IsMove is %t; got %t", e, g)
	}
	if e, g := p.X, dummy.MonsterPosition.X; e != g {
		t.Fatalf("expected MonsterPosition.X is %f; got %f", e, g)
	}
	if e, g := p.Y, dummy.MonsterPosition.Y; e != g {
		t.Fatalf("expected MonsterPosition.Y is %f; got %f", e, g)
	}

}

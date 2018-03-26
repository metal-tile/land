package main

import (
	"context"

	"github.com/metal-tile/land/firedb"
)

// DummyMonsterStore is UnitTestのためのMonsterStore Dummy実装
type DummyMonsterStore struct {
	UpdatePositionCount int
	MonsterPosition     *firedb.MonsterPosition
}

func (s *DummyMonsterStore) UpdatePosition(ctx context.Context, p *firedb.MonsterPosition) error {
	s.UpdatePositionCount++
	s.MonsterPosition = p

	return nil
}

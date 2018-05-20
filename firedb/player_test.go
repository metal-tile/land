package firedb

import (
	"context"
	"sync"
	"testing"
	"time"
)

type dummyPlayerStore struct {
	UpdateActiveUserCount int
}

func (s *dummyPlayerStore) Watch(ctx context.Context, path string) error {
	return nil
}

func (s *dummyPlayerStore) GetPosition(id string) *PlayerPosition {
	return nil
}

func (s *dummyPlayerStore) GetPositionMap() *sync.Map {
	return nil
}

func (s *dummyPlayerStore) UpdateActiveUser(ctx context.Context, id string, active bool) error {
	s.UpdateActiveUserCount++
	return nil
}

func TestIsChangeActiveStatus(t *testing.T) {
	candidates := []struct {
		id        string
		playerMap map[string]*User
		change    bool
	}{
		{
			id:        "hogeUserID1",
			playerMap: map[string]*User{},
			change:    true,
		},
		{
			id:        "hogeUserID2",
			playerMap: map[string]*User{"hogeUserID2": &User{Active: false, UpdatedAt: time.Now()}},
			change:    true,
		},
		{
			id:        "hogeUserID2",
			playerMap: map[string]*User{"hogeUserID2": &User{Active: true, UpdatedAt: time.Now()}},
			change:    false,
		},
		{
			id:        "hogeUserID3",
			playerMap: map[string]*User{"hogeUserID3": &User{Active: true, UpdatedAt: time.Now().Add(time.Minute * -11)}},
			change:    true,
		},
	}

	for i, v := range candidates {
		if e, g := v.change, isChangeActiveStatus(v.playerMap, v.id); e != g {
			t.Fatalf("%d : expected %t; got %t", i, e, g)
		}
	}
}

package firedb

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

// PlayerStore is PlayerStore
type PlayerStore interface {
	Watch(ctx context.Context, path string) error
	GetPosition(id string) *PlayerPosition
	GetPositionMap() *sync.Map
}

// defaultPlayerStore is Default PlayerStore Functions
type defaultPlayerStore struct {
	positionMap *sync.Map
}

var playerStore PlayerStore

// NewPlayerStore is NewPlayerStore
func NewPlayerStore() PlayerStore {
	if playerStore == nil {
		playerStore = &defaultPlayerStore{
			positionMap: &sync.Map{},
		}
	}
	return playerStore
}

// SetPlayerStore is 実装を差し替えたいときに利用する
func SetPlayerStore(s PlayerStore) {
	playerStore = s
}

// PlayerPosition is Player Position Struct
// TODO IDをstructの中に持つか、Mapで持つようにするか悩ましい
type PlayerPosition struct {
	ID     string  `firestore:"-" json:"id"`
	Angle  float64 `json:"angle" firestore:"angle"`
	IsMove bool    `json:"isMove" firestore:"isMove"`
	X      float64 `json:"x" firestore:"x"`
	Y      float64 `json:"y" firestore:"y"`
}

// Watch is PlayerPosition Sync Firestore
func (s *defaultPlayerStore) Watch(ctx context.Context, path string) error {
	iter := db.Collection(path).Snapshots(ctx)
	defer iter.Stop()
	for {
		dociter, err := iter.Next()
		if err != nil {
			return errors.WithStack(err)
		}
		dslist, err := dociter.GetAll()
		if err != nil {
			return errors.WithStack(err)
		}
		for _, v := range dslist {
			var pp PlayerPosition
			pp.ID = v.Ref.ID
			err := v.DataTo(&pp)
			if err != nil {
				return errors.WithStack(err)
			}
			s.positionMap.Store(pp.ID, &pp)
		}
	}
}

func (s *defaultPlayerStore) GetPositionMap() *sync.Map {
	return s.positionMap
}

// GetPosition is 指定したIDのプレイヤーのポジションを取得
func (s *defaultPlayerStore) GetPosition(id string) *PlayerPosition {
	pp, ok := s.positionMap.Load(id)
	if ok == false {
		return nil
	}
	return pp.(*PlayerPosition)
}

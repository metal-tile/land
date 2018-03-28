package firedb

import (
	"context"
)

// PlayerStore is PlayerStore
type PlayerStore interface {
	GetPlayerPositions(ctx context.Context) ([]*PlayerPosition, error)
}

// PlayerStoreImple is PlayerStoreImple
type PlayerStoreImple struct{}

var playerStore *PlayerStoreImple

// NewPlayerStore is NewPlayerStore
func NewPlayerStore() PlayerStore {
	if playerStore != nil {
		return playerStore
	}
	s := PlayerStoreImple{}
	return &s
}

// SetPlayerStoreImple is 実装を差し替えたいときに利用する
func SetPlayerStoreImple(s PlayerStoreImple) {
	playerStore = &s
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

// GetPlayerPositions is PlayerPositionをFirestoreから取得する
func (s *PlayerStoreImple) GetPlayerPositions(ctx context.Context) ([]*PlayerPosition, error) {
	ds, err := db.Collection("world-default-player-position").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	var pps []*PlayerPosition
	for _, v := range ds {
		var pp PlayerPosition
		pp.ID = v.Ref.ID
		err := v.DataTo(&pp)
		if err != nil {
			return nil, err
		}
		pps = append(pps, &pp)
	}

	return pps, nil
}

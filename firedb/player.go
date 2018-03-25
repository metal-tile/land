package firedb

import (
	"context"
	"sync"

	"cloud.google.com/go/firestore"
)

var mu sync.RWMutex
var db *firestore.Client

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
	Angle  float64 `json:"angle"`
	IsMove bool    `json:"isMove"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
}

// SetUp is SetUp
func SetUp(ctx context.Context, projectID string) error {
	return createWithSetClient(ctx, projectID)
}

func createWithSetClient(ctx context.Context, projectID string) error {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	db = client

	return nil
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

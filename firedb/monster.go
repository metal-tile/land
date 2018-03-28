package firedb

import "context"

// MonsterStore is Monsterに関するFirestoreとのやりとりの役割を持つ
type MonsterStore interface {
	UpdatePosition(ctx context.Context, p *MonsterPosition) error
}

type monsterStoreImple struct{}

var monsterStore MonsterStore

// NewMonsterStore is MonsterStoreを生成する
func NewMonsterStore() MonsterStore {
	if monsterStore != nil {
		return monsterStore
	}
	return &monsterStoreImple{}
}

// SetMonsterStore is MonsterStoreの実装を差し替える
// Unit Testのために利用する
func SetMonsterStore(s MonsterStore) {
	monsterStore = s
}

// MonsterPosition is Monster Position struct
type MonsterPosition struct {
	ID     string  `firestore:"-" json:"id"`
	Speed  float64 `json:"speed" firestore:"speed"`
	Angle  float64 `json:"angle" firestore:"angle"`
	IsMove bool    `json:"isMove" firestore:"isMove"`
	X      float64 `json:"x" firestore:"x"`
	Y      float64 `json:"y" firestore:"y"`
}

// UpdatePosition is MonsterのPositionを更新する
func (s *monsterStoreImple) UpdatePosition(ctx context.Context, p *MonsterPosition) error {
	_, err := db.Collection("world-default-land-home-monster-position").Doc(p.ID).Set(ctx, p)
	if err != nil {
		return err
	}

	return nil
}

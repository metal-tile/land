package firedb

import "context"

// MonsterPosition is Monster Position struct
type MonsterPosition struct {
	ID     string  `firestore:"-" json:"id"`
	Angle  float64 `json:"angle"`
	IsMove bool    `json:"isMove"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
}

// UpdateMonsterPositions is MonsterのPositionを更新する
func UpdateMonsterPositions(ctx context.Context, p *MonsterPosition) error {
	_, err := db.Collection("world-default-land-home-monster-position").Doc(p.ID).Set(ctx, p)
	if err != nil {
		return err
	}

	return nil
}

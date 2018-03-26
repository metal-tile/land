package main

import (
	"context"
	"sync"
	"time"

	"github.com/metal-tile/land/dqn"
	"github.com/metal-tile/land/firedb"
	"github.com/pkg/errors"
	"github.com/sinmetal/slog"
)

var monsterPositionMap *sync.Map

func init() {
	monsterPositionMap = &sync.Map{}
}

// MonsterClient is Monsterに関連する処理を行うClient
type MonsterClient struct {
	DQN dqn.Client
}

// RunControlMonster is MonsterのControlを開始する
func RunControlMonster(client *MonsterClient) error {
	// TODO dummy monsterをdebugのために追加する
	const monsterID = "dummy"
	monsterPositionMap.Store(monsterID, &firedb.MonsterPosition{
		ID:    monsterID,
		X:     950,
		Y:     1000,
		Angle: 180,
		Speed: 4,
	})

	for {
		t := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-t.C:
				log := slog.Start(time.Now())

				// TODO getMonsterPosition()があったほうがいいかもしれない
				v, ok := monsterPositionMap.Load(monsterID)
				if !ok {
					log.Infof("%s is not found monsterPositionMap.", monsterID)
					continue
				}
				mob, ok := v.(*firedb.MonsterPosition)
				if !ok {
					log.Infof("%s is not cast monsterPositionMap.", monsterID)
					continue
				}
				err := client.UpdateMonster(&log, mob, playerPositionMap)
				if err != nil {
					log.Errorf("failed UpdateMonster. %+v", err)
				}

				log.Flush()
			}
		}
	}
}

// UpdateMonster is DQN Predictionに基づき、Firestore上のMonsterの位置を更新する
func (client *MonsterClient) UpdateMonster(log *slog.Log, mob *firedb.MonsterPosition, playerPositionMap *sync.Map) error {
	dp, err := buildDQNPayload(log, mob, playerPositionMap)
	if err != nil {
		return errors.Wrap(err, "failed buildDQNPayload")
	}
	ans, err := client.DQN.Prediction(log, dp)
	if err != nil {
		log.Infof("DQN.Payload %#v", dp)
		return errors.Wrap(err, "failed DQN.Prediction")
	}
	log.Infof("DQNAnswer:%+v", ans)

	ctx := context.Background()
	ms := firedb.NewMonsterStore()

	mob.X += ans.X * mob.Speed
	mob.Y += ans.Y * mob.Speed
	mob.IsMove = ans.Paused
	mob.Angle = ans.Angle
	monsterPositionMap.Store(mob.ID, mob)
	return ms.UpdatePosition(ctx, mob)
}

func buildDQNPayload(log *slog.Log, mp *firedb.MonsterPosition, playerPositionMap *sync.Map) (*dqn.Payload, error) {
	const dqnLayer = 0
	const playerLayer = 1

	payload := &dqn.Payload{
		Instances: []dqn.Instance{
			dqn.Instance{},
		},
	}
	// Monsterが中心ぐらいにいる状態
	payload.Instances[0].State[(dqn.SenseRangeRow / 2)][(dqn.SenseRangeCol / 2)][dqnLayer] = 1

	mobRow, mobCol := ConvertXYToRowCol(mp.X, mp.Y, 1.0)
	playerPositionMap.Range(func(key, value interface{}) bool {
		p, ok := value.(firedb.PlayerPosition)
		if !ok {
			return true
		}
		plyRow, plyCol := ConvertXYToRowCol(p.X, p.Y, 1.0)

		row := plyRow - mobRow + (dqn.SenseRangeRow / 2)
		if row < 0 || row >= dqn.SenseRangeRow {
			// 索敵範囲外にいる
			return true
		}
		col := plyCol - mobCol + (dqn.SenseRangeCol / 2)
		if col < 0 || col >= dqn.SenseRangeCol {
			// 索敵範囲外にいる
			return true
		}

		payload.Instances[0].State[row][col][playerLayer] = 1
		return true
	})

	return payload, nil
}

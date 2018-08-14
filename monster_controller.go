package main

import (
	"context"
	"fmt"
	"time"

	"github.com/metal-tile/land/dqn"
	"github.com/metal-tile/land/firedb"
	"github.com/pkg/errors"
	"github.com/sinmetal/slog"
	"github.com/sinmetal/stime"
	"github.com/tenntenn/sync/recoverable"
)

var monsterPositionMap map[string]*firedb.MonsterPosition

func init() {
	monsterPositionMap = make(map[string]*firedb.MonsterPosition)
}

// MonsterClient is Monsterに関連する処理を行うClient
type MonsterClient struct {
	DQN dqn.Client
	firedb.PlayerStore
}

// RunControlMonster is MonsterのControlを開始する
func RunControlMonster(client *MonsterClient) error {
	// TODO dummy monsterをdebugのために追加する
	const monsterID = "dummy"
	monsterPositionMap[monsterID] = &firedb.MonsterPosition{
		ID:    monsterID,
		X:     950,
		Y:     1000,
		Angle: 180,
		Speed: 4,
	}

	for {
		t := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-t.C:
				ctx := slog.WithLog(context.Background())

				f := recoverable.Func(func() {
					if err := handleMonster(ctx, client, monsterID); err != nil {
						panic(err) // TODO どうしようかな？
					}
				})
				if err := f(); err != nil {
					v, ok := recoverable.RecoveredValue(err)
					if ok {
						panic(v)
					}
					slog.Info(ctx, "FailedHandleMonster", fmt.Sprintf("%+v", err))
				}

				slog.Flush(ctx)
			}
		}
	}
}

func handleMonster(ctx context.Context, client *MonsterClient, monsterID string) error {
	if firedb.ExistsActivePlayer(client.PlayerStore.GetPlayerMap()) == false {
		return nil
	}

	mob, ok := monsterPositionMap[monsterID]
	if !ok {
		slog.Info(ctx, "NotFoundMonster", fmt.Sprintf("%s is not found monsterPositionMap.", monsterID))
		return nil
	}
	ppm := client.PlayerStore.GetPositionMapSnapshot()
	dp, err := BuildDQNPayload(ctx, mob, ppm)
	if err != nil {
		slog.Info(ctx, "FailedBuildDQNPayload", fmt.Sprintf("failed BuildDQNPayload. %+v,%+v,%+v", mob, ppm, err)) // TODO LogLevelを変えるか？
		return nil
	}
	err = client.UpdateMonster(ctx, mob, dp)
	if err != nil {
		slog.Info(ctx, "FailedUpdateMonster", fmt.Sprintf("failed UpdateMonster. %+v", err)) // TODO LogLevelを変えるか？
		return nil
	}

	return nil
}

// UpdateMonster is DQN Predictionに基づき、Firestore上のMonsterの位置を更新する
func (client *MonsterClient) UpdateMonster(ctx context.Context, mob *firedb.MonsterPosition, dp *dqn.Payload) error {
	ans, err := client.DQN.Prediction(ctx, dp)
	if err != nil {
		slog.Info(ctx, "DQNPayload", slog.KV{"DQNPayload", dp})
		return errors.Wrap(err, "failed DQN.Prediction")
	}
	slog.Info(ctx, "DQNAnswer", slog.KV{"DQNAnswer", ans})

	ms := firedb.NewMonsterStore()

	mob.X += ans.X * mob.Speed
	mob.Y += ans.Y * mob.Speed
	mob.IsMove = ans.IsMove
	mob.Angle = ans.Angle
	monsterPositionMap[mob.ID] = mob
	return ms.UpdatePosition(ctx, mob)
}

// BuildDQNPayload is DQNに渡すPayloadを構築する
func BuildDQNPayload(ctx context.Context, mp *firedb.MonsterPosition, playerPositionMap map[string]*firedb.PlayerPosition) (*dqn.Payload, error) {
	payload := &dqn.Payload{
		Instances: []dqn.Instance{
			dqn.Instance{},
		},
	}
	// Monsterが中心ぐらいにいる状態
	payload.Instances[0].State[(dqn.SenseRangeRow / 2)][(dqn.SenseRangeCol / 2)][dqn.MonsterLayer] = 1

	mobRow, mobCol := ConvertXYToRowCol(mp.X, mp.Y, 1.0)
	slog.Info(ctx, "StartPlayerPositionMapRange", "Start playerPositionMap.Range.")
	for _, p := range playerPositionMap {
		if stime.InTime(stime.Now(), p.FirestoreUpdateAt, 10*time.Second) == false {
			continue
		}
		plyRow, plyCol := ConvertXYToRowCol(p.X, p.Y, 1.0)

		row := plyRow - mobRow + (dqn.SenseRangeRow / 2)
		if row < 0 || row >= dqn.SenseRangeRow {
			// 索敵範囲外にいる
			slog.Info(ctx, "DQN.TargetIsFarAway", slog.KV{"row", row})
			continue
		}
		col := plyCol - mobCol + (dqn.SenseRangeCol / 2)
		if col < 0 || col >= dqn.SenseRangeCol {
			slog.Info(ctx, "DQN.TargetIsFarAway", slog.KV{"col", col})
			// 索敵範囲外にいる
			continue
		}

		slog.Info(ctx, "DQNPayloadPlayerPosition", fmt.Sprintf("DQN.Payload.PlayerPosition row=%d,col=%d", row, col))
		payload.Instances[0].State[row][col][dqn.PlayerLayer] = 1
	}

	return payload, nil
}

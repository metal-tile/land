package main

import (
	"context"
	"time"

	"github.com/metal-tile/land/dqn"
	"github.com/metal-tile/land/firedb"
	"github.com/pkg/errors"
	"github.com/sinmetal/slog"
)

// MonsterClient is Monsterに関連する処理を行うClient
type MonsterClient struct {
	// SLog *slog.Log
	DQN *dqn.Client
}

// RunControlMonster is MonsterのControlを開始する
func RunControlMonster(client *MonsterClient) error {
	for {
		t := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-t.C:
				log := slog.Start(time.Now())

				err := client.UpdateMonster(&log)
				if err != nil {
					log.Errorf("failed UpdateMonster. %+v", err)
				}

				// client.SLog.Flush() // TODO ctx baseでログをまとめるようにする
				log.Flush()
			}
		}
	}
}

// UpdateMonster is DQN Predictionに基づき、Firestore上のMonsterの位置を更新する
func (client *MonsterClient) UpdateMonster(slog *slog.Log) error {
	dp, err := buildDQNPayload(slog)
	if err != nil {
		return errors.Wrap(err, "failed buildDQNPayload")
	}
	ans, err := client.DQN.Prediction(slog, dp)
	if err != nil {
		slog.Infof("DQN.Payload %#v", dp)
		return errors.Wrap(err, "failed DQN.Prediction")
	}
	slog.Infof("DQNAnswer", ans)

	// TODO 適当に値を入れてみる
	ctx := context.Background()
	return firedb.UpdateMonsterPositions(ctx, &firedb.MonsterPosition{
		ID:    "dummy",
		Angle: 180,
		X:     1000,
		Y:     1000,
	})
}

func buildDQNPayload(log *slog.Log) (*dqn.Payload, error) {
	const dqnLayer = 0
	const playerLayer = 1

	payload := &dqn.Payload{
		Instances: []dqn.Instance{
			dqn.Instance{},
		},
	}

	// TODO Playerが常に右隣にいる状態
	payload.Instances[0].State[(dqn.SenseRangeRow / 2)][(dqn.SenseRangeCol / 2)][playerLayer] = 1

	// DQNが中心ぐらいにいる状態
	payload.Instances[0].State[(dqn.SenseRangeRow / 2)][(dqn.SenseRangeCol / 2)][dqnLayer] = 1

	return payload, nil
}

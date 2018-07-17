package main

import (
	"context"
	"testing"

	"github.com/metal-tile/land/dqn"
	"github.com/metal-tile/land/firedb"
	"github.com/sinmetal/slog"
	"github.com/sinmetal/stime"
)

func TestMonsterClient_UpdateMonster(t *testing.T) {
	dqnDummy := &DQNDummyClient{
		DummyAnswer: &dqn.Answer{
			X:      -1,
			Y:      0,
			IsMove: true,
			Angle:  0,
			Speed:  4,
		},
	}
	dqn.SetDummyClient(dqnDummy)

	msDummy := &DummyMonsterStore{}
	firedb.SetMonsterStore(msDummy)

	d := dqn.NewClient()
	ctx := slog.WithLog(context.Background())
	client := MonsterClient{
		DQN: d,
	}

	playerPositionMap := make(map[string]*firedb.PlayerPosition)
	mob := &firedb.MonsterPosition{
		ID:    "dummy",
		X:     950,
		Y:     1000,
		Angle: 180,
		Speed: 4,
	}
	dp, err := BuildDQNPayload(ctx, mob, playerPositionMap)
	if err != nil {
		t.Fatalf("failed BuildDQNPayload. err=%+v", err)
	}

	if err := client.UpdateMonster(ctx, mob, dp); err != nil {
		t.Fatalf("failed UpdateMonster. err=%+v", err)
	}

	// とりあえず呼び出されていることだけを確認
	if e, g := 1, dqnDummy.PredictionCount; e != g {
		t.Fatalf("expected DQN.PredictionCount is %d; gpt %d", e, g)
	}
	if e, g := 1, msDummy.UpdatePositionCount; e != g {
		t.Fatalf("expected MonsterStore.UpdatePositionCount is %d; gpt %d", e, g)
	}
}

func TestBuildDQNPayload(t *testing.T) {
	candidates := []struct {
		playerPositionMap map[string]*firedb.PlayerPosition
		monsterPosition   *firedb.MonsterPosition
		row               int
		col               int
		target            float64
		targetNothing     bool
	}{
		{
			playerPositionMap: make(map[string]*firedb.PlayerPosition),
			monsterPosition: &firedb.MonsterPosition{
				ID:    "dummy",
				X:     950,
				Y:     1000,
				Angle: 180,
				Speed: 4,
			},
			row:    0,
			col:    -1,
			target: 1.0, // 右にプレイヤーがいるので、COLが中心より1少ない値になる
		},
		{
			playerPositionMap: make(map[string]*firedb.PlayerPosition),
			monsterPosition: &firedb.MonsterPosition{
				ID:    "dummy",
				X:     950,
				Y:     1000,
				Angle: 180,
				Speed: 4,
			},
			row:           0,
			col:           0,
			target:        0.0,
			targetNothing: true, // プレイヤーがいないので、動かない
		},
	}
	candidates[0].playerPositionMap["sinmetal"] = &firedb.PlayerPosition{
		ID:                "sinmetal",
		X:                 900,
		Y:                 1000,
		Angle:             180,
		FirestoreUpdateAt: stime.Now(),
	}

	// FirestoreUpdateAt が古いので無視されるプレイヤー
	candidates[1].playerPositionMap["sinmetal"] = &firedb.PlayerPosition{
		ID:    "sinmetal",
		X:     900,
		Y:     1000,
		Angle: 180,
	}

	ctx := slog.WithLog(context.Background())
	for i, v := range candidates {
		dp, err := BuildDQNPayload(ctx, v.monsterPosition, v.playerPositionMap)
		if err != nil {
			t.Fatalf("failed BuildDQNPayload. err=%+v", err)
		}
		if e, g := v.target, dp.Instances[0].State[dqn.SenseRangeRow/2+v.row][dqn.SenseRangeCol/2+v.col][dqn.PlayerLayer]; e != g {
			t.Fatalf("%d : expected v = %f; got %f", i, e, g)
		}

		if v.targetNothing {
			for row := 0; row < dqn.SenseRangeRow; row++ {
				for col := 0; col < dqn.SenseRangeCol; col++ {
					if e, g := 0.0, dp.Instances[0].State[row][col][dqn.PlayerLayer]; e != g {
						t.Fatalf("%d : expected [%d][%d] = %f; got %f", i, row, col, e, g)
					}
				}
			}
		}
	}
}

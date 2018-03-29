package main

import (
	"sync"
	"testing"

	"github.com/metal-tile/land/dqn"
	"github.com/metal-tile/land/firedb"
	"github.com/sinmetal/slog"
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
	l := &slog.Log{}
	client := MonsterClient{
		DQN: d,
	}

	playerPositionMap := &sync.Map{}
	mob := &firedb.MonsterPosition{
		ID:    "dummy",
		X:     950,
		Y:     1000,
		Angle: 180,
		Speed: 4,
	}
	dp, err := BuildDQNPayload(l, mob, playerPositionMap)
	if err != nil {
		t.Fatalf("failed BuildDQNPayload. err=%+v", err)
	}

	if err := client.UpdateMonster(l, mob, dp); err != nil {
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
	l := &slog.Log{}

	playerPositionMap := &sync.Map{}
	playerPositionMap.Store("sinmetal", &firedb.PlayerPosition{
		ID:    "sinmetal",
		X:     900,
		Y:     1000,
		Angle: 180,
	})

	mob := &firedb.MonsterPosition{
		ID:    "dummy",
		X:     950,
		Y:     1000,
		Angle: 180,
		Speed: 4,
	}
	dp, err := BuildDQNPayload(l, mob, playerPositionMap)
	if err != nil {
		t.Fatalf("failed BuildDQNPayload. err=%+v", err)
	}
	// 右にプレイヤーがいるので、COLが中心より1少ない値になる
	if e, g := 1.0, dp.Instances[0].State[dqn.SenseRangeRow/2][dqn.SenseRangeCol/2-1][1]; e != g {
		t.Fatalf("expected v = %f; got %f", e, g)
	}
}

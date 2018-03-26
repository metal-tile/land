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
			Paused: false,
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

	if err := client.UpdateMonster(l, &firedb.MonsterPosition{
		ID:    "dummy",
		X:     950,
		Y:     1000,
		Angle: 180,
		Speed: 4,
	}, playerPositionMap); err != nil {
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

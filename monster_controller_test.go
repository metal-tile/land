package main

import (
	"testing"

	"github.com/metal-tile/land/dqn"
	"github.com/metal-tile/land/firedb"
	"github.com/sinmetal/slog"
)

func TestMonsterClient_UpdateMonster(t *testing.T) {
	dqnDummy := &DQNDummyClient{}
	dqn.SetDummyClient(dqnDummy)

	msDummy := &DummyMonsterStore{}
	firedb.SetMonsterStore(msDummy)

	d := dqn.NewClient()
	l := &slog.Log{}
	client := MonsterClient{
		DQN: d,
	}

	if err := client.UpdateMonster(l); err != nil {
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

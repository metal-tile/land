package main

import (
	"github.com/metal-tile/land/dqn"
	"github.com/sinmetal/slog"
)

// DQNDummyClient is UnitTestのためのDQN Dummy実装
type DQNDummyClient struct {
	PredictionCount int
	Body            *dqn.Payload
	DummyAnswer     *dqn.Answer
}

func (client *DQNDummyClient) Prediction(log *slog.Log, body *dqn.Payload) (*dqn.Answer, error) {
	client.PredictionCount++
	client.Body = body
	return client.DummyAnswer, nil
}

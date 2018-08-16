package main

import (
	"context"

	"github.com/metal-tile/land/dqn"
)

// DQNDummyClient is UnitTestのためのDQN Dummy実装
type DQNDummyClient struct {
	PredictionCount int
	Body            *dqn.Payload
	DummyAnswer     *dqn.Answer
}

func (client *DQNDummyClient) Prediction(ctx context.Context, body *dqn.Payload) (*dqn.Answer, error) {
	client.PredictionCount++
	client.Body = body
	return client.DummyAnswer, nil
}

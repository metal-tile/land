package dqn

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sinmetal/slog"
	"go.opencensus.io/trace"
)

const (
	// SenseRangeRow is AIが感知できる範囲Row
	SenseRangeRow = 8
	// SenseRangeCol is AIが感知できる範囲Col
	SenseRangeCol = 8

	// AngleLeft is 左向きの角度
	AngleLeft = 270.0
	// AngleRight is 右向きの角度
	AngleRight = 90.0
	// AngleUp is 上向きの角度
	AngleUp = 0.0
	// AngleDown is 下向きの角度
	AngleDown = 180.0

	speed = 4.0

	// MonsterLayer is DQN PayloadのMonsterの座標の添字
	MonsterLayer = 0
	// PlayerLayer is DQN PayloadのPlayerの座標の添字
	PlayerLayer = 1
)

// ErrDQNAPIResponse is DQN ServerからのError時に利用する
var ErrDQNAPIResponse = errors.New("dqn: api error")

// Payload is DQN APIへ送るためのリクエストのPayload
type Payload struct {
	Instances []Instance `json:"instances"`
}

// Instance is DQNが判断するためのMapの情報
// [0] 追いかける奴の場所が1で他が0になっている[8 x 8]の配列
// [1] 追いかけられる奴の場所が1で他が0になっている[8 x 8]の配列
// [2] 障害物がある場所が1で他が0になっている[8 x 8]の配列
type Instance struct {
	State [SenseRangeRow][SenseRangeCol][3]float64 `json:"state"`
	Key   int                                      `json:"key"`
}

// apiResponse is DQN APIからのResponseの型
type apiResponse struct {
	Predictions []predictions `json:"predictions"`
}

// predictions is DQN APIのPrediction結果
// Q Score: [0]何もしない,[1]左,[2]右,[3]上,[4]下
type predictions struct {
	Q   []float64 `json:"q"`
	Key int       `json:"key"`
}

// Answer is DQN APIのResponseを元に、モンスターの行動を決定した内容
type Answer struct {
	X      float64
	Y      float64
	Angle  float64
	IsMove bool
	Speed  float64
}

// Client is DQN APIを実行するClient
type Client interface {
	Prediction(ctx context.Context, body *Payload) (*Answer, error)
}

var client Client

// dqnImpl is DQN APIのためのデフォルト実装
type dqnImpl struct{}

// NewClient is Clientを返す
func NewClient() Client {
	if client != nil {
		return client
	}
	return &dqnImpl{}
}

// SetDummyClient is UnitTestのために実装を差し替えるためのもの
func SetDummyClient(dummy Client) {
	client = dummy
}

// Prediction is DQN APIを実行する実装
func (d *dqnImpl) Prediction(ctx context.Context, body *Payload) (*Answer, error) {
	ctx, span := trace.StartSpan(ctx, "/dqn")
	defer span.End()

	b, err := json.Marshal(body)
	if err != nil {
		slog.Info(ctx, "FailedDQNPrediction", err.Error())
		return nil, err
	}

	client := new(http.Client)
	req, err := http.NewRequest(
		"POST",
		"http://dqn-service.default.svc.cluster.local:8081/dqn",
		strings.NewReader(string(b)),
	)
	if err != nil {
		slog.Info(ctx, "FailedDQNPredictionRequest", err.Error())
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		slog.Info(ctx, "FailedDQNClientDo", err.Error())
		return nil, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		slog.Info(ctx, "FailedReadDQNResponseBody", err.Error())
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		slog.Info(ctx, "FailedDQN", fmt.Sprintf("DQN Response Code = %d, Body = %s", res.StatusCode, resBody))
		return nil, ErrDQNAPIResponse
	}

	var dqnRes = apiResponse{}
	err = json.Unmarshal(resBody, &dqnRes)
	if err != nil {
		slog.Info(ctx, "FailedDQNResponseJsonUnmarshal", fmt.Sprintf("err = %s, body = %s", err.Error(), resBody))
		return nil, err
	}

	return buildDQNAnswer(&dqnRes)
}

func buildDQNAnswer(res *apiResponse) (*Answer, error) {
	score := &struct {
		None  float64
		Left  float64
		Right float64
		Up    float64
		Down  float64
	}{
		None:  res.Predictions[0].Q[0],
		Left:  res.Predictions[0].Q[1],
		Right: res.Predictions[0].Q[2],
		Up:    res.Predictions[0].Q[3],
		Down:  res.Predictions[0].Q[4],
	}

	if score.None > score.Left && score.None > score.Right && score.None > score.Up && score.None > score.Down {
		// 何もしない
		return &Answer{
			X:      0,
			Y:      0,
			IsMove: false,
			Angle:  AngleDown,
			Speed:  0,
		}, nil
	} else if score.Left > score.None && score.Left > score.Right && score.Left > score.Up && score.Left > score.Down {
		// 左
		return &Answer{
			X:      -1,
			Y:      0,
			IsMove: true,
			Angle:  AngleLeft,
			Speed:  speed,
		}, nil
	} else if score.Right > score.None && score.Right > score.Left && score.Right > score.Up && score.Right > score.Down {
		// 右
		return &Answer{
			X:      1,
			Y:      0,
			IsMove: true,
			Angle:  AngleRight,
			Speed:  speed,
		}, nil
	} else if score.Up > score.None && score.Up > score.Left && score.Up > score.Right && score.Up > score.Down {
		// 上
		return &Answer{
			X:      0,
			Y:      -1,
			IsMove: true,
			Angle:  AngleUp,
			Speed:  speed,
		}, nil
	} else {
		// 下
		return &Answer{
			X:      0,
			Y:      1,
			IsMove: true,
			Angle:  AngleDown,
			Speed:  speed,
		}, nil
	}

}

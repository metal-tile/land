package dqn

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sinmetal/slog"
)

const (
	// SenseRangeRow is AIが感知できる範囲Row
	SenseRangeRow = 8
	// SenseRangeCol is AIが感知できる範囲Col
	SenseRangeCol = 8

	// AngleLeft is 左向きの角度
	AngleLeft = 180.0
	// AngleRight is 右向きの角度
	AngleRight = 0.0
	// AngleUp is 上向きの角度
	AngleUp = 90.0
	// AngleDown is 下向きの角度
	AngleDown = 270.0

	speed = 4.0
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
	Paused bool
	Speed  float64
}

// Client is DQN APIを実行するClient
type Client interface {
	Prediction(log *slog.Log, body *Payload) (*Answer, error)
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
func (d *dqnImpl) Prediction(log *slog.Log, body *Payload) (*Answer, error) {
	b, err := json.Marshal(body)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	client := new(http.Client)
	req, err := http.NewRequest(
		"POST",
		"http://dqn-service.default.svc.cluster.local:8081/dqn",
		strings.NewReader(string(b)),
	)
	if err != nil {
		log.Errorf("dqn request. err = %s", err.Error())
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		log.Errorf("dqn client.Do err = %s", err.Error())
		return nil, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorf("dqn request.Body %s", err.Error())
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		log.Errorf("DQN Response Code = %d, Body = %s", res.StatusCode, resBody)
		return nil, ErrDQNAPIResponse
	}

	var dqnRes = apiResponse{}
	err = json.Unmarshal(resBody, &dqnRes)
	if err != nil {
		log.Errorf("dqn response json unmarshal error. err = %s, body = %s", err.Error(), resBody)
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
			Paused: true,
			Angle:  AngleDown,
			Speed:  0,
		}, nil
	} else if score.Left > score.None && score.Left > score.Right && score.Left > score.Up && score.Left > score.Down {
		// 左
		return &Answer{
			X:      1,
			Y:      0,
			Paused: false,
			Angle:  AngleLeft,
			Speed:  speed,
		}, nil
	} else if score.Right > score.None && score.Right > score.Left && score.Right > score.Up && score.Right > score.Down {
		// 右
		return &Answer{
			X:      -1,
			Y:      0,
			Paused: false,
			Angle:  AngleRight,
			Speed:  speed,
		}, nil
	} else if score.Up > score.None && score.Up > score.Left && score.Up > score.Right && score.Up > score.Down {
		// 上
		return &Answer{
			X:      0,
			Y:      1,
			Paused: false,
			Angle:  AngleUp,
			Speed:  speed,
		}, nil
	} else {
		// 下
		return &Answer{
			X:      0,
			Y:      -1,
			Paused: false,
			Angle:  AngleDown,
			Speed:  speed,
		}, nil
	}

}

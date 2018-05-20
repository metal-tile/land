package firedb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/pkg/errors"
	"github.com/sinmetal/stime"
)

// PlayerStore is PlayerStore
type PlayerStore interface {
	Watch(ctx context.Context, path string) error
	GetPosition(id string) *PlayerPosition
	GetPlayerMap() map[string]*User
	GetPositionMap() *sync.Map
	SetPassiveUser(ctx context.Context, id string) error
	UpdateActiveUser(ctx context.Context, id string, active bool) error
}

// defaultPlayerStore is Default PlayerStore Functions
type defaultPlayerStore struct {
	playerMap   map[string]*User
	positionMap *sync.Map
}

var playerStore PlayerStore

// NewPlayerStore is NewPlayerStore
func NewPlayerStore() PlayerStore {
	if playerStore == nil {
		playerStore = &defaultPlayerStore{
			playerMap:   make(map[string]*User),
			positionMap: &sync.Map{},
		}
	}
	return playerStore
}

// SetPlayerStore is 実装を差し替えたいときに利用する
func SetPlayerStore(s PlayerStore) {
	playerStore = s
}

// User is `world-{world}-users`
type User struct {
	Name      string    `firestore:"name"`
	Active    bool      `firestore:"active"`
	UpdatedAt time.Time `firestore:"updatedAt"`
}

// PlayerPosition is Player Position Struct
// TODO IDをstructの中に持つか、Mapで持つようにするか悩ましい
type PlayerPosition struct {
	ID     string  `firestore:"-" json:"id"`
	Angle  float64 `json:"angle" firestore:"angle"`
	IsMove bool    `json:"isMove" firestore:"isMove"`
	X      float64 `json:"x" firestore:"x"`
	Y      float64 `json:"y" firestore:"y"`
}

// Watch is PlayerPosition Sync Firestore
func (s *defaultPlayerStore) Watch(ctx context.Context, path string) error {
	iter := db.Collection(path).Snapshots(ctx)
	defer iter.Stop()
	for {
		dociter, err := iter.Next()
		if err != nil {
			return errors.WithStack(err)
		}
		dslist, err := dociter.GetAll()
		if err != nil {
			return errors.WithStack(err)
		}
		for _, v := range dslist {
			var pp PlayerPosition
			pp.ID = v.Ref.ID
			err := v.DataTo(&pp)
			if err != nil {
				return errors.WithStack(err)
			}
			s.positionMap.Store(pp.ID, &pp)

			if isChangeActiveStatus(s.playerMap, pp.ID) == false {
				s.SetActiveUser(ctx, pp.ID)
			}
		}
	}
}

func (s *defaultPlayerStore) GetPlayerMap() map[string]*User {
	return s.playerMap
}

func (s *defaultPlayerStore) GetPositionMap() *sync.Map {
	return s.positionMap
}

// GetPosition is 指定したIDのプレイヤーのポジションを取得
func (s *defaultPlayerStore) GetPosition(id string) *PlayerPosition {
	pp, ok := s.positionMap.Load(id)
	if ok == false {
		return nil
	}
	return pp.(*PlayerPosition)
}

// SetActiveUser is 移動しているなどアクティブであることが計測されたユーザの状態を更新する
func (s *defaultPlayerStore) SetActiveUser(ctx context.Context, id string) error {
	_, ok := s.playerMap[id]
	if !ok {
		s.playerMap[id] = &User{}
	}

	s.playerMap[id].Active = true
	s.playerMap[id].UpdatedAt = stime.Now()

	if err := s.UpdateActiveUser(ctx, id, true); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// SetPassiveUser is ユーザをパッシブ状態にする
func (s *defaultPlayerStore) SetPassiveUser(ctx context.Context, id string) error {
	v, ok := s.playerMap[id]
	if !ok {
		s.playerMap[id] = &User{}
	}
	v.Active = false
	v.UpdatedAt = stime.Now()

	s.playerMap[id] = v

	if err := s.UpdateActiveUser(ctx, id, false); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *defaultPlayerStore) UpdateActiveUser(ctx context.Context, id string, active bool) error {
	ref := db.Doc(fmt.Sprintf("world-default-users/%s", id))
	err := db.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref)
		if err != nil {
			return err
		}
		var u User
		doc.DataTo(&u)
		u.Active = active
		u.UpdatedAt = time.Now()
		return tx.Set(ref, &u)
	})
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("id = %s", id))
	}

	return nil
}

func isChangeActiveStatus(playerMap map[string]*User, id string) bool {
	u, ok := playerMap[id]
	if !ok {
		return true
	}
	if u.Active == false {
		return true
	}
	t := u.UpdatedAt.Add(time.Minute * 10)
	if t.Before(stime.Now()) {
		return true
	}

	return false
}

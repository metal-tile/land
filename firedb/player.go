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
	GetPlayerMapSnapshot() map[string]*User
	GetPositionMapSnapshot() map[string]*PlayerPosition
	SetPassiveUser(ctx context.Context, id string) error
	UpdateActiveUser(ctx context.Context, id string, active bool) error
}

// defaultPlayerStore is Default PlayerStore Functions
type defaultPlayerStore struct {
	playerMap        map[string]*User
	positionMap      map[string]*PlayerPosition
	playerMapMutex   *sync.RWMutex
	positionMapMutex *sync.RWMutex
}

var playerStore PlayerStore

// NewPlayerStore is NewPlayerStore
func NewPlayerStore() PlayerStore {
	if playerStore == nil {
		playerStore = &defaultPlayerStore{
			playerMap:        make(map[string]*User),
			positionMap:      make(map[string]*PlayerPosition),
			playerMapMutex:   &sync.RWMutex{},
			positionMapMutex: &sync.RWMutex{},
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
	ID                string    `firestore:"-" json:"id"`
	Angle             float64   `json:"angle" firestore:"angle"`
	IsMove            bool      `json:"isMove" firestore:"isMove"`
	X                 float64   `json:"x" firestore:"x"`
	Y                 float64   `json:"y" firestore:"y"`
	FirestoreUpdateAt time.Time `firestore:"-"` // FirestoreのUpdateTime
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
		dslist := dociter.Changes
		if err != nil {
			return errors.WithStack(err)
		}
		for _, v := range dslist {
			// ReadTimeはアプリケーションが読み込んだ時間、 UpdateTimeはそのデータがFirestoreで読み込んだ時間のようだ
			if stime.InTime(stime.Now(), v.Doc.UpdateTime, 10*time.Second) == false {
				// 対象のデータが古い場合は、スルーする
				continue
			}

			var pp PlayerPosition
			pp.ID = v.Doc.Ref.ID
			err := v.Doc.DataTo(&pp)
			if err != nil {
				return errors.WithStack(err)
			}
			pp.FirestoreUpdateAt = v.Doc.UpdateTime
			s.positionMap[pp.ID] = &pp

			if isChangeActiveStatus(s.playerMap, pp.ID) {
				fmt.Printf("%s is Active\n", pp.ID)
				if err := s.SetActiveUser(ctx, pp.ID); err != nil {
					return errors.WithStack(err)
				}
			}
		}
	}
}

func (s *defaultPlayerStore) GetPlayerMapSnapshot() map[string]*User {
	s.playerMapMutex.RLock()
	defer s.playerMapMutex.RUnlock()

	playerMap := make(map[string]*User)
	for k, v := range s.playerMap {
		playerMap[k] = v
	}
	return playerMap
}

// GetPositionMapSnapshot is PlayerPositionMapをCopyして返す
// Copyしているのは、複数goroutineで使うことを考慮しているため。
// Map全体を見る処理が軽い場合は、Copyせずに直接Lockを取った方が良いが、重たい処理をする時のためにSnapshotを取っている。
func (s *defaultPlayerStore) GetPositionMapSnapshot() map[string]*PlayerPosition {
	s.positionMapMutex.RLock()
	defer s.positionMapMutex.RUnlock()

	positionMap := make(map[string]*PlayerPosition)
	for k, v := range s.positionMap {
		positionMap[k] = v
	}

	return positionMap
}

// GetPosition is 指定したIDのプレイヤーのポジションを取得
func (s *defaultPlayerStore) GetPosition(id string) *PlayerPosition {
	pp, ok := s.positionMap[id]
	if ok == false {
		return nil
	}
	return pp
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

// ExistsActivePlayer is Activeなプレイヤーが存在するかどうかを返す
func ExistsActivePlayer(playerMap map[string]*User) bool {
	for _, v := range playerMap {
		if v.Active {
			return true
		}
	}
	return false
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

package main

import (
	"context"
	"time"

	"github.com/metal-tile/land/firedb"
	"github.com/sinmetal/stime"
)

// WatchPassivePlayer is まったく動いていないプレイヤーを探して、パッシブ状態にDBを変更する
func WatchPassivePlayer() error {
	ps := firedb.NewPlayerStore()
	for {
		t := time.NewTicker(60 * time.Second)
		for {
			select {
			case <-t.C:
				ctx := context.Background()
				pm := ps.GetPlayerMap()
				for k, v := range pm {
					if v.Active == false {
						continue
					}
					if IsPlayerPassive(v) {
						ps.SetPassiveUser(ctx, k)
					}
				}
			}
		}
	}
}

// IsPlayerPassive is プレイヤーがパッシブかどうかの判定
func IsPlayerPassive(user *firedb.User) bool {
	t := user.UpdatedAt.Add(time.Minute * 15)
	return t.Before(stime.Now())
}

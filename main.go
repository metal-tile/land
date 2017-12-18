package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sinmetal/slog"

	"github.com/metal-tile/land/firedb"
)

var playerPositionMap *sync.Map

func main() {
	hs, err := os.Hostname()
	if err != nil {
		fmt.Printf("Fail os.Hostname. %s\n", err.Error())
	}
	fmt.Printf("Hostname is %s\n", hs)
	fmt.Println("")
	fmt.Println(os.Environ())

	playerPositionMap = &sync.Map{}

	ctx := context.Background()
	firedb.SetUp(ctx, "metal-tile-dev1")

	ch := make(chan error)

	go func() {
		ch <- watchPlayerPositions()
	}()

	err = <-ch
	fmt.Println(err.Error())
}

func watchPlayerPositions() error {
	playerStore := firedb.NewPlayerStore()
	for {
		t := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-t.C:
				log := slog.Start(time.Now())
				ctx := context.Background()
				pps, err := playerStore.GetPlayerPositions(ctx)
				if err != nil {
					log.Errorf("playerStore.GetPlayerPositions. %s", err.Error())
					log.Flush()
					continue
				}

				// TODO playerStore.GetPlayerPositions の戻り値はmapの方が分かりやすい気がする
				for _, v := range pps {
					playerPositionMap.Store(v.ID, v)
				}

				// debug log
				j, err := json.Marshal(pps)
				if err != nil {
					log.Errorf("json.Marshal. %s", err.Error())
					log.Flush()
				}
				log.Infof(string(j))
				log.Flush()
			}
		}
	}
}

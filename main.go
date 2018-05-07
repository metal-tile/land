package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/profiler"
	"github.com/metal-tile/land/dqn"
	"github.com/metal-tile/land/firedb"
	"github.com/sinmetal/slog"
)

var playerPositionMap *sync.Map

func main() {
	if err := profiler.Start(profiler.Config{Service: "land", ServiceVersion: "0.0.1"}); err != nil {
		fmt.Printf("failed stackdriver.profiler.Start %+v", err)
	}

	hs, err := os.Hostname()
	if err != nil {
		fmt.Printf("Fail os.Hostname. %s\n", err.Error())
	}
	fmt.Printf("Hostname is %s\n", hs)
	fmt.Println("")
	fmt.Println(os.Environ())

	onlyFuncActivate := flag.String("onlyFuncActivate", "", "Activate only specified function")
	flag.Parse()
	fmt.Printf("onlyFuncActivate is %s\n", *onlyFuncActivate)

	playerPositionMap = &sync.Map{}

	ctx := context.Background()
	firedb.SetUp(ctx, "metal-tile-dev1")

	ch := make(chan error)

	if *onlyFuncActivate == "" || *onlyFuncActivate == "field" {
		fmt.Println("Start WatchField")
		fieldStore := firedb.NewFieldStore()
		go func() {
			ch <- fieldStore.Watch(ctx, "world-default20170908-land-home")
		}()
	}

	if *onlyFuncActivate == "" || *onlyFuncActivate == "playerPosition" {
		fmt.Println("Start WatchPlayerPositions")
		go func() {
			ch <- watchPlayerPositions()
		}()
	}

	if *onlyFuncActivate == "" || *onlyFuncActivate == "monster" {
		fmt.Println("Start Monster Control")
		go func() {
			c := &MonsterClient{
				DQN: dqn.NewClient(),
			}
			ch <- RunControlMonster(c)
		}()
	}

	err = <-ch
	fmt.Printf("%+v", err)
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
				//j, err := json.Marshal(pps)
				//if err != nil {
				//	log.Errorf("json.Marshal. %s", err.Error())
				//	log.Flush()
				//}
				//log.Infof(string(j))
				log.Flush()
			}
		}
	}
}

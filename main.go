package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"cloud.google.com/go/profiler"
	"github.com/metal-tile/land/dqn"
	"github.com/metal-tile/land/firedb"
)

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

	ctx := context.Background()
	firedb.SetUp(ctx, "metal-tile-dev1")

	ch := make(chan error)

	fieldStore := firedb.NewFieldStore()
	if *onlyFuncActivate == "" || *onlyFuncActivate == "field" {
		fmt.Println("Start WatchField")
		go func() {
			ch <- fieldStore.Watch(ctx, "world-default20170908-land-home")
		}()
	}

	playerStore := firedb.NewPlayerStore()
	if *onlyFuncActivate == "" || *onlyFuncActivate == "playerPosition" {
		fmt.Println("Start WatchPlayerPositions")
		go func() {
			ch <- playerStore.Watch(ctx, "world-default-player-position")
		}()
	}

	if *onlyFuncActivate == "" || *onlyFuncActivate == "monster" {
		fmt.Println("Start Monster Control")
		go func() {
			c := &MonsterClient{
				DQN:         dqn.NewClient(),
				PlayerStore: playerStore,
			}
			ch <- RunControlMonster(c)
		}()
	}

	// Debug HTTP Handler
	go func() {
		http.HandleFunc("/", helthHandler)
		http.HandleFunc("/field", fieldHandler)
		http.HandleFunc("/player", playerHandler)
		http.HandleFunc("/healthz", helthHandler)
		http.ListenAndServe(":8080", nil)
	}()

	err = <-ch
	fmt.Printf("%+v", err)
}

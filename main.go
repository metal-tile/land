package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"cloud.google.com/go/profiler"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/metal-tile/land/dqn"
	"github.com/metal-tile/land/firedb"
	"github.com/sinmetal/gcpmetadata"
	"go.opencensus.io/trace"
)

func main() {
	projectID, err := gcpmetadata.GetProjectID()
	if err != nil {
		panic(err)
	}

	if err := profiler.Start(profiler.Config{Service: "land", ServiceVersion: "0.0.1"}); err != nil {
		fmt.Printf("failed stackdriver.profiler.Start %+v", err)
	}
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: projectID,
	})
	if err != nil {
		panic(err)
	}
	trace.RegisterExporter(exporter)

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

	fsProject := os.Getenv("FIRESTORE_PROJECT")
	if len(fsProject) < 1 {
		fsProject = projectID
	}
	fmt.Printf("FIRESTORE_PROJECT:%s\n", fsProject)

	ctx := context.Background()
	if err := firedb.SetUp(ctx, fsProject); err != nil {
		panic(err)
	}

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

	if *onlyFuncActivate == "" || *onlyFuncActivate == "watchPassivePlayer" {
		fmt.Println("Start WatchPassivePlayer")
		go func() {
			ch <- WatchPassivePlayer()
		}()
	}

	// Debug HTTP Handler
	go func() {
		http.HandleFunc("/", helthHandler)
		http.HandleFunc("/field", fieldHandler)
		http.HandleFunc("/player", playerHandler)
		http.HandleFunc("/healthz", helthHandler)
		if err := http.ListenAndServe(":8080", nil); err != nil {
			panic(err)
		}
	}()

	err = <-ch
	fmt.Printf("%+v", err)
}

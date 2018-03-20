package slog

import (
	"fmt"
	"time"
)

func Example() {
	//if e, g := 1, len(log.Messages); e != g {
	//	t.Fatalf("log.messages.len expected %d; got %d", e, g)
	//}
	//
	//if e, g := "Hello slog World", log.Messages[0]; e != g {
	//	t.Fatalf("log.messages[0] expected %s; got %s", e, g)
	//}
}

func ExampleLog_Infof() {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	log := Start(time.Date(2017, time.April, 1, 13, 15, 30, 45, loc))
	log.Infof("Hello World %d", 1)
	log.Infof("Hello World %d", 2)
	log.Flush()
	// Output: {"timestamp":{"seconds":1491020130,"nanos":45},"message":"[\"Hello World 1\",\"Hello World 2\"]","severity":"INFO","thread":1491020130000000045}
}

func ExampleLog_Info() {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	log := Start(time.Date(2017, time.April, 1, 13, 15, 30, 45, loc))
	log.Info("Hello World 1")
	log.Info("Hello World 2")
	log.Flush()
	// Output: {"timestamp":{"seconds":1491020130,"nanos":45},"message":"[\"Hello World 1\",\"Hello World 2\"]","severity":"INFO","thread":1491020130000000045}
}

func ExampleLog_Errorf() {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	log := Start(time.Date(2017, time.April, 1, 13, 15, 30, 45, loc))
	log.Errorf("Hello World %d", 1)
	log.Errorf("Hello World %d", 2)
	log.Flush()
	// Output: {"timestamp":{"seconds":1491020130,"nanos":45},"message":"[\"Hello World 1\",\"Hello World 2\"]","severity":"ERROR","thread":1491020130000000045}
}

func ExampleLog_Error() {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	log := Start(time.Date(2017, time.April, 1, 13, 15, 30, 45, loc))
	log.Error("Hello World 1")
	log.Error("Hello World 2")
	log.Flush()
	// Output: {"timestamp":{"seconds":1491020130,"nanos":45},"message":"[\"Hello World 1\",\"Hello World 2\"]","severity":"ERROR","thread":1491020130000000045}
}

// ExampleLog_Error2 is Error Levelが優先されることを確認
func ExampleLog_Error_level() {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	log := Start(time.Date(2017, time.April, 1, 13, 15, 30, 45, loc))
	log.Info("Hello Info")
	log.Error("Hello Error")
	log.Info("Hello Info")
	log.Flush()
	// Output: {"timestamp":{"seconds":1491020130,"nanos":45},"message":"[\"Hello Info\",\"Hello Error\",\"Hello Info\"]","severity":"ERROR","thread":1491020130000000045}
}

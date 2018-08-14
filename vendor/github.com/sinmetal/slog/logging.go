package slog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"
)

// StackdriverLogEntry is Stackdriver Logging Entry
type StackdriverLogEntry struct {
	Severity    string  `json:"severity"`
	LogName     string  `json:"logName"`
	Lines       []Line  `json:"lines"`
	FlushCaller *Caller `json:"flushCaller"`
}

// Line is Application Log Entry
// Stackdriver Logging JSON Payload
type Line struct {
	Severity  string    `json:"severity"`
	Name      string    `json:"name"`
	Body      string    `json:"body"`
	Timestamp time.Time `json:"timestamp"`
}

// KV is Line Bodyに利用するKey Value struct
type KV struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// Caller is 関数を呼び出したファイルや行数, 関数名を収めるstruct
type Caller struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Func string `json:"func"`
}

type contextLogKey struct{}

// WithLog is context.ValueにLogを入れたものを返す
// Log周期開始時に利用する
func WithLog(ctx context.Context) context.Context {
	_, ok := ctx.Value(contextLogKey{}).(*StackdriverLogEntry)
	if ok {
		return ctx
	}

	l := &StackdriverLogEntry{
		Lines: []Line{},
	}

	return context.WithValue(ctx, contextLogKey{}, l)
}

// SetLogName is set LogName
func SetLogName(ctx context.Context, logName string) {
	l, ok := ctx.Value(contextLogKey{}).(*StackdriverLogEntry)
	if !ok {
		panic(fmt.Sprintf("not contain log. logName = %+v", logName))
	}
	l.LogName = logName
}

// Debug is output info level Log
func Debug(ctx context.Context, name string, body interface{}) {
	log(ctx, "DEBUG", name, body)
}

// Info is output info level Log
func Info(ctx context.Context, name string, body interface{}) {
	log(ctx, "INFO", name, body)
}

// Warning is output info level Log
func Warning(ctx context.Context, name string, body interface{}) {
	log(ctx, "WARNING", name, body)
}

// Error is output info level Log
func Error(ctx context.Context, name string, body interface{}) {
	log(ctx, "ERROR", name, body)
}

func log(ctx context.Context, logLevel string, name string, body interface{}) {
	l, ok := ctx.Value(contextLogKey{}).(*StackdriverLogEntry)
	if !ok {
		panic(fmt.Sprintf("not contain log. body = %+v", body))
	}
	l.Severity = maxSeverity(l.Severity, logLevel)
	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	l.Lines = append(l.Lines, Line{
		Severity:  logLevel,
		Name:      name,
		Body:      string(b),
		Timestamp: time.Now(),
	})
}

// Flush is ログを出力する
func Flush(ctx context.Context) {
	l, ok := ctx.Value(contextLogKey{}).(*StackdriverLogEntry)
	if ok {
		if len(l.Lines) < 1 {
			return
		}

		pt, file, line, ok := runtime.Caller(1)
		if !ok {
			fmt.Println("スタックトレースの取得失敗")
			return
		}
		funcName := runtime.FuncForPC(pt).Name()
		l.FlushCaller = &Caller{
			File: file,
			Line: line,
			Func: funcName,
		}
		encoder := json.NewEncoder(os.Stdout)
		if err := encoder.Encode(l); err != nil {
			_, err := os.Stdout.WriteString(err.Error())
			if err != nil {
				panic(err)
			}
		}
	}
}

func maxSeverity(severities ...string) (severity string) {
	severityLevel := make(map[string]int)
	severityLevel["DEFAULT"] = 0
	severityLevel["DEBUG"] = 100
	severityLevel["INFO"] = 200
	severityLevel["NOTICE"] = 300
	severityLevel["WARNING"] = 400
	severityLevel["ERROR"] = 500
	severityLevel["CRITICAL"] = 600
	severityLevel["ALERT"] = 700
	severityLevel["EMERGENCY"] = 800

	level := -1
	for _, s := range severities {
		lv, ok := severityLevel[s]
		if !ok {
			lv = -1
		}
		if lv > level {
			severity = s
			level = lv
		}
	}

	return severity
}

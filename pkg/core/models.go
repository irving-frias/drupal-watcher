package core

import "time"

type Op int

const (
	Create Op = iota
	Write
	Remove
	Rename
	Chmod
)

type FileEvent struct {
	Path      string
	Op        Op
	IsDir     bool
}

type ExecutionResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
	Command  string
}

type EngineEventType int

const (
	EventChange EngineEventType = iota
	EventCacheClear
	EventError
)

type EngineEvent struct {
	Type      EngineEventType
	File      string
	Changes   int
	Commands  string
	ExitCode  int
	Duration  time.Duration
	Stderr    string
	SiteName  string
	Error     error
	Timestamp time.Time
}

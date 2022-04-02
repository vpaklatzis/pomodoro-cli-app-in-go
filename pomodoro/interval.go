package pomodoro

import (
	"errors"
	"time"
)

// Category constants
const (
	CategoryPomodoro   = "Pomodoro"
	CategoryShortBreak = "ShortBreak"
	CategoryLongBreak  = "LongBreak"
)

// State constants
// iota operator - first value starts at 0, and then the following value 
// gets incremented by 1 for each following constant
const (
	StateNotStarted = iota
	StateDone
	StateRunning
	StatePaused
	StateCancelled
)

type Interval struct {
	ID 				int64
	StartTime       time.Time
	PlannedDuration time.Duration
	ActualDuration  time.Duration
	Category        string
	State 		    int
}

type Repository interface {
	Create(i Interval) (int64, error)
	Update(i Interval) error
	ByID(id int64) (Interval, error)
	Last() (Interval, error)
	Breaks(n int) ([]Interval, error)
}

// custom errors that may occur in business logic
var (
	ErrNoIntervals = errors.New("No intervals")
	ErrIntervalNotRunning = errors.New("Interval not running")
	ErrIntervalCompleted = errors.New("Interval is completed or cancelled")
	ErrInvalidState = errors.New("Invalid state")
	ErrInvalidID = errors.New("Invalid ID")
)

// info required to instantiate an interval
type IntervalConfig struct {
	repo 			   Repository
	PomodoroDuration   time.Duration
	ShortBreakDuration time.Duration
	LongBreakDuration  time.Duration
}

// Instantiate new IntervalConfig
func NewConfig(repo Repository, pomodoro, shortBreak, longBreak time.Duration) *IntervalConfig {
	config := &IntervalConfig{
		repo: repo,
		PomodoroDuration:   25 * time.Minute,
		ShortBreakDuration: 5 * time.Minute,
		LongBreakDuration:  15 * time.Minute,
	}

	if pomodoro > 0 { config.PomodoroDuration = pomodoro }

	if shortBreak > 0 { config.ShortBreakDuration = shortBreak }

	if longBreak > 0 { config.LongBreakDuration = longBreak }

	return config
}
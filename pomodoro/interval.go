package pomodoro

import (
	"context"
	"errors"
	"time"
)

// category constants
const (
	CategoryPomodoro   = "Pomodoro"
	CategoryShortBreak = "ShortBreak"
	CategoryLongBreak  = "LongBreak"
)

// state constants
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

// instantiate new IntervalConfig
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
/**
* Takes a reference to the repository as input and returns
* the next interval category as a string or an error.
* Retrieves the last interval from the repository
* and determines the next interval category.
* After each Pomodoro interval, there’s a short break
* and after four Pomodoros, there’s a long break.
*/
func nextCategory(r Repository) (string, error) {
	lastInterval, err := r.Last()
	if err != nil && err == ErrNoIntervals {
		return CategoryPomodoro, nil
	}

	if err != nil {
		return "", err
	}

	if lastInterval.Category == CategoryLongBreak || lastInterval.Category == CategoryShortBreak {
		return CategoryPomodoro, nil
	}

	lastBreaks, err := r.Breaks(3)
	if err != nil {
		return "", err
	}

	if len(lastBreaks) < 3 {
		return CategoryShortBreak, err
	}

	for _, i := range lastBreaks {
		if i.Category == CategoryLongBreak {
			return CategoryShortBreak, nil
		}
	}

	return CategoryLongBreak, nil
}

// used to perform tasks while the interval executes
type Callback func(Interval)

/**
* This function uses the time.Ticker type and a loop to execute actions every
* second while the interval time progresses. It uses a select statement to take
* actions, executing periodically when the time.Ticker goes off,
* finishing successfully when the interval time expires
* or canceling when a signal is received from Context.
*/
func tick(ctx context.Context, id int64, config *IntervalConfig, start, periodic, end Callback) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	interval, err := config.repo.ByID(id)
	if err != nil {
		return err
	}

	expire := time.After(interval.PlannedDuration - interval.ActualDuration)

	start(interval)

	for {
		select {
		case <- ticker.C:
			interval, err := config.repo.ByID(id)
			if err != nil {
				return err
			}

			if interval.State == StatePaused {
				return nil
			}

			interval.ActualDuration += time.Second
			if err := config.repo.Update(interval); err != nil {
				return err
			}

			periodic(interval)
		case <- expire:
			interval, err := config.repo.ByID(id)
			if err != nil {
				return err
			}

			interval.State = StateDone

			end(interval)

			return config.repo.Update(interval)
		case <- ctx.Done():
			interval, err := config.repo.ByID(id)
			if err != nil {
				return err
			}
			interval.State = StateCancelled
		}
	}
}
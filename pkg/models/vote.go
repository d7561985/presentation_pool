package models

type Step struct {
	Question string
	Option   []string
}

type Vote struct {
	Name string

	Steps []Step
}

type StatusData struct {
	VoteName string
	Step     int64
	Status   string
}

const (
	StatusIdle       = "IDLE"
	StatusInProgress = "IN_PROGRESS"
	StatusComplete   = "COMPLETE"
)

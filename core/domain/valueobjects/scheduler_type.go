package valueobjects

// SchedulerType represents the spaced repetition algorithm type
type SchedulerType string

const (
	// SchedulerTypeSM2 represents the SuperMemo 2 algorithm
	SchedulerTypeSM2 SchedulerType = "sm2"
	// SchedulerTypeFSRS represents the Free Spaced Repetition Scheduler algorithm
	SchedulerTypeFSRS SchedulerType = "fsrs"
)

// IsValid checks if the scheduler type is valid
func (s SchedulerType) IsValid() bool {
	return s == SchedulerTypeSM2 || s == SchedulerTypeFSRS
}

// String returns the string representation of the scheduler type
func (s SchedulerType) String() string {
	return string(s)
}


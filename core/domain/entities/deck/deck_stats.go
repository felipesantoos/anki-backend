package deck

// DeckStats represents the study statistics for a deck
type DeckStats struct {
	DeckID         int64
	NewCount       int
	LearningCount  int
	ReviewCount    int
	SuspendedCount int
	NotesCount     int
	DueTodayCount  int
}


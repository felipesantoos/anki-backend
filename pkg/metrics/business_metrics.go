package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// BusinessMetrics holds all business domain-related Prometheus metrics
type BusinessMetrics struct {
	DecksCreatedTotal   *prometheus.CounterVec
	CardsCreatedTotal   *prometheus.CounterVec
	NotesCreatedTotal   *prometheus.CounterVec
	StudySessionsTotal  *prometheus.CounterVec
	CardReviewsTotal    *prometheus.CounterVec
}

// NewBusinessMetrics creates a new BusinessMetrics instance with all business metrics configured
func NewBusinessMetrics() *BusinessMetrics {
	return &BusinessMetrics{
		DecksCreatedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "decks_created_total",
				Help: "Total number of decks created",
			},
			[]string{"user_id"},
		),
		CardsCreatedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cards_created_total",
				Help: "Total number of cards created",
			},
			[]string{"deck_id"},
		),
		NotesCreatedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "notes_created_total",
				Help: "Total number of notes created",
			},
			[]string{"user_id", "deck_id"},
		),
		StudySessionsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "study_sessions_total",
				Help: "Total number of study sessions",
			},
			[]string{"user_id"},
		),
		CardReviewsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "card_reviews_total",
				Help: "Total number of card reviews",
			},
			[]string{"user_id", "deck_id", "rating"},
		),
	}
}

// Register registers all business metrics with the given Prometheus registry
func (b *BusinessMetrics) Register(registry *prometheus.Registry) error {
	registry.MustRegister(b.DecksCreatedTotal)
	registry.MustRegister(b.CardsCreatedTotal)
	registry.MustRegister(b.NotesCreatedTotal)
	registry.MustRegister(b.StudySessionsTotal)
	registry.MustRegister(b.CardReviewsTotal)
	return nil
}

// RegisterBusinessMetrics registers business domain metrics with the given Prometheus registry
// This function is kept for backward compatibility but is deprecated.
// Use NewBusinessMetrics() and Register() instead.
func RegisterBusinessMetrics(registry *prometheus.Registry) error {
	businessMetrics := NewBusinessMetrics()
	return businessMetrics.Register(registry)
}

package events

import "time"

type UserEvent struct {
	Event      string    `json:"event"`
	UserID     string    `json:"user_id"`
	Email      string    `json:"email,omitempty"`
	OccurredAt time.Time `json:"occurred_at"`
	TraceID    string    `json:"trace_id"`
}

func NewUserEvent(event, userID, email, traceID string) UserEvent {
	return UserEvent{
		Event:      event,
		UserID:     userID,
		Email:      email,
		OccurredAt: time.Now().UTC(),
		TraceID:    traceID,
	}
}

package models

import "time"

type Decision struct {
	ActorUserId     string
	RecipientUserId string
	LikedRecipient  bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

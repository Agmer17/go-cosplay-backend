package shared

import (
	"time"

	"github.com/google/uuid"
)

type SessionsData struct {
	SessionId string    `json:"session_id"`
	UserId    uuid.UUID `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type UserCredential struct {
	UserId uuid.UUID `json:"user_id"`
	Status string    `json:"status"`
	Role   string    `json:"role"`
}

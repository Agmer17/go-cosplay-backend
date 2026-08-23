package shared

import (
	"time"

	"github.com/google/uuid"
)

type SessionsData struct {
	SessionId string    `json:"session_id"`
	UserId    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
}

package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Grid struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	Title     string
	Data      json.RawMessage
	CreatedAt time.Time
}

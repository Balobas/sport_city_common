package outboxEntity

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Message struct {
	Uid              uuid.UUID
	SubjectName      string
	Payload          []byte
	CreatedAt        time.Time
	UpdatedAt        time.Time
	LastErrorMessage string
	SendAt           time.Time
}

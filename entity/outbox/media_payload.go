package outboxEntity

import uuid "github.com/satori/go.uuid"

type MediaUsagePayload struct {
	BaseMsgPayload
	Domain      string      `json:"domain"`
	FirstlyUsed []uuid.UUID `json:"firstlyUsedMedia"`
	Unused      []uuid.UUID `json:"unusedMedia"`
}

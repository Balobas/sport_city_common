package outboxEntity

import uuid "github.com/satori/go.uuid"

type BaseMsgPayload struct {
	MsgUid  uuid.UUID `json:"msgUid"`
	TraceId string    `json:"traceId"`
}

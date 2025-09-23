package repositoryBaseEntityPostgres

import (
	"time"

	sq "github.com/Masterminds/squirrel"
	outboxEntity "github.com/balobas/sport_city_common/entity/outbox"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
)

type OutboxMessageRow struct {
	Uid              pgtype.UUID
	SubjectName      string
	Payload          string
	CreatedAt        pgtype.Timestamp
	UpdatedAt        pgtype.Timestamp
	LastErrorMessage string
	SendAt           pgtype.Timestamp
}

func NewOutboxMessageRow() *OutboxMessageRow {
	return &OutboxMessageRow{}
}

func (m *OutboxMessageRow) New() *OutboxMessageRow {
	return &OutboxMessageRow{}
}

func (m *OutboxMessageRow) FromEntity(mqMessage outboxEntity.Message) *OutboxMessageRow {
	m.Uid = pgtype.UUID{
		Bytes:  mqMessage.Uid,
		Status: pgtype.Present,
	}
	m.SubjectName = mqMessage.SubjectName
	m.Payload = string(mqMessage.Payload)
	m.LastErrorMessage = mqMessage.LastErrorMessage

	m.CreatedAt = pgtype.Timestamp{
		Time:   mqMessage.CreatedAt,
		Status: pgtype.Present,
	}
	m.UpdatedAt = pgtype.Timestamp{Time: mqMessage.UpdatedAt, Status: pgtype.Present}
	if mqMessage.UpdatedAt.Equal(time.Time{}) {
		m.UpdatedAt.Status = pgtype.Null
	}
	m.SendAt = pgtype.Timestamp{Time: mqMessage.SendAt, Status: pgtype.Present}
	if mqMessage.SendAt.Equal(time.Time{}) {
		m.SendAt.Status = pgtype.Null
	}

	return m
}

func (m *OutboxMessageRow) ToEntity() outboxEntity.Message {
	return outboxEntity.Message{
		Uid:              m.Uid.Bytes,
		SubjectName:      m.SubjectName,
		Payload:          []byte(m.Payload),
		LastErrorMessage: m.LastErrorMessage,
		CreatedAt:        m.CreatedAt.Time,
		UpdatedAt:        m.UpdatedAt.Time,
		SendAt:           m.SendAt.Time,
	}
}

func (m *OutboxMessageRow) IdColumnName() string {
	return "uid"
}

func (m *OutboxMessageRow) Values() []interface{} {
	return []interface{}{
		m.Uid, m.SubjectName, m.Payload, m.CreatedAt, m.UpdatedAt, m.LastErrorMessage, m.SendAt,
	}
}

func (m *OutboxMessageRow) Columns() []string {
	return []string{
		"uid", "subject_name", "payload", "created_at", "updated_at", "last_error_msg", "send_at",
	}
}

func (m *OutboxMessageRow) Table() string {
	return "outbox_messages"
}

func (m *OutboxMessageRow) Scan(row pgx.Row) error {
	return row.Scan(&m.Uid, &m.SubjectName, &m.Payload, &m.CreatedAt, &m.UpdatedAt, &m.LastErrorMessage, &m.SendAt)
}

func (m *OutboxMessageRow) ColumnsForUpdate() []string {
	return []string{
		"updated_at", "last_error_msg", "send_at",
	}
}

func (m *OutboxMessageRow) ValuesForUpdate() []interface{} {
	return []interface{}{
		m.UpdatedAt, m.LastErrorMessage, m.SendAt,
	}
}

func (m *OutboxMessageRow) ConditionUidEqual() sq.Eq {
	return sq.Eq{"uid": m.Uid}
}

func (m *OutboxMessageRow) ConditionSendAtIsNull() sq.Eq {
	return sq.Eq{"send_at": pgtype.Timestamp{Status: pgtype.Null}}
}

func NewOutboxMessageRows() *Rows[*OutboxMessageRow, outboxEntity.Message] {
	return &Rows[*OutboxMessageRow, outboxEntity.Message]{}
}

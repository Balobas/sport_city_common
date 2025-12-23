package riverOutboxPublisher

import (
	"context"
)

type Publisher interface {
	Publish(ctx context.Context, subjectName string, data []byte) error
}

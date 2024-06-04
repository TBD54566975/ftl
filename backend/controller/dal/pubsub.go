package dal

import (
	"context"

	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/internal/model"
)

func (d *DAL) PublishEventForTopic(ctx context.Context, module, topic string, payload []byte) error {
	err := d.db.PublishEventForTopic(ctx, sql.PublishEventForTopicParams{
		Key:     model.NewTopicEventKey(module, topic),
		Module:  module,
		Topic:   topic,
		Payload: payload,
	})
	if err != nil {
		return translatePGError(err)
	}
	return nil
}

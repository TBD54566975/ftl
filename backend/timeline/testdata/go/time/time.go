package time

import (
	"context"
	"time"

	"github.com/block/ftl/go-runtime/ftl"
)

type TimeRequest struct{}
type TimeResponse struct {
	Time time.Time
}

// Time returns the current time.
//
//ftl:verb export
func Time(ctx context.Context, req TimeRequest) (TimeResponse, error) {
	return TimeResponse{Time: time.Now()}, nil
}

//ftl:verb
func Internal(ctx context.Context, req TimeRequest) (TimeResponse, error) {
	return TimeResponse{Time: time.Now()}, nil
}

type PublishInvoiceRequest struct {
	Amount int
}

//ftl:verb
func PublishInvoice(ctx context.Context, req PublishInvoiceRequest, invoices Invoices) error {
	invoices.Publish(ctx, Invoice{Amount: req.Amount})
	return nil
}

type Invoice struct {
	Amount int
}

//ftl:export
type Invoices = ftl.TopicHandle[Invoice, ftl.SinglePartitionMap[Invoice]]

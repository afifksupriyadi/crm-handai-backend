package model

import (
	"time"

	"github.com/uptrace/bun"
)

// CustomerSegment represents customer segmentation based on prediction accuracy
type CustomerSegment struct {
	bun.BaseModel `bun:"table:analytics.customer_segments,alias:cs"`

	CustomerID                    int       `bun:"customer_id,pk" json:"customer_id"`
	Segment                       string    `bun:"segment,notnull" json:"segment"` // LOYAL, CHURN, REGULAR
	ConsecutiveCorrectPredictions int       `bun:"consecutive_correct_predictions,notnull,default:0" json:"consecutive_correct_predictions"`
	TotalPredictions              int       `bun:"total_predictions,notnull,default:0" json:"total_predictions"`
	TotalCorrectPredictions       int       `bun:"total_correct_predictions,notnull,default:0" json:"total_correct_predictions"`
	LastUpdatedAt                 time.Time `bun:"last_updated_at,notnull,default:current_timestamp" json:"last_updated_at"`
	UpdatedByBatchID              int       `bun:"updated_by_batch_id" json:"updated_by_batch_id,omitempty"`

	// Relation
	Customer *Customer `bun:"rel:belongs-to,join:customer_id=id" json:"customer,omitempty"`
}

// Segment constants
const (
	SegmentLoyal   = "LOYAL"
	SegmentChurn   = "CHURN"
	SegmentRegular = "REGULAR"
)

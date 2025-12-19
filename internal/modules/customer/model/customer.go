package model

import (
	"time"

	transactionModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/transactions/model"
	"github.com/uptrace/bun"
)

// Customer represents a customer in the system (operational data only)
type Customer struct {
	bun.BaseModel `bun:"table:customers,alias:c"`

	ID                int        `bun:"id,pk,autoincrement"`
	Name              string     `bun:"name,notnull"`
	Phone             string     `bun:"phone,notnull,unique"`
	CreatedAt         time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt         *time.Time `bun:"updated_at"`
	DeletedAt         *time.Time `bun:"deleted_at"`
	UpgradedFromGuest bool       `bun:"upgraded_from_guest,notnull,default:false"`
	UpgradedAt        *time.Time `bun:"upgraded_at"`
	FirstSeenAsGuest  *time.Time `bun:"first_seen_as_guest"`

	// Relations
	Transactions       []*transactionModel.Transaction `bun:"rel:has-many,join:id=customer_id"`
	CustomerMetrics    []*CustomerMetric               `bun:"rel:has-many,join:id=customer_id"`
	CustomerPrediction *CustomerPrediction             `bun:"rel:has-one,join:id=customer_id"`
	ChurnAlerts        []*ChurnAlert                   `bun:"rel:has-many,join:id=customer_id"`
}

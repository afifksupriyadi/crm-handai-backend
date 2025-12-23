package model

import (
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/util/request"
)

// GetCustomerPredictionsRequest represents request to get customer predictions
type GetCustomerPredictionsRequest struct {
	request.AuthorizedRequest
	ID    int `path:"id" validate:"required" doc:"Customer ID"`
	Limit int `query:"limit" default:"4" minimum:"1" maximum:"10" doc:"Number of predictions to return"`
}

// CustomerPredictionResponse represents a single prediction record
type CustomerPredictionResponse struct {
	ID                        int        `json:"id"`
	LastTransactionDate       string     `json:"last_transaction_date"`
	PredictedNextPurchaseDate string     `json:"predicted_next_purchase_date"`
	ActualNextPurchaseDate    *string    `json:"actual_next_purchase_date,omitempty"`
	IsPredictedCorrect        *bool      `json:"is_predicted_correct,omitempty"`
	CreatedAt                 time.Time  `json:"created_at"`
	ValidatedAt               *time.Time `json:"validated_at,omitempty"`
}

// CustomerPredictionListResponse represents list of predictions
type CustomerPredictionListResponse struct {
	Data []*CustomerPredictionResponse `json:"data"`
}

package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/repository"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
)

// ProductPredictionCalculatorServiceImpl handles product prediction calculations
type ProductPredictionCalculatorServiceImpl struct {
	predictedProductRepo repository.CustomerPredictedProductRepository
}

// NewProductPredictionCalculatorService creates a new ProductPredictionCalculatorService
func NewProductPredictionCalculatorService(
	predictedProductRepo repository.CustomerPredictedProductRepository,
) customer.ProductPredictionCalculatorService {
	return &ProductPredictionCalculatorServiceImpl{
		predictedProductRepo: predictedProductRepo,
	}
}

// ProductAnalysis represents analyzed product purchase data
type ProductAnalysis struct {
	ProductID       int
	VariantID       *int
	ProductName     string
	VariantName     *string
	PurchaseCount   int     // How many transactions included this product
	TotalQuantity   int     // Total quantity purchased
	AvgQuantity     float64 // Average quantity per transaction
	ConfidenceScore float64 // Percentage (0-100)
}

// CalculateProductPredictions analyzes last 10 transactions and predicts products
func (s *ProductPredictionCalculatorServiceImpl) CalculateProductPredictions(ctx context.Context, customerID int, predictionID int) ([]*model.CustomerPredictedProduct, error) {
	log := logger.FromContext(ctx, 2)

	// STEP 1: Get transaction products from repository (NO QUERY IN SERVICE!)
	txProducts, err := s.predictedProductRepo.GetCustomerTransactionProducts(ctx, customerID, 10)
	if err != nil {
		return nil, err
	}

	if len(txProducts) == 0 {
		log.Debug().Int("customer_id", customerID).Msg("No transaction history for product prediction")
		return []*model.CustomerPredictedProduct{}, nil
	}

	// STEP 2: Get unique transactions (last 10 only)
	uniqueTxCodes := make(map[string]bool)
	var lastTxCodes []string
	for _, tp := range txProducts {
		if !uniqueTxCodes[tp.TransactionCode] {
			uniqueTxCodes[tp.TransactionCode] = true
			lastTxCodes = append(lastTxCodes, tp.TransactionCode)
			if len(lastTxCodes) >= 10 {
				break
			}
		}
	}

	totalTransactions := len(lastTxCodes)
	log.Debug().Int("total_transactions", totalTransactions).Msg("Analyzing transactions")

	// STEP 3: Filter products to only those in last 10 transactions
	var filteredProducts []model.TransactionProduct
	for _, tp := range txProducts {
		for _, code := range lastTxCodes {
			if tp.TransactionCode == code {
				filteredProducts = append(filteredProducts, tp)
				break
			}
		}
	}

	// STEP 4: Analyze product frequency and quantity
	productMap := make(map[string]*ProductAnalysis) // Key: "productID_variantID"

	for _, tp := range filteredProducts {
		// Create unique key for product+variant combination
		key := fmt.Sprintf("%d_%v", tp.ProductID, tp.VariantID)

		if analysis, exists := productMap[key]; exists {
			// Product already tracked, increment
			analysis.TotalQuantity += tp.Quantity
			analysis.PurchaseCount++
		} else {
			// New product
			productMap[key] = &ProductAnalysis{
				ProductID:     tp.ProductID,
				VariantID:     tp.VariantID,
				ProductName:   tp.ProductName,
				VariantName:   tp.VariantName,
				PurchaseCount: 1,
				TotalQuantity: tp.Quantity,
			}
		}
	}

	// STEP 5: Calculate metrics
	var analyses []*ProductAnalysis
	for _, analysis := range productMap {
		// Average quantity per transaction
		analysis.AvgQuantity = float64(analysis.TotalQuantity) / float64(analysis.PurchaseCount)

		// Confidence score = (purchase count / total transactions) * 100
		analysis.ConfidenceScore = (float64(analysis.PurchaseCount) / float64(totalTransactions)) * 100

		analyses = append(analyses, analysis)
	}

	// STEP 6: Sort by confidence score (highest first)
	sort.Slice(analyses, func(i, j int) bool {
		return analyses[i].ConfidenceScore > analyses[j].ConfidenceScore
	})

	// STEP 7: Filter top products (min 30% confidence, max 5 products)
	var predictedProducts []*model.CustomerPredictedProduct
	maxProducts := 5
	minConfidence := 30.0

	for i, analysis := range analyses {
		if i >= maxProducts {
			break
		}

		if analysis.ConfidenceScore < minConfidence {
			break
		}

		// Round quantity to nearest integer
		predictedQty := int(analysis.AvgQuantity + 0.5)
		if predictedQty < 1 {
			predictedQty = 1
		}

		predictedProducts = append(predictedProducts, &model.CustomerPredictedProduct{
			PredictionID:      predictionID,
			ProductID:         analysis.ProductID,
			VariantID:         analysis.VariantID,
			PredictedQuantity: predictedQty,
			ConfidenceScore:   analysis.ConfidenceScore,
			PurchaseFrequency: analysis.PurchaseCount,
			AvgQuantity:       analysis.AvgQuantity,
		})
	}

	log.Info().
		Int("customer_id", customerID).
		Int("analyzed_transactions", totalTransactions).
		Int("unique_products", len(productMap)).
		Int("predicted_products", len(predictedProducts)).
		Msg("Product predictions calculated")

	return predictedProducts, nil
}

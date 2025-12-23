package handler

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/constant"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/parser"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
	"github.com/gofiber/fiber/v2"
)

type ImportHandler struct {
	importService importdata.ImportService
}

func NewImportHandler(importService importdata.ImportService) *ImportHandler {
	return &ImportHandler{
		importService: importService,
	}
}

// HandleImportBatch handles batch import (transaction required, customer optional)
func (h *ImportHandler) HandleImportBatch(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Minute)
	defer cancel()

	log := logger.FromContext(ctx, 2)

	startDateStr := c.FormValue("start_date")
	endDateStr := c.FormValue("end_date")

	if startDateStr == "" {
		errResp := response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrStartDateRequired, "start_date is required"))
		return c.Status(errResp.Status).JSON(errResp.Body)
	}

	if endDateStr == "" {
		errResp := response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrEndDateRequired, "end_date is required"))
		return c.Status(errResp.Status).JSON(errResp.Body)
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		errResp := response.BuildError(ctx, response.WrapAppError(ctx, err, response.ErrInvalidDateFormat, "Invalid start_date format, expected YYYY-MM-DD"))
		return c.Status(errResp.Status).JSON(errResp.Body)
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		errResp := response.BuildError(ctx, response.WrapAppError(ctx, err, response.ErrInvalidDateFormat, "Invalid end_date format, expected YYYY-MM-DD"))
		return c.Status(errResp.Status).JSON(errResp.Body)
	}

	if endDate.Before(startDate) {
		errResp := response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrInvalidDateRange, "end_date must be >= start_date"))
		return c.Status(errResp.Status).JSON(errResp.Body)
	}

	log.Info().Str("start_date", startDate.Format("2006-01-02")).Str("end_date", endDate.Format("2006-01-02")).Msg("Date range validated")

	// 1. Get notes (optional)
	notes := c.FormValue("notes")

	// 2. Get transaction file (REQUIRED)
	transactionFile, err := c.FormFile("file_transaction")
	if err != nil {
		errResp := response.BuildError(ctx, response.WrapAppError(ctx, err, response.ErrTransactionFileRequired, "file_transaction is required"))
		return c.Status(errResp.Status).JSON(errResp.Body)
	}

	// 3. Get customer file (OPTIONAL)
	customerFile, _ := c.FormFile("file_customer")

	// 4. VALIDATION: Validate transaction filename format
	_, err = parser.ExtractBatchDateFromFilename(transactionFile.Filename, string(constant.ImportTypeTransaction))
	if err != nil {
		errResp := response.BuildError(ctx, response.WrapAppError(ctx, err, response.ErrInvalidFilenameFormat, err.Error()))
		return c.Status(errResp.Status).JSON(errResp.Body)
	}

	// 5. VALIDATION: Validate customer filename format (if provided)
	var customerDate *time.Time
	if customerFile != nil {
		date, err := parser.ExtractBatchDateFromFilename(customerFile.Filename, string(constant.ImportTypeCustomer))
		if err != nil {
			errResp := response.BuildError(ctx, response.WrapAppError(ctx, err, response.ErrInvalidFilenameFormat, err.Error()))
			return c.Status(errResp.Status).JSON(errResp.Body)
		}
		customerDate = &date
	}

	// 6. VALIDATION: Check date order (if customer file provided)
	if customerDate != nil {
		transactionDate, _ := parser.ExtractBatchDateFromFilename(transactionFile.Filename, string(constant.ImportTypeTransaction))
		if transactionDate.After(*customerDate) {
			errMsg := fmt.Sprintf(
				"Transaction date (%s) cannot be later than customer date (%s). This could result in missing customer data",
				transactionDate.Format("2006-01-02"),
				customerDate.Format("2006-01-02"),
			)
			errResp := response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrTransactionDateExceedsCustomer, errMsg))
			return c.Status(errResp.Status).JSON(errResp.Body)
		}
	}

	// 7. Open files
	var customerSrc multipart.File
	var customerFilename string

	if customerFile != nil {
		src, err := customerFile.Open()
		if err != nil {
			errResp := response.BuildError(ctx, response.WrapAppError(ctx, err, response.ErrInvalidExcelFormat, "Failed to open customer file"))
			return c.Status(errResp.Status).JSON(errResp.Body)
		}
		defer src.Close()
		customerSrc = src
		customerFilename = customerFile.Filename
	}

	transactionSrc, err := transactionFile.Open()
	if err != nil {
		errResp := response.BuildError(ctx, response.WrapAppError(ctx, err, response.ErrInvalidExcelFormat, "Failed to open transaction file"))
		return c.Status(errResp.Status).JSON(errResp.Body)
	}
	defer transactionSrc.Close()

	data, err := h.importService.ImportBatch(
		ctx,
		customerSrc,
		transactionSrc,
		customerFilename,
		transactionFile.Filename,
		notes,
		startDate,
		endDate,
	)
	if err != nil {
		errResp := response.BuildError(ctx, err)
		return c.Status(errResp.Status).JSON(errResp.Body)
	}

	// 9. Success response
	successResp := response.BuildSuccess(data, response.SuccessImportBatch)
	return c.Status(successResp.Status).JSON(successResp.Body)
}

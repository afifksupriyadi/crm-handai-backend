package handler

import (
	"mime/multipart"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata"
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

// ==========================================
// NEW BATCH IMPORT ENDPOINT
// ==========================================

// HandleImportBatch handles batch import (transaction required, customer optional)
func (h *ImportHandler) HandleImportBatch(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get batch_date from form
	batchDate := c.FormValue("batch_date")
	if batchDate == "" {
		errResp := response.BuildError(ctx, response.WrapAppError(ctx, nil, response.ErrInvalidBatchDate, "batch_date is required"))
		return c.Status(errResp.Status).JSON(errResp.Body)
	}

	// Get notes (optional)
	notes := c.FormValue("notes")

	// Get overwrite_if_exist flag (optional, default false)
	overwriteIfExist := c.FormValue("overwrite_if_exist") == "true"

	// Get transaction file (REQUIRED)
	transactionFile, err := c.FormFile("file_transaction")
	if err != nil {
		errResp := response.BuildError(ctx, response.WrapAppError(ctx, err, response.ErrTransactionFileRequired, "file_transaction is required"))
		return c.Status(errResp.Status).JSON(errResp.Body)
	}

	// Get customer file (OPTIONAL)
	customerFile, _ := c.FormFile("file_customer") // No error check, it's optional

	var customerSrc multipart.File
	var customerFilename string

	// Open customer file if provided
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

	// Open transaction file
	transactionSrc, err := transactionFile.Open()
	if err != nil {
		errResp := response.BuildError(ctx, response.WrapAppError(ctx, err, response.ErrInvalidExcelFormat, "Failed to open transaction file"))
		return c.Status(errResp.Status).JSON(errResp.Body)
	}
	defer transactionSrc.Close()

	// Import batch
	data, err := h.importService.ImportBatch(
		ctx,
		customerSrc, // can be nil
		transactionSrc,
		customerFilename, // can be empty
		transactionFile.Filename,
		batchDate,
		notes,
		overwriteIfExist,
	)
	if err != nil {
		errResp := response.BuildError(ctx, err)
		return c.Status(errResp.Status).JSON(errResp.Body)
	}

	successResp := response.BuildSuccess(data, response.SuccessImportBatch)
	return c.Status(successResp.Status).JSON(successResp.Body)
}

// ==========================================
// LEGACY ENDPOINTS (Backward Compatibility)
// ==========================================

// HandleImportCustomers handles customer import from Excel file (LEGACY)
func (h *ImportHandler) HandleImportCustomers(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		errResp := response.BuildError(ctx, response.WrapAppError(ctx, err, response.ErrEmptyRequestBody, "File is required"))
		return c.Status(errResp.Status).JSON(errResp.Body)
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		errResp := response.BuildError(ctx, response.WrapAppError(ctx, err, response.ErrInvalidExcelFormat, "Failed to open file"))
		return c.Status(errResp.Status).JSON(errResp.Body)
	}
	defer src.Close()

	// Import customers
	data, err := h.importService.ImportCustomers(ctx, src, file.Filename)
	if err != nil {
		errResp := response.BuildError(ctx, err)
		return c.Status(errResp.Status).JSON(errResp.Body)
	}

	successResp := response.BuildSuccess(data, response.SuccessImportCustomers)
	return c.Status(successResp.Status).JSON(successResp.Body)
}

// HandleImportTransactions handles transaction import from Excel file (LEGACY)
func (h *ImportHandler) HandleImportTransactions(c *fiber.Ctx) error {
	ctx := c.Context()

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		errResp := response.BuildError(ctx, response.WrapAppError(ctx, err, response.ErrEmptyRequestBody, "File is required"))
		return c.Status(errResp.Status).JSON(errResp.Body)
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		errResp := response.BuildError(ctx, response.WrapAppError(ctx, err, response.ErrInvalidExcelFormat, "Failed to open file"))
		return c.Status(errResp.Status).JSON(errResp.Body)
	}
	defer src.Close()

	// Import transactions
	data, err := h.importService.ImportTransactions(ctx, src, file.Filename)
	if err != nil {
		errResp := response.BuildError(ctx, err)
		return c.Status(errResp.Status).JSON(errResp.Body)
	}

	successResp := response.BuildSuccess(data, response.SuccessImportTransactions)
	return c.Status(successResp.Status).JSON(successResp.Body)
}

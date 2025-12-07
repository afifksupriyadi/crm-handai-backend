// internal/modules/importdata/handler/import_handler.go

package handler

import (
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

// HandleImportCustomers handles customer import from Excel file
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

// HandleImportTransactions handles transaction import from Excel file
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

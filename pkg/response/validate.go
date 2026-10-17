package response

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/hanjo-erp/hanjo/pkg/validator"
)

// ValidateOrFail validates req using pkg/validator.
// Returns ok=false and writes HTTP error (422 for field errors, 400 for bad body) on failure.
//
// Always check !ok:
//
//	if ok, err := response.ValidateOrFail(c, req); !ok {
//	    return err
//	}
func ValidateOrFail(c fiber.Ctx, req any) (ok bool, err error) {
	verr := validator.Struct(req)
	if verr == nil {
		return true, nil
	}

	var fieldErrs validator.FieldErrors
	if errors.As(verr, &fieldErrs) {
		return false, FailWithDetails(c, fiber.StatusUnprocessableEntity, "VALIDATION_ERROR", "validation failed", fieldErrs)
	}
	return false, Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
}

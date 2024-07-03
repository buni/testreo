package wallet_test

import (
	"os"
	"testing"

	"github.com/buni/wallet/internal/pkg/render/errorhandler"
	httpin_integration "github.com/ggicci/httpin/integration" //nolint
	"github.com/go-chi/chi/v5"
)

func TestMain(m *testing.M) {
	httpin_integration.UseGochiURLParam("path", chi.URLParam)
	errorhandler.RegisterErrorHandler("validation_error_handler", errorhandler.ValidationErrorHandler)
	errorhandler.RegisterErrorHandler("validation_field_errors_handler", errorhandler.ValidationFieldErrorsHandler)
	errorhandler.RegisterErrorHandler("validation_field_error_handler", errorhandler.ValidationFieldErrorHandler)
	errorhandler.RegisterErrorHandler("not_found_error_handler", errorhandler.NotFoundErrorHandler)

	code := m.Run()
	os.Exit(code)
}

package requestvalidator

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/buni/wallet/internal/pkg/render"
	"github.com/ettle/strcase"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

var ErrBadTranslatorLocale = errors.New("bad translator locale")

type Validator interface {
	Validate(ctx context.Context, req any) error
}

type PlaygroundValidator struct {
	validator  *validator.Validate
	translator ut.Translator
}

func NewValidator() (*PlaygroundValidator, error) {
	reqValidator := validator.New()

	reqValidator.RegisterTagNameFunc(fieldTagNameFunc)

	translator := en.New()
	univTranslator := ut.New(translator, translator)
	enTranslator, ok := univTranslator.GetTranslator("en")
	if !ok {
		return nil, ErrBadTranslatorLocale
	}

	err := reqValidator.RegisterTranslation("required", enTranslator, func(ut ut.Translator) error {
		return ut.Add("required", "{0} is a required field", true) //nolint:wrapcheck
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field()) //nolint:errcheck
		return t
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register required translation: %w", err)
	}

	return &PlaygroundValidator{validator: reqValidator, translator: enTranslator}, nil
}

func (rv *PlaygroundValidator) Validate(ctx context.Context, req any) error {
	err := rv.validator.StructCtx(ctx, req)
	if err == nil {
		return nil
	}

	var validationError validator.ValidationErrors
	if !errors.As(err, &validationError) {
		return fmt.Errorf("failed to get underlying validation error type: %w", err)
	}

	fieldErrors := make(render.FieldErrors, 0, len(validationError))

	for _, validationError := range validationError {
		fieldErrors = append(fieldErrors, &render.FieldError{
			Field:   validationError.Field(),
			Message: validationError.Translate(rv.translator),
		})
	}

	renderError := render.NewValidationError(fieldErrors...)

	return renderError
}

func fieldTagNameFunc(fld reflect.StructField) string {
	fieldFragments := strings.SplitN(fld.Tag.Get("json"), ",", 2)

	if len(fieldFragments) == 0 || fieldFragments[0] == "-" {
		return strcase.ToSnake(fld.Name)
	}
	return fieldFragments[0]
}

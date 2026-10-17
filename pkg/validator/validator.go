package validator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	govalidator "github.com/go-playground/validator/v10"
	idtranslations "github.com/go-playground/validator/v10/translations/id"
)

var (
	strongPasswordUpperRe   = regexp.MustCompile(`[A-Z]`)
	strongPasswordDigitRe   = regexp.MustCompile(`[0-9]`)
	strongPasswordSpecialRe = regexp.MustCompile(`[^A-Za-z0-9]`)
)

var (
	validate *govalidator.Validate
	trans    ut.Translator
)

func init() {
	validate = govalidator.New(govalidator.WithRequiredStructEnabled())

	// Resolve JSON tag name instead of Go struct field name
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})

	idLocale := id.New()
	uni := ut.New(idLocale, idLocale)
	trans, _ = uni.GetTranslator("id")

	if err := idtranslations.RegisterDefaultTranslations(validate, trans); err != nil {
		panic(fmt.Errorf("validator: failed to register id translations: %w", err))
	}

	// strongpassword tag: min 1 uppercase, 1 digit, 1 special character
	if err := validate.RegisterValidation("strongpassword", validateStrongPassword); err != nil {
		panic(fmt.Errorf("validator: failed to register strongpassword tag: %w", err))
	}
	if err := validate.RegisterTranslation("strongpassword", trans, registerStrongPasswordTranslation, translateStrongPassword); err != nil {
		panic(fmt.Errorf("validator: failed to register strongpassword translation: %w", err))
	}
}

func validateStrongPassword(fl govalidator.FieldLevel) bool {
	val := fl.Field().String()
	return strongPasswordUpperRe.MatchString(val) &&
		strongPasswordDigitRe.MatchString(val) &&
		strongPasswordSpecialRe.MatchString(val)
}

func registerStrongPasswordTranslation(ut ut.Translator) error {
	return ut.Add("strongpassword", "{0} harus mengandung minimal 1 huruf besar, 1 angka, dan 1 karakter spesial", true)
}

func translateStrongPassword(ut ut.Translator, fe govalidator.FieldError) string {
	t, _ := ut.T("strongpassword", fe.Field())
	return t
}

// FieldErrors maps a JSON field name to its validation error messages.
type FieldErrors map[string][]string

func (fe FieldErrors) Error() string {
	messages := make([]string, 0, len(fe))
	for field, errs := range fe {
		messages = append(messages, fmt.Sprintf("%s: %s", field, strings.Join(errs, ", ")))
	}
	return strings.Join(messages, "; ")
}

// Struct validates s against its struct tags.
func Struct(s any) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrs govalidator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return err
	}

	fieldErrs := make(FieldErrors)
	for _, fe := range validationErrs {
		fieldErrs[fe.Field()] = append(fieldErrs[fe.Field()], fe.Translate(trans))
	}
	return fieldErrs
}

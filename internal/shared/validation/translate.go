package validation

import (
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

var trans ut.Translator

func InitTranslations() {
	english := en.New()
	uni := ut.New(english, english)
	trans, _ = uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, trans)
}

func FormatErrors(err error) string {
	if vErrs, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, f := range vErrs {
			messages = append(messages, f.Translate(trans))
		}
		// Join with a separator suitable for your logging or UI (e.g., "; ")
		return strings.Join(messages, "; ")
	}
	return err.Error()
}

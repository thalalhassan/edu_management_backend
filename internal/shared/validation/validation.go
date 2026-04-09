package validation

import (
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var (
	once     sync.Once
	validate *validator.Validate
)

func InitValidator() {
	once.Do(func() {
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			validate = v
			// Use JSON tag names in error messages
			validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
				name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
				if name == "-" {
					return ""
				}
				return name
			})

			v.RegisterValidation("gender", func(fl validator.FieldLevel) bool {
				g := fl.Field().String()
				return g == "male" || g == "female" || g == "other"
			})
		}

		InitTranslations()
	})
}

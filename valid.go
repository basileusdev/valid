package valid

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const (
	tagName      = "valid"
	tagSeparator = ","
)

type ValidationError struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Value   any    `json:"value,omitempty"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	errStrings := make([]string, 0, len(e))
	for _, err := range e {
		errStrings = append(errStrings, err.Error())
	}

	return strings.Join(errStrings, " | ")
}

type Rule interface {
	Validate(value any, field reflect.StructField) error
	Msg(message string) Rule
}

type Validator struct {
	Rules map[string]Rule
}

func New() *Validator {
	return &Validator{
		Rules: make(map[string]Rule),
	}
}

func (v *Validator) Check(value any) error {
	rv := reflect.ValueOf(value)
	rt := reflect.TypeOf(value)

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
		rt = rt.Elem()
	}

	if rt.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %s", rt.Kind())
	}

	validationErrors, err := v.validateStruct(rt, rv)
	if err != nil {
		return err
	}

	return validationErrors
}

// TODO: Think of a better name
func (v *Validator) validateStruct(rt reflect.Type, rv reflect.Value) (ValidationErrors, error) {
	var validationErrors ValidationErrors

	for i := 0; i < rt.NumField(); i++ {
		fieldType := rt.Field(i)
		fieldValue := rv.Field(i)

		tag := fieldType.Tag.Get(tagName)
		if tag == "" {
			return nil, fmt.Errorf("field %s has an empty tag", fieldType.Name)
		}

		for _, ruleName := range strings.Split(tag, tagSeparator) {
			rule, ok := v.Rules[ruleName]
			if !ok {
				return nil, fmt.Errorf("rule %s doesn't exist", ruleName)
			}

			err := rule.Validate(fieldValue.Interface(), fieldType)
			if err == nil {
				continue
			}

			validationError, ok := errors.AsType[ValidationError](err)
			if !ok {
				return nil, err
			}

			validationErrors = append(validationErrors, validationError)
		}

		ft := fieldType.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		if ft.Kind() == reflect.Struct {
			validationErrs, err := v.validateStruct(ft, rv)
			if err != nil {
				return nil, err
			}

			validationErrors = append(validationErrors, validationErrs...)
		}
	}

	return validationErrors, nil
}

var validator = New()

func Check(value any) error {
	return validator.Check(value)
}

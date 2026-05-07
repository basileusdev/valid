package valid

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	tagName      = "validate"
	tagSeparator = ","
)

type ViolationError struct {
	Path  string `json:"path"`
	Rule  string `json:"rule"`
	Value any    `json:"value,omitempty"`
	Msg   string `json:"msg"`
}

// TODO: Temporary body
func (e ViolationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Msg)
}

type Violation struct {
	Value any
	Msg   string
}

type ValidationErrors []ViolationError

func (e ValidationErrors) Error() string {
	errStrings := make([]string, 0, len(e))
	for _, err := range e {
		errStrings = append(errStrings, err.Error())
	}

	return strings.Join(errStrings, " | ")
}

type Rule interface {
	Validate(value any) (*Violation, error)
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

	validationErrors, err := v.validateStruct(rt, rv, "")
	if err != nil {
		return err
	}

	return validationErrors
}

// TODO: Think of a better name
func (v *Validator) validateStruct(rt reflect.Type, rv reflect.Value, path string) (ValidationErrors, error) {
	var validationErrors ValidationErrors

	for i := 0; i < rt.NumField(); i++ {
		fieldType := rt.Field(i)
		fieldValue := rv.Field(i)

		tag := fieldType.Tag.Get(tagName)
		if tag == "" {
			return nil, fmt.Errorf("field %s has an empty tag", fieldType.Name)
		}

		fieldValidationErrors, err := v.validateFieldRules(
			fieldValue.Interface(),
			formatPath(path, fieldType.Name),
			strings.Split(tag, tagSeparator),
		)
		if err != nil {
			return nil, err
		}

		validationErrors = append(validationErrors, fieldValidationErrors...)

		fv := fieldValue
		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}

		switch fv.Kind() {
		case reflect.Struct:
			validationErrs, err := v.validateStruct(fv.Type(), fv, formatPath(path, fieldType.Name))
			if err != nil {
				return nil, err
			}

			validationErrors = append(validationErrors, validationErrs...)

		case reflect.Slice:
			for j := 0; j < fv.Len(); j++ {
				elemValue := fv.Index(j)

				if elemValue.Kind() == reflect.Ptr {
					if elemValue.IsNil() {
						continue
					}
					elemValue = elemValue.Elem()
				}

				if elemValue.Kind() != reflect.Struct {
					continue
				}

				validationErrs, err := v.validateStruct(
					elemValue.Type(),
					elemValue,
					fmt.Sprintf("%s[%d]", formatPath(path, fieldType.Name), j),
				)
				if err != nil {
					return nil, err
				}

				validationErrors = append(validationErrors, validationErrs...)
			}
		}
	}

	return validationErrors, nil
}

func (v *Validator) validateFieldRules(value any, path string, rules []string) (ValidationErrors, error) {
	var validationErrors ValidationErrors

	for _, ruleName := range rules {
		rule, ok := v.Rules[ruleName]
		if !ok {
			return nil, fmt.Errorf("rule %s doesn't exist", ruleName)
		}

		violation, err := rule.Validate(value)
		if err != nil {
			return nil, err
		}

		if violation == nil {
			continue
		}

		validationErrors = append(validationErrors, ViolationError{
			Path:  path,
			Rule:  ruleName,
			Value: violation.Value,
			Msg:   violation.Msg,
		})
	}

	return validationErrors, nil
}

var validator = New()

func Check(value any) error {
	return validator.Check(value)
}

func formatPath(parent, child string) string {
	if parent == "" {
		return child
	}

	return fmt.Sprintf("%s.%s", parent, child)
}

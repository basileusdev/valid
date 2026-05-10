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

type Violation struct {
	Path string `json:"path"`
	Rule string `json:"rule"`
	Msg  string `json:"msg"`
}

// TODO: Temporary body
func (e Violation) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Msg)
}

type Violations []Violation

func (e Violations) Error() string {
	errStrings := make([]string, 0, len(e))
	for _, err := range e {
		errStrings = append(errStrings, err.Error())
	}

	return strings.Join(errStrings, " | ")
}

type Rule interface {
	Validate(value reflect.Value) (string, error)
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

func (v *Validator) validateStruct(rt reflect.Type, rv reflect.Value, path string) (Violations, error) {
	var violations Violations

	for i := 0; i < rt.NumField(); i++ {
		fieldType := rt.Field(i)
		fieldValue := rv.Field(i)

		fieldViolations, err := v.validateField(fieldType, fieldValue, path)
		if err != nil {
			return nil, err
		}

		violations = append(violations, fieldViolations...)
	}

	return violations, nil
}

func (v *Validator) validateField(
	fieldType reflect.StructField,
	fieldValue reflect.Value,
	path string,
) (Violations, error) {
	if !fieldType.IsExported() {
		return nil, nil
	}

	violations, err := v.validateFieldRules(
		fieldValue,
		formatPath(path, fieldType.Name),
		strings.Split(fieldType.Tag.Get(tagName), tagSeparator),
	)
	if err != nil {
		return nil, err
	}

	nestedViolations, err := v.validateNested(fieldType, fieldValue, path)
	if err != nil {
		return nil, err
	}

	violations = append(violations, nestedViolations...)

	return violations, nil
}

func (v *Validator) validateFieldRules(value reflect.Value, path string, rules []string) (Violations, error) {
	var violations Violations

	for _, ruleName := range rules {
		rule, ok := v.Rules[ruleName]
		if !ok {
			return nil, fmt.Errorf("rule \"%s\" isn't defined | field path: %s", ruleName, path)
		}

		violationMsg, err := rule.Validate(value)
		if err != nil {
			return nil, err
		}

		if violationMsg == "" {
			continue
		}

		violations = append(violations, Violation{
			Path: path,
			Rule: ruleName,
			Msg:  violationMsg,
		})
	}

	return violations, nil
}

func (v *Validator) validateNested(
	fieldType reflect.StructField,
	fieldValue reflect.Value,
	path string,
) (Violations, error) {
	value := deref(fieldValue)

	switch value.Kind() {
	case reflect.Struct:
		return v.validateStruct(value.Type(), value, formatPath(path, fieldType.Name))

	case reflect.Slice:
		return v.validateSlice(fieldType, value, path)

	default:
		return nil, nil
	}
}

func (v *Validator) validateSlice(
	fieldType reflect.StructField,
	val reflect.Value,
	path string,
) (Violations, error) {
	elemType := val.Type().Elem()

	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	if elemType.Kind() != reflect.Struct {
		return nil, nil
	}

	var violations Violations

	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		elem = deref(elem)

		if elem.Kind() != reflect.Struct {
			continue
		}

		elemViolations, err := v.validateStruct(
			elem.Type(),
			elem,
			fmt.Sprintf("%s[%d]", formatPath(path, fieldType.Name), i),
		)
		if err != nil {
			return nil, err
		}

		violations = append(violations, elemViolations...)
	}

	return violations, nil
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

func deref(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

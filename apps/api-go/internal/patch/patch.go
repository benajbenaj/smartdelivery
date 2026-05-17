package patch

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrInvalidTarget = errors.New("invalid patch target")
	ErrInvalidPatch  = errors.New("invalid patch")
)

func Apply(target any, patch any) error {
	targetValue := reflect.ValueOf(target)
	if !targetValue.IsValid() || targetValue.Kind() != reflect.Pointer || targetValue.IsNil() {
		return ErrInvalidTarget
	}

	targetStruct := targetValue.Elem()
	if targetStruct.Kind() != reflect.Struct {
		return ErrInvalidTarget
	}

	patchStruct, err := structValue(patch)
	if err != nil {
		return err
	}

	patchType := patchStruct.Type()
	for index := 0; index < patchStruct.NumField(); index++ {
		patchFieldType := patchType.Field(index)
		if !patchFieldType.IsExported() {
			continue
		}

		patchField := patchStruct.Field(index)
		if patchField.Kind() != reflect.Pointer {
			return fmt.Errorf("%w: field %s must be a pointer", ErrInvalidPatch, patchFieldType.Name)
		}
		if patchField.IsNil() {
			continue
		}

		targetField := targetStruct.FieldByName(patchFieldType.Name)
		if !targetField.IsValid() {
			return fmt.Errorf("%w: target missing field %s", ErrInvalidPatch, patchFieldType.Name)
		}
		if !targetField.CanSet() {
			return fmt.Errorf("%w: target field %s cannot be set", ErrInvalidPatch, patchFieldType.Name)
		}

		value := patchField.Elem()
		if value.Type().AssignableTo(targetField.Type()) {
			targetField.Set(value)
			continue
		}
		if value.Type().ConvertibleTo(targetField.Type()) {
			targetField.Set(value.Convert(targetField.Type()))
			continue
		}

		return fmt.Errorf("%w: field %s type %s cannot assign to %s", ErrInvalidPatch, patchFieldType.Name, value.Type(), targetField.Type())
	}

	return nil
}

func structValue(value any) (reflect.Value, error) {
	reflectValue := reflect.ValueOf(value)
	if !reflectValue.IsValid() {
		return reflect.Value{}, ErrInvalidPatch
	}

	if reflectValue.Kind() == reflect.Pointer {
		if reflectValue.IsNil() {
			return reflect.Value{}, ErrInvalidPatch
		}
		reflectValue = reflectValue.Elem()
	}
	if reflectValue.Kind() != reflect.Struct {
		return reflect.Value{}, ErrInvalidPatch
	}

	return reflectValue, nil
}

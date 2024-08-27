package configloader

import (
	"errors"
	"reflect"
	"strconv"

	flag "github.com/spf13/pflag"
)

func parseFlags(rv reflect.Value, rt reflect.Type) error {
	// парсинг флагов
	for i := 0; i < rt.NumField(); i++ {
		fieldT := rt.Field(i)
		fieldV := rv.Elem().FieldByName(fieldT.Name)

		// Пропуск неизменяемых полей
		if !fieldV.CanSet() || !fieldV.IsValid() {
			continue
		}

		// TODO: Попробовать переписать всё с использованием generics
		// TODO: Добавить поддержку других типов данных и вложенных структур
		//nolint:exhaustive // Все необходимые типы указаны и нет смылса обрабатывать оставшиеся (пока что)
		switch fieldT.Type.Kind() {
		case reflect.String:
			err := parseStringFlag(fieldT, fieldV)
			if err != nil {
				return err
			}
		case reflect.Int:
			err := parseIntFlag(fieldT, fieldV)
			if err != nil {
				return err
			}
		case reflect.Bool:
			err := parseBoolFlag(fieldT, fieldV)
			if err != nil {
				return err
			}
		default:
			return ErrStructContainsWrongField
		}
	}

	flag.Parse()

	return nil
}

func parseStringFlag(fieldT reflect.StructField, fieldV reflect.Value) error {
	// Получение указателя на поле
	pointer, ok := fieldV.Addr().Interface().(*string)
	if !ok {
		return ErrTypeAssertion
	}

	// Если shorthand указан, то используем флаг с шортхендом
	if fieldT.Tag.Get(tagShorthand) != "" {
		flag.StringVarP(
			pointer,
			fieldT.Tag.Get(tagName),
			fieldT.Tag.Get(tagShorthand),
			fieldT.Tag.Get(tagDefaultValue),
			fieldT.Tag.Get(tagUsage),
		)
	} else {
		flag.StringVar(
			pointer,
			fieldT.Tag.Get(tagName),
			fieldT.Tag.Get(tagDefaultValue),
			fieldT.Tag.Get(tagUsage),
		)
	}
	return nil
}

func parseIntFlag(fieldT reflect.StructField, fieldV reflect.Value) error {
	pointer, ok := fieldV.Addr().Interface().(*int)
	if !ok {
		return ErrTypeAssertion
	}

	// Получение значения по умолчанию
	var defaultValue int
	var err error // Чтобы избежать shadow переменных
	rawDefaultValue := fieldT.Tag.Get(tagDefaultValue)
	if rawDefaultValue != "" {
		defaultValue, err = strconv.Atoi(rawDefaultValue)
		if err != nil {
			return errors.Join(ErrInvalidTagValue, err)
		}
	}

	// Если shorthand указан, то используем флаг с шортхендом
	if fieldT.Tag.Get(tagShorthand) != "" {
		flag.IntVarP(
			pointer,
			fieldT.Tag.Get(tagName),
			fieldT.Tag.Get(tagShorthand),
			defaultValue,
			fieldT.Tag.Get(tagUsage),
		)
	} else {
		flag.IntVar(
			pointer,
			fieldT.Tag.Get(tagName),
			defaultValue,
			fieldT.Tag.Get(tagUsage),
		)
	}
	return nil
}

func parseBoolFlag(fieldT reflect.StructField, fieldV reflect.Value) error {
	// Получение указателя на поле
	pointer, ok := fieldV.Addr().Interface().(*bool)
	if !ok {
		return ErrTypeAssertion
	}

	// Получение значения по умолчанию
	var defaultValue bool
	var err error // чтобы избежать shadow переменных
	rawDefaultValue := fieldT.Tag.Get(tagDefaultValue)
	if rawDefaultValue != "" {
		defaultValue, err = strconv.ParseBool(rawDefaultValue)
		if err != nil {
			return errors.Join(ErrInvalidTagValue, err)
		}
	}

	// Если shorthand указан, то используем флаг с шортхендом
	if fieldT.Tag.Get(tagShorthand) != "" {
		flag.BoolVarP(
			pointer,
			fieldT.Tag.Get(tagName),
			fieldT.Tag.Get(tagShorthand),
			defaultValue,
			fieldT.Tag.Get(tagUsage),
		)
	} else {
		flag.BoolVar(
			pointer,
			fieldT.Tag.Get(tagName),
			defaultValue,
			fieldT.Tag.Get(tagUsage),
		)
	}
	return nil
}

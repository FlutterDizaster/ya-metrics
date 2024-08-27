package configloader

import (
	"log/slog"
	"os"
	"reflect"
	"strconv"
)

func parseEnvs(rv reflect.Value, rt reflect.Type) error {
	// парсинг переменных среды
	for i := 0; i < rt.NumField(); i++ {
		fieldT := rt.Field(i)
		fieldV := rv.Elem().FieldByName(fieldT.Name)

		// Пропуск неизменяемых полей
		if !fieldV.CanSet() || !fieldV.IsValid() {
			continue
		}

		//nolint:exhaustive // Все необходимые типы указаны и нет смылса обрабатывать оставшиеся (пока что)
		switch fieldT.Type.Kind() {
		case reflect.String:
			envVal, ok := os.LookupEnv(fieldT.Tag.Get(tagEnv))
			if ok {
				fieldV.SetString(envVal)
			}
		case reflect.Int:
			envVal, ok := lookupIntEnv(fieldT.Tag.Get(tagEnv))
			if ok {
				fieldV.SetInt(int64(envVal))
			}
		case reflect.Bool:
			envVal, ok := lookupBoolEnv(fieldT.Tag.Get(tagEnv))
			if ok {
				fieldV.SetBool(envVal)
			}
		default:
			return ErrStructContainsWrongField
		}
	}
	return nil
}

func lookupIntEnv(name string) (int, bool) {
	env, ok := os.LookupEnv(name)
	if !ok {
		return 0, false
	}
	val, err := strconv.Atoi(env)
	if err != nil {
		slog.Error(
			"wrong env type",
			slog.String("variable", name),
			slog.String("expected type", "integer"),
		)
		return 0, false
	}
	return val, true
}

func lookupBoolEnv(name string) (bool, bool) {
	env, ok := os.LookupEnv(name)
	if !ok {
		return false, false
	}
	val, err := strconv.ParseBool(env)
	if err != nil {
		slog.Error(
			"wrong env type",
			slog.String("variable", name),
			slog.String("expected type", "boolean"),
		)
		return false, false
	}
	return val, true
}

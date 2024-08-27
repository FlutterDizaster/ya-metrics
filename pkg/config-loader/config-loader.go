package configloader

import (
	"errors"
	"log/slog"
	"reflect"
)

var (
	ErrWrongType                = errors.New("config is not pointer to struct")
	ErrTypeAssertion            = errors.New("type assertion error")
	ErrStructContainsWrongField = errors.New(
		"struct contains wrong field. struct must contain only default types",
	)
	ErrInvalidTagValue = errors.New("invalid tag value")
)

const (
	tagEnv          = "env"
	tagName         = "name"
	tagDefaultValue = "default"
	tagUsage        = "usage"
	tagShorthand    = "short"
)

func LoadConfig(path string, config any) error {
	slog.Debug("Loading config")
	// Получение информации о значениях полей структуры
	reflectValue := reflect.ValueOf(config)

	// Если config не указывает на структуру, то возвращаем ошибку
	if reflectValue.Kind() != reflect.Pointer || reflectValue.IsNil() {
		return ErrWrongType
	}

	// Получение информации о типах полей структуры
	reflectType := reflect.TypeOf(config)

	// парсинг флагов
	err := parseFlags(reflectValue, reflectType)
	if err != nil {
		return err
	}

	// парсинг переменных среды
	err = parseEnvs(reflectValue, reflectType)
	if err != nil {
		return err
	}

	// парсинг файлов
	if path != "" {
		err = parseJSON(path, reflectValue, reflectType)
		if err != nil {
			return err
		}
	}

	return nil
}

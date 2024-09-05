package configloader

import (
	"errors"
	"log/slog"
	"reflect"
	"strconv"

	flag "github.com/spf13/pflag"
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
	tagName         = "name"
	tagShorthand    = "short"
	tagEnv          = "env"
	tagDefaultValue = "default"
	tagUsage        = "usage"
)

func LoadConfig(config any) error {
	slog.Debug("Loading config")
	// Получение информации о значениях полей структуры
	reflectValue := reflect.ValueOf(config)

	// Если config не указывает на структуру, то возвращаем ошибку
	if reflectValue.Kind() != reflect.Pointer || reflectValue.IsNil() {
		return ErrWrongType
	}

	// Получение информации о типах полей структуры
	reflectType := reflect.TypeOf(config)

	// Установка значений полей по умолчанию
	err := setDafaults(reflectValue, reflectType)
	if err != nil {
		return err
	}

	// получение пути к конфиг файлу
	// вызов парсинга происходит в функции parseFlags
	path := flag.StringP("config", "c", "", "path to .json config file")

	// парсинг флагов
	err = parseFlags(reflectValue, reflectType)
	if err != nil {
		return err
	}

	// парсинг переменных среды
	err = parseEnvs(reflectValue, reflectType)
	if err != nil {
		return err
	}

	// парсинг файлов
	if *path != "" {
		err = parseJSON(*path, reflectValue, reflectType)
		if err != nil {
			return err
		}
	}

	return nil
}

func setDafaults(rv reflect.Value, rt reflect.Type) error {
	for i := 0; i < rt.NumField(); i++ {
		fieldT := rt.Field(i)
		fieldV := rv.Elem().FieldByName(fieldT.Name)

		// Пропуск неизменяемых полей
		if !fieldV.CanSet() || !fieldV.IsValid() || fieldT.Tag.Get(tagDefaultValue) == "" {
			continue
		}

		// Присвоение значения по умолчанию
		defaultValue := fieldT.Tag.Get(tagDefaultValue)
		//nolint:exhaustive // Все необходимые типы указаны и нет смылса обрабатывать оставшиеся (пока что)
		switch fieldV.Kind() {
		case reflect.String:
			fieldV.SetString(defaultValue)
		case reflect.Int:
			val, err := strconv.Atoi(defaultValue)
			if err != nil {
				return err
			}
			fieldV.SetInt(int64(val))
		case reflect.Bool:
			val, err := strconv.ParseBool(defaultValue)
			if err != nil {
				return err
			}
			fieldV.SetBool(val)
		default:
			return ErrStructContainsWrongField
		}
	}
	return nil
}

package configloader

import (
	"encoding/json"
	"os"
	"reflect"
)

func parseJSON(path string, rv reflect.Value, rt reflect.Type) error {
	// открытие файла
	rawData, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// парсинг файла
	var data map[string]interface{}

	if err = json.Unmarshal(rawData, &data); err != nil {
		return err
	}

	for i := 0; i < rt.NumField(); i++ {
		fieldT := rt.Field(i)
		fieldV := rv.FieldByName(fieldT.Name)

		// Пропуск неизменяемых полей
		if !fieldV.CanSet() || !fieldV.IsValid() {
			continue
		}

		//nolint:exhaustive // Все необходимые типы указаны и нет смылса обрабатывать оставшиеся (пока что)
		switch fieldT.Type.Kind() {
		case reflect.String:
			if val, ok := data[fieldT.Tag.Get(tagName)]; ok {
				fieldV.SetString(val.(string))
			}
		case reflect.Int:
			if val, ok := data[fieldT.Tag.Get(tagName)]; ok {
				fieldV.SetInt(int64(val.(int)))
			}
		case reflect.Bool:
			if val, ok := data[fieldT.Tag.Get(tagName)]; ok {
				fieldV.SetBool(val.(bool))
			}
		default:
			return ErrStructContainsWrongField
		}
	}

	return nil
}

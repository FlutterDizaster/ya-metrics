#!/bin/bash

OUTPUT_DIR="./swagger"
GENERAL_INFO_FILE="./main.go"
DIRS="./cmd/server,./internal/server/api,./internal/view"

swag init --output "$OUTPUT_DIR" --parseInternal --generalInfo "$GENERAL_INFO_FILE" --dir "$DIRS"

if [ $? -eq 0 ]; then
    echo "Swagger документация успешно сгенерирована и сохранена в $OUTPUT_DIR."
else
    echo "Ошибка при генерации Swagger документации."
    exit 1
fi
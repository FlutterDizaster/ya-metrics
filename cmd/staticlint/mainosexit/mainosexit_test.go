package mainosexit_test

import (
	"github.com/FlutterDizaster/ya-metrics/cmd/staticlint/mainosexit"
	"golang.org/x/tools/go/analysis/multichecker"
)

//nolint:testableexamples // not needed
func Example() {
	// Для запуска анализатора необходимо установить пакет multichecker
	// и передать в функцию multichecker.Main анализатор.
	multichecker.Main(
		mainosexit.Analyzer,
	)
}

package main

import (
	"github.com/FlutterDizaster/ya-metrics/cmd/staticlint/mainosexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	// Список анализаторов из staticcheck.
	checks := map[string]bool{
		"SA":     true,
		"S1000":  true,
		"ST1001": true,
	}

	// Список анализаторов для мультианализатора.
	myChecks := []*analysis.Analyzer{
		mainosexit.Analyzer,
		appends.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		copylock.Analyzer,
		defers.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		printf.Analyzer,
		slog.Analyzer,
		unreachable.Analyzer,
	}

	// Добавление анализаторов из staticcheck в мультианализатор.
	for _, v := range staticcheck.Analyzers {
		if checks[v.Analyzer.Name] {
			myChecks = append(myChecks, v.Analyzer)
		}
	}

	// Запуск мультианализатора.
	multichecker.Main(
		myChecks...,
	)
}

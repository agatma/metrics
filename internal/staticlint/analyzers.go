package staticlint

import (
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stdversion"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/staticcheck"
)

// Analyzers build list with static analyzers
// it returns list of analyzers, which includes
//
//	all SA class analyzers of the staticcheck.io package  https://staticcheck.io/docs/checks/
//	S1009, ST1008, QF1012 analyzers of the staticcheck.io package  https://staticcheck.io/docs/checks/
//	all analyzers from golang.org/x/tools/go/analysis/passes package
func Analyzers() []*analysis.Analyzer {
	var analyzers []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") ||
			v.Analyzer.Name == "S1009" ||
			v.Analyzer.Name == "ST1008" ||
			v.Analyzer.Name == "QF1012" {
			analyzers = append(analyzers, v.Analyzer)
		}
	}
	analyzers = append(analyzers, MainExitAnalyzer)
	passes := []*analysis.Analyzer{appends.Analyzer, asmdecl.Analyzer, assign.Analyzer, atomic.Analyzer, atomicalign.Analyzer,
		bools.Analyzer, buildssa.Analyzer, buildtag.Analyzer, cgocall.Analyzer, composite.Analyzer, copylock.Analyzer, ctrlflow.Analyzer,
		deepequalerrors.Analyzer, defers.Analyzer, directive.Analyzer, errorsas.Analyzer, findcall.Analyzer,
		framepointer.Analyzer, httpmux.Analyzer, httpresponse.Analyzer, ifaceassert.Analyzer, inspect.Analyzer, loopclosure.Analyzer,
		lostcancel.Analyzer, nilfunc.Analyzer, pkgfact.Analyzer, printf.Analyzer, reflectvaluecompare.Analyzer,
		shift.Analyzer, sigchanyzer.Analyzer, slog.Analyzer, sortslice.Analyzer, stdmethods.Analyzer, stdversion.Analyzer,
		stringintconv.Analyzer, structtag.Analyzer, testinggoroutine.Analyzer, tests.Analyzer, timeformat.Analyzer, unreachable.Analyzer,
		unsafeptr.Analyzer, unusedresult.Analyzer, unusedwrite.Analyzer, usesgenerics.Analyzer,
	}
	analyzers = append(analyzers, passes...)
	return analyzers
}

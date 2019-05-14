package model

import (
	"fmt"
	"github.com/lectio/graph/observe"
)

// ProgressReporter returns a specific PR for a given type
func (o *ObservationSettings) ProgressReporter() observe.ProgressReporter {
	switch o.ProgressReporterType {
	case ProgressReporterTypeSilent:
		return observe.DefaultSilentProgressReporter
	case ProgressReporterTypeSummary:
		return observe.DefaultSummaryReporter
	case ProgressReporterTypeProgressBar:
		return observe.DefaultCommandLineProgressReporter
	default:
		fmt.Printf("Unkown ProgressReporter type %s, using Silent instead", o.ProgressReporterType)
		return observe.DefaultSilentProgressReporter
	}
}

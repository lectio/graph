package model

import (
	"github.com/lectio/graph/observe"
)

// ProgressReporter returns a specific PR for a given type
func (os ObservationSettings) ProgressReporter() observe.ProgressReporter {
	switch os.ProgressReporterType {
	case ProgressReporterTypeSilent:
		return observe.DefaultSilentProgressReporter
	case ProgressReporterTypeCommandLineProgressBar:
		return observe.DefaultCommandLineProgressReporter
	default:
		return nil
	}
}

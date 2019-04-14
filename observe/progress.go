package observe

import (
	"io"

	"gopkg.in/cheggaaa/pb.v1"
)

// ProgressReporter is one observation method for live reporting of long-running processes
type ProgressReporter interface {
	IsProgressReportingRequested() bool
	StartReportableActivity(expectedItems int)
	StartReportableReaderActivityInBytes(exepectedBytes int64, inputReader io.Reader) io.Reader
	IncrementReportableActivityProgress()
	IncrementReportableActivityProgressBy(incrementBy int)
	CompleteReportableActivityProgress(summary string)
}

// DefaultCommandLineProgressReporter returns the default CLI based progress bar
var DefaultCommandLineProgressReporter = NewCommandLineProgressReporter(true)

type progressReporter struct {
	verbose bool
	bar     *pb.ProgressBar
}

// NewCommandLineProgressReporter creates a new instance of a CLI progres bar
func NewCommandLineProgressReporter(verbose bool) ProgressReporter {
	result := new(progressReporter)
	result.verbose = verbose
	return result
}

func (pr progressReporter) IsProgressReportingRequested() bool {
	return pr.verbose
}

func (pr *progressReporter) StartReportableActivity(expectedItems int) {
	pr.bar = pb.StartNew(expectedItems)
	pr.bar.ShowCounters = true
}

func (pr *progressReporter) StartReportableReaderActivityInBytes(exepectedBytes int64, inputReader io.Reader) io.Reader {
	pr.bar = pb.New(int(exepectedBytes)).SetUnits(pb.U_BYTES)
	pr.bar.Start()
	return pr.bar.NewProxyReader(inputReader)
}

func (pr *progressReporter) IncrementReportableActivityProgress() {
	pr.bar.Increment()
}

func (pr *progressReporter) IncrementReportableActivityProgressBy(incrementBy int) {
	pr.bar.Add(incrementBy)
}

func (pr *progressReporter) CompleteReportableActivityProgress(summary string) {
	pr.bar.FinishPrint(summary)
}

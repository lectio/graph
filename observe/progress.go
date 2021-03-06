package observe

import (
	"fmt"
	"io"

	"gopkg.in/cheggaaa/pb.v1"
)

// ProgressReporter is one observation method for live reporting of long-running processes
type ProgressReporter interface {
	StartReportableActivity(summary string, expectedItems int)
	StartReportableReaderActivityInBytes(summary string, exepectedBytes int64, inputReader io.Reader) io.Reader
	IncrementReportableActivityProgress()
	IncrementReportableActivityProgressBy(incrementBy int)
	CompleteReportableActivityProgress(summary string)
}

// DefaultCommandLineProgressReporter returns the default CLI based progress bar
var DefaultCommandLineProgressReporter = NewCommandLineProgressReporter(true)

// DefaultSummaryReporter returns a PR that only provides the summary at the end (no interim progress)
var DefaultSummaryReporter = summaryReporter{}

// DefaultSilentProgressReporter returns a PR that doesn't do anything
var DefaultSilentProgressReporter = slientProgressReporter{}

type slientProgressReporter struct{}

func (pr slientProgressReporter) StartReportableActivity(summary string, expectedItems int) {
}

func (pr slientProgressReporter) StartReportableReaderActivityInBytes(summary string, exepectedBytes int64, inputReader io.Reader) io.Reader {
	return inputReader
}

func (pr slientProgressReporter) IncrementReportableActivityProgress() {
}

func (pr slientProgressReporter) IncrementReportableActivityProgressBy(incrementBy int) {
}

func (pr slientProgressReporter) CompleteReportableActivityProgress(summary string) {
}

type summaryReporter struct{}

func (pr summaryReporter) StartReportableActivity(summary string, expectedItems int) {
	fmt.Println(summary)
}

func (pr summaryReporter) StartReportableReaderActivityInBytes(summary string, exepectedBytes int64, inputReader io.Reader) io.Reader {
	fmt.Println(summary)
	return inputReader
}

func (pr summaryReporter) IncrementReportableActivityProgress() {
}

func (pr summaryReporter) IncrementReportableActivityProgressBy(incrementBy int) {
}

func (pr summaryReporter) CompleteReportableActivityProgress(summary string) {
	fmt.Println(summary)
}

type progressReporter struct {
	bar *pb.ProgressBar
}

// NewCommandLineProgressReporter creates a new instance of a CLI progres bar
func NewCommandLineProgressReporter(verbose bool) ProgressReporter {
	result := new(progressReporter)
	if !verbose {
		return DefaultSilentProgressReporter
	}
	return result
}

func (pr *progressReporter) StartReportableActivity(summary string, expectedItems int) {
	pr.bar = pb.StartNew(expectedItems)
	pr.bar.ShowCounters = true
}

func (pr *progressReporter) StartReportableReaderActivityInBytes(summary string, exepectedBytes int64, inputReader io.Reader) io.Reader {
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

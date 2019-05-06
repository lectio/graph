package pipeline

import (
	"github.com/lectio/graph/model"
	"github.com/sony/sonyflake"
	"net/url"
)

var sfSettings sonyflake.Settings
var sflake *sonyflake.Sonyflake = sonyflake.NewSonyflake(sfSettings)

// Pipeline is an abstract runner of a pre-defined pipeline
type Pipeline interface {
	URL() *url.URL
	Execute() (model.PipelineExecution, error)
}

// GenerateExecutionID returns a unique ID
func GenerateExecutionID() model.PipelineExecutionID {
	result, _ := sflake.NextID()
	return model.PipelineExecutionID(result)
}

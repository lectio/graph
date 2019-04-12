package model

type ActivityLogMessageCode string
type ActivityHumanMessage string
type ActivityMachineMessage string

type activities struct {
	activities []Activity
	errors     []ActivityLogMessage
	warnings   []ActivityLogMessage
}

type activityContext string

func (activityContext) IsActivityContext() {}

type activity struct {
}

func (a activity) IsActivity() {}

type activityLogMessage struct {
}

func (m activityLogMessage) IsActivityLogMessage() {}

func makeActivities() *activities {
	result := new(activities)
	return result
}

func (a *activities) IsActivities() {}

func (a *activities) addError(code string, message ActivityHumanMessage) {
}

package main

import (
	"fmt"
	"strings"
	"time"
)

type EventType string

const (
	EventType_Alarm         string = "alarm"
	EventType_Insight              = "insight"
	EventType_CustomInsight        = "custom-insight"
	EventType_Synthetics           = "synthetic"
	EventType_Mitigation           = "mitigation"
	EventType_Generic              = "generic"
)

type ViewModelImportance int32

const (
	ViewModelImportance_None     ViewModelImportance = 0
	ViewModelImportance_Healthy                      = 1
	ViewModelImportance_Notice                       = 2
	ViewModelImportance_Minor                        = 3
	ViewModelImportance_Warning                      = 4
	ViewModelImportance_Major                        = 5
	ViewModelImportance_Severe                       = 6
	ViewModelImportance_Critical                     = 7
)

var VieModelImportanceOrdered = [...]ViewModelImportance{
	ViewModelImportance_Critical,
	ViewModelImportance_Severe,
	ViewModelImportance_Major,
	ViewModelImportance_Warning,
	ViewModelImportance_Minor,
	ViewModelImportance_Notice,
	ViewModelImportance_Healthy,
	ViewModelImportance_None,
}

// use "export" key instead of standard json key when marshalling/unmarshalling using jsoniter (https://github.com/json-iterator/go),
// so fields are not removed per standard json tag
type EventViewModel struct {
	Type           string
	Description    string `json:",omitempty" export:"Description"`
	IsActive       bool
	StartTime      string
	EndTime        string
	CurrentState   string
	PreviousState  string
	StartTimestamp int64                 `json:"-" export:"StartTimestamp"`
	EndTimestamp   int64                 `json:"-" export:"EndTimestamp"`
	Importance     ViewModelImportance   `json:"-" export:"Importance"`
	GroupName      string                `json:"-" export:"GroupName"`
	Details        EventViewModelDetails `json:"-" export:"Details"`
}

func (event EventViewModel) IsAlarm() bool {
	return event.Type == EventType_Alarm
}

func (event EventViewModel) IsInsight() bool {
	return event.Type == EventType_Insight || event.Type == EventType_CustomInsight
}

func (event EventViewModel) IsCustomInsight() bool {
	return event.Type == EventType_CustomInsight
}

func (event EventViewModel) IsMitigation() bool {
	return event.Type == EventType_Mitigation
}

func (event EventViewModel) IsSynthetic() bool {
	return event.Type == EventType_Synthetics
}

type EventViewModelDetail struct {
	Name  string
	Label string `json:",omitempty"`
	Value interface{}
	Tag   string `json:"-" export:"Tag"`
}

type EventViewModelDetails []*EventViewModelDetail

func (details EventViewModelDetails) WithTag(tag string) EventViewModelDetails {
	result := make(EventViewModelDetails, 0)
	for _, detail := range details {
		if detail.Tag == tag {
			result = append(result, detail)
		}
	}
	return result
}

func (details EventViewModelDetails) General() EventViewModelDetails {
	result := make(EventViewModelDetails, 0)
	for _, detail := range details {
		if detail.Tag == "" {
			result = append(result, detail)
		}
	}
	return result
}

func (details EventViewModelDetails) WithNames(names ...string) EventViewModelDetails {
	result := make(EventViewModelDetails, 0)
	for _, detail := range details {
		for _, name := range names {
			if detail.Name == name {
				result = append(result, detail)
			}
		}
	}
	return result
}

func (details EventViewModelDetails) Names() []string {
	result := make([]string, 0, len(details))
	for _, detail := range details {
		result = append(result, detail.Name)
	}
	return result
}

func (details EventViewModelDetails) Values() []interface{} {
	result := make([]interface{}, 0, len(details))
	for _, detail := range details {
		result = append(result, detail.Value)
	}
	return result
}

func (details EventViewModelDetails) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	for _, detail := range details {
		result[detail.Name] = detail.Value
	}
	return result
}

func (details EventViewModelDetails) Has(name string) bool {
	for _, detail := range details {
		if detail.Name == name {
			return true
		}
	}
	return false
}

func (details *EventViewModelDetails) HasTag(tag string) bool {
	for _, detail := range *details {
		if detail.Tag == tag {
			return true
		}
	}
	return false
}

func (event *EventViewModel) AddDetail(detail *EventViewModelDetail) {
	if event.Details == nil {
		event.Details = make(EventViewModelDetails, 0, 1)
	}
	event.Details = append(event.Details, detail)
}

func (details EventViewModelDetails) Get(name string) *EventViewModelDetail {
	for _, detail := range details {
		if detail.Name == name {
			return detail
		}
	}
	return &EventViewModelDetail{ // let's just be safe here :)
		Name:  name,
		Label: name,
		Value: nil,
	}
}

func (details EventViewModelDetails) GetValue(name string) interface{} {
	return details.Get(name).Value
}

func (detail EventViewModelDetail) LabelOrName() string {
	if detail.Label != "" {
		return detail.Label
	}
	return detail.Name
}

type NotificationViewModel struct {
	CompanyID   int
	CompanyName string    `json:"-" export:"CompanyName"`
	Now         time.Time `json:"-" export:"Now"`

	RawEvents []*EventViewModel `json:"-" export:"RawEvents"`

	Config *NotificationViewConfig `json:"-" export:"Config"`
}
type NotificationViewConfig struct {
	BaseDomain string
	EmailTo    []string
}

func (vm *NotificationViewModel) BasePortalURL() string {
	return fmt.Sprintf("https://%s", vm.Config.BaseDomain)
}

func (vm *NotificationViewModel) NotificationsSettingsURL() string {
	return fmt.Sprintf("https://%s/v4/settings/notifications", vm.Config.BaseDomain)
}

func (vm *NotificationViewModel) SyntheticsDashboardURL() string {
	return fmt.Sprintf("https://%s/v4/synthetics/dashboard", vm.Config.BaseDomain)
}

func (vm *NotificationViewModel) NowDate() string {
	return vm.Now.Format("January 2, 2006")
}

func (vm *NotificationViewModel) NowRFC3339() string {
	return vm.Now.Format(time.RFC3339)
}

func (vm *NotificationViewModel) NowDatetime() string {
	return vm.Now.Format("2006-01-02 15:04:05 UTC")
}

func (vm *NotificationViewModel) NowUnix() int64 {
	return vm.Now.Unix()
}

func (vm *NotificationViewModel) Copyrights() string {
	return fmt.Sprintf("© %d Kentik", vm.Now.Year())
}

func (vm *NotificationViewModel) IsSingleEvent() bool {
	return len(vm.RawEvents) == 1
}

func (vm *NotificationViewModel) IsMultipleEvents() bool {
	return len(vm.RawEvents) > 1
}

func (vm *NotificationViewModel) IsAtLeastOneEvent() bool {
	return len(vm.RawEvents) > 0
}

func (vm *NotificationViewModel) Event() *EventViewModel {
	return vm.RawEvents[0]
}

func (vm *NotificationViewModel) Events() []*EventViewModel {
	return vm.RawEvents
}

func (vm *NotificationViewModel) ActiveCount() int {
	var result int
	for _, event := range vm.RawEvents {
		if event.IsActive {
			result += 1
		}
	}
	return result
}

func (vm *NotificationViewModel) InactiveCount() int {
	var result int
	for _, event := range vm.RawEvents {
		if !event.IsActive {
			result += 1
		}
	}
	return result
}

func (vm *NotificationViewModel) IsInsightsOnly() bool {
	for _, evt := range vm.RawEvents {
		if !evt.IsInsight() {
			return false
		}
	}
	return true
}

func (vm *NotificationViewModel) IsSyntheticsOnly() bool {
	for _, evt := range vm.RawEvents {
		if !evt.IsSynthetic() {
			return false
		}
	}
	return true
}

func (vm *NotificationViewModel) IsSingleCustomInsightOnly() bool {
	return len(vm.RawEvents) == 1 && vm.RawEvents[0].IsCustomInsight()
}

func (vm *NotificationViewModel) IsSynthOnly() bool {
	for _, evt := range vm.RawEvents {
		if !evt.IsSynthetic() {
			return false
		}
	}
	return true
}

func (vm *NotificationViewModel) Headline() string {
	segments := make([]string, 0)
	segments = append(segments, "Kentik")

	if vm.IsInsightsOnly() {
		segments = append(segments, "Insights")
	} else if vm.IsSyntheticsOnly() {
		segments = append(segments, "Synthetics")
	}

	if vm.IsMultipleEvents() {
		segments = append(segments, "Digest")
	} else {
		segments = append(segments, "Alert")
	}
	return strings.Join(segments, " ")
}

func (vm *NotificationViewModel) Summary() string {
	if vm.IsSingleEvent() {
		return vm.Event().Description
	}
	segments := make([]string, 0)
	if vm.ActiveCount() > 0 {
		segments = append(segments, fmt.Sprintf("%d changed to unhealthy", vm.ActiveCount()))
	}
	if vm.InactiveCount() > 0 {
		segments = append(segments, fmt.Sprintf("%d changed to healthy", vm.InactiveCount()))
	}
	return strings.Join(segments, ", ")
}

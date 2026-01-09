package render

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	EventType_Alarm         string = "alarm"
	EventType_Insight       string = "insight"
	EventType_CustomInsight string = "custom-insight"
	EventType_Synthetics    string = "synthetic"
	EventType_Mitigation    string = "mitigation"
	EventType_Generic       string = "generic"
)

type ViewModelImportance int32

const (
	ViewModelImportance_None     ViewModelImportance = 0
	ViewModelImportance_Healthy  ViewModelImportance = 1
	ViewModelImportance_Notice   ViewModelImportance = 2
	ViewModelImportance_Minor    ViewModelImportance = 3
	ViewModelImportance_Warning  ViewModelImportance = 4
	ViewModelImportance_Major    ViewModelImportance = 5
	ViewModelImportance_Severe   ViewModelImportance = 6
	ViewModelImportance_Critical ViewModelImportance = 7
)

var ImportanceNames = map[ViewModelImportance]string{
	ViewModelImportance_None:     "n/a",
	ViewModelImportance_Healthy:  "healthy",
	ViewModelImportance_Notice:   "notice",
	ViewModelImportance_Minor:    "minor",
	ViewModelImportance_Warning:  "warning",
	ViewModelImportance_Major:    "major",
	ViewModelImportance_Severe:   "severe",
	ViewModelImportance_Critical: "critical",
}

var ImportanceToColors = map[ViewModelImportance]string{
	ViewModelImportance_None:     "#999999",
	ViewModelImportance_Healthy:  "#1E9E1E",
	ViewModelImportance_Notice:   "#157FF3",
	ViewModelImportance_Minor:    "#F29D49",
	ViewModelImportance_Warning:  "#EE7E0F",
	ViewModelImportance_Major:    "#DB3737",
	ViewModelImportance_Severe:   "#C23030",
	ViewModelImportance_Critical: "#A82A2A",
}

var ImportanceToEmojis = map[ViewModelImportance]string{
	ViewModelImportance_None:     "",
	ViewModelImportance_Healthy:  ":warning: :large_green_circle:",
	ViewModelImportance_Notice:   ":warning: :large_blue_circle:",
	ViewModelImportance_Minor:    ":warning: :large_purple_circle:",
	ViewModelImportance_Warning:  ":warning: :large_brown_circle:",
	ViewModelImportance_Major:    ":warning: :large_yellow_circle:",
	ViewModelImportance_Severe:   ":warning: :large_orange_circle: ",
	ViewModelImportance_Critical: ":warning: :red_circle:",
}

type EventViewModel struct {
	Type           string                `description:"Event type (alarm, insight, synthetic, mitigation, generic)"`
	Description    string                `json:",omitempty" description:"Human-readable event description"`
	IsActive       bool                  `description:"Whether the event is currently active"`
	StartTime      string                `description:"Formatted start time string"`
	EndTime        string                `description:"Formatted end time string"`
	CurrentState   string                `description:"Current state of the event"`
	PreviousState  string                `description:"Previous state of the event"`
	StartTimestamp int64                 `json:"-" description:"Unix timestamp of event start"`
	EndTimestamp   int64                 `json:"-" description:"Unix timestamp of event end"`
	Importance     ViewModelImportance   `json:"-" description:"Severity level (0-7)"`
	GroupName      string                `json:"-" description:"Name of the event group"`
	Details        EventViewModelDetails `json:"-" description:"List of event detail key-value pairs"`
}

func (e *EventViewModel) UnmarshalJSON(data []byte) error {
	type EvmAsInput EventViewModel
	aux := &struct {
		StartTimestamp int64                 `json:"StartTimestamp"`
		EndTimestamp   int64                 `json:"EndTimestamp"`
		Importance     ViewModelImportance   `json:"Importance"`
		GroupName      string                `json:"GroupName"`
		Details        EventViewModelDetails `json:"Details"`
		*EvmAsInput
	}{
		EvmAsInput: (*EvmAsInput)(e),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	e.StartTimestamp = aux.StartTimestamp
	e.EndTimestamp = aux.EndTimestamp
	e.Importance = aux.Importance
	e.GroupName = aux.GroupName
	e.Details = aux.Details
	return nil
}

// IsAlarm returns true if event type is alarm.
func (event EventViewModel) IsAlarm() bool {
	return event.Type == EventType_Alarm
}

// IsInsight returns true if event type is insight or custom-insight.
func (event EventViewModel) IsInsight() bool {
	return event.Type == EventType_Insight || event.Type == EventType_CustomInsight
}

// IsCustomInsight returns true if event type is custom-insight.
func (event EventViewModel) IsCustomInsight() bool {
	return event.Type == EventType_CustomInsight
}

// IsMitigation returns true if event type is mitigation.
func (event EventViewModel) IsMitigation() bool {
	return event.Type == EventType_Mitigation
}

// IsSynthetic returns true if event type is synthetic.
func (event EventViewModel) IsSynthetic() bool {
	return event.Type == EventType_Synthetics
}

type EventViewModelDetail struct {
	Name  string      `description:"Detail field name/key"`
	Label string      `json:",omitempty" description:"Human-readable label for the detail"`
	Value interface{} `description:"Detail value (can be any type)"`
	Tag   string      `json:"-" description:"Categorization tag (metric, dimension, url, device, etc.)"`
}

func (d *EventViewModelDetail) UnmarshalJSON(data []byte) error {
	type EvmDetailAsInput EventViewModelDetail
	aux := &struct {
		Tag string `json:"Tag"`
		*EvmDetailAsInput
	}{
		EvmDetailAsInput: (*EvmDetailAsInput)(d),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	d.Tag = aux.Tag
	return nil
}

type EventViewModelDetails []*EventViewModelDetail

// WithTag filters details by the specified tag.
func (details EventViewModelDetails) WithTag(tag string) EventViewModelDetails {
	result := make(EventViewModelDetails, 0)
	for _, detail := range details {
		if detail.Tag == tag {
			result = append(result, detail)
		}
	}
	return result
}

// General returns details with an empty tag.
func (details EventViewModelDetails) General() EventViewModelDetails {
	result := make(EventViewModelDetails, 0)
	for _, detail := range details {
		if detail.Tag == "" {
			result = append(result, detail)
		}
	}
	return result
}

// WithNames filters details by the given names.
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

// Names returns all detail names.
func (details EventViewModelDetails) Names() []string {
	result := make([]string, 0, len(details))
	for _, detail := range details {
		result = append(result, detail.Name)
	}
	return result
}

// Values returns all detail values.
func (details EventViewModelDetails) Values() []interface{} {
	result := make([]interface{}, 0, len(details))
	for _, detail := range details {
		result = append(result, detail.Value)
	}
	return result
}

// ToMap converts details to a name-to-value map.
func (details EventViewModelDetails) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	for _, detail := range details {
		result[detail.Name] = detail.Value
	}
	return result
}

// Has checks if a detail with the given name exists.
func (details EventViewModelDetails) Has(name string) bool {
	for _, detail := range details {
		if detail.Name == name {
			return true
		}
	}
	return false
}

// HasTag checks if any detail has the specified tag.
func (details *EventViewModelDetails) HasTag(tag string) bool {
	for _, detail := range *details {
		if detail.Tag == tag {
			return true
		}
	}
	return false
}

// AddDetail adds a detail to the event's Details collection.
func (event *EventViewModel) AddDetail(detail *EventViewModelDetail) {
	if event.Details == nil {
		event.Details = make(EventViewModelDetails, 0, 1)
	}
	event.Details = append(event.Details, detail)
}

// Get retrieves a detail by name.
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

// GetValue retrieves a value by name.
func (details EventViewModelDetails) GetValue(name string) interface{} {
	return details.Get(name).Value
}

// LabelOrName returns Label if set, otherwise returns Name.
func (detail EventViewModelDetail) LabelOrName() string {
	if detail.Label != "" {
		return detail.Label
	}
	return detail.Name
}

type NotificationViewModel struct {
	CompanyID   int                     `description:"Unique identifier for the company"`
	CompanyName string                  `json:"-" description:"Name of the company"`
	Now         time.Time               `json:"-" description:"Current timestamp when notification is generated"`
	RawEvents   []*EventViewModel       `json:"-" description:"List of all events in this notification"`
	Config      *NotificationViewConfig `json:"-" description:"Notification configuration settings"`
}

func (vm *NotificationViewModel) UnmarshalJSON(data []byte) error {
	type NvmAsInput NotificationViewModel
	aux := &struct {
		CompanyName string                  `json:"CompanyName"`
		Now         time.Time               `json:"Now"`
		RawEvents   []*EventViewModel       `json:"Events"`
		Config      *NotificationViewConfig `json:"Config"`
		*NvmAsInput
	}{
		NvmAsInput: (*NvmAsInput)(vm),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	vm.CompanyName = aux.CompanyName
	vm.Now = aux.Now
	vm.RawEvents = aux.RawEvents
	vm.Config = aux.Config
	return nil
}

type NotificationViewConfig struct {
	BaseDomain string   `description:"Portal base domain (e.g., portal.kentik.com)"`
	EmailTo    []string `description:"List of email recipients"`
}

// BasePortalURL returns the portal base URL.
func (vm *NotificationViewModel) BasePortalURL() string {
	return fmt.Sprintf("https://%s", vm.Config.BaseDomain)
}

// NotificationsSettingsURL returns the notifications settings URL.
func (vm *NotificationViewModel) NotificationsSettingsURL() string {
	return fmt.Sprintf("https://%s/v4/settings/notifications", vm.Config.BaseDomain)
}

// SyntheticsDashboardURL returns the synthetics dashboard URL.
func (vm *NotificationViewModel) SyntheticsDashboardURL() string {
	return fmt.Sprintf("https://%s/v4/synthetics/dashboard", vm.Config.BaseDomain)
}

// NowDate returns the current date formatted as 'January 2, 2006'.
func (vm *NotificationViewModel) NowDate() string {
	return vm.Now.Format("January 2, 2006")
}

// NowRFC3339 returns the current time in RFC3339 format.
func (vm *NotificationViewModel) NowRFC3339() string {
	return vm.Now.Format(time.RFC3339)
}

// NowDatetime returns the current time as '2006-01-02 15:04:05 UTC'.
func (vm *NotificationViewModel) NowDatetime() string {
	return vm.Now.Format("2006-01-02 15:04:05 UTC")
}

// NowUnix returns the current time as Unix timestamp.
func (vm *NotificationViewModel) NowUnix() int64 {
	return vm.Now.Unix()
}

// Copyrights returns the copyright string with current year.
func (vm *NotificationViewModel) Copyrights() string {
	return fmt.Sprintf("© %d Kentik", vm.Now.Year())
}

// IsSingleEvent returns true if exactly one event.
func (vm *NotificationViewModel) IsSingleEvent() bool {
	return len(vm.RawEvents) == 1
}

// IsMultipleEvents returns true if more than one event.
func (vm *NotificationViewModel) IsMultipleEvents() bool {
	return len(vm.RawEvents) > 1
}

// IsAtLeastOneEvent returns true if at least one event exists.
func (vm *NotificationViewModel) IsAtLeastOneEvent() bool {
	return len(vm.RawEvents) > 0
}

// Event returns the first event or nil if empty.
func (vm *NotificationViewModel) Event() *EventViewModel {
	if len(vm.RawEvents) == 0 {
		return nil
	}
	return vm.RawEvents[0]
}

// Events returns all events as a slice.
func (vm *NotificationViewModel) Events() []*EventViewModel {
	return vm.RawEvents
}

// ActiveCount returns the count of currently active events.
func (vm *NotificationViewModel) ActiveCount() int {
	var result int
	for _, event := range vm.RawEvents {
		if event.IsActive {
			result += 1
		}
	}
	return result
}

// InactiveCount returns the count of inactive events.
func (vm *NotificationViewModel) InactiveCount() int {
	var result int
	for _, event := range vm.RawEvents {
		if !event.IsActive {
			result += 1
		}
	}
	return result
}

// IsInsightsOnly returns true if all events are insights.
func (vm *NotificationViewModel) IsInsightsOnly() bool {
	for _, evt := range vm.RawEvents {
		if !evt.IsInsight() {
			return false
		}
	}
	return true
}

// IsSyntheticsOnly returns true if all events are synthetics.
func (vm *NotificationViewModel) IsSyntheticsOnly() bool {
	for _, evt := range vm.RawEvents {
		if !evt.IsSynthetic() {
			return false
		}
	}
	return true
}

// IsSingleCustomInsightOnly returns true if single custom insight event.
func (vm *NotificationViewModel) IsSingleCustomInsightOnly() bool {
	return len(vm.RawEvents) == 1 && vm.RawEvents[0].IsCustomInsight()
}

// IsSynthOnly is an alias for IsSyntheticsOnly.
func (vm *NotificationViewModel) IsSynthOnly() bool {
	for _, evt := range vm.RawEvents {
		if !evt.IsSynthetic() {
			return false
		}
	}
	return true
}

// Headline returns the generated headline text.
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

// Summary returns the generated summary text.
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

// PrettifiedMetrics returns metric details with formatted values.
func (details EventViewModelDetails) PrettifiedMetrics() EventViewModelDetails {
	result := make(EventViewModelDetails, 0)
	for _, detail := range details {
		if detail.Tag != "metric" {
			continue
		}

		floatValue, err := toFloat(detail.Value)
		if err != nil {
			result = append(result, detail)
			continue
		}

		label := detail.Label

		// format bits with proper unit
		if strings.HasPrefix(detail.Name, "bits") {
			var prefix string
			floatValue, prefix = formatBits(floatValue)
			label = fmt.Sprintf("%sbits/s", prefix)
		}

		// prevent showing fractions when unnecessary
		stringValue := fmt.Sprintf("%.2f", floatValue)
		if _, fraction := math.Modf(floatValue); fraction < 0.05 {
			stringValue = fmt.Sprintf("%.0f", floatValue)
		}

		formatted := &EventViewModelDetail{
			Name:  detail.Name,
			Label: label,
			Tag:   detail.Tag,
			Value: stringValue,
		}

		result = append(result, formatted)
	}
	return result
}

func toFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("don't know how to convert %T to float64", v)
	}
}

func formatBits(bits float64) (float64, string) {
	const unit = 1024
	const suffixes = "KMGTPE"

	exp := math.Floor(math.Log(bits) / math.Log(unit))
	suffix := ""
	if exp > 0 {
		suffix = string(suffixes[int(exp)-1])
	}
	value := bits / math.Pow(unit, exp)

	return value, suffix
}

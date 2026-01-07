package render

import (
	"encoding/json"
	"testing"
)

func TestEventViewModelJSON(t *testing.T) {
	jsonInput := `{
		"Type": "alarm",
		"StartTimestamp": 1672531200,
		"EndTimestamp": 1672534800,
		"IsTestEvent": true,
		"BaseDomain": "example.com"
	}`

	var event EventViewModel
	err := json.Unmarshal([]byte(jsonInput), &event)
	if err != nil {
		t.Fatalf("Failed to unmarshal EventViewModel: %v", err)
	}

	if event.Type != "alarm" {
		t.Errorf("Expected Type 'alarm', got '%s'", event.Type)
	}
	if event.StartTimestamp != 1672531200 {
		t.Errorf("Expected StartTimestamp 1672531200, got %d", event.StartTimestamp)
	}
	if event.EndTimestamp != 1672534800 {
		t.Errorf("Expected EndTimestamp 1672534800, got %d", event.EndTimestamp)
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal EventViewModel: %v", err)
	}

	var outputMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &outputMap); err != nil {
		t.Fatalf("Failed to unmarshal output JSON: %v", err)
	}

	if outputMap["Type"] != "alarm" {
		t.Errorf("Expected Type to be present in output")
	}

	hiddenFields := []string{"StartTimestamp", "EndTimestamp", "IsTestEvent", "BaseDomain"}
	for _, field := range hiddenFields {
		if _, exists := outputMap[field]; exists {
			t.Errorf("Field '%s' should be hidden but was found in output JSON", field)
		}
	}
}

func TestEventViewModelDetailJSON(t *testing.T) {
	jsonInput := `{
		"Name": "cpu_usage",
		"Value": 85,
		"Tag": "metric"
	}`

	var detail EventViewModelDetail
	err := json.Unmarshal([]byte(jsonInput), &detail)
	if err != nil {
		t.Fatalf("Failed to unmarshal EventViewModelDetail: %v", err)
	}

	if detail.Name != "cpu_usage" {
		t.Errorf("Expected Name 'cpu_usage', got '%s'", detail.Name)
	}
	if detail.Tag != "metric" {
		t.Errorf("Expected Tag 'metric', got '%s'", detail.Tag)
	}

	jsonData, err := json.Marshal(detail)
	if err != nil {
		t.Fatalf("Failed to marshal EventViewModelDetail: %v", err)
	}

	var outputMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &outputMap); err != nil {
		t.Fatalf("Failed to unmarshal output JSON: %v", err)
	}

	if outputMap["Name"] != "cpu_usage" {
		t.Errorf("Expected Name to be present")
	}
	if _, exists := outputMap["Tag"]; exists {
		t.Errorf("Field 'Tag' should be hidden but was found in output JSON")
	}
}

func TestNotificationViewModelJSON(t *testing.T) {
	jsonInput := `{
		"CompanyID": 12345,
		"CompanyName": "Acme Corp",
		"Config": {
			"BaseDomain": "portal.kentik.com"
		},
		"Events": [
			{
				"Type": "alarm",
				"Description": "High CPU"
			}
		]
	}`

	var vm NotificationViewModel
	err := json.Unmarshal([]byte(jsonInput), &vm)
	if err != nil {
		t.Fatalf("Failed to unmarshal NotificationViewModel: %v", err)
	}

	if vm.CompanyID != 12345 {
		t.Errorf("Expected CompanyID 12345, got %d", vm.CompanyID)
	}
	if vm.CompanyName != "Acme Corp" {
		t.Errorf("Expected CompanyName 'Acme Corp', got '%s'", vm.CompanyName)
	}
	if vm.Config == nil || vm.Config.BaseDomain != "portal.kentik.com" {
		t.Errorf("Expected Config.BaseDomain 'portal.kentik.com', got %v", vm.Config)
	}
	if len(vm.RawEvents) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(vm.RawEvents))
	}
	if vm.RawEvents[0].Type != "alarm" {
		t.Errorf("Expected first event type 'alarm', got '%s'", vm.RawEvents[0].Type)
	}

	jsonData, err := json.Marshal(vm)
	if err != nil {
		t.Fatalf("Failed to marshal NotificationViewModel: %v", err)
	}

	var outputMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &outputMap); err != nil {
		t.Fatalf("Failed to unmarshal output JSON: %v", err)
	}

	if val, ok := outputMap["CompanyID"]; !ok || int(val.(float64)) != 12345 {
		t.Errorf("Expected CompanyID to be present and equal to 12345")
	}

	hiddenFields := []string{"CompanyName", "Config", "Events", "RawEvents"}
	for _, field := range hiddenFields {
		if _, exists := outputMap[field]; exists {
			t.Errorf("Field '%s' should be hidden but was found in output JSON", field)
		}
	}
}

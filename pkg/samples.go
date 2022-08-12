package main

import "time"

var TestingViewModels = map[string]*NotificationViewModel{
	"insight":    TestInsight,
	"alarm":      TestAlarm,
	"synthetics": TestSynth,
	"mitigation": TestMitigation,
	"digest":     TestDigest,
}

var timeNow = timeParseOrPanic(time.RFC3339, "2022-04-13T19:50:05Z")

func timeParseOrPanic(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}

var TestInsight = &NotificationViewModel{
	CompanyID:   1001,
	CompanyName: "Kentik Test Company",
	Now:         timeNow,

	RawEvents: []*EventViewModel{
		{
			Type:           "insight",
			Description:    "Insight for Total Traffic Today",
			IsActive:       true,
			StartTime:      "2021-11-11 18:33:53 UTC",
			EndTime:        "2021-11-11 18:33:53 UTC",
			CurrentState:   "n/a",
			PreviousState:  "n/a",
			StartTimestamp: 1636655633,
			EndTimestamp:   1636655633,
			Importance:     ViewModelImportance(4),
			GroupName:      "Total Traffic Today",
			Details: EventViewModelDetails{
				&EventViewModelDetail{
					Name:  "InsightName",
					Label: "System Name",
					Value: "interconnection.costs.bpsDayOverDay",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "InsightID",
					Label: "ID",
					Value: "k123456",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "InsightDataSourceType",
					Label: "Source",
					Value: "ksql",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "InsightPlainDescription",
					Label: "Description",
					Value: "You sent and received 26% more traffic (+155 Gbits/s) this week compared to last week.",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "InsightDetailsURL",
					Label: "Open Details",
					Value: "https://portal.kentik.com/v4/operate/insights/k123456",
					Tag:   "url",
				},
				&EventViewModelDetail{
					Name:  "InsightsMainURL",
					Label: "Open Insights Dashboard",
					Value: "https://portal.kentik.com/v4/operate/insights",
					Tag:   "url",
				},
			},
		},
	},
	Config: &NotificationViewConfig{
		BaseDomain: "portal.kentik.com",
		EmailTo:    []string{"your@email.address"},
	}}

var TestDigest = &NotificationViewModel{
	CompanyID:   1001,
	CompanyName: "Kentik Test Company",
	Now:         timeNow,

	RawEvents: []*EventViewModel{
		{
			Type:           "insight",
			Description:    "Insight for Total Traffic Today",
			IsActive:       true,
			StartTime:      "2021-11-11 18:33:53 UTC",
			EndTime:        "2021-11-11 18:33:53 UTC",
			CurrentState:   "n/a",
			PreviousState:  "n/a",
			StartTimestamp: 1636655633,
			EndTimestamp:   1636655633,
			Importance:     ViewModelImportance(4),
			GroupName:      "Total Traffic Today",
			Details: EventViewModelDetails{
				&EventViewModelDetail{
					Name:  "InsightName",
					Label: "System Name",
					Value: "interconnection.costs.bpsDayOverDay",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "InsightID",
					Label: "ID",
					Value: "k123456",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "InsightDataSourceType",
					Label: "Source",
					Value: "ksql",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "InsightPlainDescription",
					Label: "Description",
					Value: "You sent and received 26% more traffic (+155 Gbits/s) this week compared to last week.",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "InsightDetailsURL",
					Label: "Open Details",
					Value: "https://portal.kentik.com/v4/operate/insights/k123456",
					Tag:   "url",
				},
				&EventViewModelDetail{
					Name:  "InsightsMainURL",
					Label: "Open Insights Dashboard",
					Value: "https://portal.kentik.com/v4/operate/insights",
					Tag:   "url",
				},
			},
		},
		{
			Type:           "insight",
			Description:    "Insight for Total Traffic Today",
			IsActive:       true,
			StartTime:      "2021-11-11 18:33:53 UTC",
			EndTime:        "2021-11-11 18:33:53 UTC",
			CurrentState:   "n/a",
			PreviousState:  "n/a",
			StartTimestamp: 1636655633,
			EndTimestamp:   1636655633,
			Importance:     ViewModelImportance(4),
			GroupName:      "Total Traffic Today",
			Details: EventViewModelDetails{
				&EventViewModelDetail{
					Name:  "InsightName",
					Label: "System Name",
					Value: "custom.insight.UDP Fragments Attack",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "InsightID",
					Label: "ID",
					Value: "a197790252",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "InsightDataSourceType",
					Label: "Source",
					Value: "alerting",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "InsightPlainDescription",
					Label: "Description",
					Value: "An alarm was triggered for Dest IP/CIDR: 209.50.158.100",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "bits",
					Label: "",
					Value: 58555.9140625,
					Tag:   "metric",
				},
				&EventViewModelDetail{
					Name:  "packets",
					Label: "",
					Value: 11.200035095214844,
					Tag:   "metric",
				},
				&EventViewModelDetail{
					Name:  "unique_src_ip",
					Label: "",
					Value: 1,
					Tag:   "metric",
				},
				&EventViewModelDetail{
					Name:  "InsightDetailsURL",
					Label: "Open Details",
					Value: "https://portal.kentik.com/v4/operate/insights/a197790252",
					Tag:   "url",
				},
				&EventViewModelDetail{
					Name:  "InsightsMainURL",
					Label: "Open Insights Dashboard",
					Value: "https://portal.kentik.com/v4/operate/insights",
					Tag:   "url",
				},
			},
		},
	},
	Config: &NotificationViewConfig{
		BaseDomain: "portal.kentik.com",
		EmailTo:    []string{"your@email.address"},
	}}

var TestAlarm = &NotificationViewModel{
	CompanyID:   1002,
	CompanyName: "ACME Incorporated",
	Now:         timeNow,
	RawEvents: []*EventViewModel{
		{
			Type:           "alarm",
			Description:    "Alarm for UDP Fragments Attack Active",
			IsActive:       true,
			StartTime:      "2021-11-17 10:29:32 UTC",
			EndTime:        "ongoing",
			CurrentState:   "alarm",
			PreviousState:  "new",
			StartTimestamp: 1637144972,
			EndTimestamp:   0,
			Importance:     ViewModelImportance(5),
			GroupName:      "Alarm for UDP Fragments Attack",
			Details: EventViewModelDetails{
				&EventViewModelDetail{
					Name:  "AlarmID",
					Label: "ID",
					Value: 197790252,
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "AlarmSeverity",
					Label: "Severity",
					Value: "major",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "AlarmPolicyName",
					Label: "Source Policy Name",
					Value: "UDP Fragments Attack",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "AlarmPolicyID",
					Label: "Policy ID",
					Value: 432,
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "AlarmThresholdID",
					Label: "Threshold ID",
					Value: 14444,
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "AlarmBaselineSource",
					Label: "Baseline Source",
					Value: "ACT_BASELINE_MISSING_DEFAULT_INSTEAD_OF_LOWEST",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "AlarmBaselineDescription",
					Label: "Baseline Source Info",
					Value: "No baseline value was found for this key and this key's current value exceeded the default value and there were no other (lowest) values in the baseline available.",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "IP_dst",
					Label: "Dest IP/CIDR",
					Value: "209.50.158.100",
					Tag:   "dimension",
				},
				&EventViewModelDetail{
					Name:  "i_device_id",
					Label: "Device ID",
					Value: 1234,
					Tag: "device",
				},
				&EventViewModelDetail{
					Name:  "i_device_labels",
					Label: "Device Labels",
					Value: "ACME1, ACME2",
					Tag: "device_labels",
				},
				&EventViewModelDetail{
					Name:  "DashboardAlarmURL",
					Label: "Open in Dashboard",
					Value: "https://portal.kentik.com/v4/library/dashboards/49",
					Tag:   "url",
				},
				&EventViewModelDetail{
					Name:  "InsightAlarmURL",
					Label: "Open Insight",
					Value: "https://portal.kentik.com/v4/core/insights/a197790252",
					Tag:   "url",
				},
				&EventViewModelDetail{
					Name:  "AttackLogURL",
					Label: "Open Log",
					Value: "https://portal.kentik.com/v4/protect/ddos/analyze/log/197790252",
					Tag:   "url",
				},
			},
		},
	},
	Config: &NotificationViewConfig{
		BaseDomain: "portal.kentik.com",
		EmailTo:    []string{"your@email.address"},
	}}

var TestSynth = &NotificationViewModel{
	CompanyID:   1003,
	CompanyName: "ACME Incorporated",
	Now:         timeNow,
	RawEvents: []*EventViewModel{
		{
			Type:           "synthetic",
			Description:    "Test this is dns grid test Unhealthy",
			IsActive:       true,
			StartTime:      "2021-11-29 11:43:31 UTC",
			EndTime:        "2021-11-29 11:43:31 UTC",
			CurrentState:   "clear",
			PreviousState:  "alarm",
			StartTimestamp: 1638186211,
			EndTimestamp:   1638186211,
			Importance:     ViewModelImportance(5),
			GroupName:      "this is dns grid test",
			Details: EventViewModelDetails{
				&EventViewModelDetail{
					Name:  "Health",
					Label: "",
					Value: "Unhealthy",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "TestID",
					Label: "",
					Value: 1228,
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "Test",
					Label: "Test Name",
					Value: "this is dns grid test",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "TestType",
					Label: "Test Type",
					Value: "dns-grid",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "level",
					Label: "",
					Value: 300.0,
					Tag:   "metric",
				},
				&EventViewModelDetail{
					Name:  "SyntheticsTestURL",
					Label: "Open Test Details",
					Value: "https://portal.kentik.com/v4/synthetics/tests/1228/results?start=1638186211&end=1638186211",
					Tag:   "url",
				},
			},
		},
	},
	Config: &NotificationViewConfig{
		BaseDomain: "portal.kentik.com",
		EmailTo:    []string{"your@email.address"},
	}}

var TestMitigation = &NotificationViewModel{
	CompanyID:   1001,
	CompanyName: "ACME Incorporated",
	Now:         timeNow,
	RawEvents: []*EventViewModel{
		{
			Type:           "mitigation",
			Description:    "Mitigation for Policy UDP Fragments Attack Clear",
			IsActive:       false,
			StartTime:      "2020-12-18 20:49:51 UTC",
			EndTime:        "2020-12-18 22:01:23 UTC",
			CurrentState:   "archived",
			PreviousState:  "ackRequired",
			StartTimestamp: 1608324591,
			EndTimestamp:   1608328883,
			Importance:     ViewModelImportance(1),
			GroupName:      "Mitigation for Policy UDP Fragments Attack",
			Details: EventViewModelDetails{
				&EventViewModelDetail{
					Name:  "MitigationID",
					Label: "ID",
					Value: 12345,
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "MitigationPolicyID",
					Label: "Policy ID",
					Value: 7890,
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "MitigationPolicyName",
					Label: "Policy Name",
					Value: "UDP Fragments Attack",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "MitigationPlatformID",
					Label: "Platform ID",
					Value: 1747,
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "MitigationMethodID",
					Label: "Method ID",
					Value: 775,
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "MitigationAlarmID",
					Label: "Alarm ID",
					Value: 197790252,
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "MitigationAlertIp",
					Label: "IP/CIDR Address",
					Value: "92.204.191.35/32",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "LastMitigationEvent",
					Label: "",
					Value: "skipWait",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "IP_dst",
					Label: "",
					Value: "209.50.158.100",
					Tag:   "dimension",
				},
				&EventViewModelDetail{
					Name:  "AlarmSeverity",
					Label: "",
					Value: "major",
					Tag:   "",
				},
				&EventViewModelDetail{
					Name:  "MitigationURL",
					Label: "Open Mitigation Details",
					Value: "https://portal.kentik.com/v4/protect/mitigations/12345",
					Tag:   "url",
				},
			},
		},
	},
	Config: &NotificationViewConfig{
		BaseDomain: "portal.kentik.com",
		EmailTo:    []string{"your@email.address"},
	}}

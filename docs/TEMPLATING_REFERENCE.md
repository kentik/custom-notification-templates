# Using Custom Webhook Templating

## Introduction

**You can define templates for custom webhooks using [Go Template Syntax](https://pkg.go.dev/text/template). We highly recommend that you familiarize yourself with the [official docs](https://pkg.go.dev/text/template) or this [Hashicorp article](https://learn.hashicorp.com/tutorials/nomad/go-template-syntax). Nowadays, a vast majority of system integrations use JSON as the de-facto standard, therefore examples here focus on this format. Still, Go templating is flexible enough to use other output formats (e.g. XML or syslog).**

## Helper Functions

Before digging into the notification data structure, it's worth getting to know the helper functions that are added on top of Go templates.

### String Functions

String functions transform string values for the notification context.

- `toUpper` - Converts all string letters to uppercase.
- `title` - Converts all letters at the beginning of words to uppercase.
- `trimSpace` - Drops spaces from the beginning and ending of the string.

### JSON Functions

JSON functions help build valid JSON payloads in a flexible manner.

- `toJSON` (also with alias: `j`) - Translates the object into a JSON-compliant value. It is crucial that you use this function for EventDetails API (that is elaborated on later).
- `explodeJSONKeys` (also with alias: `x`) - Converts a JSON-compliant object value while extracting the properties. Useful to combine different levels of the context into a single one. Use this with caution, as JSON format is strict when it comes to comma separation, and the engine that renders the templates does not provide any kind of JSON sanitization.

### Array Helper Functions

Joining functions are especially handy for the proper and easy construction of an array of items (objects), adding separators between arrayed items.

- `join index` - Prints a comma between all items, unless index is equal to 0.
- `joinWith index string` - Like `join` but the separator can be customized to use a character other than comma (also useful for cases other than JSON).

## Notification Structure

Kentik Portal supports the following notification structures:

- **immediate notifications** - A notification is sent as soon as the system detects the condition for which the subscriber wishes to be informed.
- **digest notifications** - A summary is sent periodically according to a user-specified schedule (currently just as emails).

The data structure available within the template body (also known as a **context**) reflects both of the above cases.

Property names are TitleCamelCased, ASCII strings. There are both static fields that provide the data, as well as convenience fields (methods and functions that simplify creating various types of templates).

### Root Context

Top-level root context is event-agnostic and focuses on the circumstances of when the notification is published. The following fields are available:

- `Headline` *string* - Main title of the notification which reflects the types of both the event's source and the notification itself (if it is an immediate alert or digest), e.g. `Kentik Synthetics Alert`.
- `Summary` *string* - A description of the notification; either a description of a single message or a summary of how many issues are active and inactive.
- `CompanyID` *integer* - The company ID from which this notification is sent.
- `CompanyName` *string* - The corresponding company name.
- `NowUnix` *integer* - Unix timestamp of when the notification was last published (note, it is not the timestamp of a triggered event; it is when the notification went out).
- `NowDate` *string* - Formats `Now` timestamp into long US date, e.g. January 6, 2021.
- `NowRFC3339` *string* - Formats `Now` timestamp using [RFC 3339](https://medium.com/easyread/understanding-about-rfc-3339-for-datetime-formatting-in-software-engineering-940aa5d5f68a) (e.g. 2019-19-12T07:20:50.52A).
- `NowDatetime` *string* - Formats `Now` timestamp using readable universal format (e.g. 2006-01-02 15:04:05 UTC).
- `BasePortalURL` *string* - Kentik Portal's base URL (https://portal.kentik.com).
- `NotificationsSettingsURL` *string* - URL to Notifications settings within the Kentik Portal v4.
- `IsSingleEvent` *boolean* - Value indicating if there is just a single event within the context (true/false).
- `IsMultipleEvents` *boolean* - Value indicating if there are multiple events (true/false).
- `Events` *array of objects* - An array of the event(s).
- `Event` *object* - An alias to to the first element of the Events array.
- `ActiveCount` *integer* - The number of events that are still considered active (ongoing).
- `InactiveCount` *integer* - The number of events that are no longer considered active (past).

**Example - Native HTML template:**

```go-template
<!doctype html>
<html>
<head><title> {{ .Headline | toUpper }}: {{ .Summary | title }}</title></head>
<body>
<h1>{{.Headline}} for {{.CompanyName}}: {{ .Summary }}</h1>
<h2>{{ .NowRFC3339 }}</h2>
<p><a href="{{- .NotificationsSettingsURL -}}">Manage your notifications</a></p>
</body>
</html>
```

This will result in the following HTML payload on render:

```html
<!doctype html>
<html>
<head><title> KENTIK ALERT: Alarm For DDoS Protect Policy Active</title></head>
<body>
<h1>Kentik Alert for ACME Incorporated: Alarm for DDoS Protect Policy Active</h1>
<h2>2021-12-01T12:37:49Z</h2>
<p><a href="https://portal.kentik.com/v4/settings/notifications">Manage your notifications</a></p>
</body>
</html>
```

**Example - Basic JSON template that supports both immediate and digest notifications:**

```go-template
{
  "headline": "{{.Headline}}",
  "timestamp": "{{.NowDatetime}}",
  "summary": "{{.Summary}}",
  {{- if .IsSingleEvent  -}}
    {{- with .Event -}}
      "event": {
        "type": "{{.Type}}",
        "description": "{{.Description}}",
        "state": "{{.CurrentState}}"
      }
    {{- end -}}
  {{- else -}}
    "events": [
      {{- range $index, $event := .Events -}}
        {{- join $index -}}
        {
          "type": "{{.Type}}",
          "description": "{{.Description}}",
          "state": "{{.CurrentState}}"
        }
      {{- end -}}
    ]
  {{- end -}}
}
```

This will result in the following JSON payload on render:

```json
{
  "date": "December 1, 2021",
  "event": {
    "description": "Alarm for DDoS Protect Policy Active",
    "state": "alarm",
    "type": "alarm"
  },
  "headline": "Kentik Alert",
  "summary": "Alarm for DDoS Protect Policy Active",
  "timestamp": "2021-12-01T12:36:27Z"
}
```

> ℹ Note: We'll discuss the custom `join` function later in this document.

### Event-level Properties

- `Type` *string* - Enumerable value indicating the type (source) of the event (e.g. `alarm` or `synthetic`). In order to check the type, we recommend that you use the following boolean properties instead:
  - `IsAlarm` *boolean*
  - `IsSynthetic` *boolean*
  - `IsMitigation` *boolean*
  - `IsInsight` *boolean*
  - `IsCustomInsight` *boolean* (special case of the insight)
- `Description` *string* - English description of the event.
- `IsActive` *boolean* - Indicates whether the event is considered active (the trigger causing the event to be propagated is ongoing).
- `StartTime` *string* - RFC-3339-formatted timestamp displays when the trigger first occurred.
- `EndTime` *string* - RFC-3339-formatted timestamp displays when the trigger stopped occurring, or displays the string `ongoing` if it is still active.
- `CurrentState` *string* - Current state of the trigger (depends on the type of state).
- `PreviousState` *string* - Previous state of the trigger (depends on the type of state).
- `StartTimestamp` *integer* - Unix timestamp of when the trigger first occurred.
- `EndTimestamp` *integer* - Unix timestamp of when the trigger stopped occurring.
- `Details` *array of objects* - Type-specific properties collection grouped by tags.

**Example - JSON template for immediate notifications with a custom header:**

```go-template
{
  {{- if .IsSingleEvent -}}
    {{- with .Event -}}
      {{- if .IsActive -}}
        {{- if .IsAlarm -}}
          "header": "Alarm raised!",
        {{- else if .IsSynthetic -}}
          "header": "Synthetic test failure!",
        {{- else if .IsMitigation -}}
          "header": "Mitigation activated!",
        {{- end -}}
      {{- else -}}
        {{- if .IsAlarm -}}
          "header": "Alarm cleared",
        {{- else if .IsSynthetic -}}
          "header": "Synthetic test healthy",
        {{- else if .IsMitigation -}}
          "header": "Mitigation deactivated",
        {{- end -}}
      {{- end -}}
      "description: "{{- .Description -}}"
    {{- end -}}
  {{- end -}}
}
```

This will result in the following JSON payload on render:

```json
{
  "description": "Alarm for DDoS Protect Policy Active",
  "header": "Alarm raised!"
}
```

### Event Details

While each event source produces a variety of data, it can be published through a single notification output channel. In order to provide sufficient flexibility on the template level, we provide an Event Details API. You can decide where those details should land without many conditions checks (e.g. if/elses which are not easy to write in the templates).

Details themselves are just an array of objects with `Name` and `Value` (optionally also with `Label`). Each detail may also have a corresponding `Tag` that associates a given detail with a semantic **noun**.

> ℹ Note: `Value` type can be one the basic types — it can be `int`, `float`, `string`, or `null`.

> ℹ Note: When constructing JSON with the Event Details API, it is important to use the `toJSON` (or `j`) function (also as pipe), as the data returned by the Event Details API are not marshaled to JSON by default (this is dynamic data).

#### Use Tags to Group Details

Details can be grouped using tags. Details without a tag are called general tags. Tags can be:

- `dimension` - An event-specific parameter (e.g. device ID, IP addresses or ranges, or prefixes).
- `metric` - A numerical parameter (e.g. bandwidth, number of packets, or IP addresses).
- `url` - Action links that allow you to take action with a given event.

#### Use Filters to Collect Details

The following filtering methods return a details object, so that other Detail methods can still be invoked.

- `Details.General` - Filter out the general details (i.e. details without any tag specified).
- `Details.WithTag tag` - Filter the details to show those with the given tag.
- `Details.WithNames ...names` - Filter the details to display only those with the given names.

**Example - Template rendering all event URLs and metrics:**

```go-template
{
  {{- with .Event -}}
    "metrics": {{ .Details.WithTag "metric" | j }},
    "links": {{ .Details.WithTag "url" | j }}
  {{- end -}}
}
```

This will result in the following JSON payload on render:

```json
{
  "links": [
    {
      "Label": "Open in Dashboard",
      "Name": "DashboardAlarmURL",
      "Value": "https://portal.kentik.com/v4/library/dashboards/11"
    },
    {
      "Label": "Open Insight",
      "Name": "InsightAlarmURL",
      "Value": "https://portal.kentik.com/v4/core/insights/a197790253"
    },
    {
      "Label": "Open Log",
      "Name": "AttackLogURL",
      "Value": "https://portal.kentik.com/v4/protect/ddos/analyze/log/197790253"
    }
  ],
  "metrics": [
    {
      "Name": "bits",
      "Value": 58555.9140625
    },
    {
      "Name": "packets",
      "Value": 1333270.75
    },
    {
      "Name": "unique_src_ip",
      "Value": 1
    }
  ]
}
```

### Working with Details Collection Items

#### Details collection transformation

- `Details.ToMap` - Converts an array into the map (object), with names becoming the property keys.
- `Details.Names` - Provides an array of detail names (the ones that are present from the given list).
- `Details.Values` - Provides an array of detail values.

#### Details collection checks

- `Details.Has name` - Checks if the details include the item with a given name.
- `Details.HasTag tag` - Checks if the details include any item of a given tag.

#### Details collection getters

- `Details.Get name` - Picks a single detail of a given name or a nullish one if it is not found.
- `Details.GetValue name` - Picks just a value of the detail or `nil` (JSON's `null`) if there aren't any.

A single detail can also have one helper method:

- `Detail.LabelOrName` - Use `Label`, if present, or `Name` otherwise.

### Details Collection Reference

#### General

General details are the ones without tag specified. Particular names of general details depend on the source type, therefore it is recommended to use `{{- .Details.General.ToMap | j -}}` instead of referring the particular details name. However, the latter is also possible and supported.

The list of general detail names:

- Insights: `InsightName`,`InsightID`, `InsightDataSourceType`, `InsightPlainDescription`
- Alarm state change: `AlarmID`, `AlarmSeverity`, `AlarmPolicyName`, `AlarmPolicyID`, `AlarmThresholdID`, `AlarmBaselineSource`, `AlarmBaselineDescription`
- Mitigation: `MitigationID`, `MitigationPolicyID`, `MitigationPolicyName`, `MitigationPlatformID`, `MitigationPlatformName`, `MitigationMethodID`, `MitigationMethodName`, `MitigationAlarmID`, `MitigationAlertIp`, `LastMitigationEvent`, `AlarmSeverity`
- Synthetics: `TestName`, `TestType`, `Health`, `TestID`

Example:

```json
[{
  "Label": "ID",
  "Name": "AlarmID",
  "Value": "216148908",
},
{
  "Label": "Severity",
  "Name": "AlarmSeverity",
  "Value": "major",
},
{
  "Label": "Threshold ID",
  "Name": "AlarmThresholdID",
  "Value": "12716",
},
{
  "Label": "Policy ID",
  "Name": "AlarmPolicyID",
  "Value": "4085",
},
{
  "Label": "Source Policy Name",
  "Name": "AlarmPolicyName",
  "Value": "V4 DDoS - UDP Flood",
}]
```

Value types of general tags depends on the name.

#### Urls

Tag: `url`

In case of url details, `Value` property represents the full URL that can be used as hyperlink to access web view of the event that triggered given notification.

Example:

```json
[{
  "Name": "DashboardAlarmURL",
  "Label": "Open in Dashboard",
  "Value": "https://portal.kentik.com/v4/library/dashboards/49",
  "Tag": "url"
},
{
  "Name": "InsightAlarmURL",
  "Label": "Open Insight",
  "Value": "https://portal.kentik.com/v4/core/insights/a197790252",
  "Tag": "url"
},
{
  "Name": "AttackLogURL",
  "Label": "Open Log",
  "Value": "https://portal.kentik.com/v4/protect/ddos/analyze/log/197790252",
  "Tag": "url"
}]
```

#### Dimensions and metrics

Tags: `dimension`, `metric`

Typical for alerting-source notifications (alarms state change and mitigations). Provide information about alert dimensions and metrics. Values are pre-formatted for good visual presentation but still should be numeric where applicable. Note the convention for a metric: the `Name` and `Label` is used to pass the unit information.

Example:

```json
[{
  "Label": "Bits/second",
  "Name": "bits",
  "Tag": "metric",
  "Value": 48878.94921875
},
{
  "Label": "Packets/second",
  "Name": "packets",
  "Tag": "metric",
  "Value": 240.266668319702148
},
{
  "Label": "Unique Source IPs",
  "Name": "unique_src_ip",
  "Tag": "metric",
  "Value": 1
},
{
  "Label": "Dest IP/CIDR",
  "Name": "IP_dst",
  "Tag": "dimension",
  "Value": "208.76.14.235"
},
{
  "Label": "Device",
  "Name": "i_device_id",
  "Tag": "dimension",
  "Value": "32650"
},
{
  "Label": "Site",
  "Name": "i_device_site_name",
  "Tag": "dimension",
  "Value": "Ashburn DC3"
}]
```

#### Labels

Tags: `label`, `device_label`

Provide information about labels related with a given notification. Tag `device_label` is used for device-related labels, and basic `label` is used in general cases.

Note that the structure of labels `Value` is an object, not a simple type (string or number) and it includes `Name` (actual label), `Color` in hex and `IsDark` - a boolean which hints whether the color is dark or not.

The `Name` property of the detail is irrelevant, please use the `Value.Name` instead.

Example:

```json
[{
  "Name": "TestLabel1",
  "Tag": "label",
  "Value": {
    "Color": "#ff0000",
    "IsDark": true,
    "Name": "ACME"
  },
},
{
  "Label": "TestLabel2",
  "Tag": "label",
  "Value": {
    "Color": "#ffffff",
    "IsDark": false,
    "Name": "Foobar"
  }
}]
```

#### Device

Tag: `device`

Information about device that is associated with a given notification (e.g. provided when device is a dimension for alerting policy).

In case of device details, the names are fully meaningful and the templates can rely on them (e.g. `DeviceName` for a name and `DeviceType` for a type).

Example:

```json
[{
  "Label": "Device ID",
  "Name": "DeviceId",
  "Tag": "device",
  "Value": "32650"
},
{
  "Label": "Device",
  "Name": "DeviceName",
  "Tag": "device",
  "Value": "QFX_123456_tee"
},
{
  "Label": "Device Type",
  "Name": "DeviceType",
  "Tag": "device",
  "Value": "router"
}]
```

#### Issues

Tag: `issue`

Compound structure that provides rich metadata for abstract _issues_ related with a given notification. Currently used by synthetics notifications.

Each issue item has the same fields within (see below the example). The `Label` and `Name` attributes has minor meaning, all neccessary information about the issue is within the `Value`.

Example:

```jsonc
[{
  "Label": "Issue #1",
  "Name": "Issue1",
  "Tag": "issue",
  "Value": {
    "Description": "foo.kentik.com: PING ⇒ 208.76.14.180 went critical from healthy", // a summary of an issue
    "DetailedInfo": [
      // array of strings that provide more details on an issue
      "Packet loss: 40.00% (critical)",
      "Latency: 0.08ms (warning)",
      "Jitter: 0.01ms (healthy)"
    ],
    "Labels": [
      // labels applicable to this issue
      {
        "Color": "#ff0000",
        "IsDark": true,
        "Name": "MyAgentLabel"
      }
    ],
    "Origin": "foo.kentik.com", // source of the issue
    "Severity": "critical",
    "Status": "critical",
    "Target": "208.76.14.180", // target of the issue
    "Type": "PING",
    "Url": "http://portal.kentik.com/v4/synthetics/tests/1234/results/agent/266?start=1655472388",
    "UrlLabel": "Open Task Details" // label for the URL
  }
}]
```

#### Statistics

Tag: `statistic`

General statistics metadata to provide context of the notification. Currently used by synthetics notifications.

Example:

```json
[{
  "Label": "Total sub-tests critical",
  "Name": "TotalSubtestsCurrentlyCritical",
  "Tag": "statistic",
  "Value": 1,
},
{
  "Label": "Total sub-tests critical on packet loss",
  "Name": "PacketlossSubtestsCurrentlyCritical",
  "Tag": "statistic",
  "Value": 1
},
{
  "Label": "Total sub-tests failing",
  "Name": "TotalSubtestsCurrentlyFailing",
  "Tag": "statistic",
  "Value": 4
},
{
  "Label": "Total sub-tests healthy",
  "Name": "TotalSubtestsCurrentlyHealthy",
  "Tag": "statistic",
  "Value": 11
}]
```

## Complete Template Examples

Please refer to [templates directory](../templates/) for more inspiration.

**Example - Using functions and ToMap**

```go-template
{
  {{- with .Event -}}
    {{- if .Details.HasTag "metric" -}}
      "metrics": {{ (.Details.WithTag "metric").ToMap | j }},
    {{- end -}}
    {{- if .Details.HasTag "metric" -}}
      "links": {{ (.Details.WithTag "url").ToMap | j }},
    {{- end -}}
    "custom": {{ .Details.General.ToMap | j }}
  {{- end -}}
}
```

This will result in the following JSON payload on render:

```json
{
  "custom": {
    "AlarmBaselineSource": "ACT_BASELINE_MISSING_DEFAULT_INSTEAD_OF_LOWEST",
    "AlarmID": 197790253,
    "AlarmPolicyID": 297,
    "AlarmPolicyName": "DDoS Protect Policy",
    "AlarmSeverity": "critical",
    "AlarmThresholdID": 14444,
    "Baseline": 3076
  },
  "links": {
    "AttackLogURL": "https://portal.kentik.com/v4/protect/ddos/analyze/log/197790253",
    "DashboardAlarmURL": "https://portal.kentik.com/v4/library/dashboards/11",
    "InsightAlarmURL": "https://portal.kentik.com/v4/core/insights/a197790253"
  },
  "metrics": {
    "bits": 58555.9140625,
    "packets": 1333270.75,
    "unique_src_ip": 1
  }
}
```

**Example - Complex example that flattens general details together with common properties**

```go-template
{
  {{- . | toJSON | explodeJSONKeys -}},
  {{- if .IsSingleEvent  -}}
    {{- with .Event -}}
      {{- . | toJSON | explodeJSONKeys -}},
      {{- .Details.General.ToMap | toJSON | explodeJSONKeys -}},
      "Metrics": {{- (.Details.WithTag "metric").ToMap | toJSON -}},
      "Dimensions": {{- (.Details.WithTag "dimension").ToMap | toJSON -}},
      "Links": {{- (.Details.WithTag "url").ToMap | toJSON -}}
    {{- end -}}
  {{- else -}}
  "Events": [
    {{- range $index, $event := .Events -}}
      {{- join $index -}}
      {
        {{- . | toJSON | explodeJSONKeys -}},
        {{- .Details.General.ToMap | toJSON | explodeJSONKeys -}},
        "Metrics": {{- (.Details.WithTag "metric").ToMap | toJSON -}},
        "Dimensions": {{- (.Details.WithTag "dimension").ToMap | toJSON -}},
        "Links": {{- (.Details.WithTag "url").ToMap | toJSON -}}
      }
    {{- end -}}
  ]
  {{- end -}}
}
```

This will result in the following JSON payload on render:

```json
{
  "AlarmBaselineSource": "ACT_BASELINE_MISSING_DEFAULT_INSTEAD_OF_LOWEST",
  "AlarmID": 197790253,
  "AlarmPolicyID": 297,
  "AlarmPolicyName": "DDoS Protect Policy",
  "AlarmSeverity": "critical",
  "AlarmThresholdID": 14444,
  "Baseline": 3076,
  "CompanyID": 1001,
  "CurrentState": "alarm",
  "Description": "Alarm for DDoS Protect Policy Active",
  "Dimensions": {
    "IP_dst": "209.50.158.100",
    "InterfaceID_src": "37001",
    "i_src_connect_type_name": "",
    "i_src_network_bndry_name": "",
    "i_src_provider_classification": "---",
    "i_trf_origination": "outside"
  },
  "EndTime": "ongoing",
  "IsActive": true,
  "Links": {
    "AttackLogURL": "https://portal.kentik.com/v4/protect/ddos/analyze/log/197790253",
    "DashboardAlarmURL": "https://portal.kentik.com/v4/library/dashboards/11",
    "InsightAlarmURL": "https://portal.kentik.com/v4/core/insights/a197790253"
  },
  "Metrics": {
    "bits": 58555.9140625,
    "packets": 1333270.75,
    "unique_src_ip": 1
  },
  "PreviousState": "new",
  "StartTime": "2021-11-19T13:39:33Z",
  "Type": "alarm"
}
```

## Tips and tricks with go-templating syntax

### Parenthesis

The order of the go-templating operator is not obvious and parenthesis may be necessary to guide the parser. For instance, in order to pick event details with a specific tag and convert it to a map, the snippet should be defined as:

```go
(.Details.WithTag "dimension").ToMap
```

### String message within single JSON field

There are many webhooks that although use JSON format as the medium, they use a single field that can be HTML or Markdown to format nice message. This requires the output to be a single line string. The following hints might be helpful building it:

- The `{{-`, `-}}` brackets are useful for removal of whitespace characters from the template.
- Use `\n` to actually print the new line inside (when using it as the go function parameter, it has to be double-escaped: `"\\n"`).
- Use `{{- /**/ -}}` for nicely formatting the template file while getting rid of new lines.


See [Discord](../templates/discord.json.tmpl) (for Markdown) and [ServiceNow](../templates/servicenow_events.json.tmpl) (for plain text) templates for more inspiration on that.

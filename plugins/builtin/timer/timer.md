# Timer Plugin

A built-in plugin for fetching current time and date information.

## Features

- **Fetch Current Time**: Get current hour:minute:second
- **Fetch Current Date**: Get current month:day

## Usage

### Natural Language Triggers

```
what is the time now
what is the date today
```

### Node Format

```
[node:extractor] [mode:fetch] [type:time] [dist:$1]
[node:extractor] [mode:fetch] [type:date] [dist:$1]
```

### Parameters

| Parameter | Values         | Description                              |
|-----------|----------------|------------------------------------------|
| `mode`    | `fetch`        | Operation mode (fetch current time/date) |
| `type`    | `time`, `date` | Type of information to retrieve          |
| `dist`    | variable name  | Destination variable to store result     |

### Response Example

```
User: what is the time now
System: the time is 14:30:25 now

User: what is the date today
System: the date is 03:26 today
```

## Implementation Details

- **Extractor**: Returns time/date as a `variable.Item`
- **Checker**: Placeholder (returns false)
- **Executor**: Not implemented
- **Storage**: Uses `storage.Repository` for persistence

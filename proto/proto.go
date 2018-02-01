package proto // import "tapr.space/proto"

import (
	"tapr.space/log"
)

// To regenerate the protocol buffer output for this package, run
//	go generate

//go:generate protoc tapr.proto --go_out=.

// LogLevelProto converts a log.Level to a proto.LogLevel.
func LogLevelProto(lvl log.Level) LogLevel {
	switch lvl {
	case log.DebugLevel:
		return LogLevel_Debug
	case log.InfoLevel:
		return LogLevel_Info
	case log.WarningLevel:
		return LogLevel_Warning
	case log.ErrorLevel:
		return LogLevel_Error
	case log.DisabledLevel:
		return LogLevel_Disabled
	default:
		panic("unknown log level")
	}
}

// LogEventProto converts a log.Event to a proto.LogEvent.
func LogEventProto(e *log.Event) *LogEvent {
	return &LogEvent{
		Level:   LogLevelProto(e.Level),
		Message: e.Message,
	}
}

// TaprLogLevel converts a proto.Loglevel to a log.Level.
func TaprLogLevel(pb LogLevel) log.Level {
	switch pb {
	case LogLevel_Debug:
		return log.DebugLevel
	case LogLevel_Info:
		return log.InfoLevel
	case LogLevel_Warning:
		return log.WarningLevel
	case LogLevel_Error:
		return log.ErrorLevel
	case LogLevel_Disabled:
		return log.DisabledLevel
	default:
		panic("unknown log level")
	}
}

// TaprLogEvent converts a proto.LogEvent to a log.Event.
func TaprLogEvent(pb *LogEvent) *log.Event {
	return &log.Event{
		Level:   TaprLogLevel(pb.Level),
		Message: pb.Message,
	}
}

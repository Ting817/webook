package logger

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type LoggerV1 interface {
	Debug(msg string, args ...Field)
	Info(msg string, args ...Field)
	Warn(msg string, args ...Field)
	Error(msg string, args ...Field)
	// With args 会加入进去 LoggerV1 的任何打印出来的日志里面
	With(args ...Field) LoggerV1
}

type Field struct {
	Key   string
	Value any
}

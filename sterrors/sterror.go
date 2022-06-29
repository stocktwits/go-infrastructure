package sterrors

import "fmt"

type ErrorCode int

type ErrorData struct {
	Type      string
	Message   string
	Http_code int
}

type ErrorConfig map[ErrorCode]ErrorData

type Error struct {
	Err       error
	Code      ErrorCode
	Message   string
	Http_code int
}

type ErrorFactory struct {
	config          ErrorConfig
	defaultMessage  string
	defaultHttpCode int
}

func NewFactory(config ErrorConfig, defMsg string, defHttpCode int) *ErrorFactory {
	return &ErrorFactory{
		config:          config,
		defaultMessage:  defMsg,
		defaultHttpCode: defHttpCode,
	}
}

func (e *ErrorFactory) NewError(code ErrorCode, err error) error {
	return &Error{
		Err:       err,
		Code:      code,
		Message:   e.getMessage(code),
		Http_code: e.getHttpCode(code),
	}
}

func (s *Error) Error() string {
	if s.Err != nil {
		return fmt.Sprintf("http error: %d, with internal code: %d, message: %s, %s", s.Http_code, s.Code, s.Message, s.Err.Error())
	}

	return fmt.Sprintf("http error: %d, with internal code: %d, message: %s", s.Http_code, s.Code, s.Message)
}

func (e *ErrorFactory) getMessage(code ErrorCode) string {
	if data, ok := e.config[code]; ok {
		return data.Message
	}

	return e.defaultMessage
}

func (e *ErrorFactory) getHttpCode(code ErrorCode) int {
	if data, ok := e.config[code]; ok {
		return data.Http_code
	}

	return e.defaultHttpCode
}

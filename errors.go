
import (
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"net/http"
)

// ErrorType is the type of an error
type ErrorType uint

const (
	// NoType error
	ErrorTypeNoType ErrorType = iota
	ErrorTypeBadRequest
	ErrorTypeNotFound
	ErrorTypeInternal
	ErrorTypeNotLogged
)

type myError struct {
	errorType     ErrorType
	originalError error
	customError   string
}

func NewError(errorType ErrorType, msg string) error {
	return myError{errorType: errorType, originalError: errors.New(msg)}
}

func ErrorBadRequest(msg string) error {
	return myError{errorType: ErrorTypeBadRequest, originalError: errors.New(msg)}
}

func ErrorNotLogged(msg string) error {
	return myError{errorType: ErrorTypeNotLogged, originalError: errors.New(msg)}
}

func ErrorVeritabani(err error, msg string, args ...interface{}) error {
	wrappedError := errors.Wrapf(err, msg, args...)
	return myError{errorType: ErrorTypeInternal, originalError: wrappedError, customError: "veritabanı hatası"}
}

func ErrorInternalError(err error, msg string, args ...interface{}) error {
	wrappedError := errors.Wrapf(err, msg, args...)
	return myError{errorType: ErrorTypeInternal, originalError: wrappedError, customError: "sunucu içi hata"}
}

// Error returns the mssage of a myError
func (error myError) Error() string {
	if error.originalError == nil {
		return error.customError
	}
	return error.originalError.Error()
}

// New creates a no type error
func New(msg string) error {
	return myError{errorType: ErrorTypeNoType, originalError: errors.New(msg)}
}

// GetType returns the error type
func GetType(err error) ErrorType {
	if customErr, ok := err.(myError); ok {
		return customErr.errorType
	}

	return ErrorTypeNoType
}

// ErrorHandler fiber için
func ErrorHandler(c *fiber.Ctx, err error) error {
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	respModel := &context.ResponseModel{
		HataVarMi: true,
		Message:   "internal server error",
	}

	ctxRqId := c.Get("requestid", "")
	logger := config.Logger(ctxRqId)

	if GetType(err) == ErrorTypeNoType {
		if err != nil {
			logger.Error(err.Error())
		}
		return c.Status(http.StatusInternalServerError).JSON(respModel)
	}
	myerr := err.(myError)
	switch myerr.errorType {
	case ErrorTypeNotFound:
		respModel.Message = "bulunamadı"
		return c.Status(http.StatusNotFound).JSON(respModel)
	case ErrorTypeBadRequest:
		respModel.Message = myerr.Error()
		return c.Status(http.StatusBadRequest).JSON(respModel)
	case ErrorTypeInternal:
		respModel.Message = myerr.customError
		logger.Error(myerr.Error())
		return c.Status(http.StatusInternalServerError).JSON(respModel)
	case ErrorTypeNotLogged:
		respModel.Message = myerr.Error()
		return c.Status(http.StatusUnauthorized).JSON(respModel)
	}

	logger.Error(myerr.Error())
	return c.Status(http.StatusInternalServerError).JSON(respModel)
}

package interfaces

import (
	"errors"
	"fmt"
	"time"
)

// Базовые ошибки
var (
	ErrNotFound       = errors.New("not_found")
	ErrConflict       = errors.New("conflict")
	ErrQuotaExceeded  = errors.New("quota_exceeded")
	ErrInvalidInput   = errors.New("invalid_input")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrTimeout        = errors.New("timeout")
	ErrUnavailable    = errors.New("unavailable")
	ErrInternal       = errors.New("internal_error")
	ErrNotImplemented = errors.New("not_implemented")
)

// TemporaryError представляет временную ошибку
type TemporaryError struct {
	Err error
}

func (e TemporaryError) Error() string {
	return e.Err.Error()
}

func (e TemporaryError) Temporary() bool {
	return true
}

func (e TemporaryError) Unwrap() error {
	return e.Err
}

// RetryableError представляет ошибку, которую можно повторить
type RetryableError struct {
	Err        error
	RetryAfter time.Duration
}

func (e RetryableError) Error() string {
	return e.Err.Error()
}

func (e RetryableError) Temporary() bool {
	return true
}

func (e RetryableError) Unwrap() error {
	return e.Err
}

// ValidationError представляет ошибку валидации
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field %s: %s", e.Field, e.Message)
}

// ConflictError представляет ошибку конфликта
type ConflictError struct {
	Resource string
	ID       string
	Message  string
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("conflict for %s %s: %s", e.Resource, e.ID, e.Message)
}

// QuotaExceededError представляет ошибку превышения квоты
type QuotaExceededError struct {
	Resource string
	Limit    int64
	Current  int64
}

func (e QuotaExceededError) Error() string {
	return fmt.Sprintf("quota exceeded for %s: current=%d, limit=%d", e.Resource, e.Current, e.Limit)
}

// TimeoutError представляет ошибку таймаута
type TimeoutError struct {
	Operation string
	Timeout   time.Duration
}

func (e TimeoutError) Error() string {
	return fmt.Sprintf("timeout for operation %s after %v", e.Operation, e.Timeout)
}

func (e TimeoutError) Temporary() bool {
	return true
}

// NotFoundError представляет ошибку "не найдено"
type NotFoundError struct {
	Resource string
	ID       string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s with id %s not found", e.Resource, e.ID)
}

// UnauthorizedError представляет ошибку авторизации
type UnauthorizedError struct {
	Resource string
	Action   string
}

func (e UnauthorizedError) Error() string {
	return fmt.Sprintf("unauthorized to perform %s on %s", e.Action, e.Resource)
}

// ForbiddenError представляет ошибку запрета доступа
type ForbiddenError struct {
	Resource string
	Action   string
	Reason   string
}

func (e ForbiddenError) Error() string {
	return fmt.Sprintf("forbidden to perform %s on %s: %s", e.Action, e.Resource, e.Reason)
}

// InvalidInputError представляет ошибку неверного ввода
type InvalidInputError struct {
	Field   string
	Value   interface{}
	Reason  string
}

func (e InvalidInputError) Error() string {
	return fmt.Sprintf("invalid input for field %s: %s", e.Field, e.Reason)
}

// InternalError представляет внутреннюю ошибку
type InternalError struct {
	Operation string
	Details   string
}

func (e InternalError) Error() string {
	return fmt.Sprintf("internal error in %s: %s", e.Operation, e.Details)
}

// NotImplementedError представляет ошибку "не реализовано"
type NotImplementedError struct {
	Feature string
}

func (e NotImplementedError) Error() string {
	return fmt.Sprintf("feature %s not implemented", e.Feature)
}

// UnavailableError представляет ошибку недоступности
type UnavailableError struct {
	Service string
	Reason  string
}

func (e UnavailableError) Error() string {
	return fmt.Sprintf("service %s unavailable: %s", e.Service, e.Reason)
}

func (e UnavailableError) Temporary() bool {
	return true
}

// Создание типизированных ошибок

// NewTemporaryError создает временную ошибку
func NewTemporaryError(err error) error {
	return TemporaryError{Err: err}
}

// NewRetryableError создает ошибку с возможностью повтора
func NewRetryableError(err error, retryAfter time.Duration) error {
	return RetryableError{Err: err, RetryAfter: retryAfter}
}

// NewValidationError создает ошибку валидации
func NewValidationError(field string, value interface{}, message string) error {
	return ValidationError{Field: field, Value: value, Message: message}
}

// NewConflictError создает ошибку конфликта
func NewConflictError(resource, id, message string) error {
	return ConflictError{Resource: resource, ID: id, Message: message}
}

// NewQuotaExceededError создает ошибку превышения квоты
func NewQuotaExceededError(resource string, limit, current int64) error {
	return QuotaExceededError{Resource: resource, Limit: limit, Current: current}
}

// NewTimeoutError создает ошибку таймаута
func NewTimeoutError(operation string, timeout time.Duration) error {
	return TimeoutError{Operation: operation, Timeout: timeout}
}

// NewNotFoundError создает ошибку "не найдено"
func NewNotFoundError(resource, id string) error {
	return NotFoundError{Resource: resource, ID: id}
}

// NewUnauthorizedError создает ошибку авторизации
func NewUnauthorizedError(resource, action string) error {
	return UnauthorizedError{Resource: resource, Action: action}
}

// NewForbiddenError создает ошибку запрета доступа
func NewForbiddenError(resource, action, reason string) error {
	return ForbiddenError{Resource: resource, Action: action, Reason: reason}
}

// NewInvalidInputError создает ошибку неверного ввода
func NewInvalidInputError(field string, value interface{}, reason string) error {
	return InvalidInputError{Field: field, Value: value, Reason: reason}
}

// NewInternalError создает внутреннюю ошибку
func NewInternalError(operation, details string) error {
	return InternalError{Operation: operation, Details: details}
}

// NewNotImplementedError создает ошибку "не реализовано"
func NewNotImplementedError(feature string) error {
	return NotImplementedError{Feature: feature}
}

// NewUnavailableError создает ошибку недоступности
func NewUnavailableError(service, reason string) error {
	return UnavailableError{Service: service, Reason: reason}
}

// Проверка типов ошибок

// IsTemporary проверяет, является ли ошибка временной
func IsTemporary(err error) bool {
	var temp interface{ Temporary() bool }
	return errors.As(err, &temp) && temp.Temporary()
}

// IsRetryable проверяет, можно ли повторить операцию
func IsRetryable(err error) bool {
	return IsTemporary(err) || errors.Is(err, ErrTimeout) || errors.Is(err, ErrUnavailable)
}

// IsNotFound проверяет, является ли ошибка "не найдено"
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) || errors.As(err, &NotFoundError{})
}

// IsConflict проверяет, является ли ошибка конфликтом
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict) || errors.As(err, &ConflictError{})
}

// IsQuotaExceeded проверяет, является ли ошибка превышением квоты
func IsQuotaExceeded(err error) bool {
	return errors.Is(err, ErrQuotaExceeded) || errors.As(err, &QuotaExceededError{})
}

// IsValidationError проверяет, является ли ошибка ошибкой валидации
func IsValidationError(err error) bool {
	return errors.As(err, &ValidationError{})
}

// IsTimeoutError проверяет, является ли ошибка ошибкой таймаута
func IsTimeoutError(err error) bool {
	return errors.Is(err, ErrTimeout) || errors.As(err, &TimeoutError{})
}

// IsUnauthorized проверяет, является ли ошибка ошибкой авторизации
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized) || errors.As(err, &UnauthorizedError{})
}

// IsForbidden проверяет, является ли ошибка ошибкой запрета доступа
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden) || errors.As(err, &ForbiddenError{})
}

// IsInvalidInput проверяет, является ли ошибка ошибкой неверного ввода
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput) || errors.As(err, &InvalidInputError{})
}

// IsInternalError проверяет, является ли ошибка внутренней ошибкой
func IsInternalError(err error) bool {
	return errors.Is(err, ErrInternal) || errors.As(err, &InternalError{})
}

// IsNotImplemented проверяет, является ли ошибка ошибкой "не реализовано"
func IsNotImplemented(err error) bool {
	return errors.Is(err, ErrNotImplemented) || errors.As(err, &NotImplementedError{})
}

// IsUnavailable проверяет, является ли ошибка ошибкой недоступности
func IsUnavailable(err error) bool {
	return errors.Is(err, ErrUnavailable) || errors.As(err, &UnavailableError{})
}

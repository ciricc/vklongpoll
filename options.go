package vklongpoll

import (
	"time"

	"github.com/ciricc/vkapiexecutor/request"
)

// Режим работы
type Mode int

// Получать вложения
const Attachments Mode = 2

// Возвращать расширенный набор событий
const Extended Mode = 8

// Возвращать поле pts для дальнейшей работы
const ReturnPts Mode = 32

// Возвращать поля $extra
const ExtraFields Mode = 64

// Возвращать поле random_id
const ReturnRandomId Mode = 128

// Длительность запроса по умолчанию
var DefaultWait time.Duration = 90 * time.Second

// Режим работы по умолчанию
var DefaultMode Mode = 0

// Версия Long Poll по умолчанию
var DefaultVersion = 3

type VkLongPollOptions struct {
	Wait             time.Duration
	GetServerRequest *request.Request
	Mode             Mode
	Version          int
}

type VkLongPollOption func(v *VkLongPollOptions)

// Создает опции по умолчанию
func NewOptions() *VkLongPollOptions {
	return &VkLongPollOptions{
		Wait:             DefaultWait,
		GetServerRequest: nil,
		Mode:             DefaultMode,
		Version:          DefaultVersion,
	}
}

// Устанавливает запрос VK API, который будет отправляться для получения URL сервера (messages.getLongPollServer, например)
// Используйте заголовки и кастомные параметры запроса, если есть такая необходимость
func WithGetServerRequest(req *request.Request) VkLongPollOption {
	return func(v *VkLongPollOptions) {
		v.GetServerRequest = req
	}
}

// Устанавливает длительность одного запроса
func WithWait(wait time.Duration) VkLongPollOption {
	return func(v *VkLongPollOptions) {
		v.Wait = wait
	}
}

// Находит сумму всех указанных режимов и устанавливает в настройки подключения
func WithModeSum(modes ...Mode) VkLongPollOption {
	return func(v *VkLongPollOptions) {
		var s Mode = 0
		for _, mode := range modes {
			s = s + mode
		}
		WithMode(s)(v)
	}
}

// Устанавливает режим работы в натсройки подключения
func WithMode(mode Mode) VkLongPollOption {
	return func(v *VkLongPollOptions) {
		v.Mode = mode
	}
}

// Устанавливает версию Long Poll (2, 3, 4, 5)
func WithVersion(version int) VkLongPollOption {
	return func(v *VkLongPollOptions) {
		v.Version = version
	}
}

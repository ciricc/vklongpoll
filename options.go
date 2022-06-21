package vklongpoll

import (
	"net/url"
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
var DefaultWait = 90 * time.Second

// Режим работы по умолчанию
var DefaultMode Mode = 0

// Версия Long Poll по умолчанию
var DefaultVersion = 3

type VkLongPollOptions struct {
	Wait             time.Duration
	GetServerRequest *request.Request
	Mode             Mode
	Version          int
	Params           url.Values
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

// Суммирует все функциональные опции и возвращает структуру опций
func BuildOptions(opts ...VkLongPollOption) *VkLongPollOptions {
	opt := NewOptions()
	for _, option := range opts {
		option(opt)
	}

	return opt
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

// Суммирует режимы работы
func SumModes(modes ...Mode) Mode {
	var s Mode = 0
	for _, mode := range modes {
		s = s + mode
	}
	return s
}

// Находит сумму всех указанных режимов и устанавливает в настройки подключения
func WithModeSum(modes ...Mode) VkLongPollOption {
	return func(v *VkLongPollOptions) {
		WithMode(SumModes(modes...))(v)
	}
}

// Устанавливает кастомные параметры в запрос к Long Poll серверу
// (перезаписывает уже установленные, вроде ts, mode, wait и т.д)
func WithParams(params url.Values) VkLongPollOption {
	return func(v *VkLongPollOptions) {
		v.Params = params
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

package vklongpoll

import (
	"context"
	"net/url"
	"time"
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

// Путь параметра, где в ответе сервера хранятся обновления
var DefaultUpdatesJsonPath = []string{"updates"}

type ServerUpdater func(ctx context.Context) (*ServerCredentials, error)

type ParamsMerger func(u url.Values)
type VkLongPollOptions struct {
	Wait            time.Duration
	ServerUpdater   ServerUpdater
	Mode            Mode
	Version         int
	ParamsMerger    ParamsMerger
	UpdatesJsonPath []string
}

type ServerCredentials struct {
	Ts        int64    // Новое значение TS
	ServerURL *url.URL // Новый URL сервера
	Key       string   // Новый ключ
}

type VkLongPollOption func(v *VkLongPollOptions)

// Создает опции по умолчанию
func NewOptions() *VkLongPollOptions {
	return &VkLongPollOptions{
		Wait:            DefaultWait,
		ServerUpdater:   nil,
		Mode:            DefaultMode,
		Version:         DefaultVersion,
		UpdatesJsonPath: DefaultUpdatesJsonPath,
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

// Устанавливает обработчик для обновления информации о Long Poll соединении
// Здесь, например, вы можете делать запрос на messages.getLongPollServer и получить значение ts, key и т.д
// Если у вас нет цели писать собственный обработчик, то используйте UniversalServerUpdater
func WithServerUpdater(updater ServerUpdater) VkLongPollOption {
	return func(v *VkLongPollOptions) {
		v.ServerUpdater = updater
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
func WithParamsMerger(merger ParamsMerger) VkLongPollOption {
	return func(v *VkLongPollOptions) {
		v.ParamsMerger = merger
	}
}

// Устанавливает режим работы в натсройки подключения
func WithMode(mode Mode) VkLongPollOption {
	return func(v *VkLongPollOptions) {
		v.Mode = mode
	}
}

// Устанавливает путь параметра updates в JSON схеме
func WithUpdatesJsonPath(path ...string) VkLongPollOption {
	return func(v *VkLongPollOptions) {
		v.UpdatesJsonPath = path
	}
}

// Устанавливает версию Long Poll (2, 3, 4, 5)
func WithVersion(version int) VkLongPollOption {
	return func(v *VkLongPollOptions) {
		v.Version = version
	}
}

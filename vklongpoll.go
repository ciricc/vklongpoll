package vklongpoll

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/buger/jsonparser"
	"github.com/ciricc/vkapiexecutor/executor"
)

var DefaultExecutor = executor.New()

type VkLongPoll struct {
	VkApiExecutor *executor.Executor
	HttpClient    *http.Client
	key           string
	serverUrl     *url.URL
	Ts            int64
	pts           *Pts
	mx            *sync.Mutex
}

type Pts int64
type Update []byte

// Создает инстанс лонгполла
// Принимает executor - структура для отправки запросов к VK API
func New(executor *executor.Executor) *VkLongPoll {
	lp := VkLongPoll{
		VkApiExecutor: executor,
		HttpClient:    http.DefaultClient,
		mx:            &sync.Mutex{},
	}

	if executor == nil {
		lp.VkApiExecutor = DefaultExecutor
	}

	return &lp
}

// Возвращает последнее полученное значение pts
func (v *VkLongPoll) Pts() *Pts {
	return v.pts
}

// Делает запрос на получение списка новых событий
// Возвращает список событий в [][]byte, а также значение поля pts, если оно есть в ответе
// Автоматически переподключается в случае ошибки
//
// Чтение событий только синхронное, чтобы иметь всегда определенный ts (последний)
// Если есть задача читать события одновременно на один и тот же источник -
// лучше всего создать новое поделючение Long Poll
func (v *VkLongPoll) Recv(ctx context.Context, opts ...VkLongPollOption) ([]Update, error) {
	return v.RecvOpt(ctx, BuildOptions(opts...))
}

// То же самое, что Recv, но опции - ссылка на структуру
func (v *VkLongPoll) RecvOpt(ctx context.Context, opt *VkLongPollOptions) ([]Update, error) {
	v.mx.Lock()
	defer v.mx.Unlock()

	if v.serverUrl == nil {
		err := v.updateServer(ctx, opt)
		if err != nil {
			return nil, err
		}
	}

	requestUrl := v.serverUrl
	requestUrlQuery := requestUrl.Query()

	requestUrlQuery.Set("key", v.key)
	requestUrlQuery.Set("ts", strconv.FormatInt(v.Ts, 10))
	requestUrlQuery.Set("act", "a_check")
	requestUrlQuery.Set("wait", strconv.Itoa(int(opt.Wait.Seconds())))
	requestUrlQuery.Set("version", strconv.Itoa(opt.Version))

	if opt.Mode != 0 {
		requestUrlQuery.Set("mode", strconv.Itoa(int(opt.Mode)))
	}

	requestUrl.RawQuery = requestUrlQuery.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", requestUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := v.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("poll request error: %s; requestUrl=%s", err.Error(), requestUrl)
	}

	defer res.Body.Close()

	resBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response error: %s", err.Error())
	}

	failed, _ := jsonparser.GetInt(resBytes, "failed")

	ptsInt, err := jsonparser.GetInt(resBytes, "pts")
	if err == nil {
		v.pts = (*Pts)(&ptsInt)
	} else {
		ptsStr, _ := jsonparser.GetString(resBytes, "pts")
		ptsInt, err := strconv.ParseInt(ptsStr, 10, 64)
		if err == nil {
			v.pts = (*Pts)(&ptsInt)
		} else {
			v.pts = nil
		}
	}

	if failed != 0 {
		switch failed {
		case 2, 3:
			err = v.updateServer(ctx, opt)
			if err != nil {
				return nil, err
			}
		case 4:
			return nil, fmt.Errorf("invalid version")
		}
	}

	v.Ts, err = getTs(resBytes)

	if err != nil {
		return nil, err
	}

	updatesBytes, _, _, err := jsonparser.Get(resBytes, "updates")
	if err != nil {
		return nil, err
	}

	updates := []Update{}

	jsonparser.ArrayEach(updatesBytes, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if dataType == jsonparser.String { // ну вдруг ))
			updates = append(updates, []byte(`"`+string(value)+`"`))
		} else {
			updates = append(updates, value)
		}
	})

	return updates, nil
}

// Возвращает поле ts (если число - преобразует в строку) из json
func getTs(b []byte) (int64, error) {
	tsInt, err := jsonparser.GetInt(b, "ts")
	if err != nil {
		ts, _ := jsonparser.GetString(b, "ts")
		return strconv.ParseInt(ts, 10, 64)
	}
	return tsInt, nil
}

// Обновляет настройки Long Poll соединения
func (v *VkLongPoll) updateServer(ctx context.Context, opt *VkLongPollOptions) error {

	if opt.GetServerRequest == nil {
		return errors.New("get server request is undefined in options")
	}

	server, err := v.VkApiExecutor.DoRequestCtx(ctx, opt.GetServerRequest)
	if err != nil {
		return err
	}

	res, _, _, err := jsonparser.Get(server.Body(), "response")
	if err != nil {
		return err
	}

	v.key, err = jsonparser.GetString(res, "key")
	if err != nil {
		return err
	}

	serverUrl, err := jsonparser.GetString(res, "server")
	if err != nil {
		return err
	}

	if !strings.HasPrefix(serverUrl, "https://") && !strings.HasPrefix(serverUrl, "http://") {
		serverUrl = "https://" + serverUrl
	}

	v.serverUrl, err = url.Parse(serverUrl)
	if err != nil {
		return errors.New("parse server url error: " + err.Error() + "; serverUrl=" + serverUrl)
	}

	v.Ts, err = getTs(res)
	if err != nil {
		return err
	}

	return nil
}

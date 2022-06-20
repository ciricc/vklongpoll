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

	"github.com/buger/jsonparser"
	"github.com/ciricc/vkapiexecutor/executor"
)

var DefaultExecutor = executor.New()

type VkLongPoll struct {
	VkApiExecutor *executor.Executor
	HttpClient    *http.Client
	key           string
	serverUrl     *url.URL
	ts            string
}

type Pts int64
type Update []byte

// Создает инстанс лонгполла
// Принимает executor - структура для отправки запросов к VK API
func New(executor *executor.Executor) *VkLongPoll {
	lp := VkLongPoll{
		VkApiExecutor: executor,
		HttpClient:    http.DefaultClient,
	}

	if executor == nil {
		lp.VkApiExecutor = DefaultExecutor
	}

	return &lp
}

// Делает запрос на получение списка новых событий
// Возвращает список событий в [][]byte, а также значение поля pts, если оно есть в ответе
// Автоматически переподключается в случае ошибки
func (v *VkLongPoll) Recv(ctx context.Context, opts ...VkLongPollOption) ([]Update, *Pts, error) {
	opt := NewOptions()
	for _, option := range opts {
		option(opt)
	}

	return v.RecvOpt(ctx, opt)
}

// То же самое, что Recv, но опции - ссылка на структуру
func (v *VkLongPoll) RecvOpt(ctx context.Context, opt *VkLongPollOptions) ([]Update, *Pts, error) {

	if v.serverUrl == nil {
		err := v.updateServer(ctx, opt)
		if err != nil {
			return nil, nil, err
		}
	}

	requestUrl := v.serverUrl
	requestUrlQuery := requestUrl.Query()

	requestUrlQuery.Set("key", v.key)
	requestUrlQuery.Set("ts", v.ts)
	requestUrlQuery.Set("act", "a_check")
	requestUrlQuery.Set("wait", strconv.Itoa(int(opt.Wait.Seconds())))
	requestUrlQuery.Set("version", strconv.Itoa(opt.Version))

	if opt.Mode != 0 {
		requestUrlQuery.Set("mode", strconv.Itoa(int(opt.Mode)))
	}

	requestUrl.RawQuery = requestUrlQuery.Encode()

	res, err := v.HttpClient.Get(requestUrl.String())
	if err != nil {
		return nil, nil, fmt.Errorf("poll request error: %s; requestUrl=%s", err.Error(), requestUrl)
	}

	defer res.Body.Close()

	resBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response error: %s", err.Error())
	}

	failed, _ := jsonparser.GetInt(resBytes, "failed")

	var pts Pts
	ptsInt, err := jsonparser.GetInt(resBytes, "pts")
	if err != nil {
		pts = Pts(ptsInt)
	}

	if failed != 0 {
		switch failed {
		case 2, 3:
			err = v.updateServer(ctx, opt)
			if err != nil {
				return nil, &pts, err
			}
		case 4:
			return nil, &pts, fmt.Errorf("invalid version")
		}
	}

	v.ts, err = getTs(resBytes)
	if err != nil {
		return nil, &pts, err
	}

	updatesBytes, _, _, err := jsonparser.Get(resBytes, "updates")
	if err != nil {
		return nil, &pts, err
	}

	updates := []Update{}
	jsonparser.ArrayEach(updatesBytes, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		updates = append(updates, value)
	})

	return updates, &pts, nil
}

// Возвращает поле ts (если число - преобразует в строку) из json
func getTs(b []byte) (string, error) {
	tsInt, err := jsonparser.GetInt(b, "ts")
	if err != nil {
		ts, err := jsonparser.GetString(b, "ts")
		if err != nil {
			return "", err
		}
		return ts, nil
	}
	return strconv.FormatInt(tsInt, 10), nil
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

	v.ts, err = getTs(res)
	if err != nil {
		return err
	}
	return nil
}

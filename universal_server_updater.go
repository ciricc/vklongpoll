package vklongpoll

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/ciricc/vkapiexecutor/executor"
	"github.com/ciricc/vkapiexecutor/request"
)

// Универсальный модуль для обновления сервера
// Подходит для Long Poll сообществ, а также для LongPoll пользователя
// Обработчик использует executor, в нем он выполняет указанный запрос req
// Урпавление токенами, сессиями и параметарми запроса происходит уже на уровне создания запроса вами
func UniversalServerUpdater(req *request.Request, exec *executor.Executor) ServerUpdater {
	return func(ctx context.Context) (*ServerCredentials, error) {
		creds := ServerCredentials{}
		server, err := exec.DoRequestCtx(ctx, req)
		if err != nil {
			return nil, err
		}

		res, _, _, err := jsonparser.Get(server.Body(), "response")
		if err != nil {
			return nil, err
		}

		creds.Key, err = jsonparser.GetString(res, "key")
		if err != nil {
			return nil, err
		}

		serverUrl, err := jsonparser.GetString(res, "server")
		if err != nil {
			return nil, err
		}

		if !strings.HasPrefix(serverUrl, "https://") && !strings.HasPrefix(serverUrl, "http://") {
			serverUrl = "https://" + serverUrl
		}

		creds.ServerURL, err = url.Parse(serverUrl)
		if err != nil {
			return nil, errors.New("parse server url error: " + err.Error() + "; serverUrl=" + serverUrl)
		}

		creds.Ts, err = getTs(res)
		if err != nil {
			return nil, err
		}

		return &creds, nil
	}
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

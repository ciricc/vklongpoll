# VK LongPoll

Данный модуль предназначен для настройки соединения Long Poll с VK.
Внутри происходит только базовая поддержка соединения и парсинг только общих полей - `ts`, `key`, `failed`, `pts`

Модуль возвращает события списком из байтов `[]byte` для дальнейшего самостоятельного парсинга структур. Внутри используется модуль [vkapiexecutor](https://github.com/ciricc/vkapiexecutor) для отправки запроса к VK API на получение URL сервера.

- Модуль создает универсальное подключение Long Poll, есть поддержка как для сообществ, так и для пользовательсокго Long Poll.
- Есть возможность изменять `http.Client` для настройки своих собственных заголовков, а также для настройки прокси или кастомного сжатия.
- Есть возможность получить поле `pts` для обработки событий вручную.

## Установка

```shell
go get github.com/ciricc/vklongpoll
```

## Пример

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/ciricc/vkapiexecutor/executor"
	"github.com/ciricc/vkapiexecutor/request"
	"github.com/ciricc/vklongpoll"
)

func main() {
	exec := executor.New()

	getServerRequest := request.New()

	getServerRequest.GetParams().AccessToken("BOT_TOKEN")
	getServerRequest.GetParams().Set("group_id", "group")

	getServerRequest.Method("groups.getLongPollServer")

	lp := vklongpoll.New(exec)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			log.Println("Longpoll timed out!")
			return
		default:
			updates, err := lp.Recv(ctx, vklongpoll.WithGetServerRequest(getServerRequest))
			if err != nil {
				log.Println("get updates error", err)
			} else {
				log.Println("updates", updates)
				log.Println("ts", lp.Ts)
			}
		}
	}
}
```
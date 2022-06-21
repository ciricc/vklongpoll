package vklongpoll_test

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/ciricc/vkapiexecutor/executor"
	"github.com/ciricc/vkapiexecutor/request"
	"github.com/ciricc/vklongpoll"
)

type ServerCredentials struct {
	Server string `json:"server"`
	Key    string `json:"key"`
	Ts     string `json:"ts"`
}

type LongPollServerResponse struct {
	Ts      string        `json:"ts"`
	Updates []interface{} `json:"updates"`
	Pts     *string       `json:"pts"`
}

type GetServerResponse struct {
	Response ServerCredentials `json:"response"`
}

func getServerResponse(lpServerUrl string) *GetServerResponse {
	return &GetServerResponse{
		Response: ServerCredentials{
			Server: lpServerUrl,
			Key:    "longpoll_server_key",
			Ts:     "1",
		},
	}
}

func TestVkLongPoll(t *testing.T) {

	lpServerResponse := &LongPollServerResponse{
		Ts:      "2",
		Updates: make([]interface{}, 0),
		Pts:     nil,
	}

	longPollServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, err := json.Marshal(lpServerResponse)
		if err != nil {
			t.Error(err)
		}

		log.Println("lp server requested")
		w.Write(res)
	}))

	defer longPollServer.Close()

	expectedGetServerResponse := getServerResponse(longPollServer.URL)
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, err := json.Marshal(expectedGetServerResponse)
		if err != nil {
			t.Error(err)
		}
		log.Println("api server requested")
		w.Write(res)
	}))

	defer apiServer.Close()

	request.DefaultBaseRequestUrl = apiServer.URL
	exec := executor.New()

	lp := vklongpoll.New(exec)

	getServerRequest := request.New()
	getServerRequest.Method("get_server_url")

	t.Run("returns error if no get server request specified", func(t *testing.T) {
		_, err := lp.Recv(context.Background())
		if err == nil {
			t.Error("expected error but got nil")
		}
	})

	// Sync and save server
	_, err := lp.Recv(context.Background(), vklongpoll.WithGetServerRequest(getServerRequest))
	if err != nil {
		t.Error(err)
	}

	t.Run("no need request get server after server got", func(t *testing.T) {
		_, err := lp.Recv(context.Background())
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("lp returns zero updates", func(t *testing.T) {
		updates, err := lp.Recv(context.Background())
		if err != nil {
			t.Error(err)
		}
		if len(updates) != 0 {
			t.Errorf("expected zero updates gut got: len=%d", len(updates))
		}
	})

	t.Run("lp updated ts", func(t *testing.T) {
		_, err := lp.Recv(context.Background(), vklongpoll.WithGetServerRequest(getServerRequest))
		if err != nil {
			t.Error(err)
		}
		expectedTs, _ := strconv.ParseInt(lpServerResponse.Ts, 10, 64)
		if lp.Ts != expectedTs {
			t.Errorf("expected new ts: %d but got %d\n", expectedTs, lp.Ts)
		}
	})

	t.Run("lp updates reading correctly", func(t *testing.T) {
		expectedUpdates := []interface{}{1, LongPollServerResponse{
			Ts: "1",
		}, true, "Hi"}
		expectedUpdatesRaws := make([]string, len(expectedUpdates))

		for i, update := range expectedUpdates {
			b, err := json.Marshal(update)
			if err != nil {
				t.Error(err)
			}
			expectedUpdatesRaws[i] = string(b)
		}

		lpServerResponse = &LongPollServerResponse{
			Ts:      "3",
			Updates: expectedUpdates,
		}

		updates, err := lp.Recv(context.Background(), vklongpoll.WithGetServerRequest(getServerRequest))
		if err != nil {
			t.Error(err)
		}

		if len(updates) != len(expectedUpdates) {
			t.Errorf("expected updates length is %d but got %d", len(expectedUpdates), len(updates))
		}

		updatesRaws := make([]string, len(updates))
		for i, update := range updates {
			updatesRaws[i] = string(update)
		}

		if !reflect.DeepEqual(updatesRaws, expectedUpdatesRaws) {
			t.Errorf("expected updates %v but got %v", expectedUpdatesRaws, updatesRaws)
		}
	})

	t.Run("lp updates pts correctly", func(t *testing.T) {
		lpServerResponse = &LongPollServerResponse{
			Ts:      "3",
			Updates: make([]interface{}, 0),
		}
		_, err := lp.Recv(context.Background())
		if err != nil {
			t.Error(err)
		}
		if lp.Pts() != nil {
			t.Errorf("expected pts nil but got %d", lp.Pts())
		}
	})

	t.Run("lp updates pts correctly in int", func(t *testing.T) {
		ptsVal := "1"
		lpServerResponse = &LongPollServerResponse{
			Ts:      "3",
			Updates: make([]interface{}, 0),
			Pts:     &ptsVal,
		}
		_, err := lp.Recv(context.Background())
		if err != nil {
			t.Error(err)
		}

		var ptsValInt int64 = 1
		if lp.Pts() == nil {
			t.Errorf("pts is nil")
		} else if *lp.Pts() != vklongpoll.Pts(ptsValInt) {
			t.Errorf("expected pts %d but got %d", ptsValInt, *lp.Pts())
		}
	})

	t.Run("lp updates pts correctly to nil", func(t *testing.T) {
		lpServerResponse = &LongPollServerResponse{
			Ts:      "3",
			Updates: make([]interface{}, 0),
		}
		_, err := lp.Recv(context.Background())
		if err != nil {
			t.Error(err)
		}
		if lp.Pts() != nil {
			t.Errorf("expected pts nil but got %d", lp.Pts())
		}
	})
}

func TestLongpollContext(t *testing.T) {
	lpServerResponse := &LongPollServerResponse{
		Ts:      "2",
		Updates: make([]interface{}, 0),
		Pts:     nil,
	}

	longPollServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, err := json.Marshal(lpServerResponse)
		if err != nil {
			t.Error(err)
		}

		log.Println("lp server requested")
		time.Sleep(2 * time.Second)
		w.Write(res)
	}))

	defer longPollServer.Close()

	expectedGetServerResponse := getServerResponse(longPollServer.URL)
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		res, err := json.Marshal(expectedGetServerResponse)
		if err != nil {
			t.Error(err)
		}

		log.Println("api server requested")
		w.Write(res)
	}))

	defer apiServer.Close()

	request.DefaultBaseRequestUrl = apiServer.URL
	exec := executor.New()

	getServerRequest := request.New()
	lp := vklongpoll.New(exec)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err := lp.Recv(ctx, vklongpoll.WithGetServerRequest(getServerRequest))
	if err == nil {
		t.Errorf("expected error but got nil")
	}

}

func TestLongPollParams(t *testing.T) {
	lpServerResponse := &LongPollServerResponse{
		Ts:      "2",
		Updates: make([]interface{}, 0),
		Pts:     nil,
	}

	longPollServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, err := json.Marshal(lpServerResponse)
		if err != nil {
			t.Error(err)
		}

		expectedMethod := "GET"
		if r.Method != expectedMethod {
			t.Errorf("expected long poll request method %q but got %q", expectedMethod, r.Method)
		}

		expectedQuery := url.Values{
			"act":     {"a_check"},
			"key":     {"longpoll_server_key"},
			"mode":    {"2"},
			"ts":      {"1"},
			"version": {"1"},
			"wait":    {"43"},
			"foo":     {"bar"},
		}.Encode()

		if r.URL.Query().Encode() != expectedQuery {
			t.Errorf("expected url query: %q but got %q", expectedQuery, r.URL.Query().Encode())
		}

		log.Println("lp server requested")
		w.Write(res)
	}))

	defer longPollServer.Close()

	expectedGetServerResponse := getServerResponse(longPollServer.URL)
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		res, err := json.Marshal(expectedGetServerResponse)
		if err != nil {
			t.Error(err)
		}

		expectedPath := "/get_long_poll_server"
		if r.URL.Path != expectedPath {
			t.Errorf("exepcted get server path %q but got %q", expectedPath, r.URL.Path)
		}

		log.Println("api server requested")
		w.Write(res)
	}))

	defer apiServer.Close()

	request.DefaultBaseRequestUrl = apiServer.URL
	exec := executor.New()

	getServerRequest := request.New()
	getServerRequest.Method("get_long_poll_server")
	lp := vklongpoll.New(exec)

	_, err := lp.Recv(context.Background(),
		vklongpoll.WithGetServerRequest(getServerRequest),
		vklongpoll.WithMode(vklongpoll.Attachments),
		vklongpoll.WithWait(43*time.Second),
		vklongpoll.WithVersion(1),
		vklongpoll.WithParams(url.Values{
			"foo": {"bar"},
		}),
	)

	if err != nil {
		t.Error(err)
	}
}

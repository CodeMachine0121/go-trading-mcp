package tradingservice_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
	"github.com/CodeMachine0121/go-trading-mcp/internal/infrastructure/tradingservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// arrival is what the stand-in trading service saw, so that a test can assert the ask
// was worded the way the ability declared it.
type arrival struct {
	method        string
	path          string
	rawQuery      string
	body          string
	authorization string
	accept        string
}

func standingInFor(t *testing.T, answer func(http.ResponseWriter)) (*tradingservice.TradingServiceProxy, *arrival) {
	t.Helper()

	seen := &arrival{}
	standIn := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		*seen = arrival{
			method:        request.Method,
			path:          request.URL.Path,
			rawQuery:      request.URL.RawQuery,
			body:          string(body),
			authorization: request.Header.Get("Authorization"),
			accept:        request.Header.Get("Accept"),
		}
		answer(writer)
	}))
	t.Cleanup(standIn.Close)

	return tradingservice.NewTradingServiceProxy(standIn.URL, 5*time.Second), seen
}

func TestEachVerbLeavesAsTheMethodThatMeansIt(t *testing.T) {
	testCases := []struct {
		verb           vo.RequestVerb
		expectedMethod string
	}{
		{vo.RequestVerbRead, http.MethodGet},
		{vo.RequestVerbSubmit, http.MethodPost},
		{vo.RequestVerbReplace, http.MethodPut},
		{vo.RequestVerbRemove, http.MethodDelete},
	}

	for _, testCase := range testCases {
		t.Run(string(testCase.verb), func(t *testing.T) {
			proxy, seen := standingInFor(t, func(writer http.ResponseWriter) {
				_, _ = writer.Write([]byte(`{}`))
			})

			_, sendError := proxy.Send(context.Background(),
				vo.TradingServiceRequestVo{Verb: testCase.verb, Path: "/k-candles"}, "")

			require.NoError(t, sendError)
			assert.Equal(t, testCase.expectedMethod, seen.method)
		})
	}
}

func TestAnAskArrivesWordedTheWayItWasBuilt(t *testing.T) {
	proxy, seen := standingInFor(t, func(writer http.ResponseWriter) {
		_, _ = writer.Write([]byte(`{"ok":true}`))
	})

	_, sendError := proxy.Send(context.Background(), vo.TradingServiceRequestVo{
		Verb:  vo.RequestVerbSubmit,
		Path:  "/k-candles/history",
		Query: map[string]string{"symbol": "BTCUSDT"},
		Body:  []byte(`{"lookbackDays":30}`),
	}, "james-token")

	require.NoError(t, sendError)
	assert.Equal(t, "/k-candles/history", seen.path)
	assert.Equal(t, "symbol=BTCUSDT", seen.rawQuery)
	assert.JSONEq(t, `{"lookbackDays":30}`, seen.body)
	assert.Equal(t, "Bearer james-token", seen.authorization)
}

func TestAnAskWithNoProofCarriesNoAuthorizationAtAll(t *testing.T) {
	proxy, seen := standingInFor(t, func(writer http.ResponseWriter) {
		_, _ = writer.Write([]byte(`{}`))
	})

	_, _ = proxy.Send(context.Background(),
		vo.TradingServiceRequestVo{Verb: vo.RequestVerbRead, Path: "/health"}, "")

	assert.Empty(t, seen.authorization)
}

func TestTheVerdictFollowsWhatTheTradingServiceAnswered(t *testing.T) {
	testCases := []struct {
		name            string
		statusCode      int
		expectedOutcome vo.TradingServiceOutcome
	}{
		{"做到了", http.StatusOK, vo.TradingServiceSucceeded},
		{"收下了，還沒做完", http.StatusAccepted, vo.TradingServiceSucceeded},
		{"規則不通過", http.StatusBadRequest, vo.TradingServiceRefused},
		{"不認得這份身分", http.StatusUnauthorized, vo.TradingServiceIdentityNotRecognized},
		{"不是你的", http.StatusForbidden, vo.TradingServiceRefused},
		{"找不到", http.StatusNotFound, vo.TradingServiceRefused},
		{"已經在跑了", http.StatusConflict, vo.TradingServiceRefused},
		{"算式跑不動", http.StatusUnprocessableEntity, vo.TradingServiceRefused},
		{"額度用完", http.StatusTooManyRequests, vo.TradingServiceRefused},
		{"資料庫讀寫失敗", http.StatusBadGateway, vo.TradingServiceRefused},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			proxy, _ := standingInFor(t, func(writer http.ResponseWriter) {
				writer.WriteHeader(testCase.statusCode)
				_, _ = writer.Write([]byte(`{"message":"交易服務說的話"}`))
			})

			response, sendError := proxy.Send(context.Background(),
				vo.TradingServiceRequestVo{Verb: vo.RequestVerbRead, Path: "/k-candles"}, "")

			require.NoError(t, sendError)
			assert.Equal(t, testCase.expectedOutcome, response.Outcome)
			assert.Equal(t, `{"message":"交易服務說的話"}`, response.Content)
		})
	}
}

func TestSomethingDoneWithNothingToShowStillSaysItWasDone(t *testing.T) {
	proxy, _ := standingInFor(t, func(writer http.ResponseWriter) {
		writer.WriteHeader(http.StatusNoContent)
	})

	response, sendError := proxy.Send(context.Background(),
		vo.TradingServiceRequestVo{Verb: vo.RequestVerbRemove, Path: "/watchlist/BTCUSDT"}, "")

	require.NoError(t, sendError)
	assert.Equal(t, vo.TradingServiceSucceeded, response.Outcome)
	assert.NotEmpty(t, response.Content, "空白的成功讀起來像壞掉了")
}

func TestATradingServiceThatIsNotThereIsNotARefusal(t *testing.T) {
	proxy := tradingservice.NewTradingServiceProxy("http://127.0.0.1:1", time.Second)

	_, sendError := proxy.Send(context.Background(),
		vo.TradingServiceRequestVo{Verb: vo.RequestVerbRead, Path: "/health"}, "")

	assert.ErrorIs(t, sendError, domains.ErrTradingServiceUnreachable)
}

func TestWatchingStopsAtTheFirstUpdate(t *testing.T) {
	proxy, seen := standingInFor(t, func(writer http.ResponseWriter) {
		writer.Header().Set("Content-Type", "text/event-stream")
		_, _ = writer.Write([]byte("data: {\"symbol\":\"BTCUSDT\",\"status\":\"forming\"}\n\n" +
			"data: {\"symbol\":\"BTCUSDT\",\"status\":\"closed\"}\n\n"))
	})

	response, sendError := proxy.Send(context.Background(), vo.TradingServiceRequestVo{
		Verb:                vo.RequestVerbRead,
		Path:                "/k-candles/live",
		Query:               map[string]string{"symbol": "BTCUSDT"},
		LiveUpdateWaitLimit: 2 * time.Second,
	}, "")

	require.NoError(t, sendError)
	assert.Equal(t, vo.TradingServiceSucceeded, response.Outcome)
	assert.JSONEq(t, `{"symbol":"BTCUSDT","status":"forming"}`, response.Content)
	assert.Equal(t, "text/event-stream", seen.accept)
}

func TestWatchingAQuietMarketIsNotAFailure(t *testing.T) {
	proxy, _ := standingInFor(t, func(writer http.ResponseWriter) {
		writer.Header().Set("Content-Type", "text/event-stream")
		writer.(http.Flusher).Flush()
		time.Sleep(400 * time.Millisecond)
	})

	response, sendError := proxy.Send(context.Background(), vo.TradingServiceRequestVo{
		Verb:                vo.RequestVerbRead,
		Path:                "/k-candles/live",
		LiveUpdateWaitLimit: 150 * time.Millisecond,
	}, "")

	require.NoError(t, sendError)
	assert.Equal(t, vo.TradingServiceSucceeded, response.Outcome)
	assert.Contains(t, response.Content, "沒有收到任何即時更新")
}

func TestWatchingSomethingTheTradingServiceRefusesIsStillARefusal(t *testing.T) {
	proxy, _ := standingInFor(t, func(writer http.ResponseWriter) {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(`{"message":"symbol 不得為空"}`))
	})

	response, sendError := proxy.Send(context.Background(), vo.TradingServiceRequestVo{
		Verb:                vo.RequestVerbRead,
		Path:                "/k-candles/live",
		LiveUpdateWaitLimit: time.Second,
	}, "")

	require.NoError(t, sendError)
	assert.Equal(t, vo.TradingServiceRefused, response.Outcome)
	assert.Contains(t, response.Content, "symbol 不得為空")
}

func TestWatchingSomethingThatIsNotThereIsNotARefusalEither(t *testing.T) {
	proxy := tradingservice.NewTradingServiceProxy("http://127.0.0.1:1", time.Second)

	_, sendError := proxy.Send(context.Background(), vo.TradingServiceRequestVo{
		Verb:                vo.RequestVerbRead,
		Path:                "/k-candles/live",
		LiveUpdateWaitLimit: time.Second,
	}, "")

	assert.ErrorIs(t, sendError, domains.ErrTradingServiceUnreachable)
}

func TestAnAddressThatCannotBeBuiltIsSaidRatherThanSent(t *testing.T) {
	proxy := tradingservice.NewTradingServiceProxy("://not-an-address", time.Second)

	_, sendError := proxy.Send(context.Background(),
		vo.TradingServiceRequestVo{Verb: vo.RequestVerbRead, Path: "/health"}, "")

	assert.ErrorIs(t, sendError, domains.ErrTradingServiceUnreachable)
}

func TestAStreamThatBreaksMidWayIsNotReportedAsAQuietMarket(t *testing.T) {
	standIn := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, _ *http.Request) {
			writer.Header().Set("Content-Type", "text/event-stream")
			writer.(http.Flusher).Flush()
			// 一行長到讀不完——對讀的人來說，這跟「線斷了」是同一件事。
			_, _ = writer.Write([]byte("data: " + strings.Repeat("x", 128*1024)))
		}))
	t.Cleanup(standIn.Close)

	proxy := tradingservice.NewTradingServiceProxy(standIn.URL, 5*time.Second)

	_, sendError := proxy.Send(context.Background(), vo.TradingServiceRequestVo{
		Verb:                vo.RequestVerbRead,
		Path:                "/k-candles/live",
		LiveUpdateWaitLimit: 2 * time.Second,
	}, "")

	assert.ErrorIs(t, sendError, domains.ErrTradingServiceUnreachable,
		"說成「這段時間沒有更新」會讓人去看市場，而該看的是線路")
}

func TestAnAnswerIsReadOnlyUpToACeiling(t *testing.T) {
	proxy, _ := standingInFor(t, func(writer http.ResponseWriter) {
		writer.Header().Set("Content-Type", "application/json")
		for range 40 {
			_, _ = writer.Write([]byte(strings.Repeat("x", 1<<20)))
		}
	})

	response, sendError := proxy.Send(context.Background(),
		vo.TradingServiceRequestVo{Verb: vo.RequestVerbRead, Path: "/k-candles"}, "")

	require.NoError(t, sendError)
	assert.LessOrEqual(t, len(response.Content), 16<<20,
		"沒有上限的讀取會讓一個意外的大回應變成整個外掛的死亡——連帶帶走每個人的登入")
}

func TestAnAskThatWaitsLongerOutlastsTheUsualWait(t *testing.T) {
	testCases := []struct {
		name              string
		responseWaitLimit time.Duration
		wantsAnswer       bool
	}{
		{name: "the usual wait gives up on a slow answer", responseWaitLimit: 0, wantsAnswer: false},
		{name: "a longer wait receives it", responseWaitLimit: 2 * time.Second, wantsAnswer: true},
		{name: "a longer wait still gives up past its own limit", responseWaitLimit: 100 * time.Millisecond, wantsAnswer: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// The trading service takes 300ms; the usual wait is 150ms.
			standIn := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				time.Sleep(300 * time.Millisecond)
				_, _ = writer.Write([]byte(`{"summary":{}}`))
			}))
			t.Cleanup(standIn.Close)
			proxy := tradingservice.NewTradingServiceProxy(standIn.URL, 150*time.Millisecond)

			response, sendError := proxy.Send(context.Background(), vo.TradingServiceRequestVo{
				Verb: vo.RequestVerbSubmit, Path: "/backtests", ResponseWaitLimit: testCase.responseWaitLimit,
			}, "")

			if testCase.wantsAnswer {
				require.NoError(t, sendError)
				assert.Equal(t, vo.TradingServiceSucceeded, response.Outcome)

				return
			}
			assert.ErrorIs(t, sendError, domains.ErrTradingServiceUnreachable)
		})
	}
}

func TestInspectingAConnectorAuthorizationPostsTheTokenAsAFormAndReadsTheJudgement(t *testing.T) {
	proxy, seen := standingInFor(t, func(writer http.ResponseWriter) {
		_, _ = writer.Write([]byte(
			`{"active":true,"sub":"42","aud":"https://trading-mcp.example.com/mcp","exp":1790000000}`))
	})

	inspection, inspectionError := proxy.InspectConnectorAuthorization(context.Background(), "a-token")

	require.NoError(t, inspectionError)
	assert.Equal(t, http.MethodPost, seen.method)
	assert.Equal(t, "/oauth/introspection", seen.path)
	assert.Equal(t, "token=a-token", seen.body)
	assert.Empty(t, seen.authorization)
	assert.True(t, inspection.IsActive)
	assert.Equal(t, "42", inspection.Subject)
	assert.Equal(t, "https://trading-mcp.example.com/mcp", inspection.Audience)
	assert.Equal(t, time.Unix(1790000000, 0), inspection.ExpiresAt)
}

func TestAnInactiveJudgementIsAJudgementRatherThanAFailure(t *testing.T) {
	proxy, _ := standingInFor(t, func(writer http.ResponseWriter) {
		_, _ = writer.Write([]byte(`{"active":false}`))
	})

	inspection, inspectionError := proxy.InspectConnectorAuthorization(context.Background(), "a-token")

	require.NoError(t, inspectionError)
	assert.False(t, inspection.IsActive)
	assert.True(t, inspection.ExpiresAt.IsZero())
}

func TestNotGettingAJudgementIsBeingUnableToReachTheTradingService(t *testing.T) {
	testCases := []struct {
		name   string
		answer func(http.ResponseWriter)
	}{
		{"被限流", func(writer http.ResponseWriter) { writer.WriteHeader(http.StatusTooManyRequests) }},
		{"看不懂的回覆", func(writer http.ResponseWriter) { _, _ = writer.Write([]byte(`not json`)) }},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			proxy, _ := standingInFor(t, testCase.answer)

			_, inspectionError := proxy.InspectConnectorAuthorization(context.Background(), "a-token")

			assert.ErrorIs(t, inspectionError, domains.ErrTradingServiceUnreachable)
		})
	}
}

func TestInspectingWhenTheTradingServiceIsNotThereIsNotAJudgement(t *testing.T) {
	testCases := []struct {
		name    string
		baseUrl string
	}{
		{"沒有人在聽", "http://127.0.0.1:1"},
		{"位址組不出來", "://not-an-address"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			proxy := tradingservice.NewTradingServiceProxy(testCase.baseUrl, time.Second)

			_, inspectionError := proxy.InspectConnectorAuthorization(context.Background(), "a-token")

			assert.ErrorIs(t, inspectionError, domains.ErrTradingServiceUnreachable)
		})
	}
}

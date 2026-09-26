package tradingservice

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/domains"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

// answerSizeCeiling is the most of one answer this connector will read.
//
// The trading service is trusted, but "trusted" is about intent and this is about
// accident: a query that matches far more than expected, a stuck stream, a reply
// that is not what it claims. Reading without a ceiling turns any of those into this
// process running out of memory.
//
// Sixteen mebibytes is far above the largest honest answer here (a thousand candles
// is a few hundred kilobytes; a replay with its trade detail, a few megabytes).
const answerSizeCeiling = 16 << 20

// methodsPerVerbs is the one place an ask becomes an HTTP method.
//
// One place, so that the domain never has to hold an HTTP word, and so that moving to
// a trading service that speaks something else is a change here and nowhere else.
var methodsPerVerbs = map[vo.RequestVerb]string{
	vo.RequestVerbRead:    http.MethodGet,
	vo.RequestVerbSubmit:  http.MethodPost,
	vo.RequestVerbReplace: http.MethodPut,
	vo.RequestVerbRemove:  http.MethodDelete,
}

// TradingServiceProxy is how this connector talks to the trading service, and the
// only thing in it that knows HTTP exists.
//
// It hides two ways of talking behind one: an ordinary ask, and staying on a line for
// a while to see what comes through. The domain writes a wait limit on the request and
// knows nothing else about it — which is what stops "is this the streaming one?" from
// becoming a branch that every future change has to remember to keep in step.
//
// How long an ask waits is decided per ask rather than once for the whole client: the
// usual wait for most abilities, and a longer one for those the request says take
// longer. A single client-wide timeout would cut every replay off at the usual wait.
type TradingServiceProxy struct {
	baseUrl        string
	httpClient     *http.Client
	requestTimeout time.Duration
}

func NewTradingServiceProxy(baseUrl string, requestTimeout time.Duration) *TradingServiceProxy {
	return &TradingServiceProxy{
		baseUrl:        strings.TrimSuffix(baseUrl, "/"),
		httpClient:     &http.Client{},
		requestTimeout: requestTimeout,
	}
}

// Send carries out one ask.
func (tradingServiceProxy *TradingServiceProxy) Send(
	ctx context.Context,
	request vo.TradingServiceRequestVo,
	accessToken string,
) (vo.TradingServiceResponseVo, error) {
	if request.LiveUpdateWaitLimit > 0 {
		return tradingServiceProxy.peekLiveUpdates(ctx, request, accessToken)
	}

	// The wait covers reading the answer as well as sending the ask, so it is held
	// until this returns.
	responseWaitLimit := tradingServiceProxy.requestTimeout
	if request.ResponseWaitLimit > 0 {
		responseWaitLimit = request.ResponseWaitLimit
	}
	askCtx, stopWaiting := context.WithTimeout(ctx, responseWaitLimit)
	defer stopWaiting()

	httpResponse, sendError := tradingServiceProxy.send(askCtx, request, accessToken, "application/json")
	if sendError != nil {
		return vo.TradingServiceResponseVo{}, sendError
	}
	defer httpResponse.Body.Close()

	content, readError := io.ReadAll(io.LimitReader(httpResponse.Body, answerSizeCeiling))
	if readError != nil {
		return vo.TradingServiceResponseVo{}, fmt.Errorf(
			"%w：%s", domains.ErrTradingServiceUnreachable, readError.Error())
	}

	return tradingServiceProxy.responseOf(httpResponse.StatusCode, string(content)), nil
}

// InspectConnectorAuthorization asks the trading service's introspection endpoint
// how it judges one connector authorization.
func (tradingServiceProxy *TradingServiceProxy) InspectConnectorAuthorization(
	ctx context.Context,
	accessToken string,
) (vo.ConnectorAuthorizationInspectionVo, error) {
	askCtx, stopWaiting := context.WithTimeout(ctx, tradingServiceProxy.requestTimeout)
	defer stopWaiting()

	httpRequest, buildError := http.NewRequestWithContext(
		askCtx, http.MethodPost, tradingServiceProxy.baseUrl+"/oauth/introspection",
		strings.NewReader(url.Values{"token": {accessToken}}.Encode()))
	if buildError != nil {
		return vo.ConnectorAuthorizationInspectionVo{}, fmt.Errorf(
			"%w：%s", domains.ErrTradingServiceUnreachable, buildError.Error())
	}

	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpRequest.Header.Set("Accept", "application/json")

	httpResponse, sendError := tradingServiceProxy.httpClient.Do(httpRequest)
	if sendError != nil {
		return vo.ConnectorAuthorizationInspectionVo{}, fmt.Errorf(
			"%w：%s", domains.ErrTradingServiceUnreachable, sendError.Error())
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return vo.ConnectorAuthorizationInspectionVo{}, fmt.Errorf(
			"%w：確認外掛授權時交易服務回了 %d", domains.ErrTradingServiceUnreachable, httpResponse.StatusCode)
	}

	var inspection connectorAuthorizationInspectionWire
	if decodeError := json.NewDecoder(
		io.LimitReader(httpResponse.Body, answerSizeCeiling)).Decode(&inspection); decodeError != nil {
		return vo.ConnectorAuthorizationInspectionVo{}, fmt.Errorf(
			"%w：交易服務回了一份看不懂的外掛授權確認", domains.ErrTradingServiceUnreachable)
	}

	return inspection.ToConnectorAuthorizationInspectionVo(), nil
}

// peekLiveUpdates stays on the line only as long as it is worth staying.
//
// It stops at the first update, because one update answers the question that was
// asked ("what does it look like now"), and everything after it is a second answer
// nobody waited for. It also stops when the wait limit runs out, with nothing — and
// nothing is a real answer here: an idle market at three in the morning is not a
// failure, and calling it one would have people restarting a connector that works.
func (tradingServiceProxy *TradingServiceProxy) peekLiveUpdates(
	ctx context.Context,
	request vo.TradingServiceRequestVo,
	accessToken string,
) (vo.TradingServiceResponseVo, error) {
	watchCtx, stopWatching := context.WithTimeout(ctx, request.LiveUpdateWaitLimit)
	defer stopWatching()

	httpResponse, sendError := tradingServiceProxy.send(
		watchCtx, request, accessToken, "text/event-stream")
	if sendError != nil {
		return vo.TradingServiceResponseVo{}, sendError
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		content, _ := io.ReadAll(io.LimitReader(httpResponse.Body, answerSizeCeiling))

		return tradingServiceProxy.responseOf(httpResponse.StatusCode, string(content)), nil
	}

	lines := bufio.NewScanner(io.LimitReader(httpResponse.Body, answerSizeCeiling))
	for lines.Scan() {
		update, isUpdate := strings.CutPrefix(lines.Text(), "data:")
		if isUpdate {
			return vo.TradingServiceResponseVo{
				Outcome: vo.TradingServiceSucceeded,
				Content: strings.TrimSpace(update),
			}, nil
		}
	}

	// Running out of time is the ordinary way to leave this loop, and it is not a
	// failure: a quiet market at three in the morning has nothing to send.
	//
	// Anything *else* that ends the stream is, and has to be told apart. A line too
	// long to read, or a connection cut mid-event, would otherwise arrive as "nothing
	// came through in ten seconds" — a sentence that is not true, and that sends
	// somebody to look at the market when they should be looking at the wire.
	if readError := lines.Err(); readError != nil && !errors.Is(readError, context.DeadlineExceeded) {
		return vo.TradingServiceResponseVo{}, fmt.Errorf(
			"%w：即時更新讀到一半斷了：%s", domains.ErrTradingServiceUnreachable, readError.Error())
	}

	return vo.TradingServiceResponseVo{
		Outcome: vo.TradingServiceSucceeded,
		Content: fmt.Sprintf(
			"在 %s 之內沒有收到任何即時更新。這不是錯誤——市場安靜、或這一分鐘還沒有成交時本來就是這樣。",
			request.LiveUpdateWaitLimit),
	}, nil
}

// send is the one place a request actually leaves this process.
func (tradingServiceProxy *TradingServiceProxy) send(
	ctx context.Context,
	request vo.TradingServiceRequestVo,
	accessToken string,
	accepting string,
) (*http.Response, error) {
	query := url.Values{}
	for name, value := range request.Query {
		query.Set(name, value)
	}

	address := tradingServiceProxy.baseUrl + request.Path
	if len(query) > 0 {
		address += "?" + query.Encode()
	}

	httpRequest, buildError := http.NewRequestWithContext(
		ctx, methodsPerVerbs[request.Verb], address, bytes.NewReader(request.Body))
	if buildError != nil {
		return nil, fmt.Errorf("%w：%s", domains.ErrTradingServiceUnreachable, buildError.Error())
	}

	httpRequest.Header.Set("Accept", accepting)
	if len(request.Body) > 0 {
		httpRequest.Header.Set("Content-Type", "application/json")
	}

	if accessToken != "" {
		httpRequest.Header.Set("Authorization", "Bearer "+accessToken)
	}

	httpResponse, sendError := tradingServiceProxy.httpClient.Do(httpRequest)
	if sendError != nil {
		return nil, fmt.Errorf("%w：%s", domains.ErrTradingServiceUnreachable, sendError.Error())
	}

	return httpResponse, nil
}

// responseOf turns one answered ask into the verdict the domain reads.
//
// Not recognised is singled out because it is the one refusal the person answers by
// reconnecting rather than by rewording. Everything else — a rule not met, a thing not found, a database that
// would not read — is the trading service speaking, and is carried through in its
// own words rather than sorted into categories it did not ask for.
func (tradingServiceProxy *TradingServiceProxy) responseOf(
	statusCode int,
	content string,
) vo.TradingServiceResponseVo {
	if statusCode == http.StatusUnauthorized {
		return vo.TradingServiceResponseVo{
			Outcome: vo.TradingServiceIdentityNotRecognized,
			Content: content,
		}
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return vo.TradingServiceResponseVo{Outcome: vo.TradingServiceRefused, Content: content}
	}

	if strings.TrimSpace(content) == "" {
		return vo.TradingServiceResponseVo{
			Outcome: vo.TradingServiceSucceeded,
			Content: "已完成（交易服務沒有回傳內容）",
		}
	}

	return vo.TradingServiceResponseVo{Outcome: vo.TradingServiceSucceeded, Content: content}
}

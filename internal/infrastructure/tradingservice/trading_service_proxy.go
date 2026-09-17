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
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

// answerSizeCeiling is the most of one answer this connector will read.
//
// The trading service is trusted, but "trusted" is about intent and this is about
// accident: a query that matches far more than expected, a stuck stream, a reply
// that is not what it claims. Reading without a ceiling turns any of those into this
// process running out of memory — and a connector that dies takes every other
// person's signed-in session with it.
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
type TradingServiceProxy struct {
	baseUrl    string
	httpClient *http.Client
}

func NewTradingServiceProxy(baseUrl string, requestTimeout time.Duration) *TradingServiceProxy {
	return &TradingServiceProxy{
		baseUrl:    strings.TrimSuffix(baseUrl, "/"),
		httpClient: &http.Client{Timeout: requestTimeout},
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

	httpResponse, sendError := tradingServiceProxy.send(ctx, request, accessToken, "application/json")
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

// SignIn exchanges an account for a pair of proofs.
func (tradingServiceProxy *TradingServiceProxy) SignIn(
	ctx context.Context,
	signInDto dto.SignInDto,
) (vo.SessionGrantVo, error) {
	credentials, _ := json.Marshal(map[string]string{
		"email":    signInDto.Email,
		"password": signInDto.Password,
	})

	return tradingServiceProxy.askForGrant(ctx, "/sessions", credentials)
}

// RenewSession spends a renewal proof on a fresh pair.
func (tradingServiceProxy *TradingServiceProxy) RenewSession(
	ctx context.Context,
	refreshToken string,
) (vo.SessionGrantVo, error) {
	renewal, _ := json.Marshal(map[string]string{"refreshToken": refreshToken})

	return tradingServiceProxy.askForGrant(ctx, "/sessions/renewal", renewal)
}

// RevokeSession voids a whole renewal chain.
func (tradingServiceProxy *TradingServiceProxy) RevokeSession(
	ctx context.Context,
	refreshToken string,
) (vo.TradingServiceResponseVo, error) {
	revocation, _ := json.Marshal(map[string]string{"refreshToken": refreshToken})

	return tradingServiceProxy.Send(ctx, vo.TradingServiceRequestVo{
		Verb: vo.RequestVerbSubmit,
		Path: "/sessions/revocation",
		Body: revocation,
	}, "")
}

// askForGrant is the shape shared by signing in and renewing: post a small thing,
// and read a pair of proofs out of the answer.
//
// The two are one method because they differ only in where they are posted and what
// is posted there. Written twice, the reading of the pair would be written twice too,
// and the half that gets a field name wrong is the half nobody exercises until a
// signing-in expires in production.
func (tradingServiceProxy *TradingServiceProxy) askForGrant(
	ctx context.Context,
	path string,
	body []byte,
) (vo.SessionGrantVo, error) {
	response, sendError := tradingServiceProxy.Send(ctx, vo.TradingServiceRequestVo{
		Verb: vo.RequestVerbSubmit,
		Path: path,
		Body: body,
	}, "")
	if sendError != nil {
		return vo.SessionGrantVo{}, sendError
	}

	if response.Outcome != vo.TradingServiceSucceeded {
		return vo.SessionGrantVo{Outcome: response.Outcome, Content: response.Content}, nil
	}

	var grantedTokens sessionTokensWire
	if decodeError := json.Unmarshal([]byte(response.Content), &grantedTokens); decodeError != nil {
		return vo.SessionGrantVo{}, fmt.Errorf(
			"%w：交易服務回了一份看不懂的憑證", domains.ErrTradingServiceUnreachable)
	}

	return vo.SessionGrantVo{
		Outcome: vo.TradingServiceSucceeded,
		Tokens:  grantedTokens.ToTokenPairVo(),
	}, nil
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
// Not recognised is singled out because it is the one refusal this connector can act
// on by itself. Everything else — a rule not met, a thing not found, a database that
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

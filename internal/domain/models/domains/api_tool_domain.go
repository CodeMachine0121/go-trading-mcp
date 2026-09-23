package domains

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/dto"
	"github.com/CodeMachine0121/go-trading-mcp/internal/domain/models/vo"
)

// ErrRequiredArgumentMissing is what a half-filled ask gets. It names the box, so the
// assistant can fill that one in rather than re-send the same thing and be refused
// again.
var ErrRequiredArgumentMissing = errors.New("必填欄位沒有填")

// ApiToolDomain is one thing this connector can do, and everything about how one
// filled-in form turns into an ask the trading service understands.
//
// **It is a declaration, not a piece of code.** Every ability shares one path
// through this connector; what differs between fifty of them is only what is written
// down here. Adding the fifty-first is adding a line — not a handler, not a branch,
// not a place for one of fifty to be forgotten.
//
// It deliberately holds no rule about what the values *mean*. Whether a lookback of
// four thousand days is allowed, whether that trading symbol exists, whether that
// strategy script is yours — all of it belongs to the trading service, and a copy
// kept here is a copy that will one day disagree with the original.
type ApiToolDomain struct {
	name           string
	description    string
	verb           vo.RequestVerb
	pathTemplate   string
	parameters     []vo.ToolParameterVo
	requiresSignIn bool
	// liveUpdateWaitLimit is how long to stay on the line for an ability that
	// watches. Zero for every ordinary one.
	liveUpdateWaitLimit time.Duration
	// responseWaitLimit is how long to wait for an answer when this ability waits
	// longer than the connector's usual. Zero for every ordinary one.
	responseWaitLimit time.Duration
	// condensesReplayResults says this ability answers with a replay result, which is
	// condensed before an assistant reads it.
	condensesReplayResults bool
}

// NewApiToolDomain declares one ability.
func NewApiToolDomain(
	name string,
	description string,
	verb vo.RequestVerb,
	pathTemplate string,
	requiresSignIn bool,
	parameters ...vo.ToolParameterVo,
) ApiToolDomain {
	return ApiToolDomain{
		name:           name,
		description:    description,
		verb:           verb,
		pathTemplate:   pathTemplate,
		parameters:     parameters,
		requiresSignIn: requiresSignIn,
	}
}

// Watching turns this ability into one that stays on the line for a while rather
// than asking once, and says for how long.
//
// It is a separate step rather than a seventh argument on the constructor because
// exactly one ability in the catalog needs it, and an argument every declaration has
// to carry is an argument forty-nine of them would carry as zero.
func (apiToolDomain ApiToolDomain) Watching(waitLimit time.Duration) ApiToolDomain {
	apiToolDomain.liveUpdateWaitLimit = waitLimit

	return apiToolDomain
}

// Name is what the assistant calls this ability.
// Waiting gives this ability its own, longer wait for an answer. A replay can run for
// as long as the trading service allows it, and waiting only the usual thirty seconds
// would hand the assistant "cannot reach the trading service" about a replay that was
// about to finish — or about to say itself that it ran out of time.
func (apiToolDomain ApiToolDomain) Waiting(responseWaitLimit time.Duration) ApiToolDomain {
	apiToolDomain.responseWaitLimit = responseWaitLimit

	return apiToolDomain
}

// CondensingReplayResults marks this ability as answering with a replay result, so
// that a long one is condensed before an assistant reads it (see ReplayResultDomain).
func (apiToolDomain ApiToolDomain) CondensingReplayResults() ApiToolDomain {
	apiToolDomain.condensesReplayResults = true

	return apiToolDomain
}

// Relayed is the trading service's answer as this ability hands it on. Only a
// successful answer from an ability that answers with replay results is condensed; a
// refusal is the trading service's own sentence and always reaches the assistant
// exactly as it was said.
func (apiToolDomain ApiToolDomain) Relayed(
	response vo.TradingServiceResponseVo,
) vo.TradingServiceResponseVo {
	if !apiToolDomain.condensesReplayResults || response.Outcome != vo.TradingServiceSucceeded {
		return response
	}

	response.Content = NewReplayResultDomain(response.Content).ToCondensedContent()

	return response
}

func (apiToolDomain ApiToolDomain) Name() string {
	return apiToolDomain.name
}

// RequiresSignIn reports whether this ability has to know who is asking.
func (apiToolDomain ApiToolDomain) RequiresSignIn() bool {
	return apiToolDomain.requiresSignIn
}

// ToDefinitionDto is this ability as the assistant is told about it.
func (apiToolDomain ApiToolDomain) ToDefinitionDto() dto.ToolDefinitionDto {
	parameterDtos := make([]dto.ToolParameterDto, 0, len(apiToolDomain.parameters))
	for _, parameter := range apiToolDomain.parameters {
		parameterDtos = append(parameterDtos, dto.ToolParameterDto{
			Name:        parameter.Name,
			Kind:        string(parameter.Kind),
			Description: parameter.Description,
			IsRequired:  parameter.IsRequired,
		})
	}

	return dto.ToolDefinitionDto{
		Name:           apiToolDomain.name,
		Description:    apiToolDomain.description,
		RequiresSignIn: apiToolDomain.requiresSignIn,
		Parameters:     parameterDtos,
	}
}

// BuildRequest turns one filled-in form into one ask.
//
// The caller says this once and gets something ready to send. Behind it: every
// required box checked, every placeholder in the address filled and escaped, every
// value routed to the place its declaration named, and whatever is left over
// gathered into the contents. A caller that had to sequence those four would be a
// caller that could sequence them differently for the fifty-first ability.
func (apiToolDomain ApiToolDomain) BuildRequest(
	arguments ToolArgumentsDomain,
) (vo.TradingServiceRequestVo, error) {
	if missingName, isMissing := apiToolDomain.firstMissingRequiredName(arguments); isMissing {
		return vo.TradingServiceRequestVo{},
			fmt.Errorf("%w：%s", ErrRequiredArgumentMissing, missingName)
	}

	path := apiToolDomain.pathTemplate
	query := map[string]string{}
	bodyNames := make([]string, 0, len(apiToolDomain.parameters))

	for _, parameter := range apiToolDomain.parameters {
		if !arguments.Has(parameter.Name) {
			continue
		}

		switch parameter.Location {
		case vo.ToolParameterInPath:
			path = strings.ReplaceAll(
				path, "{"+parameter.Name+"}", url.PathEscape(arguments.PlainTextAt(parameter.Name)))
		case vo.ToolParameterInQuery:
			query[parameter.Name] = arguments.PlainTextAt(parameter.Name)
		case vo.ToolParameterInBody:
			bodyNames = append(bodyNames, parameter.Name)
		}
	}

	return vo.TradingServiceRequestVo{
		Verb:                apiToolDomain.verb,
		Path:                path,
		Query:               query,
		Body:                arguments.EncodedSubset(bodyNames),
		CarriesIdentity:     apiToolDomain.requiresSignIn,
		LiveUpdateWaitLimit: apiToolDomain.liveUpdateWaitLimit,
		ResponseWaitLimit:   apiToolDomain.responseWaitLimit,
	}, nil
}

// firstMissingRequiredName finds the one box to complain about. One rather than all,
// because the assistant fixes them one at a time anyway and a list reads as a bigger
// problem than it is.
func (apiToolDomain ApiToolDomain) firstMissingRequiredName(
	arguments ToolArgumentsDomain,
) (string, bool) {
	for _, parameter := range apiToolDomain.parameters {
		if parameter.IsRequired && !arguments.Has(parameter.Name) {
			return parameter.Name, true
		}
	}

	return "", false
}

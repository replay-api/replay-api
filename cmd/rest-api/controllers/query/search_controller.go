package query_controllers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golobby/container/v3"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
	iam_in "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/in"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/ports/in"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/in"
)

type SearchableHandler interface {
	HandleSearchRequest(w http.ResponseWriter, r *http.Request)
}

type SearchController[T any] struct {
	common.Searchable[T]
}

type SearchableResourceMultiplexer struct {
	Handlers      map[common.ResourceType]interface{}
	ResourceTypes []common.ResourceType
}

func NewSearchMux(c *container.Container) *SearchableResourceMultiplexer {
	smux := SearchableResourceMultiplexer{
		Handlers: make(map[common.ResourceType]interface{}),
	}

	// smux.Handlers[common.ResourceTypeBadge] = NewBadgeSearchController(c)
	// smux.Handlers[common.ResourceTypeRound] = NewMatchSearchController(c)
	smux.Handlers[common.ResourceTypeReplayFile] = NewReplayFileSearchController(c)
	smux.Handlers[common.ResourceTypeMatch] = NewMatchSearchController(c)
	smux.Handlers[common.ResourceTypePlayerMetadata] = NewPlayerSearchController(c)
	smux.Handlers[common.ResourceTypePlayerProfile] = NePlayerProfileSearchController(c)

	smux.Handlers[common.ResourceTypeGameEvent] = NewEventSearchController(c)
	smux.Handlers[common.ResourceTypeProfile] = NewProfileSearchController(c)
	smux.Handlers[common.ResourceTypeMembership] = NewMembershipSearchController(c)
	smux.Handlers[common.ResourceTypePlayerProfile] = NewPlayerProfileSearchController(c)
	// smux.Handlers[common.ResourceTypeTeam] = NewTeamSearchController(c)
	smux.ResourceTypes = make([]common.ResourceType, len(smux.Handlers))

	i := 0
	for k := range smux.Handlers {
		smux.ResourceTypes[i] = k
		i++
	}

	return &smux
}

func GetResourceStringFromPath(types []common.ResourceType, path string) string {
	parts := strings.Split(strings.ToLower(path), "/")

	if len(parts) <= 1 {
		return parts[0]
	}
	resourceLeaf := parts[len(parts)-1]

	if strings.Contains(resourceLeaf, "?") {
		resourceLeaf = strings.Split(resourceLeaf, "?")[0]
	}

	for _, res := range types {
		if resourceLeaf == strings.ToLower(fmt.Sprint(res)) {
			return resourceLeaf
		}
	}

	branched := strings.Join(parts[:len(parts)-1], "/")

	return GetResourceStringFromPath(types, branched)
}

func (smux *SearchableResourceMultiplexer) Dispatch(w http.ResponseWriter, r *http.Request) {
	resString := GetResourceStringFromPath(smux.ResourceTypes, r.URL.Path)

	for _, res := range smux.ResourceTypes {
		if resString == strings.ToLower(fmt.Sprint(res)) {
			smux.Handlers[res].(SearchableHandler).HandleSearchRequest(w, r)
			return
		}
	}

	err := fmt.Sprintf("Unable to resolve search handler for [%s %s]. Make sure [%s] matches an available resource, such as: %v", r.Method, r.URL.Path, resString, smux.ResourceTypes)

	slog.ErrorContext(r.Context(), "InvalidResource", "request", r, "error", err)
	http.Error(w, fmt.Sprintf("InvalidResource: %s", err), http.StatusBadRequest)
}

func NewMatchSearchController(c *container.Container) *SearchController[replay_entity.Match] {
	var s replay_in.MatchReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Error("Cannot resolve replay_out.MatchReader for NewMatchSearchController", "err", err)
		panic(err)
	}

	return &SearchController[replay_entity.Match]{
		Searchable: s,
	}
}

func NewReplayFileSearchController(c *container.Container) *SearchController[replay_entity.ReplayFile] {
	var s replay_in.ReplayFileReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Error("Cannot resolve replay_in.ReplayFileReader for NewReplayFileSearchController", "err", err)
		panic(err)
	}

	return &SearchController[replay_entity.ReplayFile]{
		Searchable: s,
	}
}

func NewTeamSearchController(c *container.Container) *SearchController[replay_entity.Team] {
	var s replay_in.TeamReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Error("Cannot resolve replay_in.TeamReader for NewTeamSearchController", "err", err)
		panic(err)
	}

	return &SearchController[replay_entity.Team]{
		Searchable: s,
	}
}

func NewPlayerSearchController(c *container.Container) *SearchController[replay_entity.PlayerMetadata] {
	var s replay_in.PlayerMetadataReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Error("Cannot resolve replay_in.PlayerReader for NewPlayerSearchController", "err", err)
		panic(err)
	}

	return &SearchController[replay_entity.PlayerMetadata]{
		Searchable: s,
	}
}

func NePlayerProfileSearchController(c *container.Container) *SearchController[squad_entities.PlayerProfile] {
	var s squad_in.PlayerProfileReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Error("Cannot resolve replay_in.PlayerProfileReader for NePlayerProfileSearchController", "err", err)
		panic(err)
	}

	return &SearchController[squad_entities.PlayerProfile]{
		Searchable: s,
	}
}

func NewEventSearchController(c *container.Container) *SearchController[replay_entity.GameEvent] {
	var s replay_in.EventReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Error("Cannot resolve replay_in.EventReader for NewEventSearchController", "err", err)
		panic(err)
	}

	return &SearchController[replay_entity.GameEvent]{
		Searchable: s,
	}
}

func NewBadgeSearchController(c *container.Container) *SearchController[replay_entity.Badge] {
	var s replay_in.BadgeReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Error("Cannot resolve replay_in.BadgeReader for NewBadgeSearchController", "err", err)
		panic(err)
	}

	return &SearchController[replay_entity.Badge]{
		Searchable: s,
	}
}

func NewProfileSearchController(c *container.Container) *SearchController[iam_entities.Profile] {
	var s iam_in.ProfileReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Error("Cannot resolve iam_entities.ProfileReader for NewProfileSearchController", "err", err)
		panic(err)
	}

	return &SearchController[iam_entities.Profile]{
		Searchable: s,
	}
}

func NewMembershipSearchController(c *container.Container) *SearchController[iam_entities.Membership] {
	var s iam_in.MembershipReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Error("Cannot resolve iam_entities.MembershipReader for NeMembershipSearchController", "err", err)
		panic(err)
	}

	return &SearchController[iam_entities.Membership]{
		Searchable: s,
	}
}

func NewPlayerProfileSearchController(c *container.Container) *SearchController[squad_entities.PlayerProfile] {
	var s squad_in.PlayerProfileReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Error("Cannot resolve squad_in.PlayerProfileReader for NewPlayerProfileSearchController", "err", err)
		panic(err)
	}

	return &SearchController[squad_entities.PlayerProfile]{
		Searchable: s,
	}
}

func GetPathParams(r *http.Request, s *common.Search) (*common.Search, error) {
	sanitizedPath := strings.Join(strings.Split(strings.ToLower(r.URL.Path), "/search/"), "")
	parts := strings.Split(sanitizedPath, "/") // test

	for i := range parts {
		if i <= 1 || i%2 != 0 {
			continue
		}

		fieldID, err := common.GetResourceFieldID(parts[i-2])

		if err != nil {
			return nil, err
		}

		value := common.SearchableValue{
			Field:    fieldID,
			Values:   []interface{}{parts[i-1]}, // TODO: parse multiple values !!!
			Operator: common.EqualsOperator,     // TODO: use IN when multiple values
		}

		params := []common.SearchParameter{
			{
				ValueParams: []common.SearchableValue{
					value,
				},
			},
		}

		s.SearchParams = append(s.SearchParams, common.SearchAggregation{
			Params: params,
		})
	}

	return s, nil
}

func InitializeSearch(r *http.Request) *common.Search {
	requestContext := r.Context()

	requestedAudience := common.GetIntendedAudience(requestContext)

	var intendedAudience common.IntendedAudienceKey

	if requestedAudience == nil {
		slog.WarnContext(requestContext, "Missing Requested Audience on r.Context, using Intented Audience on User level", "request", r)
		intendedAudience = common.UserAudienceIDKey
	} else {
		intendedAudience = *requestedAudience
	}

	s := common.Search{
		SearchParams: make([]common.SearchAggregation, 0),
		VisibilityOptions: common.SearchVisibilityOptions{
			RequestSource:    common.GetResourceOwner(requestContext),
			IntendedAudience: intendedAudience,
		},
	}
	return &s
}

func GetSearchParams(r *http.Request) (*common.Search, error) {
	s := InitializeSearch(r)

	s, err := GetPathParams(r, s)

	if err != nil {
		return nil, err
	}

	s = GetQueryParams(r, s)

	return s, nil
}

func GetQueryParams(r *http.Request, s *common.Search) *common.Search {
	queryParams := r.URL.Query()
	aggregation := common.SearchAggregation{
		Params: []common.SearchParameter{},
	}

	joinParam := queryParams["filter"]

	var operator common.SearchOperator

	if len(joinParam) > 0 {
		switch joinParam[0] {
		case "out":
			aggregation.AggregationClause = common.AndAggregationClause
			operator = common.EqualsOperator
		case "in":
			aggregation.AggregationClause = common.OrAggregationClause
			operator = common.ContainsOperator
		}
	}

	for key, values := range queryParams {
		if key == "filter" {
			continue
		}

		value := common.SearchableValue{
			Field:    key,
			Values:   make([]interface{}, len(values)),
			Operator: operator,
		}

		for i, v := range values {
			value.Values[i] = v
		}

		param := common.SearchParameter{
			ValueParams: []common.SearchableValue{
				value,
			},
			AggregationClause: aggregation.AggregationClause,
		}

		aggregation.Params = append(aggregation.Params, param)
	}

	s.SearchParams = append(s.SearchParams, aggregation)

	return s
}

func (c *SearchController[T]) HandleSearchRequest(w http.ResponseWriter, r *http.Request) {
	s, err := GetSearchParams(r)

	if err != nil {
		slog.ErrorContext(r.Context(), "BadRequest: unable to serialize URL parameters into common.Search", "request", r, "error", err)
		http.Error(w, "BadRequest", http.StatusBadRequest)
		return
	}

	result, err := c.Searchable.Search(
		r.Context(), *s,
	)

	if err != nil {
		if strings.Contains(err.Error(), "TENANCY") {
			slog.ErrorContext(r.Context(), "Unauthorized", "request", r, "error", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		slog.ErrorContext(r.Context(), "UnprocessableEntity", "request", r, "error", err)
		http.Error(w, "UnprocessableEntity", http.StatusUnprocessableEntity)
		return
	}

	if len(result) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

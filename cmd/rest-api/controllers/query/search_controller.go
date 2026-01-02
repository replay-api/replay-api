package query_controllers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
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
	smux.Handlers[common.ResourceTypePlayerProfile] = NewPlayerProfileSearchController(c)

	smux.Handlers[common.ResourceTypeGameEvent] = NewEventSearchController(c)
	smux.Handlers[common.ResourceTypeProfile] = NewProfileSearchController(c)
	smux.Handlers[common.ResourceTypeMembership] = NewMembershipSearchController(c)
	smux.Handlers[common.ResourceTypeSquad] = NewSquadSearchController(c)
	smux.Handlers[common.ResourceTypePlan] = NewPlanSearchController(c)
	smux.Handlers[common.ResourceTypeSubscription] = NewSubscriptionSearchController(c)
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

	res := common.ResourceType(resString)

	for _, resource := range smux.ResourceTypes {
		if strings.EqualFold(string(res), string(resource)) {
			smux.Handlers[resource].(SearchableHandler).HandleSearchRequest(w, r)
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
		slog.Warn("MatchReader not available for NewMatchSearchController", "error", err)
	}

	return &SearchController[replay_entity.Match]{
		Searchable: s,
	}
}

func NewReplayFileSearchController(c *container.Container) *SearchController[replay_entity.ReplayFile] {
	var s replay_in.ReplayFileReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Warn("ReplayFileReader not available for NewReplayFileSearchController", "error", err)
	}

	return &SearchController[replay_entity.ReplayFile]{
		Searchable: s,
	}
}

func NewTeamSearchController(c *container.Container) *SearchController[replay_entity.Team] {
	var s replay_in.TeamReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Warn("TeamReader not available for NewTeamSearchController", "error", err)
	}

	return &SearchController[replay_entity.Team]{
		Searchable: s,
	}
}

func NewPlayerSearchController(c *container.Container) *SearchController[replay_entity.PlayerMetadata] {
	var s replay_in.PlayerMetadataReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Warn("PlayerMetadataReader not available for NewPlayerSearchController", "error", err)
	}

	return &SearchController[replay_entity.PlayerMetadata]{
		Searchable: s,
	}
}

func NePlayerProfileSearchController(c *container.Container) *SearchController[squad_entities.PlayerProfile] {
	var s squad_in.PlayerProfileReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Warn("PlayerProfileReader not available for NePlayerProfileSearchController", "error", err)
	}

	return &SearchController[squad_entities.PlayerProfile]{
		Searchable: s,
	}
}

func NewEventSearchController(c *container.Container) *SearchController[replay_entity.GameEvent] {
	var s replay_in.EventReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Warn("EventReader not available for NewEventSearchController", "error", err)
	}

	return &SearchController[replay_entity.GameEvent]{
		Searchable: s,
	}
}

func NewBadgeSearchController(c *container.Container) *SearchController[replay_entity.Badge] {
	var s replay_in.BadgeReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Warn("BadgeReader not available for NewBadgeSearchController", "error", err)
	}

	return &SearchController[replay_entity.Badge]{
		Searchable: s,
	}
}

func NewProfileSearchController(c *container.Container) *SearchController[iam_entities.Profile] {
	var s iam_in.ProfileReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Warn("ProfileReader not available for NewProfileSearchController", "error", err)
	}

	return &SearchController[iam_entities.Profile]{
		Searchable: s,
	}
}

func NewMembershipSearchController(c *container.Container) *SearchController[iam_entities.Membership] {
	var s iam_in.MembershipReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Warn("MembershipReader not available for NeMembershipSearchController", "error", err)
	}

	return &SearchController[iam_entities.Membership]{
		Searchable: s,
	}
}

func NewPlayerProfileSearchController(c *container.Container) *SearchController[squad_entities.PlayerProfile] {
	var s squad_in.PlayerProfileReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Warn("PlayerProfileReader not available for NewPlayerProfileSearchController", "error", err)
	}

	return &SearchController[squad_entities.PlayerProfile]{
		Searchable: s,
	}
}

func NewSquadSearchController(c *container.Container) *SearchController[squad_entities.Squad] {
	var s squad_in.SquadReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Warn("SquadReader not available for NewSquadSearchController", "error", err)
	}

	return &SearchController[squad_entities.Squad]{
		Searchable: s,
	}
}

func NewSubscriptionSearchController(c *container.Container) *SearchController[billing_entities.Subscription] {
	var s billing_in.SubscriptionReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Warn("SubscriptionReader not available for NewSubscriptionSearchController", "error", err)
	}

	return &SearchController[billing_entities.Subscription]{
		Searchable: s,
	}
}

func NewPlanSearchController(c *container.Container) *SearchController[billing_entities.Plan] {
	var s billing_in.PlanReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Warn("PlanReader not available for NewPlanSearchController", "error", err)
	}

	return &SearchController[billing_entities.Plan]{
		Searchable: s,
	}
}

func GetPathParams(r *http.Request, s *common.Search) (*common.Search, error) {
	sanitizedPath := strings.Join(strings.Split(strings.ToLower(r.URL.Path), "/search/"), "")
	parts := strings.Split(sanitizedPath, "/")

	for i := range parts {
		if i <= 1 || i%2 != 0 {
			continue
		}

		fieldID, err := common.GetResourceFieldID(parts[i-2])

		if err != nil {
			return nil, err
		}

		parsedUUID, err := uuid.Parse(parts[i-1])
		if err != nil {
			// Skip this segment if it's not a valid UUID
			slog.WarnContext(r.Context(), "Skipping invalid UUID segment", "segment", parts[i-1], "error", err)
			continue
		}

		value := common.SearchableValue{
			Field:    fieldID,
			Values:   []interface{}{parsedUUID},
			Operator: common.EqualsOperator,
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

	if s.VisibilityOptions.IntendedAudience == 0 {
		slog.ErrorContext(r.Context(), "Unauthorized: missing audience", "request", r)
		s.VisibilityOptions.IntendedAudience = common.GroupAudienceIDKey
	}

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

	limitParam := queryParams["limit"]

	skipParam := queryParams["skip"]

	fullTextSearchParam := queryParams["query"]

	includeParams := queryParams["include"]

	for key, values := range queryParams {
		if key == "filter" || key == "limit" || key == "skip" || key == "query" || key == "include" {
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

		if len(fullTextSearchParam) > 0 {
			param.FullTextSearchParam = fullTextSearchParam[0]
		}

		aggregation.Params = append(aggregation.Params, param)
	}

	s.SearchParams = append(s.SearchParams, aggregation)

	s.ResultOptions = getResultOptions(limitParam, skipParam)

	for _, includeParam := range includeParams {
		resString := common.ResourceType(includeParam)

		s.IncludeParams = append(s.IncludeParams, common.IncludeParam{
			From:         resString,
			LocalField:   "ID",                             // expects to parse to baseentity._id or _id
			ForeignField: common.ResourceKeyMap[resString], // TODO: currently using snake case, needs to map to camel case in the repos (ie: subscription_id)
			IsArray:      true,
		})
	}

	return s
}

func getResultOptions(limitParam []string, skipParam []string) common.SearchResultOptions {
	var limit uint
	var offset uint

	if len(limitParam) > 0 {
		limitInt, _ := strconv.Atoi(limitParam[0])
		if limitInt > 0 {
			limit = uint(limitInt) // #nosec G115 - bounds checked
		}
	}

	if len(skipParam) > 0 {
		offsetInt, _ := strconv.Atoi(skipParam[0])
		if offsetInt > 0 {
			offset = uint(offsetInt) // #nosec G115 - bounds checked
		}
	}

	return common.SearchResultOptions{
		Limit: limit,
		Skip:  offset,
	}
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
	w.WriteHeader(http.StatusOK) // TODO: check if have other headers, check for internal error etc
	_ = json.NewEncoder(w).Encode(result)
}

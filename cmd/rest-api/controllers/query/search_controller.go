package query_controllers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golobby/container/v3"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
	replay_in "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/ports/in"
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
	smux.Handlers[common.ResourceTypePlayer] = NewPlayerSearchController(c)

	smux.Handlers[common.ResourceTypeGameEvent] = NewEventSearchController(c)
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

func NewPlayerSearchController(c *container.Container) *SearchController[replay_entity.Player] {
	var s replay_in.PlayerReader
	err := c.Resolve(&s)

	if err != nil {
		slog.Error("Cannot resolve replay_in.PlayerReader for NewPlayerSearchController", "err", err)
		panic(err)
	}

	return &SearchController[replay_entity.Player]{
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

func GetSearchParams(r *http.Request) (*common.Search, error) {
	sanitizedPath := strings.Join(strings.Split(strings.ToLower(r.URL.Path), "/search/"), "")
	parts := strings.Split(sanitizedPath, "/") // test

	s := common.Search{
		SearchParams: make([]common.SearchAggregation, 0),
		VisibilityOptions: common.SearchVisibilityOptions{
			RequestSource:    common.GetResourceOwner(r.Context()),
			IntendedAudience: common.UserAudienceIDKey,
		},
	}

	// TODO: refact
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

	// TODO: parsear parametros da query string: De acordo com a leaf desejada, parsear com base em um dictionary?
	// params := mux.Vars(r)
	queryParams := r.URL.Query()
	aggregation := common.SearchAggregation{
		Params: []common.SearchParameter{},
	}

	joinParam := queryParams["x"]

	if len(joinParam) > 0 {
		switch joinParam[0] {
		case "in":
			aggregation.AggregationClause = common.AndAggregationClause
		case "out":
			aggregation.AggregationClause = common.OrAggregationClause
		}
	}

	for key, values := range queryParams {
		value := common.SearchableValue{
			Field:    key,
			Values:   make([]interface{}, len(values)),
			Operator: common.EqualsOperator,
		}

		for i, v := range values {
			value.Values[i] = v
		}

		param := common.SearchParameter{
			ValueParams: []common.SearchableValue{
				value,
			},
		}

		aggregation.Params = append(aggregation.Params, param)
	}

	s.SearchParams = append(s.SearchParams, aggregation)

	return &s, nil
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

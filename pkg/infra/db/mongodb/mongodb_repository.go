package db

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
	"strings"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	MAX_RECURSIVE_DEPTH = 10 // disabled, to do next. (default=10)
	MAX_PAGE_SIZE       = 200
	DEFAULT_PAGE_SIZE   = 50
)

type CacheItem map[string]string

type MongoDBRepository[T common.Entity] struct {
	mongoClient       *mongo.Client
	dbName            string
	mappingCache      map[string]CacheItem
	entityModel       reflect.Type
	collectionName    string
	entityName        string
	bsonFieldMappings map[string]string // Local mapping of field names
	queryableFields   map[string]bool
	collection        *mongo.Collection
}

type MongoDBRepositoryBuilder[T common.BaseEntity] struct {
}

func (r *MongoDBRepository[T]) InitQueryableFields(queryableFields map[string]bool, bsonFieldMappings map[string]string) {
	r.queryableFields = queryableFields

	for k, v := range bsonFieldMappings {
		r.bsonFieldMappings[k] = v
	}

	r.collection = r.mongoClient.Database(r.dbName).Collection(r.collectionName)
}

func (r *MongoDBRepository[T]) GetBSONFieldName(fieldName string) (string, error) {
	if cachedBSONName, exists := r.mappingCache[r.entityName][fieldName]; exists {
		return cachedBSONName, nil
	}

	fieldParts := strings.Split(fieldName, ".")
	currentType := r.entityModel
	bsonFieldName := ""

	for i, part := range fieldParts {
		field, ok := currentType.FieldByName(part)
		if !ok {
			return "", fmt.Errorf("field %s (of %s) not found", part, currentType.Name())
		}

		bsonTag := field.Tag.Get("bson")
		if bsonTag == "" {
			return "", fmt.Errorf("field %s (of %s) does not have bson-tag", part, currentType.Name())
		}

		if bsonFieldName != "" {
			bsonFieldName += "."
		}
		bsonFieldName += bsonTag

		if bsonTag == "*" && i < len(fieldParts)-1 {
			return bsonFieldName, nil
		}

		if field.Type.Kind() == reflect.Struct && i < len(fieldParts)-1 {
			currentType = field.Type
		}
	}

	if _, exists := r.mappingCache[r.entityName]; !exists {
		r.mappingCache[r.entityName] = make(CacheItem)
	}
	r.mappingCache[r.entityName][fieldName] = bsonFieldName

	return bsonFieldName, nil
}

func (repo *MongoDBRepository[T]) Compile(ctx context.Context, searchParams []common.SearchAggregation, resultOptions common.SearchResultOptions) (*common.Search, error) {
	err := common.ValidateSearchParameters(searchParams, repo.queryableFields)
	if err != nil {
		return nil, fmt.Errorf("error validating search parameters: %v", err)
	}

	err = repo.ValidateBSONSetup(resultOptions, repo.bsonFieldMappings)
	if err != nil {
		return nil, fmt.Errorf("error validating result options: %v", err)
	}

	s := common.NewSearchByAggregation(ctx, searchParams, resultOptions, common.UserAudienceIDKey)

	return &s, nil
}

func (repo *MongoDBRepository[T]) ValidateBSONSetup(resultOptions common.SearchResultOptions, bsonFieldMappings map[string]string) error {
	if len(resultOptions.PickFields) > 0 && len(resultOptions.OmitFields) > 0 {
		return fmt.Errorf("cannot specify both pick and omit fields")
	}

	for _, field := range resultOptions.PickFields {
		if _, ok := bsonFieldMappings[field]; !ok {
			return fmt.Errorf("field %s is not a valid field to pick", field)
		}
	}

	for _, field := range resultOptions.OmitFields {
		if _, ok := bsonFieldMappings[field]; !ok {
			return fmt.Errorf("field %s is not a valid field to omit", field)
		}
	}

	return nil
}

func (r *MongoDBRepository[T]) Query(queryCtx context.Context, s common.Search) (*mongo.Cursor, error) {
	collection := r.mongoClient.Database(r.dbName).Collection(r.collectionName)

	pipe, err := r.GetPipeline(queryCtx, s)

	slog.InfoContext(queryCtx, "Query: built pipeline", "pipeline", pipe, "collectionName", r.collectionName)

	if err != nil {
		slog.ErrorContext(queryCtx, "unable to create query pipeline", "err", err)
		return nil, err
	}

	cursor, err := collection.Aggregate(queryCtx, pipe)
	if err != nil {
		slog.ErrorContext(queryCtx, "unable to open query cursor", "err", err)
		return nil, err
	}

	return cursor, nil
}

func (r *MongoDBRepository[T]) GetPipeline(queryCtx context.Context, s common.Search) ([]bson.M, error) {
	var pipe []bson.M

	pipe, err := r.addMatch(queryCtx, pipe, s)

	if err != nil {
		slog.ErrorContext(queryCtx, "GetPipeline: unable to build $match stage", "error", err)
		return nil, err
	}

	pipe = r.addProjection(pipe, s)
	pipe = r.addSort(pipe, s)
	pipe = r.addSkip(pipe, s)

	pipe, err = r.addLimit(pipe, s)

	if err != nil {
		slog.ErrorContext(queryCtx, "GetPipeline: unable to build $limit stage", "error", err)
		return nil, err
	}

	var pipeString string
	for _, stage := range pipe {
		pipeString += fmt.Sprintf("%v\n", stage)
	}

	slog.InfoContext(queryCtx, "GetPipeline: built pipeline", "pipeline", pipeString, "collectionName", r.collectionName)

	return pipe, nil
}

func (r *MongoDBRepository[T]) addLimit(pipe []bson.M, s common.Search) ([]bson.M, error) {
	limit := s.ResultOptions.Limit

	if limit <= 0 {
		limit = DEFAULT_PAGE_SIZE // TODO: Parametrizar
	}

	if limit > MAX_PAGE_SIZE { // TODO: Parametrizar
		err := fmt.Errorf("given page size %d exceeds the maximum limit of %d records per request", s.ResultOptions.Limit, MAX_PAGE_SIZE)

		return pipe, err
	}

	pipe = append(pipe, bson.M{"$limit": limit})
	return pipe, nil
}

func (r *MongoDBRepository[T]) addSkip(pipe []bson.M, s common.Search) []bson.M {
	pipe = append(pipe, bson.M{"$skip": s.ResultOptions.Skip})
	return pipe
}

func (r *MongoDBRepository[T]) addSort(pipe []bson.M, s common.Search) []bson.M {
	sortFields := []bson.D{}
	for _, sortOption := range s.SortOptions {
		sortFields = append(sortFields, bson.D{{Key: sortOption.Field, Value: sortOption.Direction}})
	}
	sortStage := bson.M{"$sort": sortFields}

	if (sortFields != nil) && (len(sortFields) > 0) {
		pipe = append(pipe, sortStage)
	}
	return pipe
}

func (r *MongoDBRepository[T]) addMatch(queryCtx context.Context, pipe []bson.M, s common.Search) ([]bson.M, error) {
	aggregate := bson.M{}
	for _, aggregator := range s.SearchParams {
		r.setMatchValues(queryCtx, aggregator.Params, &aggregate, aggregator.AggregationClause)
	}

	aggregate, err := r.EnsureTenancy(queryCtx, aggregate, s)

	if err != nil {
		slog.ErrorContext(queryCtx, "Pipeline (addMatch) aborted due to inconsistent tenancy", "err", err.Error())
		return nil, err
	}

	pipe = append(pipe, bson.M{"$match": aggregate})
	return pipe, nil
}

func (r *MongoDBRepository[T]) setMatchValues(queryCtx context.Context, params []common.SearchParameter, aggregate *bson.M, clause common.SearchAggregationClause) {
	if r.queryableFields == nil {
		panic(fmt.Errorf("queryableFields not initialized in MongoDBRepository of %s", r.entityName))
	}

	clauses := bson.A{}
	for _, p := range params {
		// Handle ValueParams
		for _, v := range p.ValueParams {
			bsonFieldName, err := r.GetBSONFieldNameFromSearchableValue(v)
			if err != nil {
				panic(err) // Retain panic for irrecoverable errors
			}

			// Check if the prefix is allowed
			if strings.HasSuffix(v.Field, ".*") {
				prefix := strings.TrimSuffix(v.Field, ".*")
				if !isPrefixAllowed(prefix, r.queryableFields) {
					panic(fmt.Errorf("filtering on fields matching '%s.*' is not permitted", prefix))
				}
			}

			filter := buildFilterForOperator(v.Operator, v.Values)
			if filter == nil {
				continue // Skip this value if not supported
			}

			// Build filter based on operator (default to $in if not specified)
			if strings.HasSuffix(v.Field, ".*") && strings.Contains(bsonFieldName, ".") {
				// Nested field with wildcard: use $elemMatch
				clauses = append(clauses, bson.M{bsonFieldName: bson.M{"$elemMatch": filter}})
			} else {
				clauses = append(clauses, bson.M{bsonFieldName: filter})
			}

			slog.InfoContext(queryCtx, "query: %v, value: %v", bsonFieldName, v.Values)
		}

		// Handle DateParams
		for _, d := range p.DateParams {
			bsonFieldName, err := r.GetBSONFieldName(d.Field)
			if err != nil {
				panic(err) // Retain panic for irrecoverable reflection errors
			}

			dateFilter := bson.M{}
			if d.Min != nil {
				dateFilter["$gte"] = *d.Min
			}
			if d.Max != nil {
				dateFilter["$lte"] = *d.Max
			}
			clauses = append(clauses, bson.M{bsonFieldName: dateFilter})
		}

		// Handle DurationParams (similar to DateParams)
		for _, dur := range p.DurationParams {
			bsonFieldName, err := r.GetBSONFieldName(dur.Field)
			if err != nil {
				panic(err) // Retain panic for irrecoverable reflection errors
			}

			durationFilter := bson.M{}
			if dur.Min != nil {
				durationFilter["$gte"] = *dur.Min
			}
			if dur.Max != nil {
				durationFilter["$lte"] = *dur.Max
			}
			clauses = append(clauses, bson.M{bsonFieldName: durationFilter})
		}

		if p.AggregationParams == nil {
			continue
		}

		for i, v := range p.AggregationParams {
			if i+1 >= MAX_RECURSIVE_DEPTH {
				slog.WarnContext(queryCtx, "setMatchValue MaxRecursiveDepth exceeded", "depth", i, "MAX_RECURSIVE_DEPTH", MAX_RECURSIVE_DEPTH, "params", p.AggregationParams)
				break
			}

			innerAggregate := bson.M{}
			r.setMatchValues(queryCtx, v.Params, &innerAggregate, clause)
			clauses = append(clauses, innerAggregate)
		}
	}

	if len(clauses) > 0 {
		if clause == common.OrAggregationClause {
			(*aggregate)["$or"] = clauses
		} else {
			(*aggregate)["$and"] = clauses
		}
	}
}

// Helper function to build the filter based on the operator
func buildFilterForOperator(operator common.SearchOperator, values []interface{}) bson.M {
	switch operator {
	case common.EqualsOperator:
		return bson.M{"$eq": values[0]}
	case common.NotEqualsOperator:
		return bson.M{"$ne": values[0]}
	case common.GreaterThanOperator:
		return bson.M{"$gt": values[0]}
	case common.LessThanOperator:
		return bson.M{"$lt": values[0]}
	case common.GreaterThanOrEqualOperator:
		return bson.M{"$gte": values[0]}
	case common.LessThanOrEqualOperator:
		return bson.M{"$lte": values[0]}
	case common.ContainsOperator:
		return bson.M{"$regex": values[0], "$options": "i"}
	case common.StartsWithOperator:
		return bson.M{"$regex": "^" + fmt.Sprintf("%v", values[0]), "$options": "i"}
	case common.EndsWithOperator:
		return bson.M{"$regex": fmt.Sprintf("%v", values[0]) + "$", "$options": "i"}
	case common.InOperator:
		return bson.M{"$in": values}
	case common.NotInOperator:
		return bson.M{"$nin": values}
	default:
		return bson.M{"$in": values}
	}
}

// helper function to reuse code when checking prefix
func isPrefixAllowed(prefix string, queryableFields map[string]bool) bool {
	for allowedField := range queryableFields {
		if strings.HasPrefix(allowedField, prefix) {
			return true
		}
	}
	return false
}

func (r *MongoDBRepository[T]) GetBSONFieldNameFromSearchableValue(v common.SearchableValue) (string, error) {
	// Check for wildcard and handle directly
	if strings.HasSuffix(v.Field, ".*") {
		return r.GetBSONFieldName(v.Field[:len(v.Field)-2])
	}

	// Direct lookup in the mapping
	if bsonFieldName, ok := r.bsonFieldMappings[v.Field]; ok {
		return bsonFieldName, nil
	}

	slog.Warn("GetBSONFieldNameFromSearchableValue: field not found in mapping", "field", v.Field, "v", v)

	if v.Field == "" {
		return "", fmt.Errorf("empty field not allowed. cant query")
	}

	return "", fmt.Errorf("field %s not found or not queryable in Entity: %s (Collection: %s. Queryable Fields: %v)", v.Field, r.entityName, r.collectionName, r.queryableFields)
}

func (r *MongoDBRepository[T]) Create(ctx context.Context, entity *T) (*T, error) {
	_, err := r.collection.InsertOne(context.TODO(), entity)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return nil, err
	}

	return entity, nil
}

func (r *MongoDBRepository[T]) CreateMany(ctx context.Context, entities []*T) error {
	runtime.GC()
	toInsert := make([]interface{}, len(entities))
	defer runtime.GC()

	for i, e := range entities {
		toInsert[i] = e
	}

	_, err := r.collection.InsertMany(context.TODO(), toInsert)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return err
	}

	return nil
}

func (r *MongoDBRepository[T]) addProjection(pipe []bson.M, s common.Search) []bson.M {
	var projection *bson.M
	if len(s.ResultOptions.PickFields) > 0 {
		projection = &bson.M{}
		for _, field := range s.ResultOptions.PickFields {
			(*projection)[field] = 1
		}
	} else if len(s.ResultOptions.OmitFields) > 0 {
		projection = &bson.M{}
		for _, field := range s.ResultOptions.OmitFields {
			(*projection)[field] = 0
		}
	}

	if projection != nil {
		pipe = append(pipe, bson.M{"$project": *projection})
	}
	return pipe
}

func (r *MongoDBRepository[T]) EnsureTenancy(queryCtx context.Context, agg bson.M, s common.Search) (bson.M, error) {
	tenantID, ok := queryCtx.Value(common.TenantIDKey).(uuid.UUID)
	if !ok || tenantID == uuid.Nil {
		return agg, fmt.Errorf("TENANCY.RequestSource: valid tenant_id is required in queryCtx: %#v", queryCtx)
	}

	if s.VisibilityOptions.RequestSource.TenantID == uuid.Nil {
		return agg, fmt.Errorf("TENANCY.RequestSource: `tenant_id` is required but not provided in `common.Search`: %#v", s)
	} else if tenantID != s.VisibilityOptions.RequestSource.TenantID {
		return agg, fmt.Errorf("TENANCY.RequestSource: `tenant_id` in queryCtx does not match `tenant_id` in `common.Search`: %v vs %v", tenantID, s.VisibilityOptions.RequestSource.TenantID)
	}

	agg["$or"] = bson.A{
		bson.M{"resource_owner.tenant_id": tenantID},
		bson.M{"baseentity.resource_owner.tenant_id": tenantID}, // TODO: this OR on tenancy will degrade performance. choose to keep only base entity
	}

	slog.InfoContext(queryCtx, "TENANCY.RequestSource: tenant_id", "tenant_id", tenantID)

	switch s.VisibilityOptions.IntendedAudience {
	case common.ClientApplicationAudienceIDKey:
		return ensureClientID(queryCtx, agg, s)

	case common.GroupAudienceIDKey:
		return ensureUserOrGroupID(queryCtx, agg, s, s.VisibilityOptions.IntendedAudience)

	case common.UserAudienceIDKey:
		return ensureUserOrGroupID(queryCtx, agg, s, s.VisibilityOptions.IntendedAudience)

	case common.TenantAudienceIDKey:
		slog.WarnContext(queryCtx, "TENANCY.Admin: tenant audience is not allowed", "intendedAudience", s.VisibilityOptions.IntendedAudience)
		return agg, fmt.Errorf("TENANCY.Admin: tenant audience is not allowed")

	default:
		slog.WarnContext(queryCtx, "TENANCY.Unknown: intended audience is invalid", "intendedAudience", s.VisibilityOptions.IntendedAudience)
		return agg, fmt.Errorf("TENANCY.Unknown: intended audience %s is invalid in `common.Search`: %#v", string(s.VisibilityOptions.IntendedAudience), s)
	}
}

func ensureClientID(ctx context.Context, agg bson.M, s common.Search) (bson.M, error) {
	clientID, ok := ctx.Value(common.ClientIDKey).(uuid.UUID)
	if !ok || clientID == uuid.Nil {
		return agg, fmt.Errorf("TENANCY.ApplicationLevel: valid client_id is required in queryCtx: %#v", ctx)
	}

	if s.VisibilityOptions.RequestSource.ClientID == uuid.Nil {
		return agg, fmt.Errorf("TENANCY.ApplicationLevel: `client_id` is required in `common.Search`: %#v", s)
	}

	if clientID != s.VisibilityOptions.RequestSource.ClientID {
		return agg, fmt.Errorf("TENANCY.ApplicationLevel: `client_id` in queryCtx does not match `client_id` in `common.Search`: %v vs %v", clientID, s.VisibilityOptions.RequestSource.ClientID)
	}

	agg["$or"] = bson.A{
		bson.M{"resource_owner.client_id": clientID},
		bson.M{"baseentity.resource_owner.client_id": clientID}, // TODO: this OR on tenancy will degrade performance. choose to keep only base entity
	}

	slog.InfoContext(ctx, "TENANCY.ApplicationLevel: client_id", "client_id", clientID)

	return agg, nil
}

func ensureUserOrGroupID(ctx context.Context, agg bson.M, s common.Search, aud common.IntendedAudienceKey) (bson.M, error) {
	groupID := s.VisibilityOptions.RequestSource.GroupID
	userID := s.VisibilityOptions.RequestSource.UserID
	userOK := userID != uuid.Nil
	groupOK := groupID != uuid.Nil
	noParams := !(userOK || groupOK)

	if noParams {
		return agg, fmt.Errorf("TENANCY.GroupLevel: valid user_id or group_id is required in search parameters: %#v", s)
	}

	if aud == common.GroupAudienceIDKey && !groupOK {
		return agg, fmt.Errorf("TENANCY.GroupLevel: group_id is required in search parameters for intended audience: %s", string(aud))
	}

	if aud == common.UserAudienceIDKey && !userOK {
		return agg, fmt.Errorf("TENANCY.UserLevel: user_id is required in search parameters for intended audience: %s", string(aud))
	}

	if groupOK && userOK {
		agg["$or"] = bson.A{
			bson.M{"resource_owner.group_id": groupID},
			bson.M{"resource_owner.user_id": userID},
			bson.M{"baseentity.resource_owner.group_id": groupID}, // TODO: review, for performance reasons, use only baseentity
			bson.M{"baseentity.resource_owner.user_id": userID},
		}
		slog.InfoContext(ctx, "TENANCY.GroupLevel: group_id OR user_id", "group_id", groupID, "user_id", userID)
	} else if groupOK {
		agg["$or"] = bson.A{
			bson.M{"resource_owner.group_id": groupID},
			bson.M{"baseentity.resource_owner.group_id": groupID}, // TODO: this OR on tenancy will degrade performance. choose to keep only base entity
		}
		slog.InfoContext(ctx, "TENANCY.GroupLevel: group_id", "group_id", groupID)
	} else {
		agg["$or"] = bson.A{
			bson.M{"resource_owner.user_id": userID},
			bson.M{"baseentity.resource_owner.user_id": userID}, // TODO: this OR on tenancy will degrade performance. choose to keep only base entity
		}
		slog.InfoContext(ctx, "TENANCY.UserLevel: user_id", "user_id", userID)
	}

	return agg, nil
}

func (r *MongoDBRepository[T]) Search(ctx context.Context, s common.Search) ([]T, error) {
	cursor, err := r.Query(ctx, s)
	if cursor != nil {
		defer cursor.Close(ctx)
	}

	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("error querying on %s (%#v) (defaultSearch)", r.entityName, r), "err", err, "search", s)
		return nil, err
	}

	records := make([]T, 0)

	for cursor.Next(ctx) {
		var entity T
		err := cursor.Decode(&entity)

		if err != nil {
			slog.ErrorContext(ctx, "error decoding replay file metadata", "err", err)
			return nil, err
		}

		records = append(records, entity)
	}

	return records, nil
}

func (r *MongoDBRepository[T]) GetByID(queryCtx context.Context, id uuid.UUID) (*T, error) {
	var entity T

	query := bson.D{
		{Key: "_id", Value: id},
	}

	err := r.collection.FindOne(queryCtx, query).Decode(&entity)
	if err != nil {
		slog.ErrorContext(queryCtx, err.Error())
		return nil, err
	}

	return &entity, nil
}

func (r *MongoDBRepository[T]) Update(createCtx context.Context, entity *T) (*T, error) {
	id := (*entity).GetID()

	_, err := r.collection.UpdateOne(createCtx, bson.M{"_id": id}, bson.M{"$set": entity})
	if err != nil {
		slog.ErrorContext(createCtx, err.Error(), "entity", entity)
		return nil, err
	}

	return entity, nil
}

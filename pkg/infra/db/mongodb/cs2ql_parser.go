package db

import (
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

// MongoDBQuery represents a MongoDB query structure
type MongoDBQuery struct {
	Filter  bson.M
	Project bson.M
	Options *bson.M // Optional options for sorting, limiting, etc.
}

// ConvertCS2MatchQLToMongo converts a CS2MatchQL query into a MongoDB query.
func ConvertCS2MatchQLToMongo(query string) (*MongoDBQuery, error) {
	// Tokenize the query
	tokens := strings.Fields(query) // Simple tokenization for illustration

	mongoQuery := &MongoDBQuery{
		Filter:  bson.M{},
		Project: bson.M{},
	}

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		switch token {
		case "match":
			// Start of a new match query, potentially with filters
			// ... (handle filters for match, round, etc.)
		case "event":
			// Filter by event type
			i++
			eventType := strings.Trim(tokens[i], "()")
			mongoQuery.Filter["eventType"] = eventType
		case "state":
			// Filter by state property
			i++
			stateProp := tokens[i]
			i++ // Assuming a comparison operator (=, !=, etc.)
			stateValue := tokens[i]
			mongoQuery.Filter["state."+stateProp] = stateValue // Construct the state filter
		case "stats":
			// Handle stats aggregation
			// ... (complex logic depending on stats function and arguments)
		default:
			// Handle other keywords or filter conditions
			if matches := regexp.MustCompile(`(.+?)(\W+)(.+?)`).FindStringSubmatch(token); len(matches) > 0 {
				field := matches[1]
				op := matches[2]
				value := matches[3]
				switch op {
				case "=":
					mongoQuery.Filter[field] = value
				case "!=":
					mongoQuery.Filter[field] = bson.M{"$ne": value}
				case ">":
					mongoQuery.Filter[field] = bson.M{"$gt": value}
				case "<":
					mongoQuery.Filter[field] = bson.M{"$lt": value}
					// ... (handle other operators)
				}
			}
		}
	}

	return mongoQuery, nil
}

// func test() {
// 	// Example usage:
// 	cs2MatchQLQuery := `
//     match
//     where round.number = 15
//     and event(Kill)
//     and state.bombPlanted = true
//     and killer.name = 's1mple'
//     `

// 	mongoQuery, err := ConvertCS2MatchQLToMongo(cs2MatchQLQuery)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}

// 	fmt.Printf("MongoDB Query: %+v\n", mongoQuery)
// }

// expect >>>
// &MongoDBQuery{
// 	Filter: bson.M{
// 		"round.number": "15",
// 		"eventType": "Kill",
// 		"state.bombPlanted": true,
// 		"killer.name": "s1mple",
// 	},
// 	Project: bson.M{},
// 	}

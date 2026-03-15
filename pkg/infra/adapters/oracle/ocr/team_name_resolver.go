package oracle_ocr

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TeamNameResolver resolves OCR-detected team names to team IDs using MongoDB + fuzzy matching
type TeamNameResolver struct {
	collection *mongo.Collection
}

// Compile-time interface check
var _ oracle_out.TeamResolverPort = (*TeamNameResolver)(nil)

// NewTeamNameResolver creates a new team name resolver backed by MongoDB
func NewTeamNameResolver(db *mongo.Database) *TeamNameResolver {
	return &TeamNameResolver{
		collection: db.Collection("teams"),
	}
}

// ResolveTeam resolves a team name (from OCR text) to a known team in the database.
// Strategy:
// 1. Exact match on Name, ShortName, or CurrentDisplayName (case-insensitive)
// 2. Match in NameHistory (case-insensitive)
// 3. Fuzzy match using Levenshtein distance
func (r *TeamNameResolver) ResolveTeam(ctx context.Context, name string, gameID replay_common.GameIDKey) (*oracle_out.TeamRef, error) {
	if name == "" {
		return nil, fmt.Errorf("empty team name")
	}

	normalizedName := normalizeTeamName(name)

	// Strategy 1: Exact match (case-insensitive) on the main name fields
	ref, err := r.findExactMatch(ctx, normalizedName)
	if err == nil && ref != nil {
		return ref, nil
	}

	// Strategy 2: Match in NameHistory
	ref, err = r.findInHistory(ctx, normalizedName)
	if err == nil && ref != nil {
		return ref, nil
	}

	// Strategy 3: Fuzzy match — load candidate teams and compute Levenshtein
	ref, err = r.findFuzzyMatch(ctx, normalizedName)
	if err == nil && ref != nil {
		return ref, nil
	}

	return nil, fmt.Errorf("team not found: %q", name)
}

// teamDoc holds only the fields we need from the team document
type teamDoc struct {
	ID                 uuid.UUID `bson:"_id"`
	Name               string    `bson:"name"`
	ShortName          string    `bson:"short_name"`
	CurrentDisplayName string    `bson:"display_name"`
	NameHistory        []string  `bson:"name_history"`
}

func (r *TeamNameResolver) findExactMatch(ctx context.Context, name string) (*oracle_out.TeamRef, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": "^" + escapeRegex(name) + "$", "$options": "i"}},
			{"short_name": bson.M{"$regex": "^" + escapeRegex(name) + "$", "$options": "i"}},
			{"display_name": bson.M{"$regex": "^" + escapeRegex(name) + "$", "$options": "i"}},
		},
	}

	opts := options.FindOne().SetProjection(bson.M{
		"_id": 1, "name": 1, "short_name": 1, "display_name": 1,
	})

	var doc teamDoc
	if err := r.collection.FindOne(ctx, filter, opts).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	matchedName := doc.CurrentDisplayName
	if matchedName == "" {
		matchedName = doc.Name
	}

	return &oracle_out.TeamRef{
		TeamID:      doc.ID,
		MatchedName: matchedName,
		Confidence:  1.0,
	}, nil
}

func (r *TeamNameResolver) findInHistory(ctx context.Context, name string) (*oracle_out.TeamRef, error) {
	filter := bson.M{
		"name_history": bson.M{"$regex": "^" + escapeRegex(name) + "$", "$options": "i"},
	}

	opts := options.FindOne().SetProjection(bson.M{
		"_id": 1, "name": 1, "display_name": 1,
	})

	var doc teamDoc
	if err := r.collection.FindOne(ctx, filter, opts).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	matchedName := doc.CurrentDisplayName
	if matchedName == "" {
		matchedName = doc.Name
	}

	return &oracle_out.TeamRef{
		TeamID:      doc.ID,
		MatchedName: matchedName,
		Confidence:  0.85,
	}, nil
}

func (r *TeamNameResolver) findFuzzyMatch(ctx context.Context, name string) (*oracle_out.TeamRef, error) {
	// Load candidates (limit to prevent excess memory use)
	opts := options.Find().
		SetProjection(bson.M{
			"_id": 1, "name": 1, "short_name": 1, "display_name": 1, "name_history": 1,
		}).
		SetLimit(500)

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bestRef *oracle_out.TeamRef
	bestDistance := len(name) // Worst case = all chars different
	threshold := max(len(name)/3, 2) // Allow up to 33% difference

	for cursor.Next(ctx) {
		var doc teamDoc
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		// Check all name variants
		candidates := []string{doc.Name, doc.ShortName, doc.CurrentDisplayName}
		candidates = append(candidates, doc.NameHistory...)

		for _, candidate := range candidates {
			if candidate == "" {
				continue
			}

			dist := levenshteinDistance(normalizeTeamName(candidate), name)
			if dist < bestDistance && dist <= threshold {
				bestDistance = dist
				matchedName := doc.CurrentDisplayName
				if matchedName == "" {
					matchedName = doc.Name
				}

				// Confidence inversely proportional to distance
				confidence := 1.0 - float64(dist)/float64(max(len(name), len(candidate)))
				if confidence < 0.5 {
					confidence = 0.5
				}

				bestRef = &oracle_out.TeamRef{
					TeamID:      doc.ID,
					MatchedName: matchedName,
					Confidence:  confidence,
				}
			}
		}
	}

	if bestRef != nil {
		slog.DebugContext(ctx, "fuzzy team match found",
			slog.String("input", name),
			slog.String("matched", bestRef.MatchedName),
			slog.Float64("confidence", bestRef.Confidence),
			slog.Int("distance", bestDistance),
		)
	}

	return bestRef, nil
}

// normalizeTeamName normalizes a team name for comparison
func normalizeTeamName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)

	// Remove common suffixes/prefixes that OCR may or may not detect
	name = strings.TrimSuffix(name, " esports")
	name = strings.TrimSuffix(name, " gaming")
	name = strings.TrimSuffix(name, " team")

	// Remove non-alphanumeric characters (keep spaces)
	var sb strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			sb.WriteRune(r)
		}
	}

	return strings.TrimSpace(sb.String())
}

// levenshteinDistance computes the Levenshtein distance between two strings
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Use two rows instead of full matrix
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)

	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			curr[j] = min(
				curr[j-1]+1,   // insertion
				prev[j]+1,     // deletion
				prev[j-1]+cost, // substitution
			)
		}
		prev, curr = curr, prev
	}

	return prev[len(b)]
}

// escapeRegex escapes special regex characters in a string
func escapeRegex(s string) string {
	return regexp.QuoteMeta(s)
}

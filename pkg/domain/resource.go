package common

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type ResourceType string

type PlayerIDType uuid.UUID

// draft
// CS:2
// TeamA > Match - PUB #1231233 > Round #5 > Event #7
// /teams/:team_slug/matches/1827381723/rounds/5/events/7

// TeamA > Match - PUB #1231233 > Player #6 > Round #5 > Events > Event #7
// /teams/9128123881723712837/matches/1827381723/players/77777/rounds/5/events/7

// TeamA > Match - PUB #1231233 > Player #6 > Round #5 > Events > Event #7
// /teams/9128123881723712837/matches/1827381723/rounds/5/events/7

// Player #6 > TeamA > Match - PUB #1231233 > Round #5 > Events
// ---- /players/1/matches/1/rounds/5/events/7

// Tags/LAN/CS:GO/2021/NA/ESL/ProLeague/Season1/Week1/Day1/Match1

// Matches
// Rounds
// Events

const (
	ResourceTypeGameEvent      ResourceType = "GameEvents"
	ResourceTypeBadge          ResourceType = "Badges"
	ResourceTypeReplayFile     ResourceType = "ReplayFiles"
	ResourceTypeMatch          ResourceType = "Matches"
	ResourceTypeRound          ResourceType = "Rounds"
	ResourceTypeGame           ResourceType = "Games"
	ResourceTypePlayerProfile  ResourceType = "Players"        // composition of user
	ResourceTypePlayerMetadata ResourceType = "PlayerMetadata" // composition of user
	ResourceTypeTeam           ResourceType = "Teams"          // specification of user group
	ResourceTypeGroup          ResourceType = "Groups"         // system group
	ResourceTypeUser           ResourceType = "Users"          // specification of group
	ResourceTypeChannel        ResourceType = "Channels"       // specification of user group
	ResourceTypeLeague         ResourceType = "Leagues"        // specification of user group
	ResourceTypeTournament     ResourceType = "Tournaments"    // specification of user group
	ResourceTypeProfile        ResourceType = "Profiles"       // specification of user group
	ResourceTypeMembership     ResourceType = "Memberships"    // specification of user group
	ResourceTypePage           ResourceType = "Pages"          //
	ResourceTypeFriends        ResourceType = "Friends"
	ResourceTypeList           ResourceType = "List" // recurse root resources (?)
	// ResourceTypeMe         ResourceType = "Me"
	// ResourceTypeCustom     ResourceType = "Custom"
	// ResourceTypePublic     ResourceType = "Public"
	// ResourceTypePublicAny  ResourceType = "Public(Anyone with the link)"
	// ResourceTypePrivate   ResourceType = "Private"
	// ResourceTypeNamespace ResourceType = "Namespaces"
	ResourceTypeTag ResourceType = "Tags"
	// ResourceTypeBugReport ResourceType = "BugReports"
)

var ResourceKeyMap = map[ResourceType]string{
	ResourceTypeGameEvent:     "game_event_id",
	ResourceTypeBadge:         "badge_id",
	ResourceTypeReplayFile:    "replay_file_id",
	ResourceTypeMatch:         "match_id",
	ResourceTypeRound:         "round_id",
	ResourceTypeGame:          "game_id",
	ResourceTypePlayerProfile: "player_id",
	ResourceTypeTeam:          "team_id",
	ResourceTypeGroup:         "group_id",
	ResourceTypeUser:          "user_id",
	ResourceTypeProfile:       "profile_id",
}

func GetResourceFieldID(resourcePart string) (string, error) {
	for k, v := range ResourceKeyMap {
		if strings.EqualFold(fmt.Sprint(k), resourcePart) {
			return v, nil
		}
	}

	return "", fmt.Errorf("failed to parse ResourceIDField: Unknown resource %s", resourcePart)
}

type Resource struct {
	ID   uuid.UUID    `json:"id" bson:"_id"`
	Type ResourceType `json:"type" bson:"type"`
	// ResourceSlug string       `json:"resource_slug" bson:"resource_slug"`
}

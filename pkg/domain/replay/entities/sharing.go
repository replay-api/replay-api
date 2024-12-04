package entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
)

type ShareTokenStatus string

const (
	ShareTokenStatusActive  ShareTokenStatus = "Active"
	ShareTokenStatusUsed    ShareTokenStatus = "Used"
	ShareTokenStatusExpired ShareTokenStatus = "Expired"
)

type SharingResourceType string

// Anchors: Tag, File, Directory, Match (options: TickRange), Round, Event* (options: Flash/WeaponFire/Frag), Log (options: Including tick range), Player, Team, Network, Connections/Friends/Stars/Starred/Liked/* (after saving replay file, can ask if any of those players are friends, in case not integrating steam), (Tournament, Season, Channel: options: including sub-anchor filter / todo: multilevel x recursion), Region, Game* (comparing %, rates, etc**, usage..),

// TODO: Tags podem ser sugeridas ao parsear o arquivo, exmplo: clutch 1v5, ace, <player> mvp, <player> <event-slug>,
// TODO: no front, pode mostrar embaixo do textarea os chips com cada item desses para adicionar antes de salvar ***

const (
	SharingResourceContentTypeMatchStats  SharingResourceType = "MatchStats"
	SharingResourceContentTypeTeamStats   SharingResourceType = "TeamStats"
	SharingResourceContentTypePlayerStats SharingResourceType = "PlayerStats"
	SharingResourceContentTypeRoundStats  SharingResourceType = "RoundStats"
	// SharingResourceTypeAllStats           SharingResourceType = "Stats"

	SharingResourceContentTypeNetworkStats SharingResourceType = "NetworkStats"

	SharingResourceTypeBadge                 SharingResourceType = "Badge"
	SharingResourceTypeTag                   SharingResourceType = "Tag"
	SharingResourceContentTypeBugReport      SharingResourceType = "BugReport"
	SharingResourceContentTypeCommunityEvent SharingResourceType = "CommunityEvent"

	SharingResourceContentTypeDirectoryStats           SharingResourceType = "DirectoryStats"
	SharingResourceContentTypeGroupStats               SharingResourceType = "GroupStats"
	SharingResourceTypeReplayFileContent               SharingResourceType = "ReplayFileContent"
	SharingResourceContentTypeReplayFileHeader         SharingResourceType = "ReplayFileHeader"
	SharingResourceContentTypeReplayFileStats          SharingResourceType = "ReplayFileStats"
	SharingResourceContentTypeReplayFileTickRangeStats SharingResourceType = "ReplayFileTickRangeStats"
)

// TODO: incluir no base repositories para verificar os sharetokens ativos que possam ser usados para acessar recursos e adicionar na condicao do $match
// TODO: inlcuir todos os grupos que faca parte (se não especificado na consulta), (verificar todos os share tokens visiveis, também caso se grupo não especificado, ou buscar por ID / link)
type SharingActionType string

const (
	SharingActionTypeAllow   SharingActionType = "Allow"
	SharingActionTypeRequest SharingActionType = "Request"
	SharingActionTypeDeny    SharingActionType = "Deny"
)

type ShareToken struct {
	ID            uuid.UUID            `json:"token" bson:"token"`
	ResourceID    uuid.UUID            `json:"resource_id" bson:"resource_id"`
	ResourceType  SharingResourceType  `json:"resource_type" bson:"resource_type"`
	ExpiresAt     time.Time            `json:"expires_at" bson:"expires_at"`
	Uri           string               `json:"uri" bson:"uri"`
	EntityType    string               `json:"entity_type" bson:"entity_type"`
	Status        ShareTokenStatus     `json:"status" bson:"status"`
	ResourceOwner common.ResourceOwner `json:"resource_owner" bson:"resource_owner"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" bson:"updated_at"`

	// ShareToken    string               `json:"share_token" bson:"share_token"`
}

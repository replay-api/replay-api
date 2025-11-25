package billing_entities

import (
	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

type BillableEntry struct {
	common.BaseEntity
	OperationID    BillableOperationKey   `json:"operation_id" bson:"operation_id"`
	PlanID         uuid.UUID              `json:"plan_id" bson:"plan_id"`
	Amount         float64                `json:"amount" bson:"amount"`
	SubscriptionID uuid.UUID              `json:"subscription_id" bson:"subscription_id"`
	Args           map[string]interface{} `json:"args" bson:"args"`
}

type BillableOperationKey string

const (
	// REPLAY & CLOUD
	OperationTypeUploadedReplayFiles BillableOperationKey = "UploadReplayFiles"
	OperationTypeStorageLimit        BillableOperationKey = "StorageLimit"
	OperationTypeMediaAlbum          BillableOperationKey = "MediaAlbum"

	// WEB 3.0
	OperationTypeMintAssetAmount  BillableOperationKey = "MintAssetAmount"
	OperationTypeListAssetAmount  BillableOperationKey = "ListAssetAmount"
	OperationTypeSwapAssetAmount  BillableOperationKey = "SwapAssetAmount"
	OperationTypeStakeAssetAmount BillableOperationKey = "StakeAssetAmount"

	// SQUAD
	OperationTypeCreateSquadProfile      BillableOperationKey = "CreateSquadProfile"
	OperationTypeUpdateSquadProfile      BillableOperationKey = "UpdateSquadProfile"
	OperationTypeDeleteSquadProfile      BillableOperationKey = "DeleteSquadProfile"
	OperationTypePlayersPerSquadAmount   BillableOperationKey = "PlayersPerSquadAmount"
	OperationTypeAddSquadMember          BillableOperationKey = "AddSquadMember"
	OperationTypeRemoveSquadMember       BillableOperationKey = "RemoveSquadMember"
	OperationTypeUpdateSquadMemberRole   BillableOperationKey = "UpdateSquadMemberRole"
	OperationTypeCreatePlayerProfile     BillableOperationKey = "CreatePlayerProfile"
	OperationTypeUpdatePlayerProfile     BillableOperationKey = "UpdatePlayerProfile"
	OperationTypeDeletePlayerProfile     BillableOperationKey = "DeletePlayerProfile"
	OperationTypeSquadProfileBoostAmount BillableOperationKey = "SquadProfileBoostAmount"

	// MATCH-MAKING
	OperationTypeJoinMatchmakingQueue            BillableOperationKey = "JoinMatchmakingQueue"
	OperationTypeLeaveMatchmakingQueue           BillableOperationKey = "LeaveMatchmakingQueue"
	OperationTypeCreateCustomLobby               BillableOperationKey = "CreateCustomLobby"
	OperationTypeJoinLobby                       BillableOperationKey = "JoinLobby"
	OperationTypeLeaveLobby                      BillableOperationKey = "LeaveLobby"
	OperationTypeSetPlayerReady                  BillableOperationKey = "SetPlayerReady"
	OperationTypeMatchMakingQueueAmount          BillableOperationKey = "MatchMakingQueueAmount"
	OperationTypeMatchMakingCreateAmount         BillableOperationKey = "MatchMakingCreateAmount"
	OperationTypeMatchMakingPriorityQueue        BillableOperationKey = "MatchMakingPriorityQueue"
	OperationTypeMatchMakingCreateFeaturedBoosts BillableOperationKey = "MatchMakingCreateFeaturedBoosts"

	// TOURNAMENT
	OperationTypeCreateTournament                BillableOperationKey = "CreateTournament"
	OperationTypeRegisterForTournament           BillableOperationKey = "RegisterForTournament"
	OperationTypeUnregisterFromTournament        BillableOperationKey = "UnregisterFromTournament"
	OperationTypeGenerateBrackets                BillableOperationKey = "GenerateBrackets"
	OperationTypeCompleteMatch                   BillableOperationKey = "CompleteMatch"
	OperationTypeCompleteTournament              BillableOperationKey = "CompleteTournament"
	OperationTypeMatchMakingTournamentAmount     BillableOperationKey = "MatchMakingTournamentAmount" // TOURNAMENT (plus+)
	OperationTypeTournamentProfileAmount         BillableOperationKey = "TournamentProfileAmount"

	// STORE
	OperationTypeStoreAmount                    BillableOperationKey = "StoreAmount"
	OperationTypeStoreBoostAmount               BillableOperationKey = "StoreBoostAmount"
	OperationTypeStoreVoucherAmount             BillableOperationKey = "StoreVoucherAmount"
	OperationTypeStorePromotionAmount           BillableOperationKey = "StorePromotionAmount"
	OperationTypeFeatureStoreCatalogIntegration BillableOperationKey = "FeatureStoreCatalogIntegration"

	// WALLET
	OperationTypeFeatureGlobalWallet           BillableOperationKey = "FeatureGlobalWallet"
	OperationTypeGlobalWalletLimit             BillableOperationKey = "GlobalWalletLimit"
	OperationTypeGlobalPaymentProcessingAmount BillableOperationKey = "GlobalPaymentProcessingAmount" // (plus+)
	OperationTypeSquadPayout                   BillableOperationKey = "SquadPayout"                   // (plus+)

	// BETTING
	OperationTypeFeatureBetting      BillableOperationKey = "FeatureBetting"
	OperationTypeBettingCreateAmount BillableOperationKey = "BettingCreateAmount" // (plus+)
	OperationTypeBettingStakeAmount  BillableOperationKey = "BettingStakeAmount"
	OperationTypeBettingPayoutAmount BillableOperationKey = "BettingPayoutAmount" // (plus+)

	// ENGAGEMENT ITEMS
	OperationTypeFeatureAnalytics       BillableOperationKey = "FeatureAnalytics"
	OperationTypeFeatureChatAndComments BillableOperationKey = "FeatureChatAndComments"
	OperationTypeFeatureForum           BillableOperationKey = "FeatureForum"
	OperationTypeFeaturedBlogPostAmount BillableOperationKey = "FeaturedBlogPostAmount" // featured blog post on main page

	OperationTypeFeatureWiki                BillableOperationKey = "FeatureWiki"
	OperationTypeFeaturePoll                BillableOperationKey = "FeaturePoll"
	OperationTypeFeatureCalendarIntegration BillableOperationKey = "FeatureCalendarIntegration"
	OperationTypeFeatureTournament          BillableOperationKey = "FeatureTournament" // (plus+)
	OperationTypeFeatureAchievements        BillableOperationKey = "FeatureAchievements"

	OperationTypeNotifications      BillableOperationKey = "Notifications"
	OperationTypeCustomEmails       BillableOperationKey = "CustomEmails"
	OperationTypePremiumSupport     BillableOperationKey = "PremiumSupport"
	OperationTypeDiscordIntegration BillableOperationKey = "DiscordIntegration"
	OperationTypeFeatureAPI         BillableOperationKey = "FeatureAPI"      // (plus+)
	OperationTypeFeatureWebhooks    BillableOperationKey = "FeatureWebhooks" // (plus+)
	// OperationTypeCustomAvatars       BillableOperationKey = "CustomAvatars"
	OperationTypeExclusiveContent    BillableOperationKey = "ExclusiveContent"
	OperationTypeAdFreeExperience    BillableOperationKey = "AdFreeExperience"
	OperationTypeEarlyAccessFeatures BillableOperationKey = "EarlyAccessFeatures"
)

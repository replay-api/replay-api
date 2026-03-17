package ioc

import (
	"context"
	"log/slog"
	"os"
	"time"

	// env
	"github.com/joho/godotenv"

	// mongodb
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	// repositories/db
	db "github.com/replay-api/replay-api/pkg/infra/db/mongodb"
	"github.com/resource-ownership/go-mongodb/pkg/mongodb"

	// messageBroker (kafka/rabbit)
	kafka "github.com/replay-api/replay-api/pkg/infra/kafka"

	// encryption
	encryption "github.com/replay-api/replay-api/pkg/infra/crypto"

	// container
	container "github.com/golobby/container/v3"

	// local files

	// ports
	common "github.com/replay-api/replay-api/pkg/domain"
	email_entities "github.com/replay-api/replay-api/pkg/domain/email/entities"
	email_in "github.com/replay-api/replay-api/pkg/domain/email/ports/in"
	email_out "github.com/replay-api/replay-api/pkg/domain/email/ports/out"
	email_use_cases "github.com/replay-api/replay-api/pkg/domain/email/use_cases"
	google_in "github.com/replay-api/replay-api/pkg/domain/google/ports/in"
	google_out "github.com/replay-api/replay-api/pkg/domain/google/ports/out"
	google_use_cases "github.com/replay-api/replay-api/pkg/domain/google/use_cases"
	metadata "github.com/replay-api/replay-api/pkg/domain/replay/services/metadata"
	squad_entities "github.com/replay-api/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/replay-api/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/replay-api/replay-api/pkg/domain/squad/ports/out"
	squad_services "github.com/replay-api/replay-api/pkg/domain/squad/services"
	squad_usecases "github.com/replay-api/replay-api/pkg/domain/squad/usecases"

	replay_in "github.com/replay-api/replay-api/pkg/domain/replay/ports/in"
	replay_out "github.com/replay-api/replay-api/pkg/domain/replay/ports/out"

	steam_in "github.com/replay-api/replay-api/pkg/domain/steam/ports/in"
	steam_out "github.com/replay-api/replay-api/pkg/domain/steam/ports/out"
	steam_query_services "github.com/replay-api/replay-api/pkg/domain/steam/services"

	matchmaking_in "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/in"
	matchmaking_out "github.com/replay-api/replay-api/pkg/domain/matchmaking/ports/out"
	matchmaking_services "github.com/replay-api/replay-api/pkg/domain/matchmaking/services"

	matchmaking_usecases "github.com/replay-api/replay-api/pkg/domain/matchmaking/usecases"

	scores_in "github.com/replay-api/replay-api/pkg/domain/scores/ports/in"
	scores_out "github.com/replay-api/replay-api/pkg/domain/scores/ports/out"
	scores_usecases "github.com/replay-api/replay-api/pkg/domain/scores/usecases"
	scores_adapter "github.com/replay-api/replay-api/pkg/infra/adapters/scores"

	oracle_in "github.com/replay-api/replay-api/pkg/domain/oracle/ports/in"
	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	oracle_services "github.com/replay-api/replay-api/pkg/domain/oracle/services"
	oracle_usecases "github.com/replay-api/replay-api/pkg/domain/oracle/usecases"
	oracle_vo "github.com/replay-api/replay-api/pkg/domain/oracle/value-objects"
	oracle_chain "github.com/replay-api/replay-api/pkg/infra/adapters/oracle/chain"
	oracle_ocr "github.com/replay-api/replay-api/pkg/infra/adapters/oracle/ocr"
	oracle_providers "github.com/replay-api/replay-api/pkg/infra/adapters/oracle/providers"

	messaging_in "github.com/replay-api/replay-api/pkg/domain/messaging/ports/in"
	messaging_out "github.com/replay-api/replay-api/pkg/domain/messaging/ports/out"
	messaging_usecases "github.com/replay-api/replay-api/pkg/domain/messaging/usecases"

	prediction_in "github.com/replay-api/replay-api/pkg/domain/prediction/ports/in"
	prediction_out "github.com/replay-api/replay-api/pkg/domain/prediction/ports/out"
	prediction_usecases "github.com/replay-api/replay-api/pkg/domain/prediction/usecases"

	tournament_in "github.com/replay-api/replay-api/pkg/domain/tournament/ports/in"
	tournament_out "github.com/replay-api/replay-api/pkg/domain/tournament/ports/out"
	tournament_services "github.com/replay-api/replay-api/pkg/domain/tournament/services"
	tournament_usecases "github.com/replay-api/replay-api/pkg/domain/tournament/usecases"
	tournament_adapters "github.com/replay-api/replay-api/pkg/infra/adapters/tournaments"

	wallet_entities "github.com/replay-api/replay-api/pkg/domain/wallet/entities"
	wallet_in "github.com/replay-api/replay-api/pkg/domain/wallet/ports/in"
	wallet_out "github.com/replay-api/replay-api/pkg/domain/wallet/ports/out"
	wallet_services "github.com/replay-api/replay-api/pkg/domain/wallet/services"
	wallet_usecases "github.com/replay-api/replay-api/pkg/domain/wallet/usecases"

	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	billing_out "github.com/replay-api/replay-api/pkg/domain/billing/ports/out"
	billing_services "github.com/replay-api/replay-api/pkg/domain/billing/services"
	billing_usecases "github.com/replay-api/replay-api/pkg/domain/billing/usecases"

	payment_in "github.com/replay-api/replay-api/pkg/domain/payment/ports/in"
	payment_out "github.com/replay-api/replay-api/pkg/domain/payment/ports/out"
	payment_services "github.com/replay-api/replay-api/pkg/domain/payment/services"
	payment_usecases "github.com/replay-api/replay-api/pkg/domain/payment/usecases"

	stripe_adapter "github.com/replay-api/replay-api/pkg/infra/adapters/stripe"

	exchange_out "github.com/replay-api/replay-api/pkg/domain/exchange/ports/out"
	exchange_services "github.com/replay-api/replay-api/pkg/domain/exchange/services"
	exchange_usecases "github.com/replay-api/replay-api/pkg/domain/exchange/usecases"

	analytics_in "github.com/replay-api/replay-api/pkg/domain/analytics/ports/in"
	analytics_out "github.com/replay-api/replay-api/pkg/domain/analytics/ports/out"
	analytics_usecases "github.com/replay-api/replay-api/pkg/domain/analytics/usecases"

	pricefeed "github.com/replay-api/replay-api/pkg/infra/adapters/pricefeed"
	coinbase "github.com/replay-api/replay-api/pkg/infra/adapters/coinbase"
	kraken "github.com/replay-api/replay-api/pkg/infra/adapters/kraken"

	shared "github.com/resource-ownership/go-common/pkg/common"

	media_out "github.com/replay-api/replay-api/pkg/domain/media/ports/out"
	media_adapter "github.com/replay-api/replay-api/pkg/infra/adapters/media"

	websocket "github.com/replay-api/replay-api/pkg/infra/websocket"

	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
	iam_out "github.com/replay-api/replay-api/pkg/domain/iam/ports/out"
	iam_query_services "github.com/replay-api/replay-api/pkg/domain/iam/services"

	auth_in "github.com/replay-api/replay-api/pkg/domain/auth/ports/in"
	auth_out "github.com/replay-api/replay-api/pkg/domain/auth/ports/out"
	auth_services "github.com/replay-api/replay-api/pkg/domain/auth/services"

	email_adapter "github.com/replay-api/replay-api/pkg/infra/adapters/email"

	// domain
	google_entities "github.com/replay-api/replay-api/pkg/domain/google/entities"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	replay_entity "github.com/replay-api/replay-api/pkg/domain/replay/entities"
	steam_entity "github.com/replay-api/replay-api/pkg/domain/steam/entities"

	// app
	cs_app "github.com/replay-api/replay-api/pkg/app/cs"
	jobs "github.com/replay-api/replay-api/pkg/app/jobs"

	// usecases
	iam_use_cases "github.com/replay-api/replay-api/pkg/domain/iam/use_cases"
	replay_use_cases "github.com/replay-api/replay-api/pkg/domain/replay/use_cases"
	steam_use_cases "github.com/replay-api/replay-api/pkg/domain/steam/use_cases"
)

type ContainerBuilder struct {
	Container container.Container
}

// NoOpWalletCommand provides a no-op implementation of WalletCommand for basic functionality
type NoOpWalletCommand struct{}

func (n *NoOpWalletCommand) CreateWallet(ctx context.Context, cmd wallet_in.CreateWalletCommand) (*wallet_entities.UserWallet, error) {
	slog.Debug("[NoOpWalletCommand] CreateWallet called", "user_id", cmd.UserID)
	return nil, nil
}

func (n *NoOpWalletCommand) Deposit(ctx context.Context, cmd wallet_in.DepositCommand) error {
	slog.Debug("[NoOpWalletCommand] Deposit called", "user_id", cmd.UserID, "amount", cmd.Amount)
	return nil
}

func (n *NoOpWalletCommand) Withdraw(ctx context.Context, cmd wallet_in.WithdrawCommand) error {
	slog.Debug("[NoOpWalletCommand] Withdraw called", "user_id", cmd.UserID, "amount", cmd.Amount)
	return nil
}

func (n *NoOpWalletCommand) DeductEntryFee(ctx context.Context, cmd wallet_in.DeductEntryFeeCommand) error {
	slog.Debug("[NoOpWalletCommand] DeductEntryFee called", "user_id", cmd.UserID, "amount", cmd.Amount)
	return nil
}

func (n *NoOpWalletCommand) AddPrize(ctx context.Context, cmd wallet_in.AddPrizeCommand) error {
	slog.Debug("[NoOpWalletCommand] AddPrize called", "user_id", cmd.UserID, "amount", cmd.Amount)
	return nil
}

func (n *NoOpWalletCommand) Refund(ctx context.Context, cmd wallet_in.RefundCommand) error {
	slog.Debug("[NoOpWalletCommand] Refund called", "user_id", cmd.UserID, "amount", cmd.Amount)
	return nil
}

func (n *NoOpWalletCommand) DebitWallet(ctx context.Context, cmd wallet_in.DebitWalletCommand) (*wallet_entities.WalletTransaction, error) {
	slog.Debug("[NoOpWalletCommand] DebitWallet called", "user_id", cmd.UserID, "amount", cmd.Amount)
	return nil, nil
}

func (n *NoOpWalletCommand) CreditWallet(ctx context.Context, cmd wallet_in.CreditWalletCommand) (*wallet_entities.WalletTransaction, error) {
	slog.Debug("[NoOpWalletCommand] CreditWallet called", "user_id", cmd.UserID, "amount", cmd.Amount)
	return nil, nil
}

func NewContainerBuilder() *ContainerBuilder {
	c := container.New()

	b := &ContainerBuilder{
		c,
	}

	err := c.Singleton(func() container.Container {
		return b.Container
	})

	if err != nil {
		slog.Error("Failed to register *container.Container  in NewContainerBuilder.")
		panic(err)
	}

	err = c.Singleton(func() *ContainerBuilder {
		return b
	})

	if err != nil {
		slog.Error("Failed to register *ContainerBuilder in NewContainerBuilder.")
		panic(err)
	}

	return b
}

func (b *ContainerBuilder) Build() container.Container {
	return b.Container
}

func (b *ContainerBuilder) WithEnvFile() *ContainerBuilder {
	if os.Getenv("DEV_ENV") == "true" {
		if _, err := os.Stat(".env"); err == nil {
			if loadErr := godotenv.Load(); loadErr != nil {
				slog.Error("Failed to load .env file")
				panic(loadErr)
			}
		} else {
			slog.Info("No .env file found, using environment variables from system")
		}
	}

	err := b.Container.Singleton(func() (common.Config, error) {
		return EnvironmentConfig()
	})

	if err != nil {
		slog.Error("Failed to load EnvironmentConfig.")
		panic(err)
	}

	return b
}

func (b *ContainerBuilder) WithInboundPorts() *ContainerBuilder {
	c := b.Container

	err := c.Singleton(func() (replay_in.EventReader, error) {
		var gameEventReader replay_out.GameEventReader

		err := c.Resolve(&gameEventReader)
		if err != nil {
			slog.Error("Failed to resolve EventsByGameReader for EventsByGameService.", "err", err)
			return nil, err
		}

		return metadata.NewEventQueryService(gameEventReader), nil
	})

	if err != nil {
		slog.Error("Failed to load EventsByGameReader.")
		panic(err)
	}

	err = c.Singleton(func() (iam_in.CreateRIDTokenCommand, error) {
		var rIDWriter iam_out.RIDTokenWriter
		err := c.Resolve(&rIDWriter)
		if err != nil {
			slog.Error("Failed to resolve RIDWriter for OnboardRIDCommand.", "err", err)
			return nil, err
		}

		var rIDReader iam_out.RIDTokenReader
		err = c.Resolve(&rIDReader)
		if err != nil {
			slog.Error("Failed to resolve RIDReader for OnboardRIDCommand.", "err", err)
			return nil, err
		}

		return iam_use_cases.NewCreateRIDTokenUseCase(rIDWriter, rIDReader), nil
	})

	if err != nil {
		slog.Error("Failed to load iam_in.CreateRIDTokenCommand.")
		panic(err)
	}

	err = c.Singleton(func() (iam_in.OnboardOpenIDUserCommandHandler, error) {
		var userReader iam_out.UserReader
		err := c.Resolve(&userReader)
		if err != nil {
			slog.Error("Failed to resolve UserReader for OnboardOpenIDUserCommand.", "err", err)
			return nil, err
		}

		var userWriter iam_out.UserWriter
		err = c.Resolve(&userWriter)
		if err != nil {
			slog.Error("Failed to resolve UserWriter for OnboardOpenIDUserCommand.", "err", err)
			return nil, err
		}

		var profileReader iam_out.ProfileReader
		err = c.Resolve(&profileReader)
		if err != nil {
			slog.Error("Failed to resolve ProfileReader for OnboardOpenIDUserCommand.", "err", err)
			return nil, err
		}

		var profileWriter iam_out.ProfileWriter
		err = c.Resolve(&profileWriter)
		if err != nil {
			slog.Error("Failed to resolve ProfileWriter for OnboardOpenIDUserCommand.", "err", err)
			return nil, err
		}

		var groupWriter iam_out.GroupWriter
		err = c.Resolve(&groupWriter)
		if err != nil {
			slog.Error("Failed to resolve GroupWriter for OnboardOpenIDUserCommand.", "err", err)
			return nil, err
		}

		var membershipWriter iam_out.MembershipWriter
		err = c.Resolve(&membershipWriter)
		if err != nil {
			slog.Error("Failed to resolve MembershipWriter for OnboardOpenIDUserCommand.", "err", err)
			return nil, err
		}

		var createRIDTokenCommand iam_in.CreateRIDTokenCommand
		err = c.Resolve(&createRIDTokenCommand)
		if err != nil {
			slog.Error("Failed to resolve CreateRIDTokenCommand for OnboardSteamUserCommand.", "err", err)
			return nil, err
		}

		return iam_use_cases.NewOnboardOpenIDUserUseCase(userReader, userWriter, profileReader, profileWriter, groupWriter, membershipWriter, createRIDTokenCommand), nil
	})

	if err != nil {
		slog.Error("Failed to load OnboardOpenIDUserCommand.")
		panic(err)
	}

	err = c.Singleton(func() (replay_in.UploadReplayFileCommand, error) {
		var gameEventReader replay_in.EventReader
		err := c.Resolve(&gameEventReader)
		if err != nil {
			slog.Error("Failed to resolve replay_in.EventReader for replay_in.UploadReplayFileCommand.", "err", err)
			return nil, err
		}

		var replayFileMetadataReader replay_out.ReplayFileMetadataReader
		err = c.Resolve(&replayFileMetadataReader)
		if err != nil {
			slog.Error("Failed to resolve ReplayFileMetadataReader for replay_in.UploadReplayFileCommand.", "err", err)
			return nil, err
		}

		var ReplayFileMetadataWriter replay_out.ReplayFileMetadataWriter
		err = c.Resolve(&ReplayFileMetadataWriter)
		if err != nil {
			slog.Error("Failed to resolve ReplayFileMetadataWriter for replay_in.UploadReplayFileCommand.", "err", err)
			return nil, err
		}

		var replayDataWriter replay_out.ReplayFileContentWriter
		err = c.Resolve(&replayDataWriter)
		if err != nil {
			slog.Error("Failed to resolve ReplayFileContentWriter for replay_in.UploadReplayFileCommand.", "err", err)
			return nil, err
		}

		var eventPublisher replay_out.ReplayEventPublisher
		err = c.Resolve(&eventPublisher)
		if err != nil {
			slog.Warn("Failed to resolve ReplayEventPublisher for replay_in.UploadReplayFileCommand - continuing without Kafka events", "err", err)
			// Continue without event publisher - can be nil for local dev
		}

		return replay_use_cases.NewUploadReplayFileUseCase(replayFileMetadataReader, ReplayFileMetadataWriter, replayDataWriter, eventPublisher), nil
	})

	if err != nil {
		slog.Error("Failed to register replay_in.UploadReplayFileCommand with UploadReplayFileUseCase")
		panic(err)
	}

	err = c.Singleton(func() (replay_in.ProcessReplayFileCommand, error) {
		var replayFileMetadataReader replay_out.ReplayFileMetadataReader
		err = c.Resolve(&replayFileMetadataReader)
		if err != nil {
			slog.Error("Failed to resolve ReplayFileMetadataReader for ProcessReplayFileCommand.", "err", err)
			return nil, err
		}

		var replayFileDataReader replay_out.ReplayFileContentReader
		err = c.Resolve(&replayFileDataReader)
		if err != nil {
			slog.Error("Failed to resolve ReplayFileContentReader for ProcessReplayFileCommand.", "err", err)
			return nil, err
		}

		var ReplayFileMetadataWriter replay_out.ReplayFileMetadataWriter
		err = c.Resolve(&ReplayFileMetadataWriter)
		if err != nil {
			slog.Error("Failed to resolve ReplayFileMetadataWriter for ProcessReplayFileCommand.", "err", err)
			return nil, err
		}

		var replayDataWriter replay_out.ReplayFileContentWriter
		err = c.Resolve(&replayDataWriter)
		if err != nil {
			slog.Error("Failed to resolve ReplayFileContentWriter for ProcessReplayFileCommand.", "err", err)
			return nil, err
		}

		var replayCommand replay_out.ReplayParser
		err = c.Resolve(&replayCommand)
		if err != nil {
			slog.Error("Failed to resolve ReplayParser for ProcessReplayFileCommand.", "err", err)
			return nil, err
		}

		var eventWriter replay_out.GameEventWriter
		err = c.Resolve(&eventWriter)
		if err != nil {
			slog.Error("Failed to resolve GameEventWriter for ProcessReplayFileCommand.", "err", err)
			return nil, err
		}

		var playerMetadataWriter replay_out.PlayerMetadataWriter
		err = c.Resolve(&playerMetadataWriter)
		if err != nil {
			slog.Error("Failed to resolve PlayerMetadataWriter for ProcessReplayFileCommand.", "err", err)
			return nil, err
		}

		var matchMetadataWriter replay_out.MatchMetadataWriter
		err = c.Resolve(&matchMetadataWriter)
		if err != nil {
			slog.Error("Failed to resolve MatchMetadataWriter for ProcessReplayFileCommand.", "err", err)
			return nil, err
		}

		return replay_use_cases.NewProcessReplayFileUseCase(replayFileMetadataReader, replayFileDataReader, ReplayFileMetadataWriter, replayDataWriter, replayCommand, eventWriter, playerMetadataWriter, matchMetadataWriter), nil
	})

	if err != nil {
		slog.Error("Failed to load ProcessReplayFileCommand.")
		panic(err)
	}

	err = c.Singleton(func() (replay_in.UpdateReplayFileHeaderCommand, error) {
		var eventReader replay_out.GameEventReader
		err = c.Resolve(&eventReader)
		if err != nil {
			slog.Error("Failed to resolve replay_out.GameEventReader for replay_in.UpdateReplayFileHeaderCommand.", "err", err)
			return nil, err
		}

		var replayFileMetadataReader replay_out.ReplayFileMetadataReader
		err = c.Resolve(&replayFileMetadataReader)
		if err != nil {
			slog.Error("Failed to resolve replay_out.ReplayFileMetadataReader for replay_in.UpdateReplayFileHeaderCommand.", "err", err)
			return nil, err
		}

		var replayFileMetadataWriter replay_out.ReplayFileMetadataWriter
		err = c.Resolve(&replayFileMetadataWriter)
		if err != nil {
			slog.Error("Failed to resolve ReplayFileMetadataWriter for UploadReplayFileCommand.", "err", err)
			return nil, err
		}

		return replay_use_cases.NewUpdateReplayFileHeaderUseCase(eventReader, replayFileMetadataReader, replayFileMetadataWriter), nil
	})

	if err != nil {
		slog.Error("Failed to load replay_in.UpdateReplayFileHeaderCommand.")
		panic(err)
	}

	err = c.Singleton(func() (replay_in.UploadAndProcessReplayFileCommand, error) {
		var uploadReplayFileCommand replay_in.UploadReplayFileCommand
		err = c.Resolve(&uploadReplayFileCommand)
		if err != nil {
			slog.Error("Failed to resolve UploadReplayFileCommand for UploadAndProcessReplayFileCommand.", "err", err)
			return nil, err
		}

		var processReplayFileCommand replay_in.ProcessReplayFileCommand
		err = c.Resolve(&processReplayFileCommand)
		if err != nil {
			slog.Error("Failed to resolve ProcessReplayFileCommand for UploadAndProcessReplayFileCommand.", "err", err)
			return nil, err
		}

		var updateReplayFileHeaderCommand replay_in.UpdateReplayFileHeaderCommand
		err = c.Resolve(&updateReplayFileHeaderCommand)
		if err != nil {
			slog.Error("Failed to resolve replay_in.UpdateReplayFileHeaderCommand for replay_in.UploadAndProcessReplayFileCommand.", "err", err)
			return nil, err
		}

		return replay_use_cases.NewUploadAndProcessReplayFileUseCase(uploadReplayFileCommand, processReplayFileCommand, updateReplayFileHeaderCommand), nil
	})

	if err != nil {
		slog.Error("Failed to register replay_in.UploadAndProcessReplayFileCommand.")
		panic(err)
	}

	// Register UpdateReplayMetadataCommand for updating replay metadata (title, description, visibility, tags)
	err = c.Singleton(func() (replay_in.UpdateReplayMetadataCommand, error) {
		var replayFileMetadataReader replay_out.ReplayFileMetadataReader
		err := c.Resolve(&replayFileMetadataReader)
		if err != nil {
			slog.Error("Failed to resolve replay_out.ReplayFileMetadataReader for replay_in.UpdateReplayMetadataCommand.", "err", err)
			return nil, err
		}

		var replayFileMetadataWriter replay_out.ReplayFileMetadataWriter
		err = c.Resolve(&replayFileMetadataWriter)
		if err != nil {
			slog.Error("Failed to resolve replay_out.ReplayFileMetadataWriter for replay_in.UpdateReplayMetadataCommand.", "err", err)
			return nil, err
		}

		return replay_use_cases.NewUpdateReplayMetadataUseCase(replayFileMetadataReader, replayFileMetadataWriter), nil
	})

	if err != nil {
		slog.Error("Failed to register replay_in.UpdateReplayMetadataCommand.")
		panic(err)
	}

	err = c.Singleton(func() (replay_in.ReplayFileReader, error) {
		var replayFileMetadataReader replay_out.ReplayFileMetadataReader
		err := c.Resolve(&replayFileMetadataReader)

		if err != nil {
			slog.Error("Failed to resolve replay_out.ReplayFileMetadataReader for replay_in.ReplayFileMetadataReader.", "err", err)
			return nil, err
		}

		return metadata.NewReplayFileQueryService(replayFileMetadataReader), nil
	})

	if err != nil {
		slog.Error("Failed to register replay_in.ReplayFileMetadataReader.")
		panic(err)
	}

	err = c.Singleton(func() (replay_in.MatchReader, error) {
		var matchMetadataReader replay_out.MatchMetadataReader
		err := c.Resolve(&matchMetadataReader)

		if err != nil {
			slog.Error("Failed to resolve replay_out.MatchMetadataReader for replay_in.MatchReader.", "err", err)
			return nil, err
		}

		return metadata.NewMatchQueryService(matchMetadataReader), nil
	})

	if err != nil {
		slog.Error("Failed to register replay_in.MatchReader.")
		panic(err)
	}

	// TeamReader - queries team data (embedded in matches)
	err = c.Singleton(func() (replay_in.TeamReader, error) {
		var matchMetadataReader replay_out.MatchMetadataReader
		err := c.Resolve(&matchMetadataReader)

		if err != nil {
			slog.Error("Failed to resolve replay_out.MatchMetadataReader for replay_in.TeamReader.", "err", err)
			return nil, err
		}

		return metadata.NewTeamQueryService(matchMetadataReader), nil
	})

	if err != nil {
		slog.Error("Failed to register replay_in.TeamReader.")
		panic(err)
	}

	// RoundReader - queries round data (embedded in matches)
	err = c.Singleton(func() (replay_in.RoundReader, error) {
		var matchMetadataReader replay_out.MatchMetadataReader
		err := c.Resolve(&matchMetadataReader)

		if err != nil {
			slog.Error("Failed to resolve replay_out.MatchMetadataReader for replay_in.RoundReader.", "err", err)
			return nil, err
		}

		return metadata.NewRoundQueryService(matchMetadataReader), nil
	})

	if err != nil {
		slog.Error("Failed to register replay_in.RoundReader.")
		panic(err)
	}

	err = c.Singleton(func() (steam_in.OnboardSteamUserCommand, error) {
		var steamUserWriter steam_out.SteamUserWriter
		err := c.Resolve(&steamUserWriter)
		if err != nil {
			slog.Error("Failed to resolve SteamUserWriter for OnboardSteamUserCommand.", "err", err)
			return nil, err
		}

		var steamUserReader steam_out.SteamUserReader
		err = c.Resolve(&steamUserReader)
		if err != nil {
			slog.Error("Failed to resolve SteamUserReader for OnboardSteamUserCommand.", "err", err)
			return nil, err
		}

		var vHashWriter steam_out.VHashWriter
		err = c.Resolve(&vHashWriter)
		if err != nil {
			slog.Error("Failed to resolve VHashWriter for OnboardSteamUserCommand.", "err", err)
			return nil, err
		}

		var onboardOpenIDUser iam_in.OnboardOpenIDUserCommandHandler
		err = c.Resolve(&onboardOpenIDUser)
		if err != nil {
			slog.Error("Failed to resolve OnboardOpenIDUserCommandHandler for OnboardSteamUserCommand.", "err", err)
			return nil, err
		}

		return steam_use_cases.NewOnboardSteamUserUseCase(steamUserWriter, steamUserReader, vHashWriter, onboardOpenIDUser), nil
	})

	if err != nil {
		slog.Error("Failed to load OnboardSteamUserCommand.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (steam_in.SteamUserReader, error) {
		var steamUserReader steam_out.SteamUserReader
		err := c.Resolve(&steamUserReader)

		if err != nil {
			slog.Error("Failed to resolve replay_out.SteamUserReader for replay_in.SteamUserReader.", "err", err)
			return nil, err
		}

		return steam_query_services.NewSteamUserQueryService(steamUserReader), nil
	})

	if err != nil {
		slog.Error("Failed to register replay_in.ReplayFileMetadataReader.")
		panic(err)
	}

	err = c.Singleton(func() (google_in.OnboardGoogleUserCommand, error) {
		var googleUserWriter google_out.GoogleUserWriter
		err := c.Resolve(&googleUserWriter)
		if err != nil {
			slog.Error("Failed to resolve GoogleUserWriter for OnboardGoogleUserCommand.", "err", err)
			return nil, err
		}

		var googleUserReader google_out.GoogleUserReader
		err = c.Resolve(&googleUserReader)
		if err != nil {
			slog.Error("Failed to resolve GoogleUserReader for OnboardGoogleUserCommand.", "err", err)
			return nil, err
		}

		var vHashWriter google_out.VHashWriter
		err = c.Resolve(&vHashWriter)
		if err != nil {
			slog.Error("Failed to resolve VHashWriter for OnboardGoogleUserCommand.", "err", err)
			return nil, err
		}

		var onboardOpenIDUser iam_in.OnboardOpenIDUserCommandHandler
		err = c.Resolve(&onboardOpenIDUser)
		if err != nil {
			slog.Error("Failed to resolve OnboardOpenIDUserCommandHandler for OnboardGoogleUserCommand.", "err", err)
			return nil, err
		}

		return google_use_cases.NewOnboardGoogleUserUseCase(googleUserWriter, googleUserReader, vHashWriter, onboardOpenIDUser), nil
	})

	if err != nil {
		slog.Error("Failed to load OnboardGoogleUserCommand.", "err", err)
		panic(err)
	}

	// Email auth use cases
	err = c.Singleton(func() (email_in.OnboardEmailUserCommand, error) {
		var emailUserWriter email_out.EmailUserWriter
		err := c.Resolve(&emailUserWriter)
		if err != nil {
			slog.Error("Failed to resolve EmailUserWriter for OnboardEmailUserCommand.", "err", err)
			return nil, err
		}

		var emailUserReader email_out.EmailUserReader
		err = c.Resolve(&emailUserReader)
		if err != nil {
			slog.Error("Failed to resolve EmailUserReader for OnboardEmailUserCommand.", "err", err)
			return nil, err
		}

		var vHashWriter email_out.VHashWriter
		err = c.Resolve(&vHashWriter)
		if err != nil {
			slog.Error("Failed to resolve VHashWriter for OnboardEmailUserCommand.", "err", err)
			return nil, err
		}

		var passwordHasher email_out.PasswordHasher
		err = c.Resolve(&passwordHasher)
		if err != nil {
			slog.Error("Failed to resolve PasswordHasher for OnboardEmailUserCommand.", "err", err)
			return nil, err
		}

		var onboardOpenIDUser iam_in.OnboardOpenIDUserCommandHandler
		err = c.Resolve(&onboardOpenIDUser)
		if err != nil {
			slog.Error("Failed to resolve OnboardOpenIDUserCommandHandler for OnboardEmailUserCommand.", "err", err)
			return nil, err
		}

		return email_use_cases.NewOnboardEmailUserUseCase(emailUserWriter, emailUserReader, vHashWriter, passwordHasher, onboardOpenIDUser), nil
	})

	if err != nil {
		slog.Error("Failed to load OnboardEmailUserCommand.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (email_in.LoginEmailUserCommand, error) {
		var emailUserReader email_out.EmailUserReader
		err := c.Resolve(&emailUserReader)
		if err != nil {
			slog.Error("Failed to resolve EmailUserReader for LoginEmailUserCommand.", "err", err)
			return nil, err
		}

		var vHashWriter email_out.VHashWriter
		err = c.Resolve(&vHashWriter)
		if err != nil {
			slog.Error("Failed to resolve VHashWriter for LoginEmailUserCommand.", "err", err)
			return nil, err
		}

		var passwordHasher email_out.PasswordHasher
		err = c.Resolve(&passwordHasher)
		if err != nil {
			slog.Error("Failed to resolve PasswordHasher for LoginEmailUserCommand.", "err", err)
			return nil, err
		}

		var createRIDToken iam_in.CreateRIDTokenCommand
		err = c.Resolve(&createRIDToken)
		if err != nil {
			slog.Error("Failed to resolve CreateRIDTokenCommand for LoginEmailUserCommand.", "err", err)
			return nil, err
		}

		return email_use_cases.NewLoginEmailUserUseCase(emailUserReader, vHashWriter, passwordHasher, createRIDToken), nil
	})

	if err != nil {
		slog.Error("Failed to load LoginEmailUserCommand.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (iam_in.VerifyRIDKeyCommand, error) {
		var rIDWriter iam_out.RIDTokenWriter
		err := c.Resolve(&rIDWriter)
		if err != nil {
			slog.Error("Failed to resolve RIDWriter for OnboardRIDCommand.", "err", err)
			return nil, err
		}

		var rIDReader iam_out.RIDTokenReader
		err = c.Resolve(&rIDReader)
		if err != nil {
			slog.Error("Failed to resolve RIDReader for OnboardRIDCommand.", "err", err)
			return nil, err
		}

		return iam_use_cases.NewVerifyRIDUseCase(rIDWriter, rIDReader), nil
	})

	if err != nil {
		slog.Error("Failed to load iam_in.CreateRIDTokenCommand.")
		panic(err)
	}

	// RefreshRIDTokenCommand - handles token refresh for extending session
	err = c.Singleton(func() (iam_in.RefreshRIDTokenCommand, error) {
		var rIDWriter iam_out.RIDTokenWriter
		err := c.Resolve(&rIDWriter)
		if err != nil {
			slog.Error("Failed to resolve RIDWriter for RefreshRIDTokenCommand.", "err", err)
			return nil, err
		}

		var rIDReader iam_out.RIDTokenReader
		err = c.Resolve(&rIDReader)
		if err != nil {
			slog.Error("Failed to resolve RIDReader for RefreshRIDTokenCommand.", "err", err)
			return nil, err
		}

		return iam_use_cases.NewRefreshRIDTokenUseCase(rIDWriter, rIDReader), nil
	})

	if err != nil {
		slog.Error("Failed to load iam_in.RefreshRIDTokenCommand.")
		panic(err)
	}

	// RevokeRIDTokenCommand - handles token revocation (logout)
	err = c.Singleton(func() (iam_in.RevokeRIDTokenCommand, error) {
		var rIDWriter iam_out.RIDTokenWriter
		err := c.Resolve(&rIDWriter)
		if err != nil {
			slog.Error("Failed to resolve RIDWriter for RevokeRIDTokenCommand.", "err", err)
			return nil, err
		}

		var rIDReader iam_out.RIDTokenReader
		err = c.Resolve(&rIDReader)
		if err != nil {
			slog.Error("Failed to resolve RIDReader for RevokeRIDTokenCommand.", "err", err)
			return nil, err
		}

		return iam_use_cases.NewRevokeRIDTokenUseCase(rIDWriter, rIDReader), nil
	})

	if err != nil {
		slog.Error("Failed to load iam_in.RevokeRIDTokenCommand.")
		panic(err)
	}

	err = c.Singleton(func() (iam_in.ProfileReader, error) {
		var profileReader iam_out.ProfileReader
		err := c.Resolve(&profileReader)

		if err != nil {
			slog.Error("Failed to resolve iam_out.ProfileReader for iam_in.ProfileReader.", "err", err)
			return nil, err
		}

		return iam_query_services.NewProfileQueryService(profileReader), nil
	})

	if err != nil {
		slog.Error("Failed to register iam_in.ProfileReader.")
		panic(err)
	}

	err = c.Singleton(func() (iam_in.MembershipReader, error) {
		var membershipReader iam_out.MembershipReader
		err := c.Resolve(&membershipReader)

		if err != nil {
			slog.Error("Failed to resolve iam_out.MembershipReader for iam_in.MembershipReader.", "err", err)
			return nil, err
		}

		var groupReader iam_out.GroupReader
		err = c.Resolve(&groupReader)

		if err != nil {
			slog.Error("Failed to resolve iam_out.GroupReader for iam_in.MembershipReader.", "err", err)
			return nil, err
		}

		return iam_query_services.NewMembershipQueryService(membershipReader, groupReader), nil
	})

	if err != nil {
		slog.Error("Failed to register iam_in.MembershipReader.")
		panic(err)
	}

	return b
}

func (b *ContainerBuilder) WithSquadAPI() *ContainerBuilder {
	c := b.Container

	// repos
	err := c.Singleton(func() (*db.PlayerProfileRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for NamedSingleton PlayerProfileRepository as generic MongoDBRepository.", "err", err)
			return &db.PlayerProfileRepository{}, err
		}

		var config common.Config
		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for db.PlayerProfileRepository.", "err", err)
			return nil, err
		}

		repo := db.NewPlayerProfileRepository(client, config.MongoDB.DBName, squad_entities.PlayerProfile{}, "player_profiles")

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load NamedSingleton PlayerProfileRepository as generic MongoDBRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (*db.PlayerProfileHistoryRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for NamedSingleton PlayerProfileHistoryRepository as generic MongoDBRepository.", "err", err)
			return &db.PlayerProfileHistoryRepository{}, err
		}

		var config common.Config
		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for db.PlayerProfileHistoryRepository.", "err", err)
			return nil, err
		}

		repo := db.NewPlayerProfileHistoryRepository(client, config.MongoDB.DBName, squad_entities.PlayerProfileHistory{}, "player_profile_histories")

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load NamedSingleton PlayerProfileHistoryRepository as generic MongoDBRepository.", "err", err)
		panic(err)
	}

	// // OutboundPorts
	// Squad
	err = c.Singleton(func() (*db.SquadRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for NamedSingleton SquadRepository as generic MongoDBRepository.", "err", err)
			return &db.SquadRepository{}, err
		}

		var config common.Config

		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for db.SquadRepository.", "err", err)
			return nil, err
		}

		repo := db.NewSquadRepository(client, config.MongoDB.DBName, squad_entities.Squad{}, "squads")

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load NamedSingleton SquadRepository as generic MongoDBRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (squad_out.PlayerProfileHistoryWriter, error) {
		var repo *db.PlayerProfileHistoryRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileHistoryRepository for squad_out.PlayerProfileHistoryWriter.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load squad_out.PlayerProfileHistoryWriter.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (squad_out.SquadReader, error) {
		var repo *db.SquadRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve SquadRepository for squad_out.SquadReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load squad_out.SquadReader.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (squad_out.SquadWriter, error) {
		var repo *db.SquadRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve SquadRepository for squad_out.SquadWriter.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load squad_out.SquadWriter.", "err", err)
		panic(err)
	}

	// Squad History Repository
	err = c.Singleton(func() (*db.SquadHistoryRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for SquadHistoryRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for SquadHistoryRepository.", "err", err)
			return nil, err
		}

		return db.NewSquadHistoryRepository(client, config.MongoDB.DBName, squad_entities.SquadHistory{}, "squad_history"), nil
	})

	if err != nil {
		slog.Error("Failed to load SquadHistoryRepository.", "err", err)
		panic(err)
	}

	// SquadHistoryWriter interface
	err = c.Singleton(func() (squad_out.SquadHistoryWriter, error) {
		var repo *db.SquadHistoryRepository
		if err := c.Resolve(&repo); err != nil {
			slog.Error("Failed to resolve SquadHistoryRepository for SquadHistoryWriter.", "err", err)
			return nil, err
		}
		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load squad_out.SquadHistoryWriter.", "err", err)
		panic(err)
	}

	// Player Profile History Repository
	err = c.Singleton(func() (*db.PlayerProfileHistoryRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for PlayerProfileHistoryRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for PlayerProfileHistoryRepository.", "err", err)
			return nil, err
		}

		return db.NewPlayerProfileHistoryRepository(client, config.MongoDB.DBName, squad_entities.PlayerProfileHistory{}, "player_profile_history"), nil
	})

	if err != nil {
		slog.Error("Failed to load PlayerProfileHistoryRepository.", "err", err)
		panic(err)
	}

	// PlayerProfileHistoryWriter interface
	err = c.Singleton(func() (squad_out.PlayerProfileHistoryWriter, error) {
		var repo *db.PlayerProfileHistoryRepository
		if err := c.Resolve(&repo); err != nil {
			slog.Error("Failed to resolve PlayerProfileHistoryRepository for PlayerProfileHistoryWriter.", "err", err)
			return nil, err
		}
		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load squad_out.PlayerProfileHistoryWriter.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (squad_out.PlayerProfileReader, error) {
		var repo *db.PlayerProfileRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileRepository for squad_out.PlayerProfileReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load squad_out.PlayerProfileReader.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (squad_out.PlayerProfileWriter, error) {
		var repo *db.PlayerProfileRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileRepository for squad_out.PlayerProfileWriter.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load squad_out.PlayerProfileWriter.", "err", err)
		panic(err)
	}

	// squad_in.PlayerProfileReader
	err = c.Singleton(func() (squad_in.PlayerProfileReader, error) {
		var repo *db.PlayerProfileRepository
		err := c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileRepository for squad_in.PlayerProfileReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load squad_in.PlayerProfileReader.")
		panic(err)
	}

	err = c.Singleton(func() (squad_in.CreatePlayerProfileCommandHandler, error) {
		var billableOperationHandler billing_in.BillableOperationCommandHandler
		err := c.Resolve(&billableOperationHandler)
		if err != nil {
			slog.Error("Failed to resolve BillableOperationCommandHandler for CreatePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		var playerWriter squad_out.PlayerProfileWriter
		err = c.Resolve(&playerWriter)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileWriter for CreatePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		var playerProfileReader squad_out.PlayerProfileReader
		err = c.Resolve(&playerProfileReader)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileReader for CreatePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		var groupWriter iam_out.GroupWriter
		err = c.Resolve(&groupWriter)
		if err != nil {
			slog.Error("Failed to resolve GroupWriter for CreatePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		var groupReader iam_out.GroupReader
		err = c.Resolve(&groupReader)
		if err != nil {
			slog.Error("Failed to resolve GroupReader for CreatePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		var playerProfileHistoryWriter squad_out.PlayerProfileHistoryWriter
		err = c.Resolve(&playerProfileHistoryWriter)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileHistoryWriter for CreatePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		var mediaWriter media_out.MediaWriter
		err = c.Resolve(&mediaWriter)
		if err != nil {
			slog.Error("Failed to resolve MediaWriter for CreatePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		uc := squad_usecases.NewCreatePlayerProfileUseCase(billableOperationHandler, playerWriter, playerProfileReader, groupWriter, groupReader, playerProfileHistoryWriter, mediaWriter)

		return uc, nil
	})

	if err != nil {
		slog.Error("Failed to load CreatePlayerProfileCommand.", "err", err)
		panic(err)
	}

	// squad_in.UpdatePlayerProfileCommandHandler
	err = c.Singleton(func() (squad_in.UpdatePlayerProfileCommandHandler, error) {
		var billableOperationHandler billing_in.BillableOperationCommandHandler
		err := c.Resolve(&billableOperationHandler)
		if err != nil {
			slog.Error("Failed to resolve BillableOperationCommandHandler for UpdatePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		var playerProfileReader squad_in.PlayerProfileReader
		err = c.Resolve(&playerProfileReader)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileReader for UpdatePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		var playerWriter squad_out.PlayerProfileWriter
		err = c.Resolve(&playerWriter)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileWriter for UpdatePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		var playerProfileHistoryWriter squad_out.PlayerProfileHistoryWriter
		err = c.Resolve(&playerProfileHistoryWriter)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileHistoryWriter for UpdatePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		var mediaWriter media_out.MediaWriter
		err = c.Resolve(&mediaWriter)
		if err != nil {
			slog.Error("Failed to resolve MediaWriter for UpdatePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		uc := squad_usecases.NewUpdatePlayerUseCase(billableOperationHandler, playerProfileReader, playerWriter, playerProfileHistoryWriter, mediaWriter)

		return uc, nil
	})

	if err != nil {
		slog.Error("Failed to load UpdatePlayerProfileCommandHandler.", "err", err)
		panic(err)
	}

	// squad_in.DeletePlayerProfileCommandHandler
	err = c.Singleton(func() (squad_in.DeletePlayerProfileCommandHandler, error) {
		var billableOperationHandler billing_in.BillableOperationCommandHandler
		err := c.Resolve(&billableOperationHandler)
		if err != nil {
			slog.Error("Failed to resolve BillableOperationCommandHandler for DeletePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		var playerProfileReader squad_in.PlayerProfileReader
		err = c.Resolve(&playerProfileReader)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileReader for DeletePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		var playerWriter squad_out.PlayerProfileWriter
		err = c.Resolve(&playerWriter)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileWriter for DeletePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		var playerProfileHistoryWriter squad_out.PlayerProfileHistoryWriter
		err = c.Resolve(&playerProfileHistoryWriter)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileHistoryWriter for DeletePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		uc := squad_usecases.NewDeletePlayerUseCase(billableOperationHandler, playerProfileReader, playerWriter, playerProfileHistoryWriter)

		return uc, nil
	})

	if err != nil {
		slog.Error("Failed to load DeletePlayerProfileCommandHandler.", "err", err)
		panic(err)
	}

	// squad_in.CreateSquadCommandHandler
	err = c.Singleton(func() (squad_in.CreateSquadCommandHandler, error) {
		var squadWriter squad_out.SquadWriter
		err := c.Resolve(&squadWriter)
		if err != nil {
			slog.Error("Failed to resolve SquadWriter for CreateSquadCommandHandler.", "err", err)
			return nil, err
		}

		var squadHistoryWriter squad_out.SquadHistoryWriter
		err = c.Resolve(&squadHistoryWriter)
		if err != nil {
			slog.Error("Failed to resolve SquadHistoryWriter for CreateSquadCommandHandler.", "err", err)
			return nil, err
		}

		var squadReader squad_out.SquadReader
		err = c.Resolve(&squadReader)
		if err != nil {
			slog.Error("Failed to resolve SquadReader for CreateSquadCommandHandler.", "err", err)
			return nil, err
		}

		var groupWriter iam_out.GroupWriter
		err = c.Resolve(&groupWriter)
		if err != nil {
			slog.Error("Failed to resolve GroupWriter for CreateSquadCommandHandler.", "err", err)
			return nil, err
		}

		var groupReader iam_out.GroupReader
		err = c.Resolve(&groupReader)
		if err != nil {
			slog.Error("Failed to resolve GroupReader for CreateSquadCommandHandler.", "err", err)
			return nil, err
		}

		var playerProfileReader squad_in.PlayerProfileReader
		err = c.Resolve(&playerProfileReader)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileReader for CreateSquadCommandHandler.", "err", err)
			return nil, err
		}

		var mediaWriter media_out.MediaWriter
		err = c.Resolve(&mediaWriter)
		if err != nil {
			slog.Error("Failed to resolve MediaWriter for CreateSquadCommandHandler.", "err", err)
			return nil, err
		}

		var billableOperationHandler billing_in.BillableOperationCommandHandler
		err = c.Resolve(&billableOperationHandler)
		if err != nil {
			slog.Error("Failed to resolve BillableOperationCommandHandler for CreateSquadCommandHandler.", "err", err)
			return nil, err
		}

		cmdHandler := squad_usecases.NewCreateSquadUseCase(squadWriter, squadHistoryWriter, squadReader, groupWriter, groupReader, playerProfileReader, mediaWriter, billableOperationHandler)

		return cmdHandler, nil
	})

	if err != nil {
		slog.Error("Failed to load CreatePlayerProfileCommandHandler.")
		panic(err)

	}

	// InboundPorts
	err = c.Singleton(func() (squad_in.SquadReader, error) {
		var squadReader squad_out.SquadReader
		err := c.Resolve(&squadReader)
		if err != nil {
			slog.Error("Failed to resolve SquadSearchableReader for SquadQueryService.", "err", err)
			return nil, err
		}

		return squad_services.NewSquadQueryService(squadReader), nil
	})

	if err != nil {
		slog.Error("Failed to load SquadSearchableReader.")
		panic(err)
	}

	// NOTE: CreateSquadCommandHandler and CreatePlayerProfileCommandHandler are already registered above

	return b

}

func (b *ContainerBuilder) WithKafkaConsumer() *ContainerBuilder {
	// c := b.Container

	// err := c.Singleton(func() (out.KafkaConsumer, error) {
	// 	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
	// 		"bootstrap.servers":        "localhost:9092",
	// 		"acks":                     1,
	// 		"retries":                  0,
	// 		"retry.backoff.ms":         100,
	// 		"socket.timeout.ms":        6000,
	// 		"reconnect.backoff.max.ms": 3000,
	// 	})
	// 	if err != nil {
	// 		slog.Error(err.Error())
	// 		panic(err)
	// 	}

	// 	var config common.Config

	// 	err := c.Resolve(&config)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	return kafka.NewKafkaConsumer(config.Kafka), nil
	// })

	// if err != nil {
	// 	slog.Error("Failed to load KafkaConsumer.")
	// 	panic(err)
	// }

	return b
}

func (b *ContainerBuilder) WithKafka() *ContainerBuilder {
	// Dummy method for RPC API - Kafka is disabled for local development
	return b
}

func (b *ContainerBuilder) WithEventPublisher() *ContainerBuilder {
	c := b.Container

	// Register Kafka client singleton
	err := c.Singleton(func() (*kafka.Client, error) {
		kafkaConfig := kafka.NewConfigFromEnv()
		if kafkaConfig.BootstrapServers == "" || kafkaConfig.BootstrapServers == "localhost:9092" {
			// Check if Kafka is actually configured
			if os.Getenv("KAFKA_BOOTSTRAP_SERVERS") == "" {
				slog.Info("Kafka not configured, using dummy client for local development")
				return nil, nil
			}
		}
		
		client, err := kafka.NewClient(kafkaConfig)
		if err != nil {
			slog.Warn("Failed to create Kafka client, continuing without it", "err", err)
			return nil, nil
		}
		slog.Info("Kafka client created", "brokers", kafkaConfig.BootstrapServers)
		return client, nil
	})

	if err != nil {
		slog.Warn("Failed to register *kafka.Client.", "err", err)
	}

	// Kafka Event Publisher
	err = c.Singleton(func() (*kafka.EventPublisher, error) {
		var kafkaClient *kafka.Client
		_ = c.Resolve(&kafkaClient) // May be nil for local dev

		// EventPublisher handles nil client gracefully for local development
		return kafka.NewEventPublisher(kafkaClient), nil
	})

	if err != nil {
		slog.Error("Failed to load *kafka.EventPublisher.", "err", err)
		panic(err)
	}

	// Register replay event publisher adapter
	err = c.Singleton(func() (replay_out.ReplayEventPublisher, error) {
		var eventPublisher *kafka.EventPublisher
		err := c.Resolve(&eventPublisher)
		if err != nil {
			slog.Warn("Failed to resolve EventPublisher for ReplayEventPublisher adapter - continuing without it", "err", err)
			return nil, nil
		}
		return kafka.NewReplayEventPublisherAdapter(eventPublisher), nil
	})

	if err != nil {
		slog.Warn("Failed to load replay_out.ReplayEventPublisher.", "err", err)
		// Don't panic - event publishing is optional for local dev
	}

	return b
}

func InjectMongoDB(c container.Container) error {
	err := c.Singleton(func() (*mongo.Client, error) {
		var config common.Config

		err := c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for mongo.Client.", "err", err)
			return nil, err
		}

		mongoOptions := options.Client().ApplyURI(config.MongoDB.URI).SetRegistry(mongodb.MongoRegistry).SetMaxPoolSize(100)

		client, err := mongo.Connect(context.TODO(), mongoOptions)

		if err != nil {
			slog.Error("Failed to connect to MongoDB.", "err", err)
			return nil, err
		}

		return client, nil
	})

	if err != nil {
		slog.Error("Failed to load mongo.Client.")
		return err
	}

	// events repo
	err = c.Singleton(func() (*db.EventsRepository, error) {
		var config common.Config

		err := c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for db.EventsRepository.", "err", err)
			return nil, err
		}

		var client *mongo.Client
		err = c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for db.EventsRepository as generic MongoDBRepository.", "err", err)
			return &db.EventsRepository{}, err
		}

		repo := db.NewEventsRepository(client, config.MongoDB.DBName, &replay_entity.GameEvent{}, "game_events")

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load NamedSingleton EventsRepository as generic MongoDBRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (replay_out.EventsByGameReader, error) {
		var repo *db.EventsRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve EventsRepository for replay_out.EventsByGameReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load replay_out.EventsByGameReader.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (replay_out.GameEventReader, error) {
		var repo *db.EventsRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve EventsRepository for replay_out.GameEventReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load replay_out.GameEventReader.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (replay_out.GameEventWriter, error) {
		var repo *db.EventsRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve EventsRepository for replay_out.GameEventWriter.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to resolve replay_out.GameEventWriter.", "err", err)
		panic(err)
	}

	// replay

	err = c.Singleton(func() (*db.ReplayFileMetadataRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for NamedSingleton ReplayFileMetadataRepository as generic MongoDBRepository.", "err", err)
			return &db.ReplayFileMetadataRepository{}, err
		}

		var config common.Config

		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for db.ReplayFileMetadataRepository.", "err", err)
			return nil, err
		}

		repo := db.NewReplayFileMetadataRepository(client, config.MongoDB.DBName, replay_entity.ReplayFile{}, "replay_files")

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load NamedSingleton ReplayFileMetadataRepository as generic MongoDBRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (replay_out.ReplayFileMetadataReader, error) {
		var repo *db.ReplayFileMetadataRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve ReplayFileMetadataRepository for replay_out.ReplayFileMetadataReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load replay_out.ReplayFileMetadataReader.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (replay_out.ReplayFileMetadataWriter, error) {
		var repo *db.ReplayFileMetadataRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve SteamUserRepository for replay_out.ReplayFileMetadataWriter.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load replay_out.ReplayFileMetadataWriter.", "err", err)
		panic(err)
	}

	// MATCH METADATA
	err = c.Singleton(func() (*db.MatchMetadataRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for NamedSingleton MatchMetadataRepository as generic MongoDBRepository.", "err", err)
			return &db.MatchMetadataRepository{}, err
		}

		var config common.Config

		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for db.MatchMetadataRepository.", "err", err)
			return nil, err
		}

		repo := db.NewMatchMetadataRepository(client, config.MongoDB.DBName, replay_entity.Match{}, "match_metadata")

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load NamedSingleton ReplayFileMetadataRepository as generic MongoDBRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (replay_out.MatchMetadataReader, error) {
		var repo *db.MatchMetadataRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve MatchMetadataRepository for replay_out.MatchMetadataReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load replay_out.MatchMetadataReader.", "err", err)
		panic(err)
	}

	// Player Metadata Repository
	err = c.Singleton(func() (*db.PlayerMetadataRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for PlayerRepository as generic MongoDBRepository.", "err", err)
			return &db.PlayerMetadataRepository{}, err
		}

		var config common.Config

		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for db.PlayerRepository.", "err", err)
			return nil, err
		}

		repo := db.NewPlayerMetadataRepository(client, config.MongoDB.DBName, replay_entity.PlayerMetadata{}, "player_metadata")

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load PlayerRepository as generic MongoDBRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (replay_out.PlayerMetadataReader, error) {
		var repo *db.PlayerMetadataRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve PlayerRepository for replay_out.PlayerMetadataReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load replay_out.PlayerMetadataReader.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (replay_out.PlayerMetadataWriter, error) {
		var repo *db.PlayerMetadataRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve PlayerRepository for replay_out.PlayerMetadataWriter.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load replay_out.PlayerMetadataWriter.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (replay_out.MatchMetadataWriter, error) {
		var repo *db.MatchMetadataRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve PlayerRepository for replay_out.MatchMetadataWriter.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load replay_out.MatchMetadataWriter.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (replay_in.PlayerMetadataReader, error) {
		var repo *db.PlayerMetadataRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve PlayerRepository for replay_in.PlayerProfileReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load replay_in.PlayerProfileReader.", "err", err)
		panic(err)
	}

	// SHARE TOKEN REPOSITORY
	err = c.Singleton(func() (*db.ShareTokenRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for ShareTokenRepository.", "err", err)
			return nil, err
		}

		var config common.Config
		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for ShareTokenRepository.", "err", err)
			return nil, err
		}

		repo := db.NewShareTokenRepository(client, config.MongoDB.DBName, replay_entity.ShareToken{}, "share_tokens")
		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load ShareTokenRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (replay_out.ShareTokenReader, error) {
		var repo *db.ShareTokenRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve ShareTokenRepository for replay_out.ShareTokenReader.", "err", err)
			return nil, err
		}
		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load replay_out.ShareTokenReader.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (replay_out.ShareTokenWriter, error) {
		var repo *db.ShareTokenRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve ShareTokenRepository for replay_out.ShareTokenWriter.", "err", err)
			return nil, err
		}
		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load replay_out.ShareTokenWriter.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (replay_in.ShareTokenReader, error) {
		var shareTokenReader replay_out.ShareTokenReader
		err := c.Resolve(&shareTokenReader)
		if err != nil {
			slog.Error("Failed to resolve ShareTokenReader for replay_in.ShareTokenReader.", "err", err)
			return nil, err
		}
		return metadata.NewShareTokenQueryService(shareTokenReader), nil
	})

	if err != nil {
		slog.Error("Failed to load replay_in.ShareTokenReader.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (replay_in.ShareTokenCommand, error) {
		var shareTokenWriter replay_out.ShareTokenWriter
		err := c.Resolve(&shareTokenWriter)
		if err != nil {
			slog.Error("Failed to resolve ShareTokenWriter for replay_in.ShareTokenCommand.", "err", err)
			return nil, err
		}
		return metadata.NewShareTokenCommandService(shareTokenWriter), nil
	})

	if err != nil {
		slog.Error("Failed to load replay_in.ShareTokenCommand.", "err", err)
		panic(err)
	}

	// err = c.Singleton(func() (replay_out.BadgeReader, error) {
	// 	var repo *db.BadgeRepository
	// 	err = c.Resolve(&repo)
	// 	if err != nil {
	// 		slog.Error("Failed to resolve BadgeRepository for replay_out.BadgeReader.", "err", err)
	// 		return nil, err
	// 	}

	// 	return repo, nil
	// })

	// if err != nil {
	// 	slog.Error("Failed to load replay_out.BadgeReader.", "err", err)
	// 	panic(err)
	// }

	err = c.Singleton(func() (replay_out.ReplayFileContentWriter, error) {
		var client *mongo.Client

		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for ReplayFileContentWriter.", "err", err)
			return nil, err
		}

		var config common.Config

		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for replay_out.ReplayFileContentWriter.", "err", err)
			return nil, err
		}

		// return s3.NewS3Adapter(config.S3), nil
		// return local_files.NewLocalFileAdapter(), nil
		return db.NewReplayFileContentRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load S3Adapter.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (replay_out.ReplayFileContentReader, error) {
		var config common.Config

		err := c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for ReplayFileContentReader.", "err", err)
			return nil, err
		}

		// return blob.NewS3Adapter(config.S3), nil
		// return local_files.NewLocalFileAdapter(), nil

		var client *mongo.Client

		err = c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for ReplayFileContentReader.", "err", err)
			return nil, err
		}

		return db.NewReplayFileContentRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load S3Adapter.")
		panic(err)
	}

	err = c.Singleton(func() replay_out.ReplayParser {
		return cs_app.NewCS2ReplayAdapter()
	})

	if err != nil {
		slog.Error("Failed to load CS2ReplayAdapter.", "err", err)
		panic(err)
	}

	// steam repo
	err = c.Singleton(func() (*db.SteamUserRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for NamedSingleton SteamUserRepository as generic MongoDBRepository.", "err", err)
			return nil, err
		}

		var config common.Config

		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for db.SteamUserRepository.", "err", err)
			return nil, err
		}

		repo := db.NewSteamUserMongoDBRepository(client, config.MongoDB.DBName, steam_entity.SteamUser{}, "steam_users")

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load NamedSingleton SteamUserRepository as generic MongoDBRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (steam_out.SteamUserWriter, error) {
		var repo *db.SteamUserRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve SteamUserRepository for steam_out.SteamUserWriter.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load SteamUserWriter.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (steam_out.SteamUserReader, error) {
		var repo *db.SteamUserRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve SteamUserRepository for steam_out.SteamUserReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load SteamUserReader.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (steam_out.VHashWriter, error) {
		var config common.Config

		err := c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for steam_out.VHashWriter.", "err", err)
			return nil, err
		}

		return encryption.NewSHA256VHasherAdapter(config.Auth.SteamConfig.VHashSource), nil
	})

	if err != nil {
		slog.Error("Failed to load VHashWriter.", "err", err)
		panic(err)
	}

	// end-steam

	// GOOGLE repo
	err = c.Singleton(func() (*db.GoogleUserRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for NamedSingleton GoogleUserRepository as generic MongoDBRepository.", "err", err)
			return nil, err
		}

		var config common.Config

		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for db.GoogleUserRepository.", "err", err)
			return nil, err
		}

		repo := db.NewGoogleUserMongoDBRepository(client, config.MongoDB.DBName, google_entities.GoogleUser{}, "google_users")

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load NamedSingleton GoogleUserRepository as generic MongoDBRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (google_out.GoogleUserWriter, error) {
		var repo *db.GoogleUserRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve GoogleUserRepository for google_out.GoogleUserWriter.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load GoogleUserWriter.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (google_out.GoogleUserReader, error) {
		var repo *db.GoogleUserRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve GoogleUserRepository for google_out.GoogleUserReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load GoogleUserReader.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (google_out.VHashWriter, error) {
		var config common.Config

		err := c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for google_out.VHashWriter.", "err", err)
			return nil, err
		}

		return encryption.NewSHA256VHasherAdapter(config.Auth.SteamConfig.VHashSource), nil
	})

	if err != nil {
		slog.Error("Failed to load VHashWriter.", "err", err)
		panic(err)
	}

	// end-google

	// EMAIL repo
	err = c.Singleton(func() (*db.EmailUserRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for NamedSingleton EmailUserRepository as generic MongoDBRepository.", "err", err)
			return nil, err
		}

		var config common.Config

		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for db.EmailUserRepository.", "err", err)
			return nil, err
		}

		repo := db.NewEmailUserMongoDBRepository(client, config.MongoDB.DBName, email_entities.EmailUser{}, "email_users")

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load NamedSingleton EmailUserRepository as generic MongoDBRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (email_out.EmailUserWriter, error) {
		var repo *db.EmailUserRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve EmailUserRepository for email_out.EmailUserWriter.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load EmailUserWriter.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (email_out.EmailUserReader, error) {
		var repo *db.EmailUserRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve EmailUserRepository for email_out.EmailUserReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load EmailUserReader.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (email_out.VHashWriter, error) {
		var config common.Config

		err := c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for email_out.VHashWriter.", "err", err)
			return nil, err
		}

		return encryption.NewSHA256VHasherAdapter(config.Auth.SteamConfig.VHashSource), nil
	})

	if err != nil {
		slog.Error("Failed to load email VHashWriter.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (email_out.PasswordHasher, error) {
		return encryption.NewBcryptPasswordHasherAdapter(10), nil
	})

	if err != nil {
		slog.Error("Failed to load PasswordHasher.", "err", err)
		panic(err)
	}

	// Email verification infrastructure
	err = c.Singleton(func() (auth_out.EmailVerificationRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for EmailVerificationRepository.", "err", err)
			return nil, err
		}

		var config common.Config
		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for EmailVerificationRepository.", "err", err)
			return nil, err
		}

		return db.NewEmailVerificationMongoDBRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load EmailVerificationRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (auth_out.EmailSender, error) {
		return email_adapter.NewNoopEmailSender(true), nil
	})

	if err != nil {
		slog.Error("Failed to load EmailSender.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (auth_out.EmailUserVerifier, error) {
		var repo *db.EmailUserRepository
		err := c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve EmailUserRepository for auth_out.EmailUserVerifier.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load EmailUserVerifier.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (auth_in.EmailVerificationCommand, error) {
		var verificationRepo auth_out.EmailVerificationRepository
		err := c.Resolve(&verificationRepo)
		if err != nil {
			slog.Error("Failed to resolve EmailVerificationRepository for EmailVerificationCommand.", "err", err)
			return nil, err
		}

		var emailSender auth_out.EmailSender
		err = c.Resolve(&emailSender)
		if err != nil {
			slog.Error("Failed to resolve EmailSender for EmailVerificationCommand.", "err", err)
			return nil, err
		}

		var emailUserVerifier auth_out.EmailUserVerifier
		err = c.Resolve(&emailUserVerifier)
		if err != nil {
			slog.Error("Failed to resolve EmailUserVerifier for EmailVerificationCommand.", "err", err)
			return nil, err
		}

		return auth_services.NewEmailVerificationService(verificationRepo, emailSender, emailUserVerifier), nil
	})

	if err != nil {
		slog.Error("Failed to load EmailVerificationCommand.", "err", err)
		panic(err)
	}

	// end-email

	// rid
	err = c.Singleton(func() (*db.RIDTokenRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for NamedSingleton RIDTokenRepository as generic MongoDBRepository.", "err", err)
			return &db.RIDTokenRepository{}, err
		}

		var config common.Config

		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for db.SteamUserRepository.", "err", err)
			return nil, err
		}

		repo := db.NewRIDTokenRepository(client, config.MongoDB.DBName, iam_entities.RIDToken{}, "rid")

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load NamedSingleton RIDTokenRepository as generic MongoDBRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (iam_out.RIDTokenWriter, error) {
		var repo *db.RIDTokenRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve RIDTokenRepository for iam_out.RIDTokenWriter.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load iam_out.RIDTokenWriter.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (iam_out.RIDTokenReader, error) {
		var repo *db.RIDTokenRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve RIDTokenRepository for iam_out.RIDTokenReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load iam_out.RIDTokenReader.", "err", err)
		panic(err)
	}

	// -----

	// User
	err = c.Singleton(func() (*db.UserRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for NamedSingleton UserRepository as generic MongoDBRepository.", "err", err)
			return &db.UserRepository{}, err
		}

		var config common.Config

		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for db.UserRepository.", "err", err)
			return nil, err
		}

		repo := db.NewUserRepository(client, config.MongoDB.DBName, &iam_entities.User{}, "users")

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load NamedSingleton UserRepository as generic MongoDBRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (iam_out.UserReader, error) {
		var repo *db.UserRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve UserRepository for iam_out.UserReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load iam_out.UserReader.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (iam_out.UserWriter, error) {
		var repo *db.UserRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve UserRepository for iam_out.UserWriter.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load iam_out.UserWriter.", "err", err)
		panic(err)
	}

	// -----

	// Group
	err = c.Singleton(func() (*db.GroupRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for NamedSingleton GroupRepository as generic MongoDBRepository.", "err", err)
			return &db.GroupRepository{}, err
		}

		var config common.Config

		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for db.GroupRepository.", "err", err)
			return nil, err
		}

		repo := db.NewGroupRepository(client, config.MongoDB.DBName, &iam_entities.Group{}, "groups")

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load NamedSingleton GroupRepository as generic MongoDBRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (iam_out.GroupReader, error) {
		var repo *db.GroupRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve GroupRepository for iam_out.GroupReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load iam_out.GroupReader.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (iam_out.GroupWriter, error) {
		var repo *db.GroupRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve GroupRepository for iam_out.GroupWriter.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load iam_out.GroupWriter.", "err", err)
		panic(err)
	}

	// -----

	// Profile
	err = c.Singleton(func() (*db.ProfileRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for NamedSingleton ProfileRepository as generic MongoDBRepository.", "err", err)
			return &db.ProfileRepository{}, err
		}

		var config common.Config

		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for db.ProfileRepository.", "err", err)
			return nil, err
		}

		repo := db.NewProfileRepository(client, config.MongoDB.DBName, iam_entities.Profile{}, "profiles")

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load NamedSingleton ProfileRepository as generic MongoDBRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (iam_out.ProfileReader, error) {
		var repo *db.ProfileRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve ProfileRepository for iam_out.ProfileReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load iam_out.ProfileReader.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (iam_out.ProfileWriter, error) {
		var repo *db.ProfileRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve ProfileRepository for iam_out.ProfileWriter.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load iam_out.ProfileWriter.", "err", err)
		panic(err)
	}

	// -----

	// Membership
	err = c.Singleton(func() (*db.MembershipRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for NamedSingleton MembershipRepository as generic MongoDBRepository.", "err", err)
			return &db.MembershipRepository{}, err
		}

		var config common.Config

		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for db.MembershipRepository.", "err", err)
			return nil, err
		}

		repo := db.NewMembershipRepository(client, config.MongoDB.DBName, &iam_entities.Membership{}, "memberships")

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load NamedSingleton MembershipRepository as generic MongoDBRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (iam_out.MembershipReader, error) {
		var repo *db.MembershipRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve MembershipRepository for iam_out.MembershipReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load iam_out.MembershipReader.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (iam_out.MembershipWriter, error) {
		var repo *db.MembershipRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve MembershipRepository for iam_out.MembershipWriter.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load iam_out.MembershipWriter.", "err", err)
		panic(err)
	}

	// ----- Matchmaking & Wallet (Prize Pool System) -----

	// WebSocket Hub (Singleton)
	err = c.Singleton(func() *websocket.WebSocketHub {
		hub := websocket.NewWebSocketHub()
		return hub
	})

	if err != nil {
		slog.Error("Failed to load *websocket.WebSocketHub.", "err", err)
		panic(err)
	}

	// Lobby Repository
	err = c.Singleton(func() (matchmaking_out.LobbyRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve *mongo.Client for MongoLobbyRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for MongoLobbyRepository.", "err", err)
			return nil, err
		}

		return db.NewMongoLobbyRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load matchmaking_out.LobbyRepository.", "err", err)
		panic(err)
	}

	// Prize Pool Repository
	err = c.Singleton(func() (matchmaking_out.PrizePoolRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve *mongo.Client for MongoPrizePoolRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for MongoPrizePoolRepository.", "err", err)
			return nil, err
		}

		return db.NewMongoPrizePoolRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load matchmaking_out.PrizePoolRepository.", "err", err)
		panic(err)
	}

	// Tournament Repository
	err = c.Singleton(func() (tournament_out.TournamentRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve *mongo.Client for MongoTournamentRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for MongoTournamentRepository.", "err", err)
			return nil, err
		}

		return db.NewMongoTournamentRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load tournament_out.TournamentRepository.", "err", err)
		panic(err)
	}

	// Wallet Repository
	err = c.Singleton(func() (wallet_out.WalletRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve *mongo.Client for MongoWalletRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for MongoWalletRepository.", "err", err)
			return nil, err
		}

		return db.NewMongoWalletRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load wallet_out.WalletRepository.", "err", err)
		panic(err)
	}

	// Ledger Repository
	err = c.Singleton(func() (wallet_out.LedgerRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for LedgerRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for LedgerRepository.", "err", err)
			return nil, err
		}

		return db.NewLedgerRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load wallet_out.LedgerRepository.", "err", err)
		panic(err)
	}

	// Idempotency Repository
	err = c.Singleton(func() (wallet_out.IdempotencyRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for IdempotencyRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for IdempotencyRepository.", "err", err)
			return nil, err
		}

		database := client.Database(config.MongoDB.DBName)
		return db.NewIdempotencyRepository(database), nil
	})

	if err != nil {
		slog.Error("Failed to load wallet_out.IdempotencyRepository.", "err", err)
		panic(err)
	}

	// No-op AuditTrailService for basic functionality (must be registered before LedgerService)
	err = c.Singleton(func() (billing_in.AuditTrailCommand, error) {
		return billing_services.NewNoOpAuditTrailService(), nil
	})

	if err != nil {
		slog.Error("Failed to load billing_in.AuditTrailCommand (no-op).", "err", err)
		panic(err)
	}

	// Ledger Service (must be registered before TransactionCoordinator)
	// Uses real MongoDB LedgerServiceRepository for financial-grade persistence
	err = c.Singleton(func() (*wallet_services.LedgerService, error) {
		var auditTrail billing_in.AuditTrailCommand
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&auditTrail); err != nil {
			slog.Error("Failed to resolve billing_in.AuditTrailCommand for LedgerService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for LedgerServiceRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for LedgerServiceRepository.", "err", err)
			return nil, err
		}

		ledgerRepo := db.NewLedgerServiceRepository(client, config.MongoDB.DBName)
		slog.Info("LedgerService using real MongoDB LedgerServiceRepository")
		return wallet_services.NewLedgerService(ledgerRepo, auditTrail), nil
	})

	if err != nil {
		slog.Error("Failed to load *wallet_services.LedgerService.", "err", err)
		panic(err)
	}

	// Transaction Coordinator
	err = c.Singleton(func() (*wallet_services.TransactionCoordinator, error) {
		var walletRepo wallet_out.WalletRepository
		var ledgerService *wallet_services.LedgerService

		if err := c.Resolve(&walletRepo); err != nil {
			slog.Error("Failed to resolve wallet_out.WalletRepository for TransactionCoordinator.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&ledgerService); err != nil {
			slog.Error("Failed to resolve *wallet_services.LedgerService for TransactionCoordinator.", "err", err)
			return nil, err
		}

		return wallet_services.NewTransactionCoordinator(walletRepo, ledgerService), nil
	})

	if err != nil {
		slog.Error("Failed to load *wallet_services.TransactionCoordinator.", "err", err)
		panic(err)
	}

	// Wallet Query Service (must be registered before WalletQuery)
	err = c.Singleton(func() (*wallet_services.WalletQueryService, error) {
		var walletRepo wallet_out.WalletRepository

		if err := c.Resolve(&walletRepo); err != nil {
			slog.Error("Failed to resolve wallet_out.WalletRepository for WalletQueryService.", "err", err)
			return nil, err
		}

		return wallet_services.NewWalletQueryService(walletRepo), nil
	})

	if err != nil {
		slog.Error("Failed to load *wallet_services.WalletQueryService.", "err", err)
		panic(err)
	}

	// GetWalletBalanceUseCase
	err = c.Singleton(func() (*wallet_usecases.GetWalletBalanceUseCase, error) {
		var walletQuerySvc *wallet_services.WalletQueryService

		if err := c.Resolve(&walletQuerySvc); err != nil {
			slog.Error("Failed to resolve WalletQueryService for GetWalletBalanceUseCase.", "err", err)
			return nil, err
		}

		return wallet_usecases.NewGetWalletBalanceUseCase(walletQuerySvc), nil
	})

	if err != nil {
		slog.Error("Failed to load GetWalletBalanceUseCase.", "err", err)
		panic(err)
	}

	// GetTransactionsUseCase
	err = c.Singleton(func() (*wallet_usecases.GetTransactionsUseCase, error) {
		var walletRepo wallet_out.WalletRepository
		var walletQuerySvc *wallet_services.WalletQueryService
		var ledgerRepo wallet_out.LedgerRepository

		if err := c.Resolve(&walletRepo); err != nil {
			slog.Error("Failed to resolve wallet_out.WalletRepository for GetTransactionsUseCase.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&walletQuerySvc); err != nil {
			slog.Error("Failed to resolve WalletQueryService for GetTransactionsUseCase.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&ledgerRepo); err != nil {
			slog.Error("Failed to resolve wallet_out.LedgerRepository for GetTransactionsUseCase.", "err", err)
			return nil, err
		}

		return wallet_usecases.NewGetTransactionsUseCase(walletRepo, walletQuerySvc, ledgerRepo), nil
	})

	if err != nil {
		slog.Error("Failed to load GetTransactionsUseCase.", "err", err)
		panic(err)
	}

	// Wallet Command Service (uses real WalletService with Kafka event publishing)
	err = c.Singleton(func() (wallet_in.WalletCommand, error) {
		var walletRepo wallet_out.WalletRepository
		var walletQuerySvc *wallet_services.WalletQueryService
		var coordinator *wallet_services.TransactionCoordinator

		if err := c.Resolve(&walletRepo); err != nil {
			slog.Warn("Failed to resolve WalletRepository for WalletCommand, using NoOp", "err", err)
			return &NoOpWalletCommand{}, nil
		}

		if err := c.Resolve(&walletQuerySvc); err != nil {
			slog.Warn("Failed to resolve WalletQueryService for WalletCommand, using NoOp", "err", err)
			return &NoOpWalletCommand{}, nil
		}

		if err := c.Resolve(&coordinator); err != nil {
			slog.Warn("Failed to resolve TransactionCoordinator for WalletCommand, using NoOp", "err", err)
			return &NoOpWalletCommand{}, nil
		}

		// Try to wire Kafka event publisher for financial-grade event streaming
		var eventPublisher *kafka.EventPublisher
		if err := c.Resolve(&eventPublisher); err != nil {
			slog.Info("EventPublisher not available for WalletService — events will not be published", "err", err)
		}

		walletEventAdapter := kafka.NewWalletEventPublisherAdapter(eventPublisher)

		if walletEventAdapter != nil {
			slog.Info("WalletCommand using real WalletService with Kafka event publishing")
			return wallet_services.NewWalletServiceWithEvents(walletRepo, walletQuerySvc, coordinator, walletEventAdapter), nil
		}

		slog.Info("WalletCommand using real WalletService without event publishing")
		return wallet_services.NewWalletService(walletRepo, walletQuerySvc, coordinator), nil
	})

	if err != nil {
		slog.Error("Failed to load wallet_in.WalletCommand.", "err", err)
		panic(err)
	}

	// Wallet Query Service
	err = c.Singleton(func() (wallet_in.WalletQuery, error) {
		var getBalanceUseCase *wallet_usecases.GetWalletBalanceUseCase
		var getTransactionsUseCase *wallet_usecases.GetTransactionsUseCase

		if err := c.Resolve(&getBalanceUseCase); err != nil {
			slog.Error("Failed to resolve GetWalletBalanceUseCase for WalletQueryService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&getTransactionsUseCase); err != nil {
			slog.Error("Failed to resolve GetTransactionsUseCase for WalletQueryService.", "err", err)
			return nil, err
		}

		return wallet_usecases.NewWalletQueryService(getBalanceUseCase, getTransactionsUseCase), nil
	})

	if err != nil {
		slog.Error("Failed to load wallet_in.WalletQuery.", "err", err)
		panic(err)
	}

	// ═══════════════════════════════════════════════════════════════════════════════
	// TEAM VAULT (Multisig) SERVICES
	// ═══════════════════════════════════════════════════════════════════════════════

	// Team Vault Repository
	err = c.Singleton(func() (wallet_out.TeamVaultRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for TeamVaultRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for TeamVaultRepository.", "err", err)
			return nil, err
		}

		return db.NewMongoTeamVaultRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load wallet_out.TeamVaultRepository.", "err", err)
		panic(err)
	}

	// Vault Proposal Repository
	err = c.Singleton(func() (wallet_out.VaultProposalRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for VaultProposalRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for VaultProposalRepository.", "err", err)
			return nil, err
		}

		return db.NewMongoVaultProposalRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load wallet_out.VaultProposalRepository.", "err", err)
		panic(err)
	}

	// Vault Activity Repository
	err = c.Singleton(func() (wallet_out.VaultActivityRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for VaultActivityRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for VaultActivityRepository.", "err", err)
			return nil, err
		}

		return db.NewMongoVaultActivityRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load wallet_out.VaultActivityRepository.", "err", err)
		panic(err)
	}

	// Inventory Item Repository
	err = c.Singleton(func() (wallet_out.InventoryItemRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for InventoryItemRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for InventoryItemRepository.", "err", err)
			return nil, err
		}

		return db.NewMongoInventoryItemRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load wallet_out.InventoryItemRepository.", "err", err)
		panic(err)
	}

	// SquadRepository (needed by VaultService - may also be registered in WithSquadAPI)
	_ = c.Singleton(func() (*db.SquadRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for SquadRepository (InjectMongoDB).", "err", err)
			return &db.SquadRepository{}, err
		}

		var config common.Config
		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for SquadRepository (InjectMongoDB).", "err", err)
			return nil, err
		}

		repo := db.NewSquadRepository(client, config.MongoDB.DBName, squad_entities.Squad{}, "squads")
		return repo, nil
	})

	// squad_out.SquadReader (needed by VaultService - may also be registered in WithSquadAPI)
	_ = c.Singleton(func() (squad_out.SquadReader, error) {
		var repo *db.SquadRepository
		err := c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve SquadRepository for squad_out.SquadReader (InjectMongoDB).", "err", err)
			return nil, err
		}
		return repo, nil
	})

	// Team Vault Command Service (implements both TeamVaultCommand and TeamVaultQuery)
	err = c.Singleton(func() (*wallet_usecases.VaultService, error) {
		var vaultRepo wallet_out.TeamVaultRepository
		var proposalRepo wallet_out.VaultProposalRepository
		var activityRepo wallet_out.VaultActivityRepository
		var inventoryRepo wallet_out.InventoryItemRepository
		var walletRepo wallet_out.WalletRepository
		var squadReader squad_out.SquadReader

		if err := c.Resolve(&vaultRepo); err != nil {
			slog.Error("Failed to resolve TeamVaultRepository for VaultService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&proposalRepo); err != nil {
			slog.Error("Failed to resolve VaultProposalRepository for VaultService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&activityRepo); err != nil {
			slog.Error("Failed to resolve VaultActivityRepository for VaultService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&inventoryRepo); err != nil {
			slog.Error("Failed to resolve InventoryItemRepository for VaultService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&walletRepo); err != nil {
			slog.Error("Failed to resolve WalletRepository for VaultService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&squadReader); err != nil {
			slog.Error("Failed to resolve SquadReader for VaultService.", "err", err)
			return nil, err
		}

		return wallet_usecases.NewVaultService(vaultRepo, proposalRepo, activityRepo, inventoryRepo, walletRepo, squadReader), nil
	})

	if err != nil {
		slog.Error("Failed to load *wallet_usecases.VaultService.", "err", err)
		panic(err)
	}

	// TeamVaultCommand interface binding
	err = c.Singleton(func() (wallet_in.TeamVaultCommand, error) {
		var vaultService *wallet_usecases.VaultService

		if err := c.Resolve(&vaultService); err != nil {
			slog.Error("Failed to resolve VaultService for TeamVaultCommand.", "err", err)
			return nil, err
		}

		return vaultService, nil
	})

	if err != nil {
		slog.Error("Failed to load wallet_in.TeamVaultCommand.", "err", err)
		panic(err)
	}

	// TeamVaultQuery interface binding
	err = c.Singleton(func() (wallet_in.TeamVaultQuery, error) {
		var vaultService *wallet_usecases.VaultService

		if err := c.Resolve(&vaultService); err != nil {
			slog.Error("Failed to resolve VaultService for TeamVaultQuery.", "err", err)
			return nil, err
		}

		return vaultService, nil
	})

	if err != nil {
		slog.Error("Failed to load wallet_in.TeamVaultQuery.", "err", err)
		panic(err)
	}

	// Media Writer (no-op for now)
	err = c.Singleton(func() (media_out.MediaWriter, error) {
		return media_adapter.NewNoopMediaAdapter(), nil
	})

	if err != nil {
		slog.Error("Failed to load media_out.MediaWriter.", "err", err)
		panic(err)
	}

	// Billing - BillableEntry Repository
	err = c.Singleton(func() (*db.BillableEntryRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for BillableEntryRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for BillableEntryRepository.", "err", err)
			return nil, err
		}

		return db.NewBillableEntryRepository(client, config.MongoDB.DBName, billing_entities.BillableEntry{}, "billable_entries"), nil
	})

	if err != nil {
		slog.Error("Failed to load BillableEntryRepository.", "err", err)
		panic(err)
	}

	// BillableEntryWriter interface
	err = c.Singleton(func() (billing_out.BillableEntryWriter, error) {
		var repo *db.BillableEntryRepository
		if err := c.Resolve(&repo); err != nil {
			slog.Error("Failed to resolve BillableEntryRepository for BillableEntryWriter.", "err", err)
			return nil, err
		}
		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load billing_out.BillableEntryWriter.", "err", err)
		panic(err)
	}

	// BillableEntryReader interface
	err = c.Singleton(func() (billing_out.BillableEntryReader, error) {
		var repo *db.BillableEntryRepository
		if err := c.Resolve(&repo); err != nil {
			slog.Error("Failed to resolve BillableEntryRepository for BillableEntryReader.", "err", err)
			return nil, err
		}
		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load billing_out.BillableEntryReader.", "err", err)
		panic(err)
	}

	// Billing - Subscription Repository
	err = c.Singleton(func() (*db.SubscriptionRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for SubscriptionRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for SubscriptionRepository.", "err", err)
			return nil, err
		}

		return db.NewSubscriptionRepository(client, config.MongoDB.DBName, billing_entities.Subscription{}, "subscriptions"), nil
	})

	if err != nil {
		slog.Error("Failed to load SubscriptionRepository.", "err", err)
		panic(err)
	}

	// SubscriptionWriter interface
	err = c.Singleton(func() (billing_out.SubscriptionWriter, error) {
		var repo *db.SubscriptionRepository
		if err := c.Resolve(&repo); err != nil {
			slog.Error("Failed to resolve SubscriptionRepository for SubscriptionWriter.", "err", err)
			return nil, err
		}
		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load billing_out.SubscriptionWriter.", "err", err)
		panic(err)
	}

	// SubscriptionReader interface
	err = c.Singleton(func() (billing_out.SubscriptionReader, error) {
		var repo *db.SubscriptionRepository
		if err := c.Resolve(&repo); err != nil {
			slog.Error("Failed to resolve SubscriptionRepository for SubscriptionReader.", "err", err)
			return nil, err
		}
		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load billing_out.SubscriptionReader.", "err", err)
		panic(err)
	}

	// Billing - Plan Repository
	err = c.Singleton(func() (*db.PlanRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for PlanRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for PlanRepository.", "err", err)
			return nil, err
		}

		return db.NewPlanRepository(client, config.MongoDB.DBName, billing_entities.Plan{}, "plans"), nil
	})

	if err != nil {
		slog.Error("Failed to load PlanRepository.", "err", err)
		panic(err)
	}

	// PlanReader interface
	err = c.Singleton(func() (billing_out.PlanReader, error) {
		var repo *db.PlanRepository
		if err := c.Resolve(&repo); err != nil {
			slog.Error("Failed to resolve PlanRepository for PlanReader.", "err", err)
			return nil, err
		}
		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load billing_out.PlanReader.", "err", err)
		panic(err)
	}

	// billing_in.PlanReader (query service)
	err = c.Singleton(func() (billing_in.PlanReader, error) {
		var planReader billing_out.PlanReader
		if err := c.Resolve(&planReader); err != nil {
			slog.Error("Failed to resolve billing_out.PlanReader for billing_in.PlanReader.", "err", err)
			return nil, err
		}
		return billing_services.NewPlanQueryService(planReader), nil
	})

	if err != nil {
		slog.Error("Failed to load billing_in.PlanReader.", "err", err)
		panic(err)
	}

	// billing_in.SubscriptionReader (query service)
	err = c.Singleton(func() (billing_in.SubscriptionReader, error) {
		var subscriptionReader billing_out.SubscriptionReader
		if err := c.Resolve(&subscriptionReader); err != nil {
			slog.Error("Failed to resolve billing_out.SubscriptionReader for billing_in.SubscriptionReader.", "err", err)
			return nil, err
		}
		return billing_services.NewSubscriptionQueryService(subscriptionReader), nil
	})

	if err != nil {
		slog.Error("Failed to load billing_in.SubscriptionReader.", "err", err)
		panic(err)
	}

	// BillableOperationCommandHandler
	err = c.Singleton(func() (billing_in.BillableOperationCommandHandler, error) {
		var billableEntryWriter billing_out.BillableEntryWriter
		var billableEntryReader billing_out.BillableEntryReader
		var subscriptionWriter billing_out.SubscriptionWriter
		var subscriptionReader billing_out.SubscriptionReader
		var planReader billing_out.PlanReader
		var groupReader iam_out.GroupReader

		if err := c.Resolve(&billableEntryWriter); err != nil {
			slog.Error("Failed to resolve BillableEntryWriter for BillableOperationCommandHandler.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&billableEntryReader); err != nil {
			slog.Error("Failed to resolve BillableEntryReader for BillableOperationCommandHandler.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&subscriptionWriter); err != nil {
			slog.Error("Failed to resolve SubscriptionWriter for BillableOperationCommandHandler.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&subscriptionReader); err != nil {
			slog.Error("Failed to resolve SubscriptionReader for BillableOperationCommandHandler.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&planReader); err != nil {
			slog.Error("Failed to resolve PlanReader for BillableOperationCommandHandler.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&groupReader); err != nil {
			slog.Error("Failed to resolve GroupReader for BillableOperationCommandHandler.", "err", err)
			return nil, err
		}

		return billing_usecases.NewCreateBillableEntryUseCase(
			billableEntryWriter,
			billableEntryReader,
			subscriptionWriter,
			subscriptionReader,
			planReader,
			groupReader,
		), nil
	})

	if err != nil {
		slog.Error("Failed to load billing_in.BillableOperationCommandHandler.", "err", err)
		panic(err)
	}

	// UpgradeSubscriptionCommandHandler
	err = c.Singleton(func() (billing_in.UpgradeSubscriptionCommandHandler, error) {
		var subscriptionReader billing_out.SubscriptionReader
		var subscriptionWriter billing_out.SubscriptionWriter
		var planReader billing_out.PlanReader

		if err := c.Resolve(&subscriptionReader); err != nil {
			slog.Error("Failed to resolve SubscriptionReader for UpgradeSubscriptionCommandHandler.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&subscriptionWriter); err != nil {
			slog.Error("Failed to resolve SubscriptionWriter for UpgradeSubscriptionCommandHandler.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&planReader); err != nil {
			slog.Error("Failed to resolve PlanReader for UpgradeSubscriptionCommandHandler.", "err", err)
			return nil, err
		}

		return billing_usecases.NewUpgradeSubscriptionUseCase(subscriptionReader, subscriptionWriter, planReader), nil
	})

	if err != nil {
		slog.Error("Failed to load billing_in.UpgradeSubscriptionCommandHandler.", "err", err)
		panic(err)
	}

	// DowngradeSubscriptionCommandHandler
	err = c.Singleton(func() (billing_in.DowngradeSubscriptionCommandHandler, error) {
		var subscriptionReader billing_out.SubscriptionReader
		var subscriptionWriter billing_out.SubscriptionWriter
		var planReader billing_out.PlanReader
		var billableEntryReader billing_out.BillableEntryReader

		if err := c.Resolve(&subscriptionReader); err != nil {
			slog.Error("Failed to resolve SubscriptionReader for DowngradeSubscriptionCommandHandler.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&subscriptionWriter); err != nil {
			slog.Error("Failed to resolve SubscriptionWriter for DowngradeSubscriptionCommandHandler.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&planReader); err != nil {
			slog.Error("Failed to resolve PlanReader for DowngradeSubscriptionCommandHandler.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&billableEntryReader); err != nil {
			slog.Error("Failed to resolve BillableEntryReader for DowngradeSubscriptionCommandHandler.", "err", err)
			return nil, err
		}

		return billing_usecases.NewDowngradeSubscriptionUseCase(subscriptionReader, subscriptionWriter, planReader, billableEntryReader), nil
	})

	if err != nil {
		slog.Error("Failed to load billing_in.DowngradeSubscriptionCommandHandler.", "err", err)
		panic(err)
	}

	// Payment - PaymentRepository (MongoDB)
	err = c.Singleton(func() (payment_out.PaymentRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for PaymentRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for PaymentRepository.", "err", err)
			return nil, err
		}

		return db.NewPaymentRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load payment_out.PaymentRepository.", "err", err)
		panic(err)
	}

	// StripeAdapter (PaymentProviderAdapter)
	err = c.Singleton(func() (payment_out.PaymentProviderAdapter, error) {
		return stripe_adapter.NewStripeAdapter(), nil
	})

	if err != nil {
		slog.Error("Failed to load payment_out.PaymentProviderAdapter (Stripe).", "err", err)
		panic(err)
	}

	// PaymentCommand (PaymentService)
	err = c.Singleton(func() (payment_in.PaymentCommand, error) {
		var paymentRepo payment_out.PaymentRepository
		var walletCommand wallet_in.WalletCommand
		var stripeAdapter payment_out.PaymentProviderAdapter

		if err := c.Resolve(&paymentRepo); err != nil {
			slog.Error("Failed to resolve PaymentRepository for PaymentService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&walletCommand); err != nil {
			slog.Error("Failed to resolve WalletCommand for PaymentService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&stripeAdapter); err != nil {
			slog.Error("Failed to resolve PaymentProviderAdapter for PaymentService.", "err", err)
			return nil, err
		}

		return payment_services.NewPaymentService(paymentRepo, walletCommand, stripeAdapter), nil
	})

	if err != nil {
		slog.Error("Failed to load payment_in.PaymentCommand.", "err", err)
		panic(err)
	}

	// PaymentQuery (PaymentQueryService)
	err = c.Singleton(func() (payment_in.PaymentQuery, error) {
		var paymentRepo payment_out.PaymentRepository

		if err := c.Resolve(&paymentRepo); err != nil {
			slog.Error("Failed to resolve PaymentRepository for PaymentQueryService.", "err", err)
			return nil, err
		}

		return payment_usecases.NewPaymentQueryService(paymentRepo), nil
	})

	if err != nil {
		slog.Error("Failed to load payment_in.PaymentQuery.", "err", err)
		panic(err)
	}

	// CheckoutSubscriptionCommandHandler
	err = c.Singleton(func() (billing_in.CheckoutSubscriptionCommandHandler, error) {
		var subscriptionReader billing_out.SubscriptionReader
		var subscriptionWriter billing_out.SubscriptionWriter
		var planReader billing_out.PlanReader
		var paymentRepo payment_out.PaymentRepository

		if err := c.Resolve(&subscriptionReader); err != nil {
			slog.Error("Failed to resolve SubscriptionReader for CheckoutSubscriptionCommandHandler.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&subscriptionWriter); err != nil {
			slog.Error("Failed to resolve SubscriptionWriter for CheckoutSubscriptionCommandHandler.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&planReader); err != nil {
			slog.Error("Failed to resolve PlanReader for CheckoutSubscriptionCommandHandler.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&paymentRepo); err != nil {
			slog.Error("Failed to resolve PaymentRepository for CheckoutSubscriptionCommandHandler.", "err", err)
			return nil, err
		}

		return billing_usecases.NewCheckoutSubscriptionUseCase(subscriptionReader, subscriptionWriter, planReader, paymentRepo), nil
	})

	if err != nil {
		slog.Error("Failed to load billing_in.CheckoutSubscriptionCommandHandler.", "err", err)
		panic(err)
	}

	// Billing - Audit Trail Repository (commented out for basic functionality)
	// err = c.Singleton(func() (*db.AuditTrailRepository, error) {
	// 	var client *mongo.Client
	// 	var config common.Config

	// 	if err := c.Resolve(&client); err != nil {
	// 		slog.Error("Failed to resolve mongo.Client for AuditTrailRepository.", "err", err)
	// 		return nil, err
	// 	}

	// 	if err := c.Resolve(&config); err != nil {
	// 		slog.Error("Failed to resolve common.Config for AuditTrailRepository.", "err", err)
	// 		return nil, err
	// 	}

	// 	return db.NewAuditTrailRepository(client, config.MongoDB.DBName, billing_entities.AuditTrailEntry{}, "audit_trail"), nil
	// })

	// if err != nil {
	// 	slog.Error("Failed to load AuditTrailRepository.", "err", err)
	// 	panic(err)
	// }

	// AuditTrailWriter interface (commented out for basic functionality)
	// err = c.Singleton(func() (billing_out.AuditTrailWriter, error) {
	// 	var repo *db.AuditTrailRepository
	// 	if err := c.Resolve(&repo); err != nil {
	// 		slog.Error("Failed to resolve AuditTrailRepository for AuditTrailWriter.", "err", err)
	// 		return nil, err
	// 	}
	// 	return repo, nil
	// })

	// if err != nil {
	// 	slog.Error("Failed to load billing_out.AuditTrailWriter.", "err", err)
	// 	panic(err)
	// }

	// AuditTrailReader interface (commented out for basic functionality)
	// err = c.Singleton(func() (billing_out.AuditTrailReader, error) {
	// 	var repo *db.AuditTrailRepository
	// 	if err := c.Resolve(&repo); err != nil {
	// 		slog.Error("Failed to resolve AuditTrailRepository for AuditTrailReader.", "err", err)
	// 		return nil, err
	// 	}
	// 	return repo, nil
	// })

	// if err != nil {
	// 	slog.Error("Failed to load billing_out.AuditTrailReader.", "err", err)
	// 	panic(err)
	// }

	// ===== Exchange Domain =====

	// Price Feed Providers
	err = c.Singleton(func() ([]exchange_out.PriceFeedProvider, error) {
		var providers []exchange_out.PriceFeedProvider
		providers = append(providers, pricefeed.NewCoinGeckoAdapter(os.Getenv("COINGECKO_API_KEY")))
		providers = append(providers, pricefeed.NewCoinbasePriceAdapter())
		providers = append(providers, pricefeed.NewKrakenPriceAdapter())
		return providers, nil
	})

	if err != nil {
		slog.Error("Failed to load []exchange_out.PriceFeedProvider.", "err", err)
		panic(err)
	}

	// Exchange Adapters (Coinbase + Kraken)
	err = c.Singleton(func() ([]exchange_out.ExchangeAdapter, error) {
		var adapters []exchange_out.ExchangeAdapter
		if key := os.Getenv("COINBASE_API_KEY"); key != "" {
			adapters = append(adapters, coinbase.NewCoinbaseAdapter(key, os.Getenv("COINBASE_API_SECRET")))
		}
		if key := os.Getenv("KRAKEN_API_KEY"); key != "" {
			adapters = append(adapters, kraken.NewKrakenAdapter(key, os.Getenv("KRAKEN_API_SECRET")))
		}
		return adapters, nil
	})

	if err != nil {
		slog.Error("Failed to load []exchange_out.ExchangeAdapter.", "err", err)
		panic(err)
	}

	// Exchange - Kafka Event Publisher (if kafka client available)
	err = c.Singleton(func() (exchange_out.ExchangeEventPublisher, error) {
		var kafkaClient *kafka.Client
		if err := c.Resolve(&kafkaClient); err != nil {
			slog.Warn("Kafka client not available for ExchangeEventPublisher — exchange events will not be published", "err", err)
			return nil, nil
		}
		return kafka.NewExchangeEventPublisherAdapter(kafkaClient), nil
	})

	if err != nil {
		slog.Error("Failed to load exchange_out.ExchangeEventPublisher.", "err", err)
		panic(err)
	}

	// Exchange - MongoDB Repositories
	err = c.Singleton(func() (exchange_out.OrderRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for ExchangeOrderRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for ExchangeOrderRepository.", "err", err)
			return nil, err
		}

		return db.NewExchangeOrderRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load exchange_out.OrderRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (exchange_out.QuoteRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for ExchangeQuoteRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for ExchangeQuoteRepository.", "err", err)
			return nil, err
		}

		return db.NewExchangeQuoteRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load exchange_out.QuoteRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (exchange_out.ExchangeRateRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for ExchangeRateRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for ExchangeRateRepository.", "err", err)
			return nil, err
		}

		return db.NewExchangeRateRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load exchange_out.ExchangeRateRepository.", "err", err)
		panic(err)
	}

	// Exchange Services
	err = c.Singleton(func() (*exchange_services.PricingService, error) {
		var providers []exchange_out.PriceFeedProvider
		if err := c.Resolve(&providers); err != nil {
			slog.Error("Failed to resolve []exchange_out.PriceFeedProvider for PricingService.", "err", err)
			return nil, err
		}

		var rateRepo exchange_out.ExchangeRateRepository
		if err := c.Resolve(&rateRepo); err != nil {
			slog.Error("Failed to resolve exchange_out.ExchangeRateRepository for PricingService.", "err", err)
			return nil, err
		}

		// RateCache is optional (Dragonfly/Redis)
		var rateCache exchange_out.RateCache
		if err := c.Resolve(&rateCache); err != nil {
			slog.Info("RateCache not available for PricingService — caching disabled", "err", err)
		}

		resourceOwner := shared.ResourceOwner{} // platform-level (system) resource owner
		return exchange_services.NewPricingService(providers, rateCache, rateRepo, resourceOwner), nil
	})

	if err != nil {
		slog.Error("Failed to load *exchange_services.PricingService.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (*exchange_services.FeeService, error) {
		// TODO: wire SubscriptionPlanResolver when available
		return exchange_services.NewFeeService(nil), nil
	})

	if err != nil {
		slog.Error("Failed to load *exchange_services.FeeService.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (*exchange_services.SmartRouter, error) {
		var adapters []exchange_out.ExchangeAdapter
		if err := c.Resolve(&adapters); err != nil {
			slog.Error("Failed to resolve []exchange_out.ExchangeAdapter for SmartRouter.", "err", err)
			return nil, err
		}

		return exchange_services.NewSmartRouter(adapters), nil
	})

	if err != nil {
		slog.Error("Failed to load *exchange_services.SmartRouter.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (*exchange_services.OrderService, error) {
		var orderRepo exchange_out.OrderRepository
		var quoteRepo exchange_out.QuoteRepository
		var router *exchange_services.SmartRouter
		var pricing *exchange_services.PricingService
		var fees *exchange_services.FeeService

		if err := c.Resolve(&orderRepo); err != nil {
			slog.Error("Failed to resolve exchange_out.OrderRepository for OrderService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&quoteRepo); err != nil {
			slog.Error("Failed to resolve exchange_out.QuoteRepository for OrderService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&router); err != nil {
			slog.Error("Failed to resolve *exchange_services.SmartRouter for OrderService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&pricing); err != nil {
			slog.Error("Failed to resolve *exchange_services.PricingService for OrderService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&fees); err != nil {
			slog.Error("Failed to resolve *exchange_services.FeeService for OrderService.", "err", err)
			return nil, err
		}

		var eventPublisher exchange_out.ExchangeEventPublisher
		if err := c.Resolve(&eventPublisher); err != nil {
			slog.Warn("ExchangeEventPublisher not available for OrderService — events will not be published", "err", err)
		}

		resourceOwner := shared.ResourceOwner{} // platform-level (system) resource owner
		// Wallet and Stripe ops - these are interface adapters
		// TODO: wire concrete implementations
		return exchange_services.NewOrderService(orderRepo, quoteRepo, router, pricing, fees, nil, nil, eventPublisher, resourceOwner), nil
	})

	if err != nil {
		slog.Error("Failed to load *exchange_services.OrderService.", "err", err)
		panic(err)
	}

	// Exchange Use Cases
	err = c.Singleton(func() (*exchange_usecases.GetQuoteUseCase, error) {
		var pricing *exchange_services.PricingService
		var fees *exchange_services.FeeService
		var quoteRepo exchange_out.QuoteRepository

		if err := c.Resolve(&pricing); err != nil {
			slog.Error("Failed to resolve *exchange_services.PricingService for GetQuoteUseCase.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&fees); err != nil {
			slog.Error("Failed to resolve *exchange_services.FeeService for GetQuoteUseCase.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&quoteRepo); err != nil {
			slog.Error("Failed to resolve exchange_out.QuoteRepository for GetQuoteUseCase.", "err", err)
			return nil, err
		}

		var eventPublisher exchange_out.ExchangeEventPublisher
		if err := c.Resolve(&eventPublisher); err != nil {
			slog.Warn("ExchangeEventPublisher not available for GetQuoteUseCase — events will not be published", "err", err)
		}

		resourceOwner := shared.ResourceOwner{} // platform-level (system) resource owner
		return exchange_usecases.NewGetQuoteUseCase(pricing, fees, quoteRepo, eventPublisher, resourceOwner), nil
	})

	if err != nil {
		slog.Error("Failed to load *exchange_usecases.GetQuoteUseCase.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (*exchange_usecases.GetExchangeRatesUseCase, error) {
		var pricing *exchange_services.PricingService
		if err := c.Resolve(&pricing); err != nil {
			slog.Error("Failed to resolve *exchange_services.PricingService for GetExchangeRatesUseCase.", "err", err)
			return nil, err
		}

		return exchange_usecases.NewGetExchangeRatesUseCase(pricing), nil
	})

	if err != nil {
		slog.Error("Failed to load *exchange_usecases.GetExchangeRatesUseCase.", "err", err)
		panic(err)
	}

	// Matchmaking Session Repository
	err = c.Singleton(func() (matchmaking_out.MatchmakingSessionRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for MatchmakingSessionRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for MatchmakingSessionRepository.", "err", err)
			return nil, err
		}

		return db.NewMatchmakingSessionRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load matchmaking_out.MatchmakingSessionRepository.", "err", err)
		panic(err)
	}

	// Player Rating Repository
	err = c.Singleton(func() (matchmaking_out.PlayerRatingRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for PlayerRatingRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for PlayerRatingRepository.", "err", err)
			return nil, err
		}

		return db.NewPlayerRatingMongoDBRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load matchmaking_out.PlayerRatingRepository.", "err", err)
		panic(err)
	}

	// Matchmaking Pool Repository
	err = c.Singleton(func() (matchmaking_out.MatchmakingPoolRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for MatchmakingPoolRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for MatchmakingPoolRepository.", "err", err)
			return nil, err
		}

		return db.NewMatchmakingPoolRepository(client, config.MongoDB.DBName), nil
	})

	if err != nil {
		slog.Error("Failed to load matchmaking_out.MatchmakingPoolRepository.", "err", err)
		panic(err)
	}

	// Lobby Orchestration Service
	err = c.Singleton(func() (matchmaking_in.LobbyCommand, error) {
		var lobbyRepo matchmaking_out.LobbyRepository
		var poolRepo matchmaking_out.PrizePoolRepository
		var walletCmd wallet_in.WalletCommand
		var wsHub *websocket.WebSocketHub

		if err := c.Resolve(&lobbyRepo); err != nil {
			slog.Error("Failed to resolve matchmaking_out.LobbyRepository for LobbyOrchestrationService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&poolRepo); err != nil {
			slog.Error("Failed to resolve matchmaking_out.PrizePoolRepository for LobbyOrchestrationService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&walletCmd); err != nil {
			slog.Error("Failed to resolve wallet_in.WalletCommand for LobbyOrchestrationService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&wsHub); err != nil {
			slog.Error("Failed to resolve *websocket.WebSocketHub for LobbyOrchestrationService.", "err", err)
			return nil, err
		}

		return matchmaking_services.NewLobbyOrchestrationService(lobbyRepo, poolRepo, walletCmd, wsHub), nil
	})

	if err != nil {
		slog.Error("Failed to load matchmaking_in.LobbyCommand.", "err", err)
		panic(err)
	}

	// Prize Pool Query Service
	err = c.Singleton(func() (*matchmaking_services.PrizePoolQueryService, error) {
		var poolRepo matchmaking_out.PrizePoolRepository

		if err := c.Resolve(&poolRepo); err != nil {
			slog.Error("Failed to resolve matchmaking_out.PrizePoolRepository for PrizePoolQueryService.", "err", err)
			return nil, err
		}

		return matchmaking_services.NewPrizePoolQueryService(poolRepo), nil
	})

	if err != nil {
		slog.Error("Failed to load *matchmaking_services.PrizePoolQueryService.", "err", err)
		panic(err)
	}

	// Matchmaking Session Query Service
	err = c.Singleton(func() (*matchmaking_services.MatchmakingSessionQueryService, error) {
		var sessionRepo matchmaking_out.MatchmakingSessionRepository

		if err := c.Resolve(&sessionRepo); err != nil {
			slog.Error("Failed to resolve matchmaking_out.MatchmakingSessionRepository for MatchmakingSessionQueryService.", "err", err)
			return nil, err
		}

		return matchmaking_services.NewMatchmakingSessionQueryService(sessionRepo), nil
	})

	if err != nil {
		slog.Error("Failed to load *matchmaking_services.MatchmakingSessionQueryService.", "err", err)
		panic(err)
	}

	// Matchmaking Pool Query Service
	err = c.Singleton(func() (*matchmaking_services.MatchmakingPoolQueryService, error) {
		var poolRepo matchmaking_out.MatchmakingPoolRepository

		if err := c.Resolve(&poolRepo); err != nil {
			slog.Error("Failed to resolve matchmaking_out.MatchmakingPoolRepository for MatchmakingPoolQueryService.", "err", err)
			return nil, err
		}

		return matchmaking_services.NewMatchmakingPoolQueryService(poolRepo), nil
	})

	if err != nil {
		slog.Error("Failed to load *matchmaking_services.MatchmakingPoolQueryService.", "err", err)
		panic(err)
	}

	// Prize Distribution Job
	err = c.Singleton(func() (*jobs.PrizeDistributionJob, error) {
		var poolRepo matchmaking_out.PrizePoolRepository
		var poolQuerySvc *matchmaking_services.PrizePoolQueryService
		var walletCmd wallet_in.WalletCommand

		if err := c.Resolve(&poolRepo); err != nil {
			slog.Error("Failed to resolve matchmaking_out.PrizePoolRepository for PrizeDistributionJob.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&poolQuerySvc); err != nil {
			slog.Error("Failed to resolve *matchmaking_services.PrizePoolQueryService for PrizeDistributionJob.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&walletCmd); err != nil {
			slog.Error("Failed to resolve wallet_in.WalletCommand for PrizeDistributionJob.", "err", err)
			return nil, err
		}

		// Run every 5 minutes
		return jobs.NewPrizeDistributionJob(poolRepo, poolQuerySvc, walletCmd, 5*time.Minute), nil
	})

	if err != nil {
		slog.Error("Failed to load *jobs.PrizeDistributionJob.", "err", err)
		panic(err)
	}

	// PlayerProfileRepository (needed by tournament use cases - may also be registered in WithSquadAPI)
	// This ensures it's available before tournament use cases are eagerly resolved
	err = c.Singleton(func() (*db.PlayerProfileRepository, error) {
		var client *mongo.Client
		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve mongo.Client for PlayerProfileRepository (tournament dependency).", "err", err)
			return nil, err
		}

		var config common.Config
		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve config for PlayerProfileRepository (tournament dependency).", "err", err)
			return nil, err
		}

		return db.NewPlayerProfileRepository(client, config.MongoDB.DBName, squad_entities.PlayerProfile{}, "player_profiles"), nil
	})
	if err != nil {
		slog.Error("Failed to load PlayerProfileRepository (tournament dependency).", "err", err)
		panic(err)
	}

	// squad_in.PlayerProfileReader (needed by RegisterForTournamentUseCase - may also be registered in WithSquadAPI)
	err = c.Singleton(func() (squad_in.PlayerProfileReader, error) {
		var repo *db.PlayerProfileRepository
		if err := c.Resolve(&repo); err != nil {
			slog.Error("Failed to resolve PlayerProfileRepository for squad_in.PlayerProfileReader (tournament dependency).", "err", err)
			return nil, err
		}

		return repo, nil
	})
	if err != nil {
		slog.Error("Failed to load squad_in.PlayerProfileReader (tournament dependency).", "err", err)
		panic(err)
	}

	// GenerateBracketsUseCase (must be registered before TournamentService which depends on it)
	err = c.Singleton(func() (*tournament_usecases.GenerateBracketsUseCase, error) {
		var billableOperationHandler billing_in.BillableOperationCommandHandler
		var tournamentRepo tournament_out.TournamentRepository

		if err := c.Resolve(&billableOperationHandler); err != nil {
			slog.Error("Failed to resolve BillableOperationCommandHandler for GenerateBracketsUseCase.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&tournamentRepo); err != nil {
			slog.Error("Failed to resolve TournamentRepository for GenerateBracketsUseCase.", "err", err)
			return nil, err
		}

		return tournament_usecases.NewGenerateBracketsUseCase(billableOperationHandler, tournamentRepo), nil
	})
	if err != nil {
		slog.Error("Failed to load GenerateBracketsUseCase.", "err", err)
		panic(err)
	}

	// Tournament Command Service
	// Tournament Authorization Adapter
	err = c.Singleton(func() (tournament_out.TournamentAuthorization, error) {
		var tournamentRepo tournament_out.TournamentRepository

		if err := c.Resolve(&tournamentRepo); err != nil {
			slog.Error("Failed to resolve TournamentRepository for TournamentAuthorization.", "err", err)
			return nil, err
		}

		return tournament_adapters.NewTournamentAuthorizationAdapter(tournamentRepo), nil
	})

	if err != nil {
		slog.Error("Failed to load tournament_out.TournamentAuthorization.", "err", err)
		panic(err)
	}

	// Tournament Event Publisher
	err = c.Singleton(func() (tournament_out.TournamentEventPublisher, error) {
		var eventPub *kafka.EventPublisher

		if err := c.Resolve(&eventPub); err != nil {
			slog.Warn("EventPublisher not available for TournamentEventPublisher", "err", err)
			return nil, nil
		}

		return kafka.NewTournamentEventPublisherAdapter(eventPub), nil
	})
	if err != nil {
		slog.Error("Failed to load tournament_out.TournamentEventPublisher.", "err", err)
		panic(err)
	}

	// Tournament Command Service
	err = c.Singleton(func() (tournament_in.TournamentCommand, error) {
		var tournamentRepo tournament_out.TournamentRepository
		var walletCmd wallet_in.WalletCommand
		var bracketGenerator *tournament_usecases.GenerateBracketsUseCase
		var authorization tournament_out.TournamentAuthorization
		var eventPublisher tournament_out.TournamentEventPublisher

		if err := c.Resolve(&tournamentRepo); err != nil {
			slog.Error("Failed to resolve tournament_out.TournamentRepository for TournamentService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&walletCmd); err != nil {
			slog.Error("Failed to resolve wallet_in.WalletCommand for TournamentService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&bracketGenerator); err != nil {
			slog.Error("Failed to resolve GenerateBracketsUseCase for TournamentService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&authorization); err != nil {
			slog.Warn("TournamentAuthorization not available, proceeding without RBAC", "err", err)
		}

		if err := c.Resolve(&eventPublisher); err != nil {
			slog.Warn("TournamentEventPublisher not available, proceeding without events", "err", err)
		}

		return tournament_services.NewTournamentService(tournamentRepo, walletCmd, bracketGenerator, authorization, eventPublisher), nil
	})

	if err != nil {
		slog.Error("Failed to load tournament_in.TournamentCommand.", "err", err)
		panic(err)
	}

	// Tournament Query Service
	err = c.Singleton(func() (*tournament_services.TournamentQueryService, error) {
		var tournamentRepo tournament_out.TournamentRepository

		if err := c.Resolve(&tournamentRepo); err != nil {
			slog.Error("Failed to resolve tournament_out.TournamentRepository for TournamentQueryService.", "err", err)
			return nil, err
		}

		return tournament_services.NewTournamentQueryService(tournamentRepo), nil
	})

	if err != nil {
		slog.Error("Failed to load *tournament_services.TournamentQueryService.", "err", err)
		panic(err)
	}

	// Tournament Reader Service
	err = c.Singleton(func() (tournament_in.TournamentReader, error) {
		var tournamentQuerySvc *tournament_services.TournamentQueryService
		var tournamentRepo tournament_out.TournamentRepository

		if err := c.Resolve(&tournamentQuerySvc); err != nil {
			slog.Error("Failed to resolve TournamentQueryService for TournamentReaderService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&tournamentRepo); err != nil {
			slog.Error("Failed to resolve TournamentRepository for TournamentReaderService.", "err", err)
			return nil, err
		}

		return tournament_services.NewTournamentReaderService(tournamentQuerySvc, tournamentRepo), nil
	})

	if err != nil {
		slog.Error("Failed to load tournament_in.TournamentReader.", "err", err)
		panic(err)
	}

	// Matchmaking Usecases
	err = c.Singleton(func() (matchmaking_in.JoinMatchmakingQueueCommandHandler, error) {
		var billableOperationHandler billing_in.BillableOperationCommandHandler
		var sessionRepo matchmaking_out.MatchmakingSessionRepository
		var lobbyRepo matchmaking_out.LobbyRepository
		var eventPublisher *kafka.EventPublisher
		var wsHub *websocket.WebSocketHub

		if err := c.Resolve(&billableOperationHandler); err != nil {
			slog.Error("Failed to resolve BillableOperationCommandHandler for JoinMatchmakingQueueUseCase.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&sessionRepo); err != nil {
			slog.Error("Failed to resolve MatchmakingSessionRepository for JoinMatchmakingQueueUseCase.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&lobbyRepo); err != nil {
			slog.Error("Failed to resolve LobbyRepository for JoinMatchmakingQueueUseCase.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&eventPublisher); err != nil {
			slog.Error("Failed to resolve EventPublisher for JoinMatchmakingQueueUseCase.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&wsHub); err != nil {
			slog.Warn("Failed to resolve WebSocketHub for JoinMatchmakingQueueUseCase (notifications disabled).", "err", err)
			// non-fatal: matching works without WebSocket
		}

		return matchmaking_usecases.NewJoinMatchmakingQueueUseCase(billableOperationHandler, sessionRepo, lobbyRepo, eventPublisher, wsHub), nil
	})
	if err != nil {
		slog.Error("Failed to load JoinMatchmakingQueueCommandHandler.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (matchmaking_in.LeaveMatchmakingQueueCommandHandler, error) {
		var billableOperationHandler billing_in.BillableOperationCommandHandler
		var sessionRepo matchmaking_out.MatchmakingSessionRepository
		var eventPublisher *kafka.EventPublisher

		if err := c.Resolve(&billableOperationHandler); err != nil {
			slog.Error("Failed to resolve BillableOperationCommandHandler for LeaveMatchmakingQueueUseCase.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&sessionRepo); err != nil {
			slog.Error("Failed to resolve MatchmakingSessionRepository for LeaveMatchmakingQueueUseCase.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&eventPublisher); err != nil {
			slog.Error("Failed to resolve EventPublisher for LeaveMatchmakingQueueUseCase.", "err", err)
			return nil, err
		}

		return matchmaking_usecases.NewLeaveMatchmakingQueueUseCase(billableOperationHandler, sessionRepo, eventPublisher), nil
	})
	if err != nil {
		slog.Error("Failed to load LeaveMatchmakingQueueCommandHandler.", "err", err)
		panic(err)
	}

	// Tournament Usecases
	err = c.Singleton(func() (*tournament_usecases.CreateTournamentUseCase, error) {
		var billableOperationHandler billing_in.BillableOperationCommandHandler
		var tournamentRepo tournament_out.TournamentRepository

		if err := c.Resolve(&billableOperationHandler); err != nil {
			slog.Error("Failed to resolve BillableOperationCommandHandler for CreateTournamentUseCase.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&tournamentRepo); err != nil {
			slog.Error("Failed to resolve TournamentRepository for CreateTournamentUseCase.", "err", err)
			return nil, err
		}

		return tournament_usecases.NewCreateTournamentUseCase(billableOperationHandler, tournamentRepo), nil
	})
	if err != nil {
		slog.Error("Failed to load CreateTournamentUseCase.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (*tournament_usecases.RegisterForTournamentUseCase, error) {
		var billableOperationHandler billing_in.BillableOperationCommandHandler
		var tournamentRepo tournament_out.TournamentRepository
		var playerProfileReader squad_in.PlayerProfileReader

		if err := c.Resolve(&billableOperationHandler); err != nil {
			slog.Error("Failed to resolve BillableOperationCommandHandler for RegisterForTournamentUseCase.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&tournamentRepo); err != nil {
			slog.Error("Failed to resolve TournamentRepository for RegisterForTournamentUseCase.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&playerProfileReader); err != nil {
			slog.Error("Failed to resolve PlayerProfileReader for RegisterForTournamentUseCase.", "err", err)
			return nil, err
		}

		return tournament_usecases.NewRegisterForTournamentUseCase(billableOperationHandler, tournamentRepo, playerProfileReader), nil
	})
	if err != nil {
		slog.Error("Failed to load RegisterForTournamentUseCase.", "err", err)
		panic(err)
	}

	// -----

	// Scores Domain - Match Result Repository
	err = c.Singleton(func() (scores_out.MatchResultRepository, error) {
		var mongoClient *mongo.Client
		var config common.Config

		if err := c.Resolve(&mongoClient); err != nil {
			slog.Error("Failed to resolve mongo.Client for MatchResultRepository.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve Config for MatchResultRepository.", "err", err)
			return nil, err
		}

		return db.NewMongoMatchResultRepository(mongoClient, config.MongoDB.DBName), nil
	})
	if err != nil {
		slog.Error("Failed to load scores_out.MatchResultRepository.", "err", err)
		panic(err)
	}

	// Scores Domain - Event Publisher Adapter
	err = c.Singleton(func() (scores_out.ScoreEventPublisher, error) {
		var eventPublisher *kafka.EventPublisher

		if err := c.Resolve(&eventPublisher); err != nil {
			slog.Warn("EventPublisher not available for ScoreEventPublisher, using nil-safe adapter", "err", err)
			return kafka.NewScoreEventPublisherAdapter(nil), nil
		}

		return kafka.NewScoreEventPublisherAdapter(eventPublisher), nil
	})
	if err != nil {
		slog.Error("Failed to load scores_out.ScoreEventPublisher.", "err", err)
		panic(err)
	}

	// Scores Domain - Prize Distribution Gateway
	err = c.Singleton(func() (scores_out.PrizeDistributionGateway, error) {
		var tournamentPrizeService *tournament_services.PrizeDistributionService
		var matchmakingPrizeRepo matchmaking_out.PrizePoolRepository
		var tournamentPrizePoolRepo tournament_services.PrizePoolRepository

		// These are optional dependencies — scores work without prize distribution
		if err := c.Resolve(&tournamentPrizeService); err != nil {
			slog.Warn("TournamentPrizeDistributionService not available for PrizeDistributionGateway", "err", err)
		}
		if err := c.Resolve(&matchmakingPrizeRepo); err != nil {
			slog.Warn("MatchmakingPrizePoolRepository not available for PrizeDistributionGateway", "err", err)
		}
		if err := c.Resolve(&tournamentPrizePoolRepo); err != nil {
			slog.Warn("TournamentPrizePoolRepository not available for PrizeDistributionGateway", "err", err)
		}

		return scores_adapter.NewPrizeDistributionAdapter(tournamentPrizeService, tournamentPrizePoolRepo, matchmakingPrizeRepo), nil
	})
	if err != nil {
		slog.Error("Failed to load scores_out.PrizeDistributionGateway.", "err", err)
		panic(err)
	}

	// Scores Domain - Command Handler
	err = c.Singleton(func() (scores_in.MatchResultCommandHandler, error) {
		var repo scores_out.MatchResultRepository
		var eventPub scores_out.ScoreEventPublisher
		var prizeGateway scores_out.PrizeDistributionGateway
		var tournamentRepo tournament_out.TournamentRepository
		var tournamentCmd tournament_in.TournamentCommand

		if err := c.Resolve(&repo); err != nil {
			slog.Error("Failed to resolve MatchResultRepository for MatchResultCommandHandler.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&eventPub); err != nil {
			slog.Error("Failed to resolve ScoreEventPublisher for MatchResultCommandHandler.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&prizeGateway); err != nil {
			slog.Warn("PrizeDistributionGateway not available for MatchResultCommandHandler", "err", err)
		}
		if err := c.Resolve(&tournamentRepo); err != nil {
			slog.Warn("TournamentRepository not available for ScoreAuthorization", "err", err)
		}

		// Create authorization adapter with tournament and match result repos
		authorization := scores_adapter.NewScoreAuthorizationAdapter(tournamentRepo, repo)

		// Create tournament match callback (score → tournament domain bridge)
		var tournamentCallback scores_out.TournamentMatchCallback
		if err := c.Resolve(&tournamentCmd); err != nil {
			slog.Warn("TournamentCommand not available for TournamentMatchCallback", "err", err)
		} else {
			tournamentCallback = scores_adapter.NewTournamentMatchCallbackAdapter(tournamentCmd)
		}

		return scores_usecases.NewMatchResultCommandHandler(repo, eventPub, prizeGateway, authorization, tournamentCallback), nil
	})
	if err != nil {
		slog.Error("Failed to load scores_in.MatchResultCommandHandler.", "err", err)
		panic(err)
	}

	// Scores Domain - Query Handler
	err = c.Singleton(func() (scores_in.MatchResultQueryHandler, error) {
		var repo scores_out.MatchResultRepository

		if err := c.Resolve(&repo); err != nil {
			slog.Error("Failed to resolve MatchResultRepository for MatchResultQueryHandler.", "err", err)
			return nil, err
		}

		return scores_usecases.NewMatchResultQueryHandler(repo), nil
	})
	if err != nil {
		slog.Error("Failed to load scores_in.MatchResultQueryHandler.", "err", err)
		panic(err)
	}

	// ========================
	// Oracle Domain
	// ========================

	// Oracle Domain - Result Repository
	err = c.Singleton(func() (oracle_out.OracleResultRepository, error) {
		var mongoClient *mongo.Client
		var config common.Config

		if err := c.Resolve(&mongoClient); err != nil {
			slog.Error("Failed to resolve mongo.Client for OracleResultRepository.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve Config for OracleResultRepository.", "err", err)
			return nil, err
		}

		return db.NewMongoOracleResultRepository(mongoClient, config.MongoDB.DBName), nil
	})
	if err != nil {
		slog.Error("Failed to load oracle_out.OracleResultRepository.", "err", err)
		panic(err)
	}

	// Oracle Domain - Event Publisher Adapter
	err = c.Singleton(func() (oracle_out.OracleEventPublisher, error) {
		var eventPublisher *kafka.EventPublisher

		if err := c.Resolve(&eventPublisher); err != nil {
			slog.Warn("EventPublisher not available for OracleEventPublisher, using nil-safe adapter", "err", err)
			return kafka.NewOracleEventPublisherAdapter(nil), nil
		}

		return kafka.NewOracleEventPublisherAdapter(eventPublisher), nil
	})
	if err != nil {
		slog.Error("Failed to load oracle_out.OracleEventPublisher.", "err", err)
		panic(err)
	}

	// Oracle Domain - Chain Score Gateway
	err = c.Singleton(func() (oracle_out.ChainScoreGateway, error) {
		chainConfig := oracle_chain.EVMChainScoreGatewayConfig{
			PolygonRPCURL:       os.Getenv("POLYGON_RPC_URL"),
			PolygonContractAddr: os.Getenv("POLYGON_ORACLE_CONTRACT"),
			AmoyRPCURL:          os.Getenv("POLYGON_AMOY_RPC_URL"),
			AmoyContractAddr:    os.Getenv("POLYGON_AMOY_ORACLE_CONTRACT"),
			PrivateKey:          os.Getenv("ORACLE_PRIVATE_KEY"),
		}

		return oracle_chain.NewEVMChainScoreGateway(chainConfig), nil
	})
	if err != nil {
		slog.Error("Failed to load oracle_out.ChainScoreGateway.", "err", err)
		panic(err)
	}

	// Oracle Domain - External Score Providers
	err = c.Singleton(func() ([]oracle_out.ExternalScorePort, error) {
		providers := make([]oracle_out.ExternalScorePort, 0)

		pandaScoreKey := os.Getenv("PANDASCORE_API_KEY")
		if pandaScoreKey != "" {
			providers = append(providers, oracle_providers.NewPandaScoreAdapter(pandaScoreKey))
			slog.Info("PandaScore provider registered")
		}

		steamKey := os.Getenv("STEAM_WEB_API_KEY")
		if steamKey != "" {
			providers = append(providers, oracle_providers.NewSteamWebAPIAdapter(steamKey))
			slog.Info("Steam Web API provider registered")
		}

		faceitKey := os.Getenv("FACEIT_API_KEY")
		if faceitKey != "" {
			providers = append(providers, oracle_providers.NewFACEITAdapter(faceitKey))
			slog.Info("FACEIT provider registered")
		}

		slog.Info("Oracle providers registered", "count", len(providers))
		return providers, nil
	})
	if err != nil {
		slog.Error("Failed to load oracle ExternalScorePort providers.", "err", err)
		panic(err)
	}

	// Oracle Domain - Consensus Engine
	err = c.Singleton(func() (*oracle_services.ConsensusEngine, error) {
		tracker := oracle_services.NewProviderReliabilityTracker()
		return oracle_services.NewConsensusEngine(tracker), nil
	})
	if err != nil {
		slog.Error("Failed to load oracle ConsensusEngine.", "err", err)
		panic(err)
	}

	// Oracle Domain - Command Handler
	err = c.Singleton(func() (oracle_in.OracleCommandHandler, error) {
		var repo oracle_out.OracleResultRepository
		var eventPub oracle_out.OracleEventPublisher
		var chainGateway oracle_out.ChainScoreGateway
		var providers []oracle_out.ExternalScorePort
		var consensusEngine *oracle_services.ConsensusEngine

		if err := c.Resolve(&repo); err != nil {
			slog.Error("Failed to resolve OracleResultRepository for OracleCommandHandler.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&eventPub); err != nil {
			slog.Error("Failed to resolve OracleEventPublisher for OracleCommandHandler.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&chainGateway); err != nil {
			slog.Error("Failed to resolve ChainScoreGateway for OracleCommandHandler.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&providers); err != nil {
			slog.Warn("ExternalScorePort providers not available for OracleCommandHandler", "err", err)
			providers = []oracle_out.ExternalScorePort{}
		}
		if err := c.Resolve(&consensusEngine); err != nil {
			slog.Error("Failed to resolve ConsensusEngine for OracleCommandHandler.", "err", err)
			return nil, err
		}

		policy := oracle_vo.StandardPolicy()

		return oracle_usecases.NewOracleCommandHandler(repo, eventPub, providers, consensusEngine, chainGateway, policy), nil
	})
	if err != nil {
		slog.Error("Failed to load oracle_in.OracleCommandHandler.", "err", err)
		panic(err)
	}

	// Oracle Domain - Query Handler
	err = c.Singleton(func() (oracle_in.OracleQueryHandler, error) {
		var repo oracle_out.OracleResultRepository

		if err := c.Resolve(&repo); err != nil {
			slog.Error("Failed to resolve OracleResultRepository for OracleQueryHandler.", "err", err)
			return nil, err
		}

		return oracle_usecases.NewOracleQueryHandler(repo), nil
	})
	if err != nil {
		slog.Error("Failed to load oracle_in.OracleQueryHandler.", "err", err)
		panic(err)
	}

	// Oracle Domain - OCR Stream Config Repository
	err = c.Singleton(func() (oracle_out.OCRStreamConfigRepository, error) {
		var mongoClient *mongo.Client
		var config common.Config

		if err := c.Resolve(&mongoClient); err != nil {
			slog.Error("Failed to resolve mongo.Client for OCRStreamConfigRepository.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve Config for OCRStreamConfigRepository.", "err", err)
			return nil, err
		}

		return db.NewMongoOCRStreamConfigRepository(mongoClient, config.MongoDB.DBName), nil
	})
	if err != nil {
		slog.Error("Failed to load oracle_out.OCRStreamConfigRepository.", "err", err)
		panic(err)
	}

	// Oracle Domain - Match Reconciliation Service (MUST be registered before GameImportCommandHandler)
	err = c.Singleton(func() (*metadata.MatchReconciliationService, error) {
		var matchReader replay_out.MatchMetadataReader
		var matchWriter replay_out.MatchMetadataWriter

		if err := c.Resolve(&matchReader); err != nil {
			slog.Error("Failed to resolve MatchMetadataReader for MatchReconciliationService.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&matchWriter); err != nil {
			slog.Error("Failed to resolve MatchMetadataWriter for MatchReconciliationService.", "err", err)
			return nil, err
		}

		return metadata.NewMatchReconciliationService(matchReader, matchWriter), nil
	})
	if err != nil {
		slog.Error("Failed to load metadata.MatchReconciliationService.", "err", err)
		panic(err)
	}

	// Oracle Domain - Game Import Command Handler
	err = c.Singleton(func() (oracle_in.GameImportCommandHandler, error) {
		var oracleCommandHandler oracle_in.OracleCommandHandler
		var oracleResultRepo oracle_out.OracleResultRepository
		var streamConfigRepo oracle_out.OCRStreamConfigRepository
		var reconciliationService *metadata.MatchReconciliationService
		var matchResultRepo scores_out.MatchResultRepository
		var eventPublisher oracle_out.OracleEventPublisher

		if err := c.Resolve(&oracleCommandHandler); err != nil {
			slog.Error("Failed to resolve OracleCommandHandler for GameImportCommandHandler.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&oracleResultRepo); err != nil {
			slog.Error("Failed to resolve OracleResultRepository for GameImportCommandHandler.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&streamConfigRepo); err != nil {
			slog.Error("Failed to resolve OCRStreamConfigRepository for GameImportCommandHandler.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&reconciliationService); err != nil {
			slog.Error("Failed to resolve MatchReconciliationService for GameImportCommandHandler.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&matchResultRepo); err != nil {
			slog.Error("Failed to resolve MatchResultRepository for GameImportCommandHandler.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&eventPublisher); err != nil {
			slog.Error("Failed to resolve OracleEventPublisher for GameImportCommandHandler.", "err", err)
			return nil, err
		}

		return oracle_usecases.NewGameImportCommandHandler(
			oracleCommandHandler,
			oracleResultRepo,
			streamConfigRepo,
			reconciliationService,
			matchResultRepo,
			eventPublisher,
		), nil
	})
	if err != nil {
		slog.Error("Failed to load oracle_in.GameImportCommandHandler.", "err", err)
		panic(err)
	}

	// Oracle Domain - Game Discovery Service
	err = c.Singleton(func() (*oracle_services.GameDiscoveryService, error) {
		var providers []oracle_out.ExternalScorePort
		var oracleResultRepo oracle_out.OracleResultRepository
		var streamConfigRepo oracle_out.OCRStreamConfigRepository
		var eventPublisher oracle_out.OracleEventPublisher

		if err := c.Resolve(&providers); err != nil {
			slog.Error("Failed to resolve ExternalScorePort providers for GameDiscoveryService.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&oracleResultRepo); err != nil {
			slog.Error("Failed to resolve OracleResultRepository for GameDiscoveryService.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&streamConfigRepo); err != nil {
			slog.Error("Failed to resolve OCRStreamConfigRepository for GameDiscoveryService.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&eventPublisher); err != nil {
			slog.Error("Failed to resolve OracleEventPublisher for GameDiscoveryService.", "err", err)
			return nil, err
		}

		config := oracle_services.DefaultGameDiscoveryConfig()
		return oracle_services.NewGameDiscoveryService(providers, oracleResultRepo, streamConfigRepo, eventPublisher, config), nil
	})
	if err != nil {
		slog.Error("Failed to load GameDiscoveryService.", "err", err)
		panic(err)
	}

	// ========================
	// Oracle OCR Domain
	// ========================

	// OCR Score Parser
	err = c.Singleton(func() (*oracle_services.OCRScoreParser, error) {
		return oracle_services.NewOCRScoreParser(), nil
	})
	if err != nil {
		slog.Error("Failed to load OCRScoreParser.", "err", err)
		panic(err)
	}

	// OCR Stream Capture Port (streamlink + ffmpeg)
	err = c.Singleton(func() (oracle_out.StreamCapturePort, error) {
		quality := os.Getenv("STREAM_QUALITY")
		return oracle_ocr.NewStreamlinkCapture(quality), nil
	})
	if err != nil {
		slog.Error("Failed to load StreamCapturePort.", "err", err)
		panic(err)
	}

	// OCR Engine Port (PaddleOCR)
	err = c.Singleton(func() (oracle_out.OCREnginePort, error) {
		pythonPath := os.Getenv("PADDLEOCR_PYTHON_PATH")
		scriptPath := os.Getenv("PADDLEOCR_SCRIPT_PATH")
		useGPU := os.Getenv("PADDLEOCR_USE_GPU") == "true"
		if scriptPath == "" {
			scriptPath = "/app/scripts/paddleocr_wrapper.py"
		}
		return oracle_ocr.NewPaddleOCRAdapter(pythonPath, scriptPath, useGPU), nil
	})
	if err != nil {
		slog.Error("Failed to load OCREnginePort.", "err", err)
		panic(err)
	}

	// Team Name Resolver (MongoDB-backed)
	err = c.Singleton(func() (oracle_out.TeamResolverPort, error) {
		var mongoClient *mongo.Client
		var config common.Config

		if err := c.Resolve(&mongoClient); err != nil {
			slog.Warn("MongoDB not available for TeamNameResolver", "err", err)
			return nil, nil
		}
		if err := c.Resolve(&config); err != nil {
			slog.Warn("Config not available for TeamNameResolver", "err", err)
			return nil, nil
		}

		db := mongoClient.Database(config.MongoDB.DBName)
		return oracle_ocr.NewTeamNameResolver(db), nil
	})
	if err != nil {
		slog.Error("Failed to load TeamResolverPort.", "err", err)
		panic(err)
	}

	// Stream Monitor
	err = c.Singleton(func() (*oracle_services.StreamMonitor, error) {
		var streamCapture oracle_out.StreamCapturePort
		var ocrEngine oracle_out.OCREnginePort
		var teamResolver oracle_out.TeamResolverPort
		var scoreParser *oracle_services.OCRScoreParser
		var commandHandler oracle_in.OracleCommandHandler

		if err := c.Resolve(&streamCapture); err != nil {
			slog.Error("Failed to resolve StreamCapturePort for StreamMonitor.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&ocrEngine); err != nil {
			slog.Error("Failed to resolve OCREnginePort for StreamMonitor.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&teamResolver); err != nil {
			slog.Warn("TeamResolverPort not available for StreamMonitor", "err", err)
		}
		if err := c.Resolve(&scoreParser); err != nil {
			slog.Error("Failed to resolve OCRScoreParser for StreamMonitor.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&commandHandler); err != nil {
			slog.Error("Failed to resolve OracleCommandHandler for StreamMonitor.", "err", err)
			return nil, err
		}

		return oracle_services.NewStreamMonitor(streamCapture, ocrEngine, teamResolver, scoreParser, commandHandler), nil
	})
	if err != nil {
		slog.Error("Failed to load StreamMonitor.", "err", err)
		panic(err)
	}

	// ========================
	// Messaging Domain
	// ========================

	// Comment Repository
	err = c.Singleton(func() (messaging_out.CommentRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve *mongo.Client for CommentRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for CommentRepository.", "err", err)
			return nil, err
		}

		return db.NewCommentMongoRepository(client.Database(config.MongoDB.DBName)), nil
	})
	if err != nil {
		slog.Error("Failed to load messaging_out.CommentRepository.", "err", err)
		panic(err)
	}

	// Direct Message Repository
	err = c.Singleton(func() (messaging_out.DirectMessageRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve *mongo.Client for DirectMessageRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for DirectMessageRepository.", "err", err)
			return nil, err
		}

		return db.NewDirectMessageMongoRepository(client.Database(config.MongoDB.DBName)), nil
	})
	if err != nil {
		slog.Error("Failed to load messaging_out.DirectMessageRepository.", "err", err)
		panic(err)
	}

	// Team Message Repository
	err = c.Singleton(func() (messaging_out.TeamMessageRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve *mongo.Client for TeamMessageRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for TeamMessageRepository.", "err", err)
			return nil, err
		}

		return db.NewTeamMessageMongoRepository(client.Database(config.MongoDB.DBName)), nil
	})
	if err != nil {
		slog.Error("Failed to load messaging_out.TeamMessageRepository.", "err", err)
		panic(err)
	}

	// Messaging Event Publisher
	err = c.Singleton(func() (messaging_out.MessagingEventPublisher, error) {
		var kafkaClient *kafka.Client
		err := c.Resolve(&kafkaClient)
		if err != nil || kafkaClient == nil {
			slog.Warn("Kafka client not available for MessagingEventPublisher, messaging events will not be published.", "err", err)
			return kafka.NewNoopMessagingEventPublisher(), nil
		}

		return kafka.NewMessagingEventPublisherAdapter(kafkaClient), nil
	})
	if err != nil {
		slog.Warn("Failed to load messaging_out.MessagingEventPublisher — messaging events disabled.", "err", err)
	}

	// Comment Command Use Case
	err = c.Singleton(func() (messaging_in.CommentCommand, error) {
		var commentRepo messaging_out.CommentRepository
		var eventPublisher messaging_out.MessagingEventPublisher

		if err := c.Resolve(&commentRepo); err != nil {
			slog.Error("Failed to resolve CommentRepository for CommentCommandUseCase.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&eventPublisher); err != nil {
			slog.Warn("MessagingEventPublisher not available, comments will work without events.", "err", err)
		}

		return messaging_usecases.NewCommentCommandUseCase(commentRepo, eventPublisher), nil
	})
	if err != nil {
		slog.Error("Failed to load messaging_in.CommentCommand.", "err", err)
		panic(err)
	}

	// Comment Query Use Case
	err = c.Singleton(func() (messaging_in.CommentQuery, error) {
		var commentRepo messaging_out.CommentRepository

		if err := c.Resolve(&commentRepo); err != nil {
			slog.Error("Failed to resolve CommentRepository for CommentQueryUseCase.", "err", err)
			return nil, err
		}

		return messaging_usecases.NewCommentQueryUseCase(commentRepo), nil
	})
	if err != nil {
		slog.Error("Failed to load messaging_in.CommentQuery.", "err", err)
		panic(err)
	}

	// Direct Message Command Use Case
	err = c.Singleton(func() (messaging_in.DirectMessageCommand, error) {
		var dmRepo messaging_out.DirectMessageRepository
		var eventPublisher messaging_out.MessagingEventPublisher

		if err := c.Resolve(&dmRepo); err != nil {
			slog.Error("Failed to resolve DirectMessageRepository for DirectMessageCommandUseCase.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&eventPublisher); err != nil {
			slog.Warn("MessagingEventPublisher not available for DMs.", "err", err)
		}

		return messaging_usecases.NewDirectMessageCommandUseCase(dmRepo, eventPublisher), nil
	})
	if err != nil {
		slog.Error("Failed to load messaging_in.DirectMessageCommand.", "err", err)
		panic(err)
	}

	// Direct Message Query Use Case
	err = c.Singleton(func() (messaging_in.DirectMessageQuery, error) {
		var dmRepo messaging_out.DirectMessageRepository

		if err := c.Resolve(&dmRepo); err != nil {
			slog.Error("Failed to resolve DirectMessageRepository for DirectMessageQueryUseCase.", "err", err)
			return nil, err
		}

		return messaging_usecases.NewDirectMessageQueryUseCase(dmRepo), nil
	})
	if err != nil {
		slog.Error("Failed to load messaging_in.DirectMessageQuery.", "err", err)
		panic(err)
	}

	// squad_in.SquadReader (needed by TeamMessageCommandUseCase - may also be registered in WithSquadAPI)
	_ = c.Singleton(func() (squad_in.SquadReader, error) {
		var squadReader squad_out.SquadReader
		err := c.Resolve(&squadReader)
		if err != nil {
			slog.Error("Failed to resolve squad_out.SquadReader for squad_in.SquadReader (InjectMongoDB).", "err", err)
			return nil, err
		}
		return squad_services.NewSquadQueryService(squadReader), nil
	})

	// Team Message Command Use Case
	err = c.Singleton(func() (messaging_in.TeamMessageCommand, error) {
		var teamMsgRepo messaging_out.TeamMessageRepository
		var squadReader squad_in.SquadReader
		var eventPublisher messaging_out.MessagingEventPublisher

		if err := c.Resolve(&teamMsgRepo); err != nil {
			slog.Error("Failed to resolve TeamMessageRepository for TeamMessageCommandUseCase.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&squadReader); err != nil {
			slog.Error("Failed to resolve SquadReader for TeamMessageCommandUseCase.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&eventPublisher); err != nil {
			slog.Warn("MessagingEventPublisher not available for team messages.", "err", err)
		}

		return messaging_usecases.NewTeamMessageCommandUseCase(teamMsgRepo, squadReader, eventPublisher), nil
	})
	if err != nil {
		slog.Error("Failed to load messaging_in.TeamMessageCommand.", "err", err)
		panic(err)
	}

	// Team Message Query Use Case
	err = c.Singleton(func() (messaging_in.TeamMessageQuery, error) {
		var teamMsgRepo messaging_out.TeamMessageRepository

		if err := c.Resolve(&teamMsgRepo); err != nil {
			slog.Error("Failed to resolve TeamMessageRepository for TeamMessageQueryUseCase.", "err", err)
			return nil, err
		}

		return messaging_usecases.NewTeamMessageQueryUseCase(teamMsgRepo), nil
	})
	if err != nil {
		slog.Error("Failed to load messaging_in.TeamMessageQuery.", "err", err)
		panic(err)
	}

	// ========================
	// Prediction Domain
	// ========================

	// Market Repository
	err = c.Singleton(func() (prediction_out.MarketRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve *mongo.Client for MarketRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for MarketRepository.", "err", err)
			return nil, err
		}

		return db.NewMarketMongoRepository(client.Database(config.MongoDB.DBName)), nil
	})
	if err != nil {
		slog.Error("Failed to load prediction_out.MarketRepository.", "err", err)
		panic(err)
	}

	// Bet Repository
	err = c.Singleton(func() (prediction_out.BetRepository, error) {
		var client *mongo.Client
		var config common.Config

		if err := c.Resolve(&client); err != nil {
			slog.Error("Failed to resolve *mongo.Client for BetRepository.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve common.Config for BetRepository.", "err", err)
			return nil, err
		}

		return db.NewBetMongoRepository(client.Database(config.MongoDB.DBName)), nil
	})
	if err != nil {
		slog.Error("Failed to load prediction_out.BetRepository.", "err", err)
		panic(err)
	}

	// Prediction Event Publisher
	err = c.Singleton(func() (prediction_out.PredictionEventPublisher, error) {
		var kafkaClient *kafka.Client
		err := c.Resolve(&kafkaClient)
		if err != nil || kafkaClient == nil {
			slog.Warn("Kafka client not available for PredictionEventPublisher, prediction events will not be published.", "err", err)
			return kafka.NewNoopPredictionEventPublisher(), nil
		}

		return kafka.NewPredictionEventPublisherAdapter(kafkaClient), nil
	})
	if err != nil {
		slog.Warn("Failed to load prediction_out.PredictionEventPublisher — prediction events disabled.", "err", err)
	}

	// Market Command Use Case
	err = c.Singleton(func() (prediction_in.MarketCommand, error) {
		var marketRepo prediction_out.MarketRepository
		var betRepo prediction_out.BetRepository
		var eventPublisher prediction_out.PredictionEventPublisher

		if err := c.Resolve(&marketRepo); err != nil {
			slog.Error("Failed to resolve MarketRepository for MarketCommandUseCase.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&betRepo); err != nil {
			slog.Error("Failed to resolve BetRepository for MarketCommandUseCase.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&eventPublisher); err != nil {
			slog.Warn("PredictionEventPublisher not available for markets.", "err", err)
		}

		return prediction_usecases.NewMarketCommandUseCase(marketRepo, betRepo, eventPublisher), nil
	})
	if err != nil {
		slog.Error("Failed to load prediction_in.MarketCommand.", "err", err)
		panic(err)
	}

	// Market Query Use Case
	err = c.Singleton(func() (prediction_in.MarketQuery, error) {
		var marketRepo prediction_out.MarketRepository

		if err := c.Resolve(&marketRepo); err != nil {
			slog.Error("Failed to resolve MarketRepository for MarketQueryUseCase.", "err", err)
			return nil, err
		}

		return prediction_usecases.NewMarketQueryUseCase(marketRepo), nil
	})
	if err != nil {
		slog.Error("Failed to load prediction_in.MarketQuery.", "err", err)
		panic(err)
	}

	// Bet Command Use Case
	err = c.Singleton(func() (prediction_in.BetCommand, error) {
		var betRepo prediction_out.BetRepository
		var marketRepo prediction_out.MarketRepository
		var eventPublisher prediction_out.PredictionEventPublisher

		if err := c.Resolve(&betRepo); err != nil {
			slog.Error("Failed to resolve BetRepository for BetCommandUseCase.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&marketRepo); err != nil {
			slog.Error("Failed to resolve MarketRepository for BetCommandUseCase.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&eventPublisher); err != nil {
			slog.Warn("PredictionEventPublisher not available for bets.", "err", err)
		}

		return prediction_usecases.NewBetCommandUseCase(betRepo, marketRepo, eventPublisher), nil
	})
	if err != nil {
		slog.Error("Failed to load prediction_in.BetCommand.", "err", err)
		panic(err)
	}

	// Bet Query Use Case
	err = c.Singleton(func() (prediction_in.BetQuery, error) {
		var betRepo prediction_out.BetRepository
		var marketRepo prediction_out.MarketRepository

		if err := c.Resolve(&betRepo); err != nil {
			slog.Error("Failed to resolve BetRepository for BetQueryUseCase.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&marketRepo); err != nil {
			slog.Error("Failed to resolve MarketRepository for BetQueryUseCase.", "err", err)
			return nil, err
		}

		return prediction_usecases.NewBetQueryUseCase(betRepo, marketRepo), nil
	})
	if err != nil {
		slog.Error("Failed to load prediction_in.BetQuery.", "err", err)
		panic(err)
	}

	// =============================================
	// Analytics - View Tracking & Insights
	// =============================================

	// Entity View Repository (implements EntityViewWriter + EntityViewReader)
	err = c.Singleton(func() (*db.EntityViewRepository, error) {
		var mongoClient *mongo.Client
		if err := c.Resolve(&mongoClient); err != nil {
			slog.Error("Failed to resolve mongo.Client for EntityViewRepository.", "err", err)
			return nil, err
		}
		var config common.Config
		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve config for EntityViewRepository.", "err", err)
			return nil, err
		}
		return db.NewEntityViewRepository(mongoClient, config.MongoDB.DBName), nil
	})
	if err != nil {
		slog.Error("Failed to load EntityViewRepository.", "err", err)
		panic(err)
	}

	// View Statistics Repository
	err = c.Singleton(func() (*db.ViewStatisticsRepository, error) {
		var mongoClient *mongo.Client
		if err := c.Resolve(&mongoClient); err != nil {
			slog.Error("Failed to resolve mongo.Client for ViewStatisticsRepository.", "err", err)
			return nil, err
		}
		var config common.Config
		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve config for ViewStatisticsRepository.", "err", err)
			return nil, err
		}
		return db.NewViewStatisticsRepository(mongoClient, config.MongoDB.DBName), nil
	})
	if err != nil {
		slog.Error("Failed to load ViewStatisticsRepository.", "err", err)
		panic(err)
	}

	// Viewer Insight Repository
	err = c.Singleton(func() (*db.ViewerInsightRepository, error) {
		var mongoClient *mongo.Client
		if err := c.Resolve(&mongoClient); err != nil {
			slog.Error("Failed to resolve mongo.Client for ViewerInsightRepository.", "err", err)
			return nil, err
		}
		var config common.Config
		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve config for ViewerInsightRepository.", "err", err)
			return nil, err
		}
		return db.NewViewerInsightRepository(mongoClient, config.MongoDB.DBName), nil
	})
	if err != nil {
		slog.Error("Failed to load ViewerInsightRepository.", "err", err)
		panic(err)
	}

	// View Privacy Repository
	err = c.Singleton(func() (*db.ViewPrivacyRepository, error) {
		var mongoClient *mongo.Client
		if err := c.Resolve(&mongoClient); err != nil {
			slog.Error("Failed to resolve mongo.Client for ViewPrivacyRepository.", "err", err)
			return nil, err
		}
		var config common.Config
		if err := c.Resolve(&config); err != nil {
			slog.Error("Failed to resolve config for ViewPrivacyRepository.", "err", err)
			return nil, err
		}
		return db.NewViewPrivacyRepository(mongoClient, config.MongoDB.DBName), nil
	})
	if err != nil {
		slog.Error("Failed to load ViewPrivacyRepository.", "err", err)
		panic(err)
	}

	// View Event Publisher (Kafka)
	err = c.Singleton(func() (analytics_out.ViewEventPublisher, error) {
		var eventPublisher *kafka.EventPublisher
		if err := c.Resolve(&eventPublisher); err != nil {
			slog.Error("Failed to resolve EventPublisher for ViewEventPublisherAdapter.", "err", err)
			return nil, err
		}
		return kafka.NewViewEventPublisherAdapter(eventPublisher), nil
	})
	if err != nil {
		slog.Error("Failed to load ViewEventPublisher.", "err", err)
		panic(err)
	}

	// RecordViewCommandHandler use case
	err = c.Singleton(func() (analytics_in.RecordViewCommandHandler, error) {
		var entityViewRepo *db.EntityViewRepository
		if err := c.Resolve(&entityViewRepo); err != nil {
			slog.Error("Failed to resolve EntityViewRepository for RecordEntityViewUseCase.", "err", err)
			return nil, err
		}
		var viewEventPublisher analytics_out.ViewEventPublisher
		if err := c.Resolve(&viewEventPublisher); err != nil {
			slog.Error("Failed to resolve ViewEventPublisher for RecordEntityViewUseCase.", "err", err)
			return nil, err
		}
		var viewPrivacyRepo *db.ViewPrivacyRepository
		if err := c.Resolve(&viewPrivacyRepo); err != nil {
			slog.Error("Failed to resolve ViewPrivacyRepository for RecordEntityViewUseCase.", "err", err)
			return nil, err
		}
		return analytics_usecases.NewRecordEntityViewUseCase(entityViewRepo, entityViewRepo, viewEventPublisher, viewPrivacyRepo), nil
	})
	if err != nil {
		slog.Error("Failed to load RecordViewCommandHandler.", "err", err)
		panic(err)
	}

	// ViewStatisticsQueryHandler use case
	err = c.Singleton(func() (analytics_in.ViewStatisticsQueryHandler, error) {
		var viewStatsRepo *db.ViewStatisticsRepository
		if err := c.Resolve(&viewStatsRepo); err != nil {
			slog.Error("Failed to resolve ViewStatisticsRepository for GetViewStatisticsUseCase.", "err", err)
			return nil, err
		}
		return analytics_usecases.NewGetViewStatisticsUseCase(viewStatsRepo), nil
	})
	if err != nil {
		slog.Error("Failed to load ViewStatisticsQueryHandler.", "err", err)
		panic(err)
	}

	// ViewInsightsQueryHandler use case
	err = c.Singleton(func() (analytics_in.ViewInsightsQueryHandler, error) {
		var viewerInsightRepo *db.ViewerInsightRepository
		if err := c.Resolve(&viewerInsightRepo); err != nil {
			slog.Error("Failed to resolve ViewerInsightRepository for GetViewInsightsUseCase.", "err", err)
			return nil, err
		}
		var viewPrivacyRepo *db.ViewPrivacyRepository
		if err := c.Resolve(&viewPrivacyRepo); err != nil {
			slog.Error("Failed to resolve ViewPrivacyRepository for GetViewInsightsUseCase.", "err", err)
			return nil, err
		}
		return analytics_usecases.NewGetViewInsightsUseCase(viewerInsightRepo, viewPrivacyRepo), nil
	})
	if err != nil {
		slog.Error("Failed to load ViewInsightsQueryHandler.", "err", err)
		panic(err)
	}

	// MyAnalyticsQueryHandler use case
	err = c.Singleton(func() (analytics_in.MyAnalyticsQueryHandler, error) {
		var viewStatsRepo *db.ViewStatisticsRepository
		if err := c.Resolve(&viewStatsRepo); err != nil {
			slog.Error("Failed to resolve ViewStatisticsRepository for GetMyAnalyticsUseCase.", "err", err)
			return nil, err
		}
		return analytics_usecases.NewGetMyAnalyticsUseCase(viewStatsRepo), nil
	})
	if err != nil {
		slog.Error("Failed to load MyAnalyticsQueryHandler.", "err", err)
		panic(err)
	}

	// UpdateViewPrivacyCommandHandler use case
	err = c.Singleton(func() (analytics_in.UpdateViewPrivacyCommandHandler, error) {
		var viewPrivacyRepo *db.ViewPrivacyRepository
		if err := c.Resolve(&viewPrivacyRepo); err != nil {
			slog.Error("Failed to resolve ViewPrivacyRepository for UpdateViewPrivacyUseCase.", "err", err)
			return nil, err
		}
		return analytics_usecases.NewUpdateViewPrivacyUseCase(viewPrivacyRepo, viewPrivacyRepo), nil
	})
	if err != nil {
		slog.Error("Failed to load UpdateViewPrivacyCommandHandler.", "err", err)
		panic(err)
	}

	// AnalyticsEventConsumer — Kafka consumer factory
	// This registers a factory function that the replay-worker can resolve
	// to instantiate the analytics consumer with a running Kafka client.
	err = c.Singleton(func() (*kafka.AnalyticsEventConsumerConfig, error) {
		var viewStatsRepo *db.ViewStatisticsRepository
		if err := c.Resolve(&viewStatsRepo); err != nil {
			slog.Error("Failed to resolve ViewStatisticsRepository for AnalyticsEventConsumer.", "err", err)
			return nil, err
		}
		var viewerInsightRepo *db.ViewerInsightRepository
		if err := c.Resolve(&viewerInsightRepo); err != nil {
			slog.Error("Failed to resolve ViewerInsightRepository for AnalyticsEventConsumer.", "err", err)
			return nil, err
		}
		return &kafka.AnalyticsEventConsumerConfig{
			GroupID:       "analytics-aggregation-group",
			StatsWriter:   viewStatsRepo,
			StatsReader:   viewStatsRepo,
			InsightWriter: viewerInsightRepo,
			InsightReader: viewerInsightRepo,
		}, nil
	})
	if err != nil {
		slog.Error("Failed to load AnalyticsEventConsumerConfig.", "err", err)
		panic(err)
	}

	return nil
}

func (b *ContainerBuilder) With(resolver interface{}) *ContainerBuilder {
	c := b.Container

	err := c.Singleton(resolver)

	if err != nil {
		slog.Error("Failed to register resolver.", "err", err)
		panic(err)
	}

	return b
}

func (b *ContainerBuilder) Close(c container.Container) {
	var client *mongo.Client
	err := c.Resolve(&client)

	if client != nil && err == nil {
		_ = client.Disconnect(context.TODO())
	}
}

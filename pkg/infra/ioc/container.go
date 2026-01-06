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
	tournament_in "github.com/replay-api/replay-api/pkg/domain/tournament/ports/in"
	tournament_out "github.com/replay-api/replay-api/pkg/domain/tournament/ports/out"
	tournament_services "github.com/replay-api/replay-api/pkg/domain/tournament/services"
	tournament_usecases "github.com/replay-api/replay-api/pkg/domain/tournament/usecases"

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

	media_out "github.com/replay-api/replay-api/pkg/domain/media/ports/out"
	media_adapter "github.com/replay-api/replay-api/pkg/infra/adapters/media"

	websocket "github.com/replay-api/replay-api/pkg/infra/websocket"

	iam_in "github.com/replay-api/replay-api/pkg/domain/iam/ports/in"
	iam_out "github.com/replay-api/replay-api/pkg/domain/iam/ports/out"
	iam_query_services "github.com/replay-api/replay-api/pkg/domain/iam/services"

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

		return replay_use_cases.NewUploadReplayFileUseCase(ReplayFileMetadataWriter, replayDataWriter), nil
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

	// Kafka Event Publisher (dummy for local development)
	err := c.Singleton(func() (*kafka.EventPublisher, error) {
		// EventPublisher handles nil client gracefully for local development
		return &kafka.EventPublisher{}, nil
	})

	if err != nil {
		slog.Error("Failed to load *kafka.EventPublisher.", "err", err)
		panic(err)
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

		repo := db.NewReplayFileMetadataRepository(client, config.MongoDB.DBName, replay_entity.ReplayFile{}, "replay_file_metadata")

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
		return db.NewReplayFileContentRepository(client), nil
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

		return db.NewReplayFileContentRepository(client), nil
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
	err = c.Singleton(func() (*wallet_services.LedgerService, error) {
		var auditTrail billing_in.AuditTrailCommand

		if err := c.Resolve(&auditTrail); err != nil {
			slog.Error("Failed to resolve billing_in.AuditTrailCommand for LedgerService.", "err", err)
			return nil, err
		}

		// Use no-op repository for now - TODO: implement proper MongoDB repository
		ledgerRepo := wallet_services.NewNoOpLedgerRepository()
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

	// Wallet Service (no-op for basic functionality)
	err = c.Singleton(func() (wallet_in.WalletCommand, error) {
		return &NoOpWalletCommand{}, nil
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

	// Tournament Command Service
	err = c.Singleton(func() (tournament_in.TournamentCommand, error) {
		var tournamentRepo tournament_out.TournamentRepository
		var walletCmd wallet_in.WalletCommand

		if err := c.Resolve(&tournamentRepo); err != nil {
			slog.Error("Failed to resolve tournament_out.TournamentRepository for TournamentService.", "err", err)
			return nil, err
		}

		if err := c.Resolve(&walletCmd); err != nil {
			slog.Error("Failed to resolve wallet_in.WalletCommand for TournamentService.", "err", err)
			return nil, err
		}

		return tournament_services.NewTournamentService(tournamentRepo, walletCmd), nil
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

		if err := c.Resolve(&tournamentQuerySvc); err != nil {
			slog.Error("Failed to resolve TournamentQueryService for TournamentReaderService.", "err", err)
			return nil, err
		}

		return tournament_services.NewTournamentReaderService(tournamentQuerySvc), nil
	})

	if err != nil {
		slog.Error("Failed to load tournament_in.TournamentReader.", "err", err)
		panic(err)
	}

	// Matchmaking Usecases
	err = c.Singleton(func() (matchmaking_in.JoinMatchmakingQueueCommandHandler, error) {
		var billableOperationHandler billing_in.BillableOperationCommandHandler
		var sessionRepo matchmaking_out.MatchmakingSessionRepository
		var eventPublisher *kafka.EventPublisher

		if err := c.Resolve(&billableOperationHandler); err != nil {
			slog.Error("Failed to resolve BillableOperationCommandHandler for JoinMatchmakingQueueUseCase.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&sessionRepo); err != nil {
			slog.Error("Failed to resolve MatchmakingSessionRepository for JoinMatchmakingQueueUseCase.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&eventPublisher); err != nil {
			slog.Error("Failed to resolve EventPublisher for JoinMatchmakingQueueUseCase.", "err", err)
			return nil, err
		}

		return matchmaking_usecases.NewJoinMatchmakingQueueUseCase(billableOperationHandler, sessionRepo, eventPublisher), nil
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
		// var playerProfileRepo *db.PlayerProfileRepository // TODO: Re-enable once PlayerProfileRepository is properly registered

		if err := c.Resolve(&billableOperationHandler); err != nil {
			slog.Error("Failed to resolve BillableOperationCommandHandler for RegisterForTournamentUseCase.", "err", err)
			return nil, err
		}
		if err := c.Resolve(&tournamentRepo); err != nil {
			slog.Error("Failed to resolve TournamentRepository for RegisterForTournamentUseCase.", "err", err)
			return nil, err
		}
		// if err := c.Resolve(&playerProfileRepo); err != nil {
		// 	slog.Error("Failed to resolve PlayerProfileRepository for RegisterForTournamentUseCase.", "err", err)
		// 	return nil, err
		// }

		return tournament_usecases.NewRegisterForTournamentUseCase(billableOperationHandler, tournamentRepo), nil
	})
	if err != nil {
		slog.Error("Failed to load RegisterForTournamentUseCase.", "err", err)
		panic(err)
	}

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

	// -----

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

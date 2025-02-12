package ioc

import (
	"context"
	"log/slog"
	"os"

	// env
	"github.com/joho/godotenv"

	// mongodb
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	// repositories/db
	db "github.com/psavelis/team-pro/replay-api/pkg/infra/db/mongodb"

	// messageBroker (kafka/rabbit)

	// encryption
	encryption "github.com/psavelis/team-pro/replay-api/pkg/infra/crypto"

	// container
	container "github.com/golobby/container/v3"

	// local files

	// ports
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	google_in "github.com/psavelis/team-pro/replay-api/pkg/domain/google/ports/in"
	google_out "github.com/psavelis/team-pro/replay-api/pkg/domain/google/ports/out"
	google_use_cases "github.com/psavelis/team-pro/replay-api/pkg/domain/google/use_cases"
	metadata "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/services/metadata"
	squad_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/entities"
	squad_in "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/in"
	squad_out "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/ports/out"
	squad_services "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/services"
	squad_usecases "github.com/psavelis/team-pro/replay-api/pkg/domain/squad/usecases"

	replay_in "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/ports/in"
	replay_out "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/ports/out"

	steam_in "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/ports/in"
	steam_out "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/ports/out"
	steam_query_services "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/services"

	iam_in "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/in"
	iam_out "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/out"
	iam_query_services "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/services"

	// domain
	google_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/google/entities"
	iam_entities "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
	replay_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/entities"
	steam_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/entities"

	// app
	cs_app "github.com/psavelis/team-pro/replay-api/pkg/app/cs"

	// usecases
	iam_use_cases "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/use_cases"
	replay_use_cases "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/use_cases"
	steam_use_cases "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/use_cases"
)

type ContainerBuilder struct {
	Container container.Container
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
		err := godotenv.Load()
		if err != nil {
			slog.Error("Failed to load .env file")
			panic(err)
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

	err = c.Singleton(func() (*db.SquadHistoryRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for NamedSingleton SquadHistoryRepository as generic MongoDBRepository.", "err", err)
			return &db.SquadHistoryRepository{}, err
		}

		var config common.Config

		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for db.SquadHistoryRepository.", "err", err)
			return nil, err
		}

		repo := db.NewSquadHistoryRepository(client, config.MongoDB.DBName, squad_entities.SquadHistory{}, "squad_histories")

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load NamedSingleton SquadHistoryRepository as generic MongoDBRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (squad_out.SquadHistoryWriter, error) {
		var repo *db.SquadHistoryRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve SquadHistoryRepository for squad_out.SquadHistoryWriter.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load squad_out.SquadHistoryWriter.", "err", err)
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
		var playerProfileReader squad_out.PlayerProfileReader
		err := c.Resolve(&playerProfileReader)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileReader for PlayerProfileQueryService.", "err", err)
			return nil, err
		}

		return squad_services.NewPlayerProfileQueryService(playerProfileReader), nil
	})

	if err != nil {
		slog.Error("Failed to load PlayerProfileReader.")
		panic(err)
	}

	err = c.Singleton(func() (squad_in.CreatePlayerProfileCommandHandler, error) {
		var playerWriter squad_out.PlayerProfileWriter
		err := c.Resolve(&playerWriter)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileWriter for CreatePlayerProfileCommandHandler.", "err", err)
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

		var playerProfileReader squad_out.PlayerProfileReader
		err = c.Resolve(&playerProfileReader)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileReader for CreatePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		uc := squad_usecases.NewCreatePlayerProfileUseCase(playerWriter, playerProfileReader, groupWriter, groupReader, playerProfileHistoryWriter)

		return uc, nil
	})

	if err != nil {
		slog.Error("Failed to load CreatePlayerProfileCommand.", "err", err)
		panic(err)
	}

	// squad_in.CreateSquadCommandHandler
	err = c.Singleton(func() (squad_in.CreateSquadCommandHandler, error) {
		var squadWriter squad_out.SquadWriter
		err := c.Resolve(&squadWriter)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileWriter for CreatePlayerProfileCommandHandler.", "err", err)
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
			slog.Error("Failed to resolve GroupWriter for CreatePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		var groupReader iam_out.GroupReader
		err = c.Resolve(&groupReader)
		if err != nil {
			slog.Error("Failed to resolve GroupReader for CreatePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		var squadHistoryWriter squad_out.SquadHistoryWriter
		err = c.Resolve(&squadHistoryWriter)
		if err != nil {
			slog.Error("Failed to resolve SquadHistoryWriter for CreateSquadCommandHandler.", "err", err)
			return nil, err
		}

		var playerProfileReader squad_in.PlayerProfileReader
		err = c.Resolve(&playerProfileReader)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileReader for CreateSquadCommandHandler.", "err", err)
			return nil, err
		}

		cmdHandler := squad_usecases.NewCreateSquadUseCase(squadWriter, squadHistoryWriter, squadReader, groupWriter, groupReader, playerProfileReader)

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

	// squad_in.CreatePlayerProfileCommandHandler
	err = c.Singleton(func() (squad_in.CreatePlayerProfileCommandHandler, error) {
		var playerProfileWriter squad_out.PlayerProfileWriter
		err := c.Resolve(&playerProfileWriter)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileWriter for CreatePlayerProfileCommandHandler.", "err", err)
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

		var playerProfileReader squad_out.PlayerProfileReader
		err = c.Resolve(&playerProfileReader)
		if err != nil {
			slog.Error("Failed to resolve PlayerProfileReader for CreatePlayerProfileCommandHandler.", "err", err)
			return nil, err
		}

		cmdHandler := squad_usecases.NewCreatePlayerProfileUseCase(playerProfileWriter, playerProfileReader, groupWriter, groupReader, playerProfileHistoryWriter)

		return cmdHandler, nil
	})

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

func InjectMongoDB(c container.Container) error {
	err := c.Singleton(func() (*mongo.Client, error) {
		var config common.Config

		err := c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for mongo.Client.", "err", err)
			return nil, err
		}

		mongoOptions := options.Client().ApplyURI(config.MongoDB.URI).SetRegistry(db.MongoRegistry).SetMaxPoolSize(100)

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

		repo := db.NewProfileRepository(client, config.MongoDB.DBName, &iam_entities.Profile{}, "profiles")

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
		client.Disconnect(context.TODO())
	}
}

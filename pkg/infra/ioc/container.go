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
	db "github.com/psavelis/team-pro/replay-api/pkg/infra/db/mongodb"

	// messageBroker (kafka/rabbit)

	// encryption
	encryption "github.com/psavelis/team-pro/replay-api/pkg/infra/crypto"

	// container
	container "github.com/golobby/container/v3"

	// local files

	// ports
	common "github.com/psavelis/team-pro/replay-api/pkg/domain"
	metadata "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/services/metadata"

	replay_in "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/ports/in"
	replay_out "github.com/psavelis/team-pro/replay-api/pkg/domain/replay/ports/out"

	steam_in "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/ports/in"
	steam_out "github.com/psavelis/team-pro/replay-api/pkg/domain/steam/ports/out"

	iam_in "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/in"
	iam_out "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/ports/out"

	// domain
	iam_entity "github.com/psavelis/team-pro/replay-api/pkg/domain/iam/entities"
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

		return steam_use_cases.NewOnboardSteamUserUseCase(steamUserWriter, steamUserReader, vHashWriter), nil
	})

	if err != nil {
		slog.Error("Failed to load OnboardSteamUserCommand.", "err", err)
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

		mongoOptions := options.Client().ApplyURI(config.MongoDB.URI)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := mongo.Connect(ctx, mongoOptions)

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

		repo := db.NewEventsRepository(client, config.MongoDB.DBName, replay_entity.GameEvent{}, "game_events")

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
	err = c.Singleton(func() (*db.PlayerRepository, error) {
		var client *mongo.Client
		err := c.Resolve(&client)
		if err != nil {
			slog.Error("Failed to resolve mongo.Client for PlayerRepository as generic MongoDBRepository.", "err", err)
			return &db.PlayerRepository{}, err
		}

		var config common.Config

		err = c.Resolve(&config)
		if err != nil {
			slog.Error("Failed to resolve config for db.PlayerRepository.", "err", err)
			return nil, err
		}

		repo := db.NewPlayerRepository(client, config.MongoDB.DBName, replay_entity.Player{}, "player_metadata")

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load PlayerRepository as generic MongoDBRepository.", "err", err)
		panic(err)
	}

	err = c.Singleton(func() (replay_out.PlayerMetadataReader, error) {
		var repo *db.PlayerRepository
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
		var repo *db.PlayerRepository
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

	err = c.Singleton(func() (replay_in.PlayerReader, error) {
		var repo *db.PlayerRepository
		err = c.Resolve(&repo)
		if err != nil {
			slog.Error("Failed to resolve PlayerRepository for replay_in.PlayerReader.", "err", err)
			return nil, err
		}

		return repo, nil
	})

	if err != nil {
		slog.Error("Failed to load replay_in.PlayerReader.", "err", err)
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

		repo := db.NewRIDTokenRepository(client, config.MongoDB.DBName, iam_entity.RIDToken{}, "rid")

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

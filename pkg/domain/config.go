package common

type SteamConfig struct {
	SteamKey    string
	PublicKey   string
	Certificate string
	VHashSource string
}

type BattleNetConfig struct {
	BattleNetKey    string
	BattleNetSecret string
}

type GitHubConfig struct {
	GitHubKey    string
	GitHubSecret string
}

type AuthConfig struct {
	SteamConfig     SteamConfig
	BattleNetConfig BattleNetConfig
	GitHubConfig    GitHubConfig
}

type MongoDBConfig struct {
	DBName      string
	URI         string
	PublicKey   string
	Certificate string
}

type Config struct {
	Auth    AuthConfig
	MongoDB MongoDBConfig
	S3      S3Config
}

type S3Config struct {
	S3Endpoint string
	// AccessKeyID     string
	Region string
	Bucket string
}

type KafkaConfig struct {
	// Kafka bootstrap brokers to connect to, as a comma separated list (ie: "kafka1:9092,kafka2:9092")
	Brokers string

	// Kafka cluster version (ie.: "2.1.1", "2.2.2", "2.3.0", ...)
	Version string

	// Kafka consumer group definition (ie: consumer group name)
	Group string

	// Kafka topics to be consumed, as a comma separated list (ie: "topic1,topic2,topic3")
	Topics string

	// Consumer group partition assignment strategy (ie: range, roundrobin, sticky)
	AssignmentStrategy string

	// Kafka consumer consume initial offset from oldest (default: true)
	Oldest bool

	// Sarama logging (default: false)
	Verbose bool
}

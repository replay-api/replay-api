![Replay API](https://media.licdn.com/dms/image/v2/D4E3DAQFMsKxbj7Rgbw/image-scale_191_1128/image-scale_191_1128/0/1737679675333/leetgaming_pro_cover?e=1739401200&v=beta&t=y4dgt-FDwO7OqEpZgDwTvbDyZqLfJanYJOvI9scDTEc)

# Replay API

Welcome to the Replay API project! This is the main project of our organization, providing robust and scalable solutions for managing and analyzing replay data from various games.

## Overview

The Replay API is designed to handle large volumes of replay data, offering features such as:

- **Replay Parsing**: Efficiently parse and process replay files from different games.
- **Data Storage**: Store parsed data in a structured format using MongoDB.
- **Search and Query**: Powerful search and query capabilities to retrieve specific data points.
- **Event Handling**: Handle various game events and generate meaningful insights.
- **Machine Learning Integration**: Evaluate player talent and performance using machine learning models.

## Features

### Replay Parsing

The Replay API supports parsing replay files from multiple games, extracting valuable data such as player actions, game events, and match statistics.

### Data Storage

Parsed data is stored in MongoDB, allowing for efficient retrieval and analysis. The repository pattern is used to abstract database operations.

### Search and Query

The API provides advanced search and query capabilities, enabling users to filter and retrieve specific data points based on various criteria.

### Event Handling

The API handles different game events, such as match start, round end, and clutch situations, generating insights and statistics for each event.

### Machine Learning Integration

The API integrates with machine learning models to evaluate player talent and performance, providing scores and insights based on in-game actions and statistics.

## Getting Started

### Prerequisites

- Go 1.16+ (Recommended using 1.23+)
- MongoDB

### Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/replay-api/replay-api.git
   cd replay-api
   ```

2. Install dependencies:
   ```sh
   go mod tidy
   ```

3. Set up MongoDB:
   - Ensure MongoDB is running and accessible.
   - Update the MongoDB connection settings in the configuration file.

### Running the API

1. Build the project:
   ```sh
   go build -o replay-api
   ```

2. Run the API:
   ```sh
   ./replay-api
   ```

### Testing

Run the tests using the following command:
```sh
go test ./...
```

### Using Docker

We recommend using Docker to run the Replay API for a consistent and isolated environment. You can build and run the Docker container using the following command:
```sh
make docker
```

### Docker Compose Setup

To set up the Replay API using Docker Compose, follow these steps:

1. Copy the `docker-compose.example.yml` file to `docker-compose.yml`:
    ```sh
    cp docker-compose.example.yml docker-compose.yml
    ```

2. Create a `.env` file and set the necessary environment variables:
    ```sh
    cp .env.example .env
    # Edit the .env file to set your environment variables
    ```

3. Start the services using Docker Compose:
    ```sh
    docker-compose up
    ```

## Contributing

We welcome contributions from the community! Please read our [Contributing Guide](CONTRIBUTING.md) to get started.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contact

For any questions or inquiries, please contact us at [staff@leetgaming.pro](mailto:staff@leetgaming.pro).

---

Thank you for using the Replay API! We hope it helps you manage and analyze your replay data effectively.

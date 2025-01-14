
## CS:GO/CS:2 Replay-API

This repository holds the source code for a project focused on processing replays from Counter-Strike: Global Offensive (CS:GO) and Counter-Strike: Source (CS:2). 

### Project Overview

This project aims to analyze and extract valuable insights from CS:GO and CS:2 replays. This can be used for various purposes, including:

* **Player and Team Performance Analysis:** Track player and team statistics, identify strengths and weaknesses, and measure progress over time.
* **Strategy Evaluation:** Analyze gameplay strategies used in replays and their effectiveness.
* **Content Creation:** Generate data-driven content for CS:GO/CS:2 communities, such as highlight reels or statistical breakdowns.
* **Machine Learning Applications:** Train machine learning models on replay data to predict outcomes, personalize experiences, or detect suspicious activity.

### Directory Structure

The project is organized into the following directories:

```
[replay-api]
│
├───.blob                      
├───.coverage       
├───.docs   
│   ├───JIRA
│   └───steam
├───.github
├───.mongodb
│   └───journal
├───.steam
│   └───edge-instance
├───.terraform
│   ├───gameserver
│   └───tencent
├───cmd
│   ├───async-api
│   ├───cli
│   └───rest-api
│       ├───controllers
│       │   ├───command
│       │   └───query
│       ├───middlewares
│       └───routing
├───mongodb
│   ├───diagnostic.data
│   └───journal
├───pkg
│   ├───app
│   │   ├───ai
│   │   │   └───evaluators
│   │   └───cs
│   │       ├───builders
│   │       ├───factories
│   │       ├───handlers
│   │       └───state
│   ├───domain
│   │   ├───challenge
│   │   ├───cs
│   │   │   └───entities
│   │   ├───iam
│   │   │   ├───entities
│   │   │   ├───ports
│   │   │   │   ├───in
│   │   │   │   └───out
│   │   │   └───use_cases
│   │   ├───replay
│   │   │   ├───entities
│   │   │   ├───ports
│   │   │   │   ├───in
│   │   │   │   └───out
│   │   │   ├───services
│   │   │   │   └───metadata
│   │   │   └───use_cases
│   │   ├───squad
│   │   │   └───entities
│   │   └───steam
│   │       ├───entities
│   │       ├───ports
│   │       │   ├───in
│   │       │   └───out
│   │       └───use_cases
│   └───infra
│       ├───blob
│       │   ├───local
│       │   └───_s3
│       ├───clients
│       ├───crypto
│       ├───db
│       │   ├───elastic
│       │   └───mongodb
│       ├───events
│       └───ioc
└───test
    ├───cmd
    │   └───rest-api-test
    └───sample_replays
        └───cs2
```
* **.blob**: Stores temporary data during processing.
* **.coverage**: Stores code coverage reports generated by test execution.
* **.docs**: Holds documentation for the project, including Jira tickets and Steam API references.
* **.github**: Contains configuration files for GitHub CI/CD workflows.
* **.mongodb**: Stores configuration files for the MongoDB connection.
* **.steam**: Holds Steam API credentials (securely stored).
* **cmd**: Contains source code for command-line tools:
    * **async-api**: Handles asynchronous APIs for processing replays.
    * **cli**: Provides command-line interface for interacting with the application.
    * **rest-api**: Code for the REST API that exposes data and functionalities.
* **mongodb**: Contains code specific to interacting with the MongoDB database:
    * **diagnostic.data**:  (Optional) Directory for storing diagnostic data from the database.
    * **journal**: Holds code for accessing and managing the MongoDB journal.
* **pkg**: Contains compiled Go code for the application:
    * **app**: Main application logic for processing replays.
        * **ai**: Code for AI-powered functionalities, like strategy evaluation (if applicable).
        * **cs**: Core functionalities related to CS:GO/CS:2 replays.
            * **builders**: Create data structures from parsed replay data.
            * **factories**: Generate objects based on extracted data.
            * **handlers**: Handle various replay actions (e.g., parsing, analysis).
            * **state**: Manages the application state during processing.
        * **domain**: Domain logic for the project.
            * **challenge**: Represents challenges (e.g., tactical analysis) and their solutions.
            * **cs**: Core CS:GO/CS:2 domain entities.
                * **entities**: Represents entities like players, teams, and rounds.
            * **iam**:  (Optional) Identity and Access Management related code.
                * **entities**: IAM related entities like users and roles.
                * **ports**: Ports for accessing IAM services.
                    * **in**: Input ports for interacting with IAM services.
                    * **out**: Output ports for receiving data from IAM services.
                * **use_cases**: Use cases for IAM functionalities.
            * **replay**: Represents replay data and its attributes.
                * **entities**: Entities related to replay data.
                * **ports**: Ports for accessing replay data.
                    * **in**: Input ports for loading and parsing replay data.
                    * **out**: Output ports for exporting analyzed replay data.
                * **services**: Services for processing replays.
                    * **metadata**: Processes replay metadata.
                * **use_cases**: Use cases for replay processing tasks.
            * **squad**: Represents squads within a team (if applicable).
                * **entities**: Entities related to squads.
            * **steam**: Steam related functionalities.
                * **entities**: Steam related entities.
                * **ports**: Ports for interacting with Steam API.
                    * **in**: Input ports for querying Steam API.
                    * **out**: Output ports for receiving data from Steam API.
                * **use_cases**: Use cases for Steam API interactions.
        * **steam**: Holds code for interacting with Steam services.
* **test**: Contains test cases for the application:

* **cmd/rest-api-test**: Tests for the REST API endpoints.
* **sample_replays**: Sample CS:GO/CS:2 replays for testing purposes.

### Getting Started

**Prerequisites:**

#### Locally
* Go installed (version 1.22 or later)
* MongoDB database

#### Docker
* Make
* Docker

**Installation:**

1. Clone the repository:
   ```bash
   git clone https://github.com/replay-api/replay-api.git
   ```
2. Install dependencies:
   ```bash
   cd replay-api
   go mod tidy
   ```
3. Set up environment variables: `~ Work in progress`

**Running the Application:**

Start
```sh
## Build & Run Services
make docker
```

Stop
```sh
## Remove all instances
make docker-down
```

### Usage

**REST API:**
#### Search API
* **Endpoint:** `/search/{query:.*}`
  * **GET:** Searchable interface for any entity (i.e.: `/search/players?tags=xyz`)

#### Replay API
* **Endpoint:** `/games/{game_id}/replays`
  * **POST:** Upload a replay file.
* **Endpoint:** `/games/{game_id}/replay/{replay_file_id}`
  * **GET:** Retrieve processed replay data.

**CLI Tool:** `Work in progress / Contributions welcome :)`

* **Process Replay:** `process-replay <replay_file>`
* **Analyze Player:** `analyze-player <player_name>`
* **Compare Teams:** `compare-teams <team1> <team2>`
-----

# Contributing

We welcome contributions to this project. Please follow these guidelines:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes and commit them.
4. Push your changes to your fork.
5. Create a pull request to the main repository.

### Slack:
Invites at https://join.slack.com/t/leetgamingpro/shared_invite/zt-2o0uze27m-InqqtOtZF3hjl_dhjzEiuA

##### Slack Relay:
n3c8f5r9e8b0p0y3@leetgamingpro.slack.com


### JIRA (Project Management & Issue Tracking)

[Project Documentation & Links at .docs folder](https://github.com/psavelis/replay-api/tree/main/.docs)

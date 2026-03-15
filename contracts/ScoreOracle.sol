// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title ScoreOracle
 * @notice On-chain score oracle for esports match results.
 *         Allows authorized publishers to submit and finalize match scores.
 * @dev Deployed on Polygon Amoy testnet (chainId 80002).
 */
contract ScoreOracle {
    // --- Types ---

    struct MatchScore {
        bytes32 externalMatchId;    // keccak256 of external match identifier
        bytes32 teamAId;            // keccak256 of team A identifier
        bytes32 teamBId;            // keccak256 of team B identifier
        uint16  teamAScore;
        uint16  teamBScore;
        bytes32 winnerId;           // bytes32(0) if draw
        bool    isDraw;
        uint32  roundsPlayed;
        string  gameId;             // e.g. "cs2"
        bytes32 sourceHash;         // hash of raw source data for provenance
        uint256 publishedAt;
        bool    finalized;
        bool    disputed;
    }

    // --- State ---

    address public owner;
    mapping(address => bool) public publishers;
    mapping(bytes32 => MatchScore) public scores;  // oracleResultId => MatchScore
    bytes32[] public publishedIds;

    uint256 public disputeWindowSeconds = 72 hours;
    uint256 public totalPublished;
    uint256 public totalFinalized;

    // --- Events ---

    event ScorePublished(
        bytes32 indexed oracleResultId,
        bytes32 indexed externalMatchId,
        uint16  teamAScore,
        uint16  teamBScore,
        string  gameId,
        uint256 publishedAt
    );

    event ScoreFinalized(
        bytes32 indexed oracleResultId,
        uint256 finalizedAt
    );

    event ScoreDisputed(
        bytes32 indexed oracleResultId,
        address indexed disputedBy,
        string  reason,
        uint256 disputedAt
    );

    event PublisherAdded(address indexed publisher);
    event PublisherRemoved(address indexed publisher);
    event DisputeWindowUpdated(uint256 oldWindow, uint256 newWindow);

    // --- Modifiers ---

    modifier onlyOwner() {
        require(msg.sender == owner, "ScoreOracle: not owner");
        _;
    }

    modifier onlyPublisher() {
        require(publishers[msg.sender] || msg.sender == owner, "ScoreOracle: not authorized");
        _;
    }

    // --- Constructor ---

    constructor() {
        owner = msg.sender;
        publishers[msg.sender] = true;
    }

    // --- Publisher Management ---

    function addPublisher(address _publisher) external onlyOwner {
        require(_publisher != address(0), "ScoreOracle: zero address");
        publishers[_publisher] = true;
        emit PublisherAdded(_publisher);
    }

    function removePublisher(address _publisher) external onlyOwner {
        publishers[_publisher] = false;
        emit PublisherRemoved(_publisher);
    }

    // --- Score Publication ---

    /**
     * @notice Publish a match score on-chain
     * @param oracleResultId Unique identifier for this oracle result
     * @param externalMatchId Hash of the external match identifier
     * @param teamAId Hash of team A identifier
     * @param teamBId Hash of team B identifier
     * @param teamAScore Team A's score
     * @param teamBScore Team B's score
     * @param winnerId Hash of winning team (bytes32(0) if draw)
     * @param isDraw Whether the match was a draw
     * @param roundsPlayed Total rounds played
     * @param gameId Game identifier (e.g. "cs2")
     * @param sourceHash Hash of raw source data
     */
    function publishScore(
        bytes32 oracleResultId,
        bytes32 externalMatchId,
        bytes32 teamAId,
        bytes32 teamBId,
        uint16  teamAScore,
        uint16  teamBScore,
        bytes32 winnerId,
        bool    isDraw,
        uint32  roundsPlayed,
        string calldata gameId,
        bytes32 sourceHash
    ) external onlyPublisher {
        require(scores[oracleResultId].publishedAt == 0, "ScoreOracle: already published");

        scores[oracleResultId] = MatchScore({
            externalMatchId: externalMatchId,
            teamAId: teamAId,
            teamBId: teamBId,
            teamAScore: teamAScore,
            teamBScore: teamBScore,
            winnerId: winnerId,
            isDraw: isDraw,
            roundsPlayed: roundsPlayed,
            gameId: gameId,
            sourceHash: sourceHash,
            publishedAt: block.timestamp,
            finalized: false,
            disputed: false
        });

        publishedIds.push(oracleResultId);
        totalPublished++;

        emit ScorePublished(
            oracleResultId,
            externalMatchId,
            teamAScore,
            teamBScore,
            gameId,
            block.timestamp
        );
    }

    // --- Finalization ---

    /**
     * @notice Finalize a score after the dispute window has passed
     * @param oracleResultId The oracle result to finalize
     */
    function finalizeScore(bytes32 oracleResultId) external onlyPublisher {
        MatchScore storage score = scores[oracleResultId];
        require(score.publishedAt > 0, "ScoreOracle: not published");
        require(!score.finalized, "ScoreOracle: already finalized");
        require(!score.disputed, "ScoreOracle: disputed");
        require(
            block.timestamp >= score.publishedAt + disputeWindowSeconds,
            "ScoreOracle: dispute window active"
        );

        score.finalized = true;
        totalFinalized++;

        emit ScoreFinalized(oracleResultId, block.timestamp);
    }

    // --- Disputes ---

    /**
     * @notice Dispute a published score
     * @param oracleResultId The oracle result to dispute
     * @param reason Human-readable dispute reason
     */
    function disputeScore(bytes32 oracleResultId, string calldata reason) external {
        MatchScore storage score = scores[oracleResultId];
        require(score.publishedAt > 0, "ScoreOracle: not published");
        require(!score.finalized, "ScoreOracle: already finalized");
        require(
            block.timestamp < score.publishedAt + disputeWindowSeconds,
            "ScoreOracle: dispute window expired"
        );

        score.disputed = true;

        emit ScoreDisputed(oracleResultId, msg.sender, reason, block.timestamp);
    }

    // --- View Functions ---

    function getScore(bytes32 oracleResultId) external view returns (MatchScore memory) {
        require(scores[oracleResultId].publishedAt > 0, "ScoreOracle: not found");
        return scores[oracleResultId];
    }

    function isFinalized(bytes32 oracleResultId) external view returns (bool) {
        return scores[oracleResultId].finalized;
    }

    function isDisputed(bytes32 oracleResultId) external view returns (bool) {
        return scores[oracleResultId].disputed;
    }

    function getPublishedCount() external view returns (uint256) {
        return totalPublished;
    }

    // --- Admin ---

    function setDisputeWindow(uint256 _seconds) external onlyOwner {
        emit DisputeWindowUpdated(disputeWindowSeconds, _seconds);
        disputeWindowSeconds = _seconds;
    }

    function transferOwnership(address _newOwner) external onlyOwner {
        require(_newOwner != address(0), "ScoreOracle: zero address");
        owner = _newOwner;
    }
}

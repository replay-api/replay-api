package oracle_services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

// ParsedScore represents structured score data extracted from OCR text
type ParsedScore struct {
	TeamAName    string `json:"team_a_name"`
	TeamBName    string `json:"team_b_name"`
	TeamAScore   int    `json:"team_a_score"`
	TeamBScore   int    `json:"team_b_score"`
	MapName      string `json:"map_name,omitempty"`
	IsHalfTime   bool   `json:"is_half_time,omitempty"`
	RoundsPlayed int    `json:"rounds_played"`
}

// OCRScoreParser extracts structured score data from OCR text blocks
type OCRScoreParser struct {
	// CS2 scoreboard patterns
	// Common formats: "NAVI 13 - 7 FaZe", "Team A 16:14 Team B", "13 | 7"
	scorePatterns []*regexp.Regexp

	// Known CS2 map names for map detection
	knownMaps map[string]bool
}

// NewOCRScoreParser creates a new parser with built-in CS2 patterns
func NewOCRScoreParser() *OCRScoreParser {
	return &OCRScoreParser{
		scorePatterns: compileScorePatterns(),
		knownMaps:     buildKnownMaps(),
	}
}

func compileScorePatterns() []*regexp.Regexp {
	patterns := []string{
		// "TeamA 13 - 7 TeamB" or "TeamA 13-7 TeamB"
		`(?i)^(.+?)\s+(\d{1,2})\s*[-–—]\s*(\d{1,2})\s+(.+?)$`,

		// "TeamA 16:14 TeamB"
		`(?i)^(.+?)\s+(\d{1,2})\s*:\s*(\d{1,2})\s+(.+?)$`,

		// "TeamA 2 | 1 TeamB" (series score)
		`(?i)^(.+?)\s+(\d)\s*\|\s*(\d)\s+(.+?)$`,

		// "13 - 7" (score only, no team names)
		`^\s*(\d{1,2})\s*[-–—]\s*(\d{1,2})\s*$`,

		// "13:7" (compact score)
		`^\s*(\d{1,2})\s*:\s*(\d{1,2})\s*$`,
	}

	compiled := make([]*regexp.Regexp, len(patterns))
	for i, p := range patterns {
		compiled[i] = regexp.MustCompile(p)
	}
	return compiled
}

func buildKnownMaps() map[string]bool {
	maps := []string{
		// CS2 competitive map pool
		"mirage", "inferno", "nuke", "overpass", "ancient", "anubis", "vertigo",
		"dust2", "dust_ii", "dust ii", "train",
	}
	m := make(map[string]bool, len(maps))
	for _, name := range maps {
		m[name] = true
	}
	return m
}

// ParseScoreboard attempts to extract structured score data from OCR text blocks
func (p *OCRScoreParser) ParseScoreboard(textBlocks []oracle_out.TextBlock, gameID replay_common.GameIDKey) (*ParsedScore, error) {
	if len(textBlocks) == 0 {
		return nil, fmt.Errorf("no text blocks to parse")
	}

	maxScore := maxScoreForGame(gameID)

	// Strategy 1: Try to find a single text block that contains the full scoreline
	for _, block := range textBlocks {
		// Skip timer blocks (M:SS) and noise
		if isCommonNoise(block.Text) {
			continue
		}
		if score := p.tryParseFullScoreline(block.Text); score != nil {
			if score.TeamAScore <= maxScore && score.TeamBScore <= maxScore {
				score.RoundsPlayed = score.TeamAScore + score.TeamBScore
				score.MapName = p.findMapName(textBlocks)
				return score, nil
			}
		}
	}

	// Strategy 2: Look for score-only blocks and nearby team name blocks
	score := p.tryAssembleFromBlocks(textBlocks)
	if score != nil {
		if score.TeamAScore <= maxScore && score.TeamBScore <= maxScore {
			score.RoundsPlayed = score.TeamAScore + score.TeamBScore
			score.MapName = p.findMapName(textBlocks)
			return score, nil
		}
	}

	return nil, fmt.Errorf("could not parse scoreboard from %d text blocks", len(textBlocks))
}

// maxScoreForGame returns the maximum reasonable score value for a game.
// CS2: max 30 (16 reg + overtime rounds). Valorant: max 25.
func maxScoreForGame(gameID replay_common.GameIDKey) int {
	switch gameID {
	case replay_common.CS2_GAME_ID, replay_common.CSGO_GAME_ID:
		return 30
	case replay_common.VLRNT_GAME_ID:
		return 25
	default:
		return 99
	}
}

// tryParseFullScoreline tries to match a single text against known score patterns
func (p *OCRScoreParser) tryParseFullScoreline(text string) *ParsedScore {
	// Normalize whitespace but preserve special characters first
	normalized := strings.Join(strings.Fields(strings.TrimSpace(text)), " ")

	// Try raw (normalized) text first, then cleaned text
	// (cleanOCRText replaces | with I which breaks pipe-based score patterns)
	variants := []string{normalized, cleanOCRText(text)}

	for _, variant := range variants {
		if result := p.matchScoreline(variant); result != nil {
			return result
		}
	}
	return nil
}

func (p *OCRScoreParser) matchScoreline(text string) *ParsedScore {
	for i, pattern := range p.scorePatterns {
		matches := pattern.FindStringSubmatch(text)
		if matches == nil {
			continue
		}

		switch i {
		case 0, 1, 2: // "TeamA SCORE sep SCORE TeamB" patterns
			if len(matches) >= 5 {
				scoreA, errA := strconv.Atoi(matches[2])
				scoreB, errB := strconv.Atoi(matches[3])
				if errA == nil && errB == nil {
					return &ParsedScore{
						TeamAName:  cleanTeamName(matches[1]),
						TeamBName:  cleanTeamName(matches[4]),
						TeamAScore: scoreA,
						TeamBScore: scoreB,
					}
				}
			}
		case 3, 4: // Score-only patterns
			if len(matches) >= 3 {
				scoreA, errA := strconv.Atoi(matches[1])
				scoreB, errB := strconv.Atoi(matches[2])
				if errA == nil && errB == nil {
					return &ParsedScore{
						TeamAScore: scoreA,
						TeamBScore: scoreB,
					}
				}
			}
		}
	}

	return nil
}

// tryAssembleFromBlocks uses spatial layout analysis to find scoreboard patterns.
// It detects CS2 HUD layouts like: TEAM_A [score] VS [score] TEAM_B
// by finding a "VS" anchor and looking for scores and team names on the same Y-line.
func (p *OCRScoreParser) tryAssembleFromBlocks(blocks []oracle_out.TextBlock) *ParsedScore {
	// Strategy A: Simple HUD layout (few blocks from cropped HUD region)
	// When there are < 10 blocks, separate numbers from text and arrange by X
	if len(blocks) <= 10 {
		if score := p.tryHUDScoreAssembly(blocks); score != nil {
			return score
		}
	}

	// Strategy B: Find "VS" block and use it as anchor
	if score := p.tryVSAnchoredAssembly(blocks); score != nil {
		return score
	}

	// Strategy C: Find two adjacent score-like numbers on the same horizontal line
	if score := p.tryAdjacentScoreAssembly(blocks); score != nil {
		return score
	}

	return nil
}

// tryHUDScoreAssembly handles the common case of a cropped CS2 HUD region
// where there are only a few blocks. It separates numbers and text, sorts by X,
// and expects the pattern: [Team A] [Score A] ... [Score B] [Team B]
func (p *OCRScoreParser) tryHUDScoreAssembly(blocks []oracle_out.TextBlock) *ParsedScore {
	type scoredBlock struct {
		text       string
		value      int
		isNumber   bool
		centerX    int
		confidence float64
	}

	var items []scoredBlock

	for _, b := range blocks {
		text := strings.TrimSpace(b.Text)
		if text == "" || b.Confidence < 0.3 {
			continue
		}
		if isCommonNoise(text) {
			continue
		}

		cx := blockCenterX(b)
		n, err := strconv.Atoi(text)
		if err == nil && n >= 0 && n <= 30 {
			// Valid CS2 round score (0-30)
			items = append(items, scoredBlock{text: text, value: n, isNumber: true, centerX: cx, confidence: b.Confidence})
		} else if len(text) >= 2 && !isMoneyAmount(text) {
			// Potential team name (2+ chars, not money)
			items = append(items, scoredBlock{text: cleanTeamName(text), isNumber: false, centerX: cx, confidence: b.Confidence})
		}
	}

	// Need at least 2 numbers (scores)
	var numbers, names []scoredBlock
	for _, it := range items {
		if it.isNumber {
			numbers = append(numbers, it)
		} else {
			names = append(names, it)
		}
	}

	if len(numbers) < 2 {
		return nil
	}

	// Sort numbers by X to get left (A) and right (B)
	for i := 1; i < len(numbers); i++ {
		for j := i; j > 0 && numbers[j].centerX < numbers[j-1].centerX; j-- {
			numbers[j], numbers[j-1] = numbers[j-1], numbers[j]
		}
	}

	// Sort names by X
	for i := 1; i < len(names); i++ {
		for j := i; j > 0 && names[j].centerX < names[j-1].centerX; j-- {
			names[j], names[j-1] = names[j-1], names[j]
		}
	}

	// Take the two numbers closest together as the score pair
	// (CS2 HUD shows both scores near the center)
	bestI, bestJ := 0, 1
	bestDist := 999999
	for i := 0; i < len(numbers); i++ {
		for j := i + 1; j < len(numbers); j++ {
			dist := numbers[j].centerX - numbers[i].centerX
			if dist < 0 {
				dist = -dist
			}
			if dist < bestDist {
				bestDist = dist
				bestI, bestJ = i, j
			}
		}
	}

	scoreA := numbers[bestI].value
	scoreB := numbers[bestJ].value
	scoreCenterX := (numbers[bestI].centerX + numbers[bestJ].centerX) / 2

	// Assign team names: names to the left of the score center = Team A, right = Team B
	var teamA, teamB string
	for _, n := range names {
		if n.centerX < scoreCenterX {
			if teamA == "" || n.confidence > 0.8 {
				teamA = n.text
			}
		} else {
			if teamB == "" || n.confidence > 0.8 {
				teamB = n.text
			}
		}
	}

	// If only one team name found, it might be to the left or right
	if teamA == "" && teamB == "" {
		// Still return the scores even without team names
	}

	return &ParsedScore{
		TeamAName:  teamA,
		TeamBName:  teamB,
		TeamAScore: scoreA,
		TeamBScore: scoreB,
	}
}

// blockCenterY returns the vertical center of a text block's bounding box
func blockCenterY(b oracle_out.TextBlock) int {
	// BoundingBox is [x, y, width, height]
	return b.BoundingBox[1] + b.BoundingBox[3]/2
}

// blockCenterX returns the horizontal center of a text block's bounding box
func blockCenterX(b oracle_out.TextBlock) int {
	return b.BoundingBox[0] + b.BoundingBox[2]/2
}

// sameRow checks if two blocks are roughly on the same horizontal line (within tolerance)
func sameRow(a, b oracle_out.TextBlock, tolerance int) bool {
	if tolerance <= 0 {
		tolerance = 20 // Default: 20px vertical tolerance
	}
	diff := blockCenterY(a) - blockCenterY(b)
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

// tryVSAnchoredAssembly finds a "VS" or "vs" block and looks for team names and
// scores on the same horizontal line, arranged as: TEAM_A [score] VS [score] TEAM_B
func (p *OCRScoreParser) tryVSAnchoredAssembly(blocks []oracle_out.TextBlock) *ParsedScore {
	for _, vsBlock := range blocks {
		vsText := strings.TrimSpace(strings.ToUpper(vsBlock.Text))
		if vsText != "VS" && vsText != "V5" && vsText != "V.S." {
			continue
		}

		if vsBlock.Confidence < 0.5 {
			continue
		}

		vsCenterX := blockCenterX(vsBlock)

		// Find all blocks on the same row
		var leftBlocks, rightBlocks []oracle_out.TextBlock
		for _, b := range blocks {
			if strings.TrimSpace(strings.ToUpper(b.Text)) == vsText {
				continue
			}
			if !sameRow(vsBlock, b, 25) {
				continue
			}

			bCenterX := blockCenterX(b)
			if bCenterX < vsCenterX {
				leftBlocks = append(leftBlocks, b)
			} else {
				rightBlocks = append(rightBlocks, b)
			}
		}

		// Sort blocks by X position
		sortBlocksByX(leftBlocks)
		sortBlocksByX(rightBlocks)

		// Extract: leftmost text = team A name, rightmost left number = team A score
		// rightmost text = team B name, leftmost right number = team B score
		teamAName, teamAScore, foundA := extractTeamAndScore(leftBlocks, true)
		teamBName, teamBScore, foundB := extractTeamAndScore(rightBlocks, false)

		if foundA && foundB {
			return &ParsedScore{
				TeamAName:  teamAName,
				TeamBName:  teamBName,
				TeamAScore: teamAScore,
				TeamBScore: teamBScore,
			}
		}
	}
	return nil
}

// extractTeamAndScore finds a team name and a score number from a set of blocks.
// For left side: score is closest to VS (rightmost number), team name is closest to score.
// For right side: score is closest to VS (leftmost number), team name is closest to score.
// CS2 HUD layout: [player_names...] TEAM_A [score] VS [score] TEAM_B [player_names...]
func extractTeamAndScore(blocks []oracle_out.TextBlock, isLeftSide bool) (string, int, bool) {
	if len(blocks) == 0 {
		return "", 0, false
	}

	// First pass: find the score number
	score := -1
	scoreIdx := -1
	for i, b := range blocks {
		text := strings.TrimSpace(b.Text)
		if text == "" || b.Confidence < 0.4 {
			continue
		}

		n, err := strconv.Atoi(text)
		if err == nil && n >= 0 && n <= 30 {
			if isLeftSide {
				score = n      // Keep overwriting = take rightmost (closest to VS)
				scoreIdx = i
			} else if score < 0 {
				score = n      // Take first = leftmost (closest to VS)
				scoreIdx = i
			}
		}
	}

	if score < 0 {
		return "", 0, false
	}

	// Second pass: find the team name closest to the score number (not a player name at the edge)
	// For left side: take the text block immediately to the left of the score
	// For right side: take the text block immediately to the right of the score
	teamName := ""
	bestDist := 999999

	for i, b := range blocks {
		if i == scoreIdx {
			continue
		}
		text := strings.TrimSpace(b.Text)
		if text == "" || b.Confidence < 0.5 || len(text) < 2 {
			continue
		}
		if _, err := strconv.Atoi(text); err == nil {
			continue // Skip numbers
		}
		if isMoneyAmount(text) || isCommonNoise(text) {
			continue
		}

		// Prefer names adjacent to the score
		if isLeftSide && i < scoreIdx {
			dist := scoreIdx - i
			if dist < bestDist {
				bestDist = dist
				teamName = cleanTeamName(text)
			}
		} else if !isLeftSide && i > scoreIdx {
			dist := i - scoreIdx
			if dist < bestDist {
				bestDist = dist
				teamName = cleanTeamName(text)
			}
		}
	}

	return teamName, score, true
}

// tryAdjacentScoreAssembly finds two number blocks close together on the same row
func (p *OCRScoreParser) tryAdjacentScoreAssembly(blocks []oracle_out.TextBlock) *ParsedScore {
	type numBlock struct {
		value int
		block oracle_out.TextBlock
	}

	var numbers []numBlock
	for _, b := range blocks {
		text := strings.TrimSpace(b.Text)
		if n, err := strconv.Atoi(text); err == nil && n >= 0 && n <= 30 && b.Confidence >= 0.5 {
			numbers = append(numbers, numBlock{value: n, block: b})
		}
	}

	// Find pairs on the same row
	for i := 0; i < len(numbers); i++ {
		for j := i + 1; j < len(numbers); j++ {
			if sameRow(numbers[i].block, numbers[j].block, 15) {
				// Check they're not too far apart (within ~30% of image width)
				dist := blockCenterX(numbers[j].block) - blockCenterX(numbers[i].block)
				if dist < 0 {
					dist = -dist
				}
				if dist < 300 && dist > 20 {
					left, right := numbers[i], numbers[j]
					if blockCenterX(left.block) > blockCenterX(right.block) {
						left, right = right, left
					}

					// Look for team names near these scores
					var teamA, teamB string
					for _, b := range blocks {
						text := strings.TrimSpace(b.Text)
						if _, err := strconv.Atoi(text); err != nil && len(text) >= 2 &&
							b.Confidence >= 0.7 && sameRow(left.block, b, 20) &&
							!isMoneyAmount(text) && !isCommonNoise(text) {
							bx := blockCenterX(b)
							if bx < blockCenterX(left.block) && teamA == "" {
								teamA = cleanTeamName(text)
							} else if bx > blockCenterX(right.block) && teamB == "" {
								teamB = cleanTeamName(text)
							}
						}
					}

					return &ParsedScore{
						TeamAName:  teamA,
						TeamBName:  teamB,
						TeamAScore: left.value,
						TeamBScore: right.value,
					}
				}
			}
		}
	}

	return nil
}

func sortBlocksByX(blocks []oracle_out.TextBlock) {
	for i := 1; i < len(blocks); i++ {
		for j := i; j > 0 && blockCenterX(blocks[j]) < blockCenterX(blocks[j-1]); j-- {
			blocks[j], blocks[j-1] = blocks[j-1], blocks[j]
		}
	}
}

func isMoneyAmount(text string) bool {
	return strings.HasPrefix(text, "$") || strings.HasPrefix(text, "€")
}

func isCommonNoise(text string) bool {
	upper := strings.ToUpper(strings.TrimSpace(text))
	noiseWords := []string{"VS", "V5", "V.S.", "TIMEOUT", "LOSS", "BONUS", "LEFT", "ROUND", "HALF"}
	for _, w := range noiseWords {
		if upper == w || strings.HasPrefix(upper, w+" ") {
			return true
		}
	}
	// Filter timer patterns like "1:43", "0:23" (M:SS format, single-digit minute)
	// But NOT score patterns like "16:14", "13:7"
	if len(upper) >= 3 && len(upper) <= 4 {
		timerPattern := regexp.MustCompile(`^\d:\d{2}$`)
		if timerPattern.MatchString(upper) {
			return true
		}
	}
	return false
}

// findMapName scans text blocks for known CS2 map names
func (p *OCRScoreParser) findMapName(blocks []oracle_out.TextBlock) string {
	for _, b := range blocks {
		text := strings.ToLower(strings.TrimSpace(b.Text))
		if isMapName(text, p.knownMaps) {
			return text
		}
	}
	return ""
}

// --- Helpers ---

func cleanOCRText(text string) string {
	// Remove common OCR artifacts
	text = strings.ReplaceAll(text, "|", "I")
	// Normalize whitespace
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}

func cleanTeamName(name string) string {
	name = strings.TrimSpace(name)
	// Remove trailing/leading punctuation
	name = strings.Trim(name, ".-_|[](){}\"'")
	// Remove common OCR noise characters
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '.' || r == '-' {
			return r
		}
		return -1
	}, name)
	return strings.TrimSpace(cleaned)
}

func isMapName(text string, knownMaps map[string]bool) bool {
	return knownMaps[strings.ToLower(text)]
}

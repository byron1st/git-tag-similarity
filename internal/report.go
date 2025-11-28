package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
)

var (
	ErrReportGeneration = errors.New("failed to generate report")
	ErrAPIRequest       = errors.New("API request failed")
	ErrReportWrite      = errors.New("failed to write report file")
)

// ClaudeMessage represents a message in the Claude API
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeRequest represents a request to the Claude API
type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
}

// ClaudeResponse represents a response from the Claude API
type ClaudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// GenerateReport creates an AI-generated markdown report analyzing the tag differences
func GenerateReport(result CompareResult, reportPath string) error {
	// Load config
	config, err := LoadConfig()
	if err != nil {
		if errors.Is(err, ErrConfigNotFound) {
			fmt.Fprintf(os.Stderr, "Warning: AI config not found. Report generation skipped. Run 'git-tag-similarity config' to set up AI.\n")
			return nil
		}
		return errors.Join(ErrReportGeneration, err)
	}

	// Validate config
	if err := config.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Invalid AI config. Report generation skipped. Error: %v\n", err)
		return nil
	}

	// Generate report content using AI based on provider
	var reportContent string
	switch config.Provider {
	case ProviderClaude:
		reportContent, err = generateReportWithClaude(result, config)
	case ProviderOpenAI:
		reportContent, err = generateReportWithOpenAI(result, config)
	case ProviderGemini:
		reportContent, err = generateReportWithGemini(result, config)
	default:
		err = fmt.Errorf("unsupported provider: %s", config.Provider)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to generate AI report: %v\n", err)
		return nil
	}

	// Write report to file
	if err := os.WriteFile(reportPath, []byte(reportContent), 0644); err != nil {
		return errors.Join(ErrReportWrite, err)
	}

	fmt.Printf("\nAI-generated report saved to: %s\n", reportPath)
	return nil
}

// generateReportWithClaude calls the Claude API to generate a report
func generateReportWithClaude(result CompareResult, config *AIConfig) (string, error) {
	// Prepare commit data for the prompt
	commitData := formatCommitDataForPrompt(result)

	// Create the prompt
	prompt := buildAnalysisPrompt(result, commitData)

	// Call Claude API
	return callClaudeAPI(prompt, config.APIKey, config.Model)
}

// formatDirectoryFilter formats the directory filter for display
func formatDirectoryFilter(directory string) string {
	if directory == "" {
		return ""
	}
	return fmt.Sprintf("Directory Filter: %s\n", directory)
}

// formatCommitDataForPrompt formats commit information for the AI prompt
func formatCommitDataForPrompt(result CompareResult) string {
	var buf strings.Builder

	// Commits only in Tag1
	if len(result.OnlyInTag1) > 0 {
		buf.WriteString(fmt.Sprintf("\nCommits only in [%s] (%d):\n", result.Config.Tag1Name, len(result.OnlyInTag1)))
		for hash := range result.OnlyInTag1 {
			commit, err := result.Repo.GetCommitObject(hash)
			if err != nil {
				buf.WriteString(fmt.Sprintf("  - %s (failed to get message)\n", hash.String()[:7]))
				continue
			}
			message := strings.Split(commit.Message, "\n")[0]
			buf.WriteString(fmt.Sprintf("  - %s: %s\n", hash.String()[:7], message))
		}
	}

	// Commits only in Tag2
	if len(result.OnlyInTag2) > 0 {
		buf.WriteString(fmt.Sprintf("\nCommits only in [%s] (%d):\n", result.Config.Tag2Name, len(result.OnlyInTag2)))
		for hash := range result.OnlyInTag2 {
			commit, err := result.Repo.GetCommitObject(hash)
			if err != nil {
				buf.WriteString(fmt.Sprintf("  - %s (failed to get message)\n", hash.String()[:7]))
				continue
			}
			message := strings.Split(commit.Message, "\n")[0]
			buf.WriteString(fmt.Sprintf("  - %s: %s\n", hash.String()[:7], message))
		}
	}

	// Add note about shared commits
	if len(result.SharedCommits) > 0 {
		buf.WriteString(fmt.Sprintf("\nNote: %d commits are shared between both tags.\n", len(result.SharedCommits)))
	}

	return buf.String()
}

// callClaudeAPI makes a request to the Claude API
func callClaudeAPI(prompt string, apiKey string, model string) (string, error) {
	apiURL := "https://api.anthropic.com/v1/messages"

	reqBody := ClaudeRequest{
		Model:     model,
		MaxTokens: 4096,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Join(ErrAPIRequest, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Join(ErrAPIRequest, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body)))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", err
	}

	if claudeResp.Error != nil {
		return "", errors.Join(ErrAPIRequest, fmt.Errorf("%s: %s", claudeResp.Error.Type, claudeResp.Error.Message))
	}

	if len(claudeResp.Content) == 0 {
		return "", errors.Join(ErrAPIRequest, fmt.Errorf("no content in response"))
	}

	return claudeResp.Content[0].Text, nil
}

// OpenAI API structures
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// generateReportWithOpenAI calls the OpenAI API to generate a report
func generateReportWithOpenAI(result CompareResult, config *AIConfig) (string, error) {
	// Prepare commit data for the prompt
	commitData := formatCommitDataForPrompt(result)

	// Create the prompt
	prompt := buildAnalysisPrompt(result, commitData)

	// Call OpenAI API
	return callOpenAIAPI(prompt, config.APIKey, config.Model)
}

// callOpenAIAPI makes a request to the OpenAI API
func callOpenAIAPI(prompt string, apiKey string, model string) (string, error) {
	apiURL := "https://api.openai.com/v1/chat/completions"

	reqBody := OpenAIRequest{
		Model: model,
		Messages: []OpenAIMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Join(ErrAPIRequest, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Join(ErrAPIRequest, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body)))
	}

	var openaiResp OpenAIResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return "", err
	}

	if openaiResp.Error != nil {
		return "", errors.Join(ErrAPIRequest, fmt.Errorf("%s: %s", openaiResp.Error.Type, openaiResp.Error.Message))
	}

	if len(openaiResp.Choices) == 0 {
		return "", errors.Join(ErrAPIRequest, fmt.Errorf("no content in response"))
	}

	return openaiResp.Choices[0].Message.Content, nil
}

// Gemini API structures
type GeminiContent struct {
	Parts []struct {
		Text string `json:"text"`
	} `json:"parts"`
	Role string `json:"role,omitempty"`
}

type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

// generateReportWithGemini calls the Gemini API to generate a report
func generateReportWithGemini(result CompareResult, config *AIConfig) (string, error) {
	// Prepare commit data for the prompt
	commitData := formatCommitDataForPrompt(result)

	// Create the prompt
	prompt := buildAnalysisPrompt(result, commitData)

	// Call Gemini API
	return callGeminiAPI(prompt, config.APIKey, config.Model)
}

// callGeminiAPI makes a request to the Gemini API
func callGeminiAPI(prompt string, apiKey string, model string) (string, error) {
	apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Join(ErrAPIRequest, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Join(ErrAPIRequest, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body)))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", err
	}

	if geminiResp.Error != nil {
		return "", errors.Join(ErrAPIRequest, fmt.Errorf("code %d: %s", geminiResp.Error.Code, geminiResp.Error.Message))
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", errors.Join(ErrAPIRequest, fmt.Errorf("no content in response"))
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// buildAnalysisPrompt creates the common analysis prompt used by all AI providers
func buildAnalysisPrompt(result CompareResult, commitData string) string {
	diffSection := ""
	if result.DiffStat != "" {
		diffSection = fmt.Sprintf("\n## File Changes (Diff Summary)\n\n```\n%s\n```\n", result.DiffStat)
	}

	return fmt.Sprintf(`You are analyzing the differences between two Git tags in a repository.

Repository: %s
Tag 1: %s
Tag 2: %s
%s
Similarity Score: %.2f%%

Summary:
- Total commits in [%s]: %d
- Total commits in [%s]: %d
- Shared commits: %d
- Unique to [%s]: %d
- Unique to [%s]: %d

%s
%s
Please create a detailed Markdown-formatted analysis report that includes:

1. Executive Summary (2-3 sentences about the overall changes)
2. Similarity Analysis (explain what the %.2f%% similarity means)
3. Key Changes (analyze the unique commits in each tag AND the file changes shown in the diff summary)
4. Impact Assessment (evaluate the significance of the differences based on both commits and actual code changes)
5. Recommendations (if applicable)

Format the output as proper Markdown with appropriate headers, lists, and formatting.
Keep the analysis concise but insightful. Focus on what the differences mean for the project.
Pay special attention to the file changes in the diff summary to understand the actual code modifications.`,
		result.Config.RepoPath,
		result.Config.Tag1Name,
		result.Config.Tag2Name,
		formatDirectoryFilter(result.Config.Directory),
		result.Similarity*100.0,
		result.Config.Tag1Name, len(result.OnlyInTag1)+len(result.SharedCommits),
		result.Config.Tag2Name, len(result.OnlyInTag2)+len(result.SharedCommits),
		len(result.SharedCommits),
		result.Config.Tag1Name, len(result.OnlyInTag1),
		result.Config.Tag2Name, len(result.OnlyInTag2),
		commitData,
		diffSection,
		result.Similarity*100.0,
	)
}

// GetCommitMessages returns commit messages for a set of commit hashes
func GetCommitMessages(repo Repository, commits map[plumbing.Hash]struct{}) ([]string, error) {
	messages := make([]string, 0, len(commits))
	for hash := range commits {
		commit, err := repo.GetCommitObject(hash)
		if err != nil {
			return nil, err
		}
		messages = append(messages, commit.Message)
	}
	return messages, nil
}

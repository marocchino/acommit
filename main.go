package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/alecthomas/kingpin/v2"
)

type Choice struct {
	Text string `json:"text"`
}

type JsonResponse struct {
	Choices []Choice `json:"choices"`
}

var (
	apiKey    = os.Getenv("OPENAI_API_KEY")
	maxTokens = kingpin.Flag("max-tokens", "Maximum number of tokens to generate.").Default("60").Int()
)

func main() {
	kingpin.Parse()

	output, err := getStagedDiff()
	if err != nil {
		fmt.Printf("Error running git diff --staged: %v\n", err)
		return
	}
	if output == "" {
		fmt.Println("No staged changes.")
	}
	p := fmt.Sprintf("You are to act as the author of a commit message in git. Your mission is to create clean and comprehensive commit messages in the gitmoji convention with emoji and explain why a change was done. I'll send you an output of 'git diff --staged' command, and you convert it into a commit message. Add a short description of WHY the changes are done after the commit message. Don't start it with 'This commit', just describe the changes. Use the present tense. Commit title must not be longer than 74 characters.\n%s", output)
	result, err := generateText(p, *maxTokens)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	text, err := parseResponse(result)
	if err != nil {
		fmt.Printf("Error unmarshalling JSON: %v\n", err)
		return
	}
	if text == "" {
		fmt.Println("No text generated.")
		return
	}
	err = commitWithEditor(text)
	if err != nil {
		fmt.Printf("Error running git commit: %v\n", err)
		return
	}
}

func getStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func generateText(prompt string, maxTokens int) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}
	url := "https://api.openai.com/v1/completions"
	payload := fmt.Sprintf(`{"model": "text-davinci-003", "prompt": %q, "max_tokens": %d}`, prompt, maxTokens)

	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Extract the generated text from the API response here.
	// You may need to use a JSON library like "encoding/json" to parse the response.
	return string(body), nil
}

func parseResponse(result string) (string, error) {
	var response JsonResponse
	err := json.Unmarshal([]byte(result), &response)
	if err != nil {
		return "", err
	}
	text := response.Choices[0].Text
	return strings.TrimPrefix(strings.Trim(text, "\n"), "Commit: "), nil
}

func commitWithEditor(message string) error {
	// Create a temporary file to hold the commit message
	tempFile, err := os.CreateTemp("", "commit-message")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())

	// Write the commit message to the temporary file
	_, err = tempFile.WriteString(message)
	if err != nil {
		return err
	}

	// Close the temp file to flush the contents
	tempFile.Close()

	cmd := exec.Command("git", "commit", "-t", tempFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

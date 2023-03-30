package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Choice struct {
	Message Message `json:"message"`
}

type JsonResponse struct {
	Choices []Choice `json:"choices"`
}

var (
	apiKey = os.Getenv("OPENAI_API_KEY")
)

func main() {

	diff, err := getStagedDiff()
	if err != nil {
		fmt.Printf("Error running git diff --staged: %v\n", err)
		return
	}
	result, err := generateText(diff)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	text, err := parseResponse(result)
	if err != nil {
		fmt.Printf("Error unmarshalling JSON: %v\n", err)
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
	result := strings.TrimSpace(string(output))
	if result == "" {
		return "", fmt.Errorf("No staged changes.")
	}

	return result, nil
}

func fetchPrompt() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(home, ".config", "acommit", "prompt.txt")

	content, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		fmt.Println("No prompt file found. Creating one at ~/.config/acommit/prompt.txt")
		const prompt = "You are to act as the author of a commit message in git. Your mission is to create clean and comprehensive commit messages in the gitmoji convention with emoji and explain why a change was done. I'll send you an output of 'git diff --staged' command, and you convert it into a commit message. Add a short description of WHY the changes are done after the commit message. Don't start it with 'This commit', just describe the changes. Use the present tense. Commit title must not be longer than 74 characters."
		// mkdir -p
		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			return "", err
		}
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		if err != nil {
			return "", err
		}
		defer file.Close()
		_, err = file.WriteString(prompt)
		if err != nil {
			return "", err
		}
		return prompt, nil
	} else if err != nil {
		return "", err
	}
	fmt.Println("Using prompt from ~/.config/acommit/prompt.txt")

	return string(content), nil
}

func generateText(diff string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable is not set. you can get it from https://platform.openai.com/account/api-keys.")
	}
	prompt, err := fetchPrompt()
	if err != nil {
		return "", err
	}
	url := "https://api.openai.com/v1/chat/completions"
	messages := []Message{
		{
			Role:    "system",
			Content: prompt,
		},
		{
			Role:    "user",
			Content: diff,
		},
	}

	data := map[string]interface{}{
		"model":    "gpt-3.5-turbo",
		"messages": messages,
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(payload)))
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
	text := string(body)
	if text == "" {
		return "", fmt.Errorf("No text generated.")
	}

	// Extract the generated text from the API response here.
	// You may need to use a JSON library like "encoding/json" to parse the response.
	return text, nil
}

func parseResponse(result string) (string, error) {
	var response JsonResponse
	err := json.Unmarshal([]byte(result), &response)
	if err != nil {
		return "", err
	}
	text := response.Choices[0].Message.Content
	return strings.Trim(text, "\n"), nil
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

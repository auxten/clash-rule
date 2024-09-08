package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

const (
	gistID      = "6d87154edc112f56c3ffe557eae7d4e9"
	clashAPIURL = "http://192.168.222.1:9090"
	authToken   = "123456"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: clash-rule <rule-type> <domain>")
		fmt.Println("Supported rule types: global-tv, direct, reject, trusted")
		os.Exit(1)
	}

	ruleType := os.Args[1]
	domain := os.Args[2]

	switch ruleType {
	case "global-tv":
		updateGist(domain, "global-tv.yaml", "gh-global-tv")
	case "direct":
		updateGist(domain, "direct.yaml", "gh-direct")
	case "reject":
		updateGist(domain, "reject.yaml", "gh-reject")
	case "trusted":
		updateGist(domain, "trusted.yaml", "gh-trusted")
	default:
		fmt.Printf("Unsupported rule type: %s\n", ruleType)
		os.Exit(1)
	}

	updateRule(ruleType)
	checkRuleStatus(ruleType)
}

func getGitHubToken() string {
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		return token
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		os.Exit(1)
	}

	patFile := filepath.Join(homeDir, ".gist_pat")
	content, err := os.ReadFile(patFile)
	if err != nil {
		fmt.Println("Error reading .gist_pat file:", err)
		os.Exit(1)
	}

	return strings.TrimSpace(string(content))
}

func updateGist(domain, fileName, providerName string) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: getGitHubToken()},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	gist, _, err := client.Gists.Get(ctx, gistID)
	if err != nil {
		fmt.Printf("Error getting gist: %v\n", err)
		os.Exit(1)
	}

	content := *gist.Files[github.GistFilename(fileName)].Content
	newContent := content + fmt.Sprintf("\n- DOMAIN-SUFFIX,%s", domain)

	gist.Files[github.GistFilename(fileName)] = github.GistFile{
		Content: github.String(newContent),
	}

	_, _, err = client.Gists.Edit(ctx, gistID, gist)
	if err != nil {
		fmt.Printf("Error updating gist: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Gist updated successfully for %s\n", providerName)
}

func updateRule(ruleType string) {
	providerName := getProviderName(ruleType)
	url := fmt.Sprintf("%s/providers/rules/%s", clashAPIURL, providerName)
	req, _ := http.NewRequest("PUT", url, nil)
	req.Header.Add("Accept", "application/json, text/plain, */*")
	req.Header.Add("Referer", "http://192.168.222.1:9090/ui/dashboard/")
	req.Header.Add("DNT", "1")
	req.Header.Add("Authorization", "Bearer "+authToken)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error updating rule: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("Rule updated successfully for %s\n", providerName)
}

func checkRuleStatus(ruleType string) {
	providerName := getProviderName(ruleType)
	url := fmt.Sprintf("%s/providers/rules", clashAPIURL)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json, text/plain, */*")
	req.Header.Add("Authorization", "Bearer "+authToken)
	req.Header.Add("DNT", "1")
	req.Header.Add("Referer", "http://192.168.222.1:9090/ui/dashboard/")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error checking rule status: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	providers := result["providers"].(map[string]interface{})
	ruleProvider := providers[providerName].(map[string]interface{})
	updatedAt := ruleProvider["updatedAt"].(string)

	parsedTime, _ := time.Parse(time.RFC3339Nano, updatedAt)
	formattedTime := parsedTime.Format("2006-01-02 15:04:05")

	fmt.Printf("Rule %s updated at: %s\n", providerName, formattedTime)
}

func getProviderName(ruleType string) string {
	switch ruleType {
	case "global-tv":
		return "gh-global-tv"
	case "direct":
		return "gh-direct"
	case "reject":
		return "gh-reject"
	case "trusted":
		return "gh-trusted"
	default:
		fmt.Printf("Unsupported rule type: %s\n", ruleType)
		os.Exit(1)
		return ""
	}
}
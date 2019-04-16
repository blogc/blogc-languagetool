package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

type ltResponse struct {
	Software struct {
		Version string `json:"version"`
	} `json:"software"`
	Matches []struct {
		Message  string `json:"message"`
		Sentence string `json:"sentence"`
		Rule     struct {
			Id          string `json:"id"`
			Description string `json:"description"`
			IssueType   string `json:"issueType"`
			Category    struct {
				Id   string `json:"id"`
				Name string `json:"name"`
			} `json:"category"`
			URLs []struct {
				Value string `json:"value"`
			} `json:"urls"`
		} `json:"rule"`
		Context struct {
			Text   string `json:"text"`
			Offset int    `json:"offset"`
			Length int    `json:"length"`
		} `json:"context"`
		Replacements []struct {
			Value string `json:"value"`
		} `json:"replacements"`
	} `json:"matches"`
}

func ltCheck(apiUrl string, text string, language string, motherTongue string, ignoreWords []string, disabledRules []string, disabledCategories []string) error {
	if apiUrl == "" {
		apiUrl = defaultApiUrl
	}
	logCtx := logrus.WithField("apiUrl", apiUrl)

	if language == "" {
		language = "auto"
	}
	logCtx = logCtx.WithField("language", language)

	form := url.Values{
		"text":        []string{text},
		"language":    []string{language},
		"enabledOnly": []string{"false"},
	}

	if motherTongue != "" {
		form.Add("motherTongue", motherTongue)
		logCtx = logCtx.WithField("motherTongue", motherTongue)
	}

	if len(disabledRules) > 0 {
		form.Add("disabledRules", strings.Join(disabledRules, ","))
		logCtx = logCtx.WithField("disabledRules", disabledRules)
	}

	if len(disabledCategories) > 0 {
		form.Add("disabledCategories", strings.Join(disabledCategories, ","))
		logCtx = logCtx.WithField("disabledCategories", disabledCategories)
	}

	logCtx.Info("Starting LanguageTool API request")
	body := strings.NewReader(form.Encode())
	resp, err := http.Post(apiUrl+"/v2/check", "application/x-www-form-urlencoded", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var obj ltResponse
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	logCtx = logrus.WithField("languagetool_version", obj.Software.Version)
	logCtx.Info("Processed successfuly by LanguageTool API")

	bar := strings.Repeat("-", 80)

	for i, m := range obj.Matches {

		found := false
		bad := m.Context.Text[m.Context.Offset : m.Context.Offset+m.Context.Length]
		for _, word := range strings.Split(bad, " ") {
			for _, ignoredWord := range ignoreWords {
				if word == ignoredWord {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if found {
			continue
		}

		fmt.Printf("Rule: %s: %s\n", m.Rule.Id, m.Rule.Description)
		fmt.Printf("      %s: %s (%s)\n", m.Rule.Category.Id, m.Rule.Category.Name, m.Rule.IssueType)
		for _, u := range m.Rule.URLs {
			fmt.Printf("      %s\n", u.Value)
		}

		fmt.Printf("Message:\n")
		fmt.Printf("      %s\n", m.Message)

		fmt.Printf("Sentence:\n")
		fmt.Printf("      %s\n", m.Sentence)

		fmt.Printf("Context:\n")
		fmt.Printf("      %s\n", m.Context.Text)
		fmt.Printf("      %s%s\n", strings.Repeat(" ", m.Context.Offset), strings.Repeat("^", m.Context.Length))

		if len(m.Replacements) > 0 {
			fmt.Printf("Replacements:\n")
			for _, u := range m.Replacements {
				fmt.Printf("      %s\n", u.Value)
			}
		}

		if i < len(obj.Matches)-1 {
			fmt.Printf("\n%s\n", bar)
		}
	}

	return nil
}

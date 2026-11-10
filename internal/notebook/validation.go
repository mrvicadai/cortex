package notebook

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"
)

func validateBody(body string) error {
	if len(body) == 0 || len(body) > 4096 || !utf8.ValidString(body) {
		return invalid("body: require 1–4096 UTF-8 bytes")
	}
	return nil
}

func validateTopics(topics []string) error {
	if topics == nil || len(topics) > 8 {
		return invalid("topics: require an array of at most eight unique values")
	}
	seen := map[string]bool{}
	for _, topic := range topics {
		if !validName(topic, 64) || seen[topic] {
			return invalid("topic")
		}
		seen[topic] = true
	}
	return nil
}

func validateEvidence(evidence []Evidence) error {
	if evidence == nil || len(evidence) > 8 {
		return invalid("evidence: require an array of at most eight pointers")
	}
	seen := map[string]bool{}
	for _, item := range evidence {
		if !validName(item.Source, 256) || seen[item.Source] {
			return invalid("evidence source")
		}
		seen[item.Source] = true
		if err := validateDate(item.VerifiedAt); err != nil {
			return err
		}
	}
	return nil
}

func validateDate(date string) error {
	if date == "" {
		return nil
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return invalid("verified_at: require YYYY-MM-DD")
	}
	return nil
}

func encodeDocument(document any) (string, error) {
	data, err := json.Marshal(document)
	if err != nil {
		return "", err
	}
	if len(data) > 6000 {
		return "", invalid("record: exceeds 6000 encoded JSON bytes; shorten or link an external artifact")
	}
	return string(data), nil
}

func validateFact(fact Fact) (string, error) {
	if !validName(fact.ID, 128) || !validName(fact.Title, 256) {
		return "", invalid("fact id or title")
	}
	if err := validateBody(fact.Body); err != nil {
		return "", err
	}
	if err := validateTopics(fact.Topics); err != nil {
		return "", err
	}
	if err := validateEvidence(fact.Evidence); err != nil {
		return "", err
	}
	return encodeDocument(fact)
}

func factSearch(fact Fact) string {
	parts := []string{fact.ID, fact.Title, fact.Body, strings.Join(fact.Topics, " ")}
	for _, item := range fact.Evidence {
		parts = append(parts, item.Source)
	}
	return strings.ToLower(strings.Join(parts, "\n"))
}

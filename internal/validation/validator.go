package validation

import (
	"strings"

	"github.com/go-resty/resty/v2"
)

type ContentValidator struct {
	Contains    []string
	NotContains []string
	Regex       string
	StatusCode  int
}

func NewContentValidator(contains, notContains []string, regex string) *ContentValidator {
	return &ContentValidator{
		Contains:    contains,
		NotContains: notContains,
		Regex:       regex,
	}
}

func (v *ContentValidator) Validate(resp *resty.Response) (bool, string, error) {
	body := string(resp.Body())

	for _, text := range v.Contains {
		if !strings.Contains(body, text) {
			return false, "missing_required_content", nil
		}
	}

	for _, text := range v.NotContains {
		if strings.Contains(body, text) {
			return false, "forbidden_content_present", nil
		}
	}

	if v.Regex != "" {
		matched, err := matchRegex(v.Regex, body)
		if err != nil {
			return false, "", err
		}
		if !matched {
			return false, "regex_not_matched", nil
		}
	}

	return true, "", nil
}

func matchRegex(pattern, text string) (bool, error) {
	return false, nil
}

type ValidationResult struct {
	Valid   bool
	Reason  string
	Details string
}

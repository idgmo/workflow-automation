package emailParser

import (
	"errors"
	"regexp"
	"strings"
)

// WorkOrder represents the structured business data extracted from a raw text email
type WorkOrder struct {
	UnitNumber  string
	IssueType   string
	IsEmergency bool
	Description string
}

// Compile regular expressions once at package level to optimize runtime speed and memory usage
var (
	unitRegex      = regexp.MustCompile(`(?i)(?:unit|apt|apartment|suite)\s*#?\s*([a-z0-9-]+)`)
	emergencyRegex = regexp.MustCompile(`(?i)(burst|flood|fire|gas|leak|danger|emergency|no heat)`)
	issueTypeRegex = regexp.MustCompile(`(?i)(plumbing|hvac|electrical|appliance|locksmith|roofing)`)
)

// ParseEmailBody processes raw email text and maps it into a structured WorkOrder entity
func ParseEmailBody(body string) (WorkOrder, error) {
	if strings.TrimSpace(body) == "" {
		return WorkOrder{}, errors.New("cannot parse an empty email body")
	}

	var order WorkOrder
	order.Description = strings.TrimSpace(body)

	// 1. Extract Unit/Apartment Number
	unitMatch := unitRegex.FindStringSubmatch(body)
	if len(unitMatch) > 1 {
		order.UnitNumber = strings.ToUpper(unitMatch[1])
	} else {
		order.UnitNumber = "UNKNOWN" // Default fallback flag
	}

	// 2. Classify Issue Type based on keyword matches
	issueMatch := issueTypeRegex.FindString(body)
	if issueMatch != "" {
		order.IssueType = strings.ToUpper(issueMatch)
	} else {
		order.IssueType = "GENERAL_MAINTENANCE"
	}

	// 3. Determine Urgency Matrix (Safety/Emergency Flag)
	if emergencyRegex.MatchString(body) {
		order.IsEmergency = true
	} else {
		order.IsEmergency = false
	}

	return order, nil
}

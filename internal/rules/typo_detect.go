package rules

import (
	"fmt"
	"strings"

	"github.com/romano/envguard/internal/parser"
)

// commonEnvKeys is a list of well-known environment variable names.
var commonEnvKeys = []string{
	"DATABASE_URL",
	"DB_HOST",
	"DB_PORT",
	"DB_USER",
	"DB_PASSWORD",
	"DB_NAME",
	"REDIS_URL",
	"REDIS_HOST",
	"REDIS_PORT",
	"API_KEY",
	"API_SECRET",
	"SECRET_KEY",
	"SECRET_KEY_BASE",
	"AWS_ACCESS_KEY_ID",
	"AWS_SECRET_ACCESS_KEY",
	"AWS_REGION",
	"AWS_BUCKET",
	"STRIPE_SECRET_KEY",
	"STRIPE_PUBLISHABLE_KEY",
	"SENDGRID_API_KEY",
	"SMTP_HOST",
	"SMTP_PORT",
	"SMTP_USER",
	"SMTP_PASSWORD",
	"PORT",
	"HOST",
	"NODE_ENV",
	"APP_ENV",
	"LOG_LEVEL",
	"DEBUG",
	"JWT_SECRET",
	"SESSION_SECRET",
	"CORS_ORIGIN",
	"ALLOWED_HOSTS",
	"SENTRY_DSN",
	"GITHUB_TOKEN",
	"GITHUB_CLIENT_ID",
	"GITHUB_CLIENT_SECRET",
	"GOOGLE_CLIENT_ID",
	"GOOGLE_CLIENT_SECRET",
	"TWITTER_API_KEY",
	"TWITTER_API_SECRET",
	"FACEBOOK_APP_ID",
	"FACEBOOK_APP_SECRET",
	"MAILGUN_API_KEY",
	"TWILIO_ACCOUNT_SID",
	"TWILIO_AUTH_TOKEN",
	"S3_BUCKET",
	"S3_REGION",
	"CLOUDINARY_URL",
	"ELASTICSEARCH_URL",
	"MONGO_URI",
	"POSTGRES_URL",
	"MYSQL_URL",
}

// TypoDetect finds keys that look like typos of common environment variable names.
// Uses Levenshtein distance: if a key is within distance 2 of a known key
// (and not an exact match), it's flagged as a possible typo.
type TypoDetect struct{}

func (td *TypoDetect) Name() string { return "typo-detect" }

func (td *TypoDetect) Run(file *parser.EnvFile) []Finding {
	var findings []Finding

	for _, e := range file.Keys() {
		key := strings.TrimSpace(e.Key)
		if key == "" {
			continue
		}

		bestMatch := ""
		bestDist := 3 // threshold: only flag if distance <= 2

		// Only compare against common/well-known keys.
		// If both DATABASE_URL and DATABASE_URI exist in the file, they're
		// intentionally different — not typos of each other.
		for _, candidate := range commonEnvKeys {
			if strings.EqualFold(candidate, key) {
				bestMatch = ""
				bestDist = 3
				break // exact match to a known key — not a typo
			}
			dist := levenshtein(strings.ToUpper(key), strings.ToUpper(candidate))
			if dist < bestDist {
				bestDist = dist
				bestMatch = candidate
			}
		}

		if bestMatch != "" {
			findings = append(findings, Finding{
				RuleName: td.Name(),
				Severity: Warning,
				Message:  fmt.Sprintf("Possible typo: %s (did you mean %s?)", key, bestMatch),
				Line:     e.LineNumber,
				Column:   1,
			})
		}
	}

	return findings
}

// levenshtein computes the Levenshtein distance between two strings.
func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Use two rows instead of full matrix for space efficiency.
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)

	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(
				curr[j-1]+1,   // insertion
				prev[j]+1,     // deletion
				prev[j-1]+cost, // substitution
			)
		}
		prev, curr = curr, prev
	}

	return prev[len(b)]
}

func min(vals ...int) int {
	m := vals[0]
	for _, v := range vals[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

package domain

// SpamAction represents the action recommended by the spam filter
type SpamAction string

const (
	SpamActionPass       SpamAction = "pass"
	SpamActionReject     SpamAction = "reject"
	SpamActionAddHeader  SpamAction = "add_header"
	SpamActionSoftReject SpamAction = "soft_reject"
)

// SpamCheckResult contains the result of a spam check
type SpamCheckResult struct {
	Action  SpamAction
	Score   float64
	Details string
	Headers map[string]string // e.g. X-Spam-Status, X-Spam-Score
}

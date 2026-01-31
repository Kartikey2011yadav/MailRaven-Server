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

// GreylistTuple represents a unique triplet for greylisting
type GreylistTuple struct {
	IPNet     string // /24 or /64 CIDR
	Sender    string // Envelope Sender (normalized)
	Recipient string // Envelope Recipient (normalized)
}

// GreylistEntry tracks the state of a greylisting tuple
type GreylistEntry struct {
	Tuple        GreylistTuple
	FirstSeenAt  int64 // Unix timestamp
	LastSeenAt   int64 // Unix timestamp
	BlockedCount int   // Metrics
}

// BayesToken represents a statistical token from training
type BayesToken struct {
	Token     string
	SpamCount int
	HamCount  int
}

// BayesGlobalStats stores global training counters
type BayesGlobalStats struct {
	TotalSpam int
	TotalHam  int
}

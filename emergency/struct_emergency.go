// Emergency alert structure
// Emergency ID
// User ID
// Emergency type
// Time
// Status
package emergency

import "time"

// Status shows where an emergency is in its life.
type Status string

const (
	StatusOpen      Status = "open"      // reported, still active
	StatusResolved  Status = "resolved"  // handled and closed
	StatusCancelled Status = "cancelled" // closed without being handled
)

// Emergency holds the data for one reported emergency.
// It only stores data. No business logic, database, or HTTP code goes here.
type Emergency struct {
	ID string `json:"id"` // unique emergency identifier

	// CallerUserID is nil when the caller is not registered.
	CallerUserID *string `json:"caller_user_id,omitempty"`
	CallerPhone  string  `json:"caller_phone"` // the number that dialed SafeLink

	Reason string `json:"reason"` // what happened, in the caller's words

	// Location. No GPS is assumed, so we use text fields.
	State     string `json:"state"`
	LGA       string `json:"lga"`
	Community string `json:"community"`
	Address   string `json:"address,omitempty"`  // street or home address, if known
	Landmark  string `json:"landmark,omitempty"` // used for the 3-alerts-per-landmark rule

	Status Status `json:"status"`

	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"` // nil until resolved
}

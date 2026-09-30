// Emergency request handler
//
// 1. Receive emergency signal
// 2. Identify user
// 3. Create emergency request
// 4. Send information to service
// 5. Return confirmation

// Emergency service
//
// 1. Receive emergency data (ReportInput)
// 2. Validate the data
// 3. Apply SafeLink rules (service sets ID, status and times)
// 4. Store or update the emergency through a Repository
// 5. Return the result (or a clear error)
//
// This file has no HTTP code and no database code.
package emergency

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

// Errors the handler can check and turn into HTTP responses later.
var (
	ErrCallerPhoneRequired = errors.New("caller phone is required")
	ErrReasonRequired      = errors.New("reason is required")
	ErrStateRequired       = errors.New("state is required")
	ErrLGARequired         = errors.New("lga is required")
	ErrCommunityRequired   = errors.New("community is required")

	// The repository should return ErrNotFound when no emergency has this ID.
	ErrNotFound = errors.New("emergency not found")

	// Returned when we try to close an emergency that is not open.
	ErrNotOpen = errors.New("emergency is already resolved or cancelled")
)

// Repository is what the service needs from storage.
// PostgreSQL (or a fake for tests) will implement this later.
// context.Context lets the database stop work if the request is cancelled.
type Repository interface {
	Create(ctx context.Context, e *Emergency) error
	GetByID(ctx context.Context, id string) (*Emergency, error)
	Update(ctx context.Context, e *Emergency) error
}

// ReportInput is the data a caller can send.
// It has no ID, Status or timestamps, because the service controls those.
type ReportInput struct {
	CallerUserID *string // nil = unregistered caller
	CallerPhone  string
	Reason       string
	State        string
	LGA          string
	Community    string
	Address      string
	Landmark     string
}

// Service holds the emergency business logic.
type Service struct {
	repo Repository // injected in NewService, never created here
}

// NewService gives the service its repository (dependency injection).
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ReportEmergency validates the input, creates a new open emergency and saves it.
//
// NOTE: The "max 3 alerts per landmark" rule is NOT enforced here.
// It is about responders, and we only know responders when we notify them.
// It will be enforced in the notification step, using the emergency's
// Landmark and a count of alerts each responder already received.
func (s *Service) ReportEmergency(ctx context.Context, in ReportInput) (*Emergency, error) {
	if err := validate(in); err != nil {
		return nil, err
	}

	id, err := newID()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	e := &Emergency{
		ID:           id,
		CallerUserID: in.CallerUserID,
		CallerPhone:  strings.TrimSpace(in.CallerPhone),
		Reason:       strings.TrimSpace(in.Reason),
		State:        strings.TrimSpace(in.State),
		LGA:          strings.TrimSpace(in.LGA),
		Community:    strings.TrimSpace(in.Community),
		Address:      strings.TrimSpace(in.Address),
		Landmark:     strings.TrimSpace(in.Landmark),
		Status:       StatusOpen, // always starts open, whatever the caller sends
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

// ResolveEmergency marks an open emergency as resolved and sets ResolvedAt.
func (s *Service) ResolveEmergency(ctx context.Context, id string) (*Emergency, error) {
	return s.close(ctx, id, StatusResolved)
}

// CancelEmergency marks an open emergency as cancelled.
func (s *Service) CancelEmergency(ctx context.Context, id string) (*Emergency, error) {
	return s.close(ctx, id, StatusCancelled)
}

// close holds the rule shared by resolve and cancel:
// only an open emergency can be closed.
func (s *Service) close(ctx context.Context, id string, newStatus Status) (*Emergency, error) {
	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if e.Status != StatusOpen {
		return nil, ErrNotOpen
	}

	now := time.Now().UTC()
	e.Status = newStatus
	e.UpdatedAt = now
	if newStatus == StatusResolved {
		e.ResolvedAt = &now // & gives a pointer, so the field is no longer nil
	}

	if err := s.repo.Update(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

// validate checks the required fields. Phone format checks come later.
func validate(in ReportInput) error {
	switch {
	case strings.TrimSpace(in.CallerPhone) == "":
		return ErrCallerPhoneRequired
	case strings.TrimSpace(in.Reason) == "":
		return ErrReasonRequired
	case strings.TrimSpace(in.State) == "":
		return ErrStateRequired
	case strings.TrimSpace(in.LGA) == "":
		return ErrLGARequired
	case strings.TrimSpace(in.Community) == "":
		return ErrCommunityRequired
	}
	return nil
}

// newID makes a random 32-character ID using only the standard library.
func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

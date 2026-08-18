package providers

import "errors"

var (
	ErrNotApproved    = errors.New("providers: provider not approved")
	ErrDisabled       = errors.New("providers: provider disabled")
	ErrKindNotAllowed = errors.New("providers: kind outside usage rights")
	ErrRateLimited    = errors.New("providers: provider rate budget spent")
)

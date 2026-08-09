package domain

import "errors"

var (
	ErrConversationNotFound = errors.New("ai conversation not found")
	ErrApprovalNotFound     = errors.New("approval request not found")
	ErrApprovalNotPending   = errors.New("approval request is not pending")
	ErrToolExecutionDenied  = errors.New("tool execution requires human approval")
)

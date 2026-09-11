package domain

import "time"

// SubmissionEmailNotification уведомление заявителю о статусе модерации.
type SubmissionEmailNotification struct {
	SubmitterName    string
	Email            string
	SubmissionID     string
	Decision         ReviewDecision
	ModeratorComment string
	Timestamp        time.Time
}

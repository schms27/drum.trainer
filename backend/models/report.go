package models

import "time"

// Report kinds and statuses.
const (
	ReportCopyright = "copyright"
	ReportAbuse     = "abuse"
	ReportOther     = "other"

	ReportOpen     = "open"
	ReportActioned = "actioned"
	ReportRejected = "rejected"
)

// ReportKinds lists the accepted values of Report.Kind.
var ReportKinds = []string{ReportCopyright, ReportAbuse, ReportOther}

// StrikeLimit is how many upheld reports cost a user the ability to share.
const StrikeLimit = 3

// Report is a complaint about a shared track, filed through its share link.
type Report struct {
	ID             int64      `json:"id"`
	TrackID        int64      `json:"track_id"`
	TrackName      string     `json:"track_name"`
	OwnerUserID    *int64     `json:"owner_user_id"`
	ShareToken     string     `json:"share_token"`
	ReporterUserID *int64     `json:"reporter_user_id"`
	ReporterName   string     `json:"reporter_name"`
	ReporterEmail  string     `json:"reporter_email"`
	Kind           string     `json:"kind"`
	WorkIdentified string     `json:"work_identified"`
	Details        string     `json:"details"`
	GoodFaith      bool       `json:"good_faith"`
	Status         string     `json:"status"`
	ResolutionNote *string    `json:"resolution_note"`
	CreatedAt      time.Time  `json:"created_at"`
	ResolvedAt     *time.Time `json:"resolved_at"`
}

// ReportWithContext is a report as the admin queue shows it, with the current
// state of the reported track and its owner.
type ReportWithContext struct {
	Report
	TrackExists   bool    `json:"track_exists"`   // false once the owner deleted it
	TrackDisabled bool    `json:"track_disabled"` // already removed by moderation
	RightsBasis   *string `json:"rights_basis"`   // the owner's declaration
	Copies        int     `json:"copies"`         // copies others made of it
	OwnerName     *string `json:"owner_name"`
	OwnerEmail    *string `json:"owner_email"`
	OwnerStrikes  int     `json:"owner_strikes"`
	OwnerBanned   bool    `json:"owner_banned"`
}

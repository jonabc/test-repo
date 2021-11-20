package gss

import (
	"context"
	"time"
)

// BannerType contains the set of possible banner types to display
type BannerType int

const (
	BannerTypeError = iota
	BannerTypeInfo
	BannerTypeWarning
)

type Banner struct {
	// Immutable Id of the banner
	ID int

	// The type of banner (Error, Info, or Warning)
	// This influences how the UI displays the banner
	Type BannerType

	// Date/time when the banner should no longer be presented to users. If not set
	// the banner does not automatically expire
	ExpiresAt *time.Time

	// Date/time when the banner was created in the system
	// This is internal metadata and not meant to be displayed in the UI
	CreatedAt *time.Time

	// Id of the GitHub repo which the banner is associated with
	RepositoryID int

	// Id of the GitHub user who created this banner
	// This is internal metadata and not meant to be displayed in the UI
	CreatorID int

	// The message contained in the banner
	// This should be displayed in the UI
	Message string
}

// Operations to interact with GitHub banners
type BannerService interface {
	AddBanner(ctx context.Context, t BannerType, expAt *time.Time, msg string) (*Banner, error)

	// Delete the banner with the given id
	DeleteBanner(ctx context.Context, id int) error

	// Update the provided banner
	UpdateBanner(ctx context.Context, b *Banner) (*Banner, error)

	// Retrieve the banner with the given id
	GetBanner(ctx context.Context, id int) (*Banner, error)

	// Retrieve the banners with the given ids (or list all banners if no ids supplied)
	// When ids is specified, the length of the returned slice is the same as the length of ids
	// and each banner in the resulting slice at some index corresponds to the requested id from the ids slice at the same index
	ListBanners(ctx context.Context, ids []int) ([]*Banner, error)
}

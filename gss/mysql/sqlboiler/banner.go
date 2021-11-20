// Package sqlboiler provides an implementation of the BannerService using SQLBoiler
package sqlboiler

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/github/go-trace"
	"github.com/jonabc/test-repo/gss"
	"github.com/jonabc/test-repo/gss/mysql/sqlboiler/internal/models"
	"github.com/pkg/errors"
	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/boil"
)

// Explicitly check that this struct satisfies the gss.BannerService interface.
var _ gss.BannerService = (*BannerService)(nil)

// BannerService implements the BannerService interface from the `gss` package.
type BannerService struct {
	db *sql.DB
}

// NewBannerService creates a new BannerService instance.
func NewBannerService(db *sql.DB) *BannerService {
	return &BannerService{
		db: db,
	}
}

// AddBanner creates a new banner and adds it to the database.
func (bs *BannerService) AddBanner(ctx context.Context, t gss.BannerType, expAt *time.Time, msg string) (*gss.Banner, error) {
	ctx, span := trace.ChildSpan(ctx)
	defer span.Finish()

	tx, err := bs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, span.WithError(err)
	}

	b := &gss.Banner{
		Type:      t,
		ExpiresAt: expAt,
		Message:   msg,
	}

	dbBanner := toDBModel(b)
	err = dbBanner.Insert(ctx, tx, boil.Infer())
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, span.WithError(errors.Wrap(err, fmt.Sprintf("unable to rollback banner insert transaction: %v", rollbackErr)))
		}
		return nil, span.WithError(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, span.WithError(err)
	}

	return b, nil
}

// DeleteBanner removes the banner from the database.
func (bs *BannerService) DeleteBanner(ctx context.Context, id int) error {
	ctx, span := trace.ChildSpan(ctx)
	defer span.Finish()

	tx, err := bs.db.BeginTx(ctx, nil)
	if err != nil {
		return span.WithError(err)
	}

	b, err := models.FindBanner(ctx, tx, uint64(id))
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return span.WithError(errors.Wrap(err, fmt.Sprintf("unable to rollback banner delete transaction: %v", rollbackErr)))
		}
		return span.WithError(err)
	}
	_, err = b.Delete(ctx, bs.db)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return span.WithError(errors.Wrap(err, fmt.Sprintf("unable to rollback banner delete transaction: %v", rollbackErr)))
		}
		return span.WithError(err)
	}

	if err := tx.Commit(); err != nil {
		return span.WithError(err)
	}
	return nil
}

// UpdateBanner accepts a banner, finds that banner by ID in the database, and updates that banner
// to match the passed-in banner object.
func (bs *BannerService) UpdateBanner(ctx context.Context, b *gss.Banner) (*gss.Banner, error) {
	ctx, span := trace.ChildSpan(ctx)
	defer span.Finish()

	tx, err := bs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, span.WithError(err)
	}

	if _, err = models.FindBanner(ctx, tx, uint64(b.ID)); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, span.WithError(errors.Wrap(err, fmt.Sprintf("unable to rollback banner update transaction: %v", rollbackErr)))
		}
		return nil, span.WithError(err)
	}

	banner := toDBModel(b)
	if _, err := banner.Update(ctx, tx, boil.Infer()); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, span.WithError(errors.Wrap(err, fmt.Sprintf("unable to rollback banner update transaction: %v", rollbackErr)))
		}
		return nil, span.WithError(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, span.WithError(err)
	}

	return toProjectModel(banner), nil
}

// GetBanner returns a banner when given an ID.
func (bs *BannerService) GetBanner(ctx context.Context, id int) (*gss.Banner, error) {
	ctx, span := trace.ChildSpan(ctx)
	defer span.Finish()

	banner, err := models.FindBanner(ctx, bs.db, uint64(id))
	if err != nil {
		return nil, span.WithError(err)
	}

	return toProjectModel(banner), nil
}

// ListBanners returns all banners that match a list of given IDs.
// If ids is nil, all banners are returned.
func (bs *BannerService) ListBanners(ctx context.Context, ids []int) ([]*gss.Banner, error) {
	ctx, span := trace.ChildSpan(ctx)
	defer span.Finish()

	banners := []*gss.Banner{}
	if ids == nil {
		dbBanners, err := models.Banners().All(ctx, bs.db)
		if err != nil {
			return nil, span.WithError(err)
		}
		for _, b := range dbBanners {
			banners = append(banners, toProjectModel(b))
		}
	} else {
		// could use bulk find query like: `models.Banners(models.BannerWhere.ID.IN(ids)).All(ctx, tx)`
		// but we'd have to iterate through ids first anyway to convert between int and uint64
		// so in terms of time complexity, we wouldn't be gaining anything
		for _, id := range ids {
			banner, err := models.FindBanner(ctx, bs.db, uint64(id))
			if err != nil {
				return nil, span.WithError(err)
			}
			banners = append(banners, toProjectModel(banner))
		}
	}

	return banners, nil
}

func toProjectModel(banner *models.Banner) *gss.Banner {
	return &gss.Banner{
		ID:           int(banner.ID),
		Type:         gss.BannerType(banner.Type),
		ExpiresAt:    &banner.ExpiresAt.Time,
		CreatedAt:    &banner.CreatedAt.Time,
		RepositoryID: int(banner.RepositoryID),
		CreatorID:    int(banner.CreatorID),
		Message:      banner.Message,
	}
}

func toDBModel(banner *gss.Banner) *models.Banner {
	return &models.Banner{
		ID:           uint64(banner.ID),
		Type:         uint(banner.Type),
		ExpiresAt:    null.TimeFrom(*banner.ExpiresAt),
		RepositoryID: uint(banner.RepositoryID),
		CreatorID:    uint(banner.CreatorID),
		Message:      banner.Message,
	}
}

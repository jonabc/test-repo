// Package vanilla provides an implementation of the BannerService using the standard
// library 'database/sql'.
package vanilla

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/github/go-trace"
	"github.com/jonabc/test-repo/gss"
)

// Explicitly check that this struct satisfies the gss.BannerService interface.
var _ gss.BannerService = (*BannerService)(nil)

// BannerService represents a vanilla connection to the mysql database.
type BannerService struct {
	db *sql.DB
}

// NewBannerService creates a BannerService instance and returns it.
func NewBannerService(db *sql.DB) *BannerService {
	return &BannerService{
		db: db,
	}
}

// AddBanner adds a banner to the database.
func (bs *BannerService) AddBanner(ctx context.Context, t gss.BannerType, expAt *time.Time, msg string) (*gss.Banner, error) {
	ctx, span := trace.ChildSpan(ctx)
	defer span.Finish()

	now := time.Now()
	banner := &gss.Banner{
		Type:         t,
		ExpiresAt:    expAt,
		CreatedAt:    &now,
		RepositoryID: 0,
		CreatorID:    0,
		Message:      msg,
	}

	result, err := bs.db.ExecContext(
		ctx,
		"INSERT INTO banners(type, expires_at, created_at, repository_id, creator_id, message) VALUES(?, ?, ?, ?, ?, ?)",
		banner.Type, banner.ExpiresAt, banner.CreatedAt, banner.RepositoryID, banner.CreatorID, banner.Message,
	)
	if err != nil {
		return nil, span.WithError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, span.WithError(err)
	}

	banner.ID = int(id)

	return banner, nil
}

// DeleteBanner removes the banner from the database.
func (bs *BannerService) DeleteBanner(ctx context.Context, id int) error {
	ctx, span := trace.ChildSpan(ctx)
	defer span.Finish()

	result, err := bs.db.ExecContext(ctx, "DELETE FROM banners where id = ?", id)
	if err != nil {
		return span.WithError(err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return span.WithError(err)
	}
	if count == 0 {
		return span.WithError(fmt.Errorf("no banner found with id %d", id))
	}

	return nil
}

// UpdateBanner accepts a banner, updates non nil/empty fields to the existing banner from the
// database
func (bs *BannerService) UpdateBanner(ctx context.Context, b *gss.Banner) (*gss.Banner, error) {
	ctx, span := trace.ChildSpan(ctx)
	defer span.Finish()

	// Because of how gosec checks for potential SQL injections,
	// we have to repeat the whole update statement for each scenario.
	// Indeed, gosec(at the moment) prevents from doing SQL string concatenations.
	var stmt string
	var values []interface{}
	switch {
	case b.ExpiresAt != nil && b.Message != "":
		stmt = "UPDATE banners SET expires_at = ?, message = ? WHERE id = ?"
		values = []interface{}{b.ExpiresAt, b.Message, b.ID}
	case b.ExpiresAt != nil && b.Message == "":
		stmt = "UPDATE banners SET expires_at = ? WHERE id = ?"
		values = []interface{}{b.ExpiresAt, b.ID}
	case b.ExpiresAt == nil && b.Message != "":
		stmt = "UPDATE banners SET message = ? WHERE id = ?"
		values = []interface{}{b.Message, b.ID}
	}

	if stmt != "" {
		// Execute the update query
		result, err := bs.db.ExecContext(ctx, stmt, values...)
		if err != nil {
			return nil, span.WithError(err)
		}

		// Check if a banner was actually updated. If not send an error.
		count, err := result.RowsAffected()
		if err != nil {
			return nil, span.WithError(err)
		}
		if count == 0 {
			return nil, span.WithError(fmt.Errorf("no banner found with id %d", b.ID))
		}
	}

	// Get the new updated bannner
	nb, err := bs.GetBanner(ctx, b.ID)
	return nb, err
}

// GetBanner returns a banner when given an ID.
func (bs *BannerService) GetBanner(ctx context.Context, id int) (*gss.Banner, error) {
	ctx, span := trace.ChildSpan(ctx)
	defer span.Finish()

	var banner gss.Banner
	row := bs.db.QueryRowContext(ctx, "SELECT id, type, expires_at, created_at, repository_id, creator_id, message from banners where id = ?", id)
	err := row.Scan(&banner.ID, &banner.Type, &banner.ExpiresAt, &banner.CreatedAt, &banner.RepositoryID, &banner.CreatorID, &banner.Message)
	if err != nil {
		return nil, span.WithError(err)
	}

	return &banner, nil
}

// ListBanners returns all banners that match a list of given IDs.
// If ids is nil, all banners are returned.
func (bs *BannerService) ListBanners(ctx context.Context, ids []int) ([]*gss.Banner, error) {
	ctx, span := trace.ChildSpan(ctx)
	defer span.Finish()

	var banners []*gss.Banner
	var rows *sql.Rows
	var err error

	if ids == nil {
		// This is not production ready, pagination using OFFSET and LIMIT should be used to avoid fetching too many rows
		rows, err = bs.db.QueryContext(ctx, "SELECT id, type, expires_at, created_at, repository_id, creator_id, message from banners")
	} else {
		values := make([]interface{}, len(ids))
		for i := range ids {
			values[i] = ids[i]
		}

		// #nosec G202
		// gosec security checker doesn't allow building SQL with string concatenation to prevent potential SQL injections.
		// In that case we need to build a dynamic SQL, with string concatenation, so we disable the gosec security check G202.
		stmt := `SELECT id, type, expires_at, created_at, repository_id, creator_id, message from banners where id in (?` + strings.Repeat(",?", len(ids)-1) + `)`
		rows, err = bs.db.Query(stmt, values...)
	}

	if err != nil {
		return nil, span.WithError(err)
	}

	defer rows.Close()
	for rows.Next() {
		var banner gss.Banner
		err := rows.Scan(&banner.ID, &banner.Type, &banner.ExpiresAt, &banner.CreatedAt, &banner.RepositoryID, &banner.CreatorID, &banner.Message)
		if err != nil {
			return nil, span.WithError(err)
		}
		banners = append(banners, &banner)
	}
	return banners, nil
}

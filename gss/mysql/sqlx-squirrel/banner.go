// Package sqlx provides an implementation of the BannerService using a combination of Squirrel
// and sqlx to access the MySQL database. Squirrel is using to build SQL queries in a fluent way.
// On the other end, sqlx is used to query the database, and map rows to predefined model structs.
package sqlx

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/iancoleman/strcase"

	sq "github.com/Masterminds/squirrel"
	"github.com/github/go-trace"
	"github.com/jmoiron/sqlx"
	"github.com/jonabc/test-repo/gss"
)

const (
	idColumn           = "id"
	typeColumn         = "type"
	expiresAtColumn    = "expires_at"
	createdAtColumn    = "created_at"
	repositoryIDColumn = "repository_id"
	creatorIDColumn    = "creator_id"
	messageColumn      = "message"
	bannersTableName   = "banners"
)

// Explicitly check that this struct satisfies the gss.BannerService interface.
var _ gss.BannerService = (*BannerService)(nil)

// BannerService represents a sqlx connection to the mysql database.
type BannerService struct {
	db *sqlx.DB
}

// NewBannerService creates a BannerService instance and returns it.
func NewBannerService(db *sql.DB) *BannerService {
	sqlxDB := sqlx.NewDb(db, "mysql")

	// Map struct field name to column name by converting field name to snake case
	sqlxDB.MapperFunc(strcase.ToSnake)

	return &BannerService{
		db: sqlxDB,
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

	sql, args, err := sq.
		Insert(bannersTableName).Columns(typeColumn, expiresAtColumn, createdAtColumn, repositoryIDColumn, creatorIDColumn, messageColumn).
		Values(banner.Type, banner.ExpiresAt, banner.CreatedAt, banner.RepositoryID, banner.CreatorID, banner.Message).
		ToSql()
	if err != nil {
		return nil, span.WithError(err)
	}

	result, err := bs.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return nil, span.WithError(err)
	}

	// Bind the newly-inserted autoincremented id to the banner ID
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

	result, err := sq.
		Delete(bannersTableName).
		Where(sq.Eq{idColumn: id}).
		RunWith(bs.db).ExecContext(ctx)
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

// UpdateBanner accepts a banner and updates non-zero-valued fields to the existing banner from the
// database
func (bs *BannerService) UpdateBanner(ctx context.Context, b *gss.Banner) (*gss.Banner, error) {
	ctx, span := trace.ChildSpan(ctx)
	defer span.Finish()

	builder := sq.Update(bannersTableName)
	if b.ExpiresAt != nil {
		builder = builder.Set(expiresAtColumn, b.ExpiresAt)
	}
	if b.Message == "" {
		builder = builder.Set(messageColumn, b.Message)
	}
	builder = builder.Where(sq.Eq{idColumn: b.ID})

	result, err := builder.RunWith(bs.db).ExecContext(ctx)
	if err != nil {
		return nil, span.WithError(err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return nil, span.WithError(err)
	}
	if count == 0 {
		return nil, span.WithError(fmt.Errorf("no banner found with id %d", b.ID))
	}

	nb, err := bs.GetBanner(ctx, b.ID)
	if err != nil {
		return nb, span.WithError(err)
	}
	return nb, nil
}

// GetBanner returns a banner when given an ID.
func (bs *BannerService) GetBanner(ctx context.Context, id int) (*gss.Banner, error) {
	ctx, span := trace.ChildSpan(ctx)
	defer span.Finish()

	banner := gss.Banner{}

	sql, args, err := sq.
		Select(idColumn, typeColumn, expiresAtColumn, createdAtColumn, repositoryIDColumn, creatorIDColumn, messageColumn).
		From(bannersTableName).
		Where(sq.Eq{idColumn: id}).
		ToSql()
	if err != nil {
		return nil, span.WithError(err)
	}

	err = bs.db.GetContext(ctx, &banner, sql, args...)
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

	builder := sq.Select(idColumn, typeColumn, expiresAtColumn, createdAtColumn, repositoryIDColumn, creatorIDColumn, messageColumn).From(bannersTableName)

	if ids != nil {
		builder = builder.Where(sq.Eq{idColumn: ids})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, span.WithError(err)
	}

	err = bs.db.SelectContext(ctx, &banners, sql, args...)
	if err != nil {
		return nil, span.WithError(err)
	}

	return banners, nil
}

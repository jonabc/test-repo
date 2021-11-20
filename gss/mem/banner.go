package mem

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jonabc/test-repo/gss"
)

// Explicitly check that this struct satisfies the gss.BannerService interface.
var _ gss.BannerService = (*BannerService)(nil)

// BannerService is the implementation of the gss.BannerService interface.
type BannerService struct {
	nextBannerID uint32
	banners      map[int]*gss.Banner
	mu           sync.Mutex // protects banners
}

// NewBannerService creates a new BannerService instance.
func NewBannerService() *BannerService {
	return &BannerService{
		nextBannerID: 1,
		banners:      make(map[int]*gss.Banner),
	}
}

func (bs *BannerService) getNextBannerID() uint32 {
	result := bs.nextBannerID
	atomic.AddUint32(&bs.nextBannerID, 1)
	return result
}

func (bs *BannerService) AddBanner(ctx context.Context, t gss.BannerType, expAt *time.Time, msg string) (*gss.Banner, error) {
	now := time.Now()

	banner := &gss.Banner{
		ID:           int(bs.getNextBannerID()),
		Type:         t,
		ExpiresAt:    expAt,
		CreatedAt:    &now,
		RepositoryID: 0, // TODO: Hook this up later
		CreatorID:    0, // TODO: Hook this up later
		Message:      msg,
	}

	bs.addBannerToCache(banner)
	return banner, nil
}

func (bs *BannerService) DeleteBanner(ctx context.Context, id int) error {
	bs.removeBannerFromCache(id)
	return nil
}

func (bs *BannerService) UpdateBanner(ctx context.Context, b *gss.Banner) (*gss.Banner, error) {
	bs.addBannerToCache(b)
	return b, nil
}

func (bs *BannerService) GetBanner(ctx context.Context, id int) (*gss.Banner, error) {
	banner, ok := bs.getBannerFromCache(id)
	if !ok {
		return nil, fmt.Errorf("unknown banner %d", id)
	}

	return banner, nil
}

func (bs *BannerService) ListBanners(ctx context.Context, ids []int) ([]*gss.Banner, error) {
	var result []*gss.Banner

	bs.mu.Lock()
	defer bs.mu.Unlock()

	// Add all banners if no ids given
	if ids == nil {
		result = []*gss.Banner{}
		for _, v := range bs.banners {
			result = append(result, v)
		}
	} else {
		result = make([]*gss.Banner, len(ids))
		for idx, id := range ids {
			result[idx] = bs.banners[id]
		}
	}

	return result, nil
}

func (bs *BannerService) addBannerToCache(b *gss.Banner) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.banners[b.ID] = b
}

func (bs *BannerService) getBannerFromCache(id int) (*gss.Banner, bool) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	banner, ok := bs.banners[id]
	return banner, ok
}

func (bs *BannerService) removeBannerFromCache(id int) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	delete(bs.banners, id)
}

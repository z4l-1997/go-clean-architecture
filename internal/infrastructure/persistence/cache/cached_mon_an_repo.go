package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"restaurant_project/internal/domain/cache"
	"restaurant_project/internal/domain/entity"
	"restaurant_project/internal/domain/repository"
)

// Đảm bảo CachedMonAnRepository implement IMonAnRepository
var _ repository.IMonAnRepository = (*CachedMonAnRepository)(nil)

// Cache keys
const (
	keyMonAnByID      = "mon_an:id:%s"
	keyMonAnAll       = "mon_an:all"
	keyMonAnConHang   = "mon_an:con_hang:%t"
	keyMonAnCount     = "mon_an:count"
	patternMonAnAll   = "mon_an:*"
)

// CachedMonAnRepository wraps IMonAnRepository với caching layer (Decorator Pattern)
// Sử dụng Cache-Aside pattern:
// - Read: Check cache → Miss → Load from DB → Cache result
// - Write: Update DB → Invalidate cache
type CachedMonAnRepository struct {
	repo  repository.IMonAnRepository // Repository gốc
	cache cache.ICacheRepository       // Cache layer
	ttl   time.Duration               // TTL mặc định
}

// NewCachedMonAnRepository tạo instance mới
func NewCachedMonAnRepository(
	repo repository.IMonAnRepository,
	cache cache.ICacheRepository,
	ttl time.Duration,
) *CachedMonAnRepository {
	if ttl == 0 {
		ttl = 5 * time.Minute
	}

	return &CachedMonAnRepository{
		repo:  repo,
		cache: cache,
		ttl:   ttl,
	}
}

// FindByID tìm món ăn theo ID với caching
func (r *CachedMonAnRepository) FindByID(ctx context.Context, id string) (*entity.MonAn, error) {
	cacheKey := fmt.Sprintf(keyMonAnByID, id)

	// Thử lấy từ cache
	data, err := r.cache.Get(ctx, cacheKey)
	if err != nil {
		// Cache error - fallback to DB
		return r.repo.FindByID(ctx, id)
	}

	if data != nil {
		// Cache hit
		var mon entity.MonAn
		if err := json.Unmarshal(data, &mon); err == nil {
			return &mon, nil
		}
		// Parse error - fallback to DB
	}

	// Cache miss - load from DB
	mon, err := r.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if mon == nil {
		return nil, nil
	}

	// Cache result
	if data, err := json.Marshal(mon); err == nil {
		r.cache.Set(ctx, cacheKey, data, r.ttl)
	}

	return mon, nil
}

// FindAll lấy tất cả món ăn với caching
func (r *CachedMonAnRepository) FindAll(ctx context.Context) ([]*entity.MonAn, error) {
	// Thử lấy từ cache
	data, err := r.cache.Get(ctx, keyMonAnAll)
	if err != nil {
		return r.repo.FindAll(ctx)
	}

	if data != nil {
		var mons []*entity.MonAn
		if err := json.Unmarshal(data, &mons); err == nil {
			return mons, nil
		}
	}

	// Cache miss
	mons, err := r.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(mons); err == nil {
		r.cache.Set(ctx, keyMonAnAll, data, r.ttl)
	}

	return mons, nil
}

// FindByConHang lấy các món theo trạng thái còn hàng với caching
func (r *CachedMonAnRepository) FindByConHang(ctx context.Context, conHang bool) ([]*entity.MonAn, error) {
	cacheKey := fmt.Sprintf(keyMonAnConHang, conHang)

	data, err := r.cache.Get(ctx, cacheKey)
	if err != nil {
		return r.repo.FindByConHang(ctx, conHang)
	}

	if data != nil {
		var mons []*entity.MonAn
		if err := json.Unmarshal(data, &mons); err == nil {
			return mons, nil
		}
	}

	mons, err := r.repo.FindByConHang(ctx, conHang)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(mons); err == nil {
		r.cache.Set(ctx, cacheKey, data, r.ttl)
	}

	return mons, nil
}

// Save lưu món ăn và invalidate cache
func (r *CachedMonAnRepository) Save(ctx context.Context, mon *entity.MonAn) error {
	// Update DB first
	if err := r.repo.Save(ctx, mon); err != nil {
		return err
	}

	// Invalidate all related caches
	r.invalidateCache(ctx, mon.ID)

	return nil
}

// Delete xóa món ăn và invalidate cache
func (r *CachedMonAnRepository) Delete(ctx context.Context, id string) error {
	// Delete from DB first
	if err := r.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate all related caches
	r.invalidateCache(ctx, id)

	return nil
}

// Count đếm tổng số món ăn với caching
func (r *CachedMonAnRepository) Count(ctx context.Context) (int64, error) {
	data, err := r.cache.Get(ctx, keyMonAnCount)
	if err != nil {
		return r.repo.Count(ctx)
	}

	if data != nil {
		var count int64
		if err := json.Unmarshal(data, &count); err == nil {
			return count, nil
		}
	}

	count, err := r.repo.Count(ctx)
	if err != nil {
		return 0, err
	}

	if data, err := json.Marshal(count); err == nil {
		r.cache.Set(ctx, keyMonAnCount, data, r.ttl)
	}

	return count, nil
}

// invalidateCache xóa tất cả cache liên quan đến món ăn
func (r *CachedMonAnRepository) invalidateCache(ctx context.Context, id string) {
	// Xóa cache theo ID
	r.cache.Delete(ctx, fmt.Sprintf(keyMonAnByID, id))

	// Xóa cache list
	r.cache.Delete(ctx, keyMonAnAll)
	r.cache.Delete(ctx, fmt.Sprintf(keyMonAnConHang, true))
	r.cache.Delete(ctx, fmt.Sprintf(keyMonAnConHang, false))
	r.cache.Delete(ctx, keyMonAnCount)
}

// InvalidateAll xóa tất cả cache liên quan đến món ăn
func (r *CachedMonAnRepository) InvalidateAll(ctx context.Context) error {
	return r.cache.DeleteByPattern(ctx, patternMonAnAll)
}

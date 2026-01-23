package db

import (
	"context"
	"fmt"
	"log"

	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/cache"
	"github.com/prudhivi99/Distributed-Systems/minisys-go/internal/models"
	"github.com/redis/go-redis/v9"
)

type CachedProductRepository struct {
	repo  *ProductRepository
	cache *cache.RedisCache
}

func NewCachedProductRepository(repo *ProductRepository, cache *cache.RedisCache) *CachedProductRepository {
	return &CachedProductRepository{
		repo:  repo,
		cache: cache,
	}
}

// Cache key helpers
func productKey(id int) string {
	return fmt.Sprintf("product:%d", id)
}

func allProductsKey() string {
	return "products:all"
}

// GetAll returns all products (with caching)
func (r *CachedProductRepository) GetAll(ctx context.Context) ([]models.Product, error) {
	cacheKey := allProductsKey()

	// Try cache first
	var products []models.Product
	err := r.cache.Get(ctx, cacheKey, &products)
	if err == nil {
		log.Println("📦 Cache HIT: all products")
		return products, nil
	}

	// Cache miss - get from database
	log.Println("💾 Cache MISS: all products - fetching from DB")
	products, err = r.repo.GetAll()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := r.cache.Set(ctx, cacheKey, products); err != nil {
		log.Printf("⚠️ Failed to cache products: %v", err)
	}

	return products, nil
}

// GetByID returns a single product (with caching)
func (r *CachedProductRepository) GetByID(ctx context.Context, id int) (*models.Product, error) {
	cacheKey := productKey(id)

	// Try cache first
	var product models.Product
	err := r.cache.Get(ctx, cacheKey, &product)
	if err == nil {
		log.Printf("📦 Cache HIT: product %d", id)
		return &product, nil
	}

	if err != redis.Nil {
		log.Printf("⚠️ Cache error: %v", err)
	}

	// Cache miss - get from database
	log.Printf("💾 Cache MISS: product %d - fetching from DB", id)
	p, err := r.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if p == nil {
		return nil, nil
	}

	// Store in cache
	if err := r.cache.Set(ctx, cacheKey, p); err != nil {
		log.Printf("⚠️ Failed to cache product: %v", err)
	}

	return p, nil
}

// Create inserts a new product and invalidates cache
func (r *CachedProductRepository) Create(ctx context.Context, req models.CreateProductRequest) (*models.Product, error) {
	// Create in database
	product, err := r.repo.Create(req)
	if err != nil {
		return nil, err
	}

	// Invalidate all products cache
	if err := r.cache.Delete(ctx, allProductsKey()); err != nil {
		log.Printf("⚠️ Failed to invalidate cache: %v", err)
	}
	log.Println("🗑️ Cache invalidated: all products")

	return product, nil
}

// Delete removes a product and invalidates cache
func (r *CachedProductRepository) Delete(ctx context.Context, id int) error {
	// Delete from database
	err := r.repo.Delete(id)
	if err != nil {
		return err
	}

	// Invalidate caches
	r.cache.Delete(ctx, productKey(id))
	r.cache.Delete(ctx, allProductsKey())
	log.Printf("🗑️ Cache invalidated: product %d and all products", id)

	return nil
}

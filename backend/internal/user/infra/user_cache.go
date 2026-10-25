package infra

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/rudradcruze/go-react-template/backend/internal/user/app"
	"github.com/rudradcruze/go-react-template/backend/internal/user/domain"
)

// userCacheTTL is the time that the cache keeps a User.
const userCacheTTL = 5 * time.Minute

var _ app.UserCache = (*RedisUserCache)(nil)

// cachedUser is the JSON form of a User in Redis.
type cachedUser struct {
	ID           uuid.UUID `json:"id"`
	EmailAddress string    `json:"emailAddress"`
	DisplayName  string    `json:"displayName"`
	CreatedAt    time.Time `json:"createdAt"`
}

// RedisUserCache keeps Users in Redis. The key has the form user:user:<id>.
type RedisUserCache struct {
	client *redis.Client
}

// NewRedisUserCache builds a cache on a Redis client. Give it a client for the cache database.
func NewRedisUserCache(client *redis.Client) *RedisUserCache {
	return &RedisUserCache{client: client}
}

func userCacheKey(id uuid.UUID) string { return "user:user:" + id.String() }

// Get gives the User with this ID. found is false when the cache does not have the User.
// A value that the cache cannot read, or that is for another ID, counts as a miss: the next Set replaces it.
func (c *RedisUserCache) Get(ctx context.Context, id uuid.UUID) (domain.User, bool, error) {
	data, err := c.client.Get(ctx, userCacheKey(id)).Bytes()
	if errors.Is(err, redis.Nil) {
		return domain.User{}, false, nil
	}
	if err != nil {
		return domain.User{}, false, err
	}
	var cached cachedUser
	if err := json.Unmarshal(data, &cached); err != nil || cached.ID != id {
		return domain.User{}, false, nil
	}
	return domain.User{
		ID:           cached.ID,
		EmailAddress: domain.EmailAddress(cached.EmailAddress),
		DisplayName:  domain.DisplayName(cached.DisplayName),
		CreatedAt:    cached.CreatedAt.UTC(),
	}, true, nil
}

// Set keeps the User in the cache for 5 minutes.
func (c *RedisUserCache) Set(ctx context.Context, user domain.User) error {
	data, err := json.Marshal(cachedUser{
		ID:           user.ID,
		EmailAddress: string(user.EmailAddress),
		DisplayName:  string(user.DisplayName),
		CreatedAt:    user.CreatedAt,
	})
	if err != nil {
		return err
	}
	return c.client.Set(ctx, userCacheKey(user.ID), data, userCacheTTL).Err()
}

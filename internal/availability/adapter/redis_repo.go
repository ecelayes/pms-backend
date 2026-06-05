package adapter

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type RedisAvailabilityRepository struct {
	client *redis.Client
}

func NewRedisAvailabilityRepository(client *redis.Client) *RedisAvailabilityRepository {
	return &RedisAvailabilityRepository{client: client}
}
func (r *RedisAvailabilityRepository) getKey(propertyID, unitID string, date time.Time) string {
	return fmt.Sprintf("availability:%s:%s:%s", propertyID, unitID, date.Format("2006-01-02"))
}
func (r *RedisAvailabilityRepository) UpdateInventory(ctx context.Context, propertyID, unitID string, date time.Time, delta int) error {
	key := r.getKey(propertyID, unitID, date)
	return r.client.IncrBy(ctx, key, int64(delta)).Err()
}
func (r *RedisAvailabilityRepository) GetInventory(ctx context.Context, propertyID, unitID string, date time.Time) (int, error) {
	key := r.getKey(propertyID, unitID, date)
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(val)
}
func (r *RedisAvailabilityRepository) GetBatchInventory(ctx context.Context, propertyID, unitID string, dates []time.Time) (map[string]int, error) {
	pipeline := r.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(dates))
	for i, date := range dates {
		key := r.getKey(propertyID, unitID, date)
		cmds[i] = pipeline.Get(ctx, key)
	}
	_, err := pipeline.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}
	result := make(map[string]int)
	for i, cmd := range cmds {
		val, err := cmd.Result()
		dateStr := dates[i].Format("2006-01-02")
		switch {
		case errors.Is(err, redis.Nil), err != nil:
			continue
		default:
			count, _ := strconv.Atoi(val)
			result[dateStr] = count
		}
	}
	return result, nil
}

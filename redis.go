
import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

var rdb *redis.Client

func Setup(cfg config.RedisConfig) {
	rdb = redis.NewClient(&redis.Options{
		Addr:         cfg.Host,
		Password:     cfg.Password,
		DB:           cfg.Db,
		WriteTimeout: cfg.WriteTimeout * time.Second,
	})
}

func Ping(ctx context.Context) error {
	return rdb.Ping(ctx).Err()
}

func Get(ctx context.Context, key string) (string, error) {
	return rdb.Get(ctx, key).Result()
}

func Set(ctx context.Context, key string, value []byte, duration time.Duration) error {
	return rdb.Set(ctx, key, value, duration).Err()
}
func Delete(ctx context.Context, key string) error {
	return rdb.Del(ctx, key).Err()
}

func Exist(ctx context.Context, key string) bool {
	res := rdb.Exists(ctx, key)
	if res != nil && res.Val() > 0 {
		return true
	} else {
		return false
	}
}

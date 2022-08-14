package utils

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

func AddEx(key string, val string, rdb *redis.Client, ctx context.Context) error {

	err := rdb.Set(ctx, key, val, time.Second*1800).Err()
	if err != nil {
		return errors.New("connection failed")
	}

	return nil
}

func AddShadow(key string, val string, rdb *redis.Client, ctx context.Context) error {

	err := rdb.Set(ctx, key, val, 0).Err()
	if err != nil {
		return errors.New("connection failed")
	}

	return nil
}

func GetKey(key string, rdb *redis.Client, ctx context.Context) (string, error) {
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return "", errors.New("get failed")
	}
	return val, nil
}

func GetPortFromPool(rdb *redis.Client, ctx context.Context) (int, error) {
	for i := 45000; i < 45100; i++ {
		_, err := rdb.Get(ctx, "portUsed:"+strconv.Itoa(i)).Result()
		if err != nil {
			rdb.Set(ctx, "portUsed:"+strconv.Itoa(i), 1, 0)
			return i, nil
		}
	}
	return -1, errors.New("all occupied")
}

func DelPort(key string, rdb *redis.Client, ctx context.Context) error {
	err := rdb.Del(ctx, "portUsed:"+key).Err()
	return err
}

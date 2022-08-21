package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

func LockByKey(c *redis.Client, key string) error {
	lockKey := fmt.Sprintf("lock_%v", key)
	ctx := context.Background()
	succeed, err := c.SetNX(ctx, lockKey, localIp, time.Second*30).Result()
	if err == nil && !succeed {
		err = fmt.Errorf("get locked failed for key : %s", key)
	}
	return err
}

func Unlock(c *redis.Client, key string) {
	lockKey := fmt.Sprintf("lock_%v", key)
	ctx := context.Background()
	val, err := c.Get(ctx, lockKey).Result()
	if err != nil || val != localIp {
		return
	}
	c.Del(ctx, lockKey)
}

func Get(c *redis.Client, key string) (string, error) {
	elemKey := fmt.Sprintf("elem_%v", key)
	ctx := context.Background()
	return c.Get(ctx, elemKey).Result()
}

func Set(c *redis.Client, key string, val string) error {
	elemKey := fmt.Sprintf("elem_%v", key)
	ctx := context.Background()
	_, err := c.Set(ctx, elemKey, val, time.Duration(0)).Result()
	return err
}

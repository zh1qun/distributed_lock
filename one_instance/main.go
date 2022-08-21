package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

var rdb *redis.Client
var wtg sync.WaitGroup
var localIp string

func init() {
	eths, er := net.Interfaces()
	if er != nil {
		fmt.Println(er)
		return
	}
	for _, eth := range eths {
		//fmt.Println(eth.Name)
		if eth.Name == "ens33" {
			addrs, _ := eth.Addrs()
			localIp = addrs[0].String()
			fmt.Println(localIp)
		}
		// 检查ip地址判断是否回环地址
		/*
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					fmt.Println(ipnet.IP.String())
				}
			}*/
	}
}

func main() {
	// 单实例redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     "0.0.0.0:6407",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("redis connect error : %s\n", err)
		return
	}

	/*
		e := LockByKey(rdb, "134")
		fmt.Println(e)
		Unlock(rdb, "134")
		e = LockByKey(rdb, "134")
		fmt.Println(e)*/

	wtg = sync.WaitGroup{}
	ch := make(chan error)
	for i := 0; i <= 20; i++ {
		// fmt.Print(i)
		wtg.Add(1)
		go SetRedisValue(ch, strconv.Itoa(i%3), strconv.Itoa(i))
	}

	go func() {
		wtg.Wait()
		close(ch)
	}()

	for e := range ch {
		fmt.Println(e)
	}
	fmt.Println("exit")

}

func SetRedisValue(ch chan error, key string, val string) {
	//自旋获取分布式锁
	locked := false

	defer func() {
		if locked {
			Unlock(rdb, key)
		}
		wtg.Done()
	}()

	for i := 0; i < 3; i++ {
		//获取分布式锁
		err := LockByKey(rdb, key)
		if err != nil {
			//log.Println(fmt.Sprintf("lock elem %v failed : %s", key, err))
			time.Sleep(time.Millisecond * time.Duration(rand.Int63n(1000)*int64(i)))
		} else {
			break
		}

		if i == 2 {
			//e := fmt.Errorf("get key %v failed for 3 times", key)
			//log.Println(e)
			ch <- err
			return
		}
	}

	locked = true
	err := Set(rdb, key, val)
	if err != nil {
		//e := fmt.Errorf("set key %v to value %v failed", key, val)
		//log.Println(e)
		ch <- err
		return
	}
	//log.Println("Set successful.")
	//return nil
}

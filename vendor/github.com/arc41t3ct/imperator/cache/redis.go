package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

type RedisCache struct {
	Conn   *redis.Pool
	Prefix string
}

type Entry map[string]interface{}

func (c *RedisCache) Has(cacheKey string) (bool, error) {
	key := fmt.Sprintf("%s:%s", c.Prefix, cacheKey)
	conn := c.Conn.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (c *RedisCache) Get(cacheKey string) (interface{}, error) {
	key := fmt.Sprintf("%s:%s", c.Prefix, cacheKey)
	conn := c.Conn.Get()
	defer conn.Close()
	cacheEntry, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}
	decoded, err := decode(string(cacheEntry))
	if err != nil {
		return nil, err
	}
	item := decoded[key]
	return item, nil
}

func (c *RedisCache) Set(cacheKey string, value interface{}, expires ...int) error {
	key := fmt.Sprintf("%s:%s", c.Prefix, cacheKey)
	conn := c.Conn.Get()
	defer conn.Close()

	entry := Entry{}
	entry[key] = value
	encoded, err := encode(entry)
	if err != nil {
		return err
	}
	if len(expires) > 0 {
		_, err := conn.Do("SETEX", key, expires[0], string(encoded))
		if err != nil {
			return err
		}
	} else {
		_, err := conn.Do("SET", key, string(encoded))
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *RedisCache) Forget(cacheKey string) error {
	key := fmt.Sprintf("%s:%s", c.Prefix, cacheKey)
	conn := c.Conn.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", key)
	if err != nil {
		return err
	}
	return nil
}

func (c *RedisCache) EmptyMatching(cacheKey string) error {
	key := fmt.Sprintf("%s:%s", c.Prefix, cacheKey)
	conn := c.Conn.Get()
	defer conn.Close()

	keys, err := c.getKeys(key)
	if err != nil {
		return err
	}

	for _, x := range keys {
		if _, err := conn.Do("DEL", x); err != nil {
			return err
		}
	}
	return nil
}

func (c *RedisCache) Empty() error {
	return c.EmptyMatching("")
}

func encode(item Entry) ([]byte, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	if err := e.Encode(item); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func decode(key string) (Entry, error) {
	item := Entry{}
	b := bytes.Buffer{}
	b.Write([]byte(key))
	d := gob.NewDecoder(&b)
	if err := d.Decode(&item); err != nil {
		return nil, err
	}
	return item, nil
}

func (c *RedisCache) getKeys(pattern string) ([]string, error) {
	conn := c.Conn.Get()
	defer conn.Close()
	iter := 0
	keys := []string{}
	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", fmt.Sprintf("%s*", pattern)))
		if err != nil {
			return keys, err
		}
		iter, _ := redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		keys = append(keys, k...)
		if iter == 0 {
			break
		}
	}

	return keys, nil
}

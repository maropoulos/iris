// Copyright (c) 2016, Gerasimos Maropoulos
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
//    this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
//	  this list of conditions and the following disclaimer
//    in the documentation and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse
//    or promote products derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER AND CONTRIBUTOR, GERASIMOS MAROPOULOS
// BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

///TODO: add doc comments
package redis

import (
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/kataras/iris/errors"
)

const (
	DefaultNetwork       = "tcp"
	DefaultAddr          = "127.0.0.1:6379"
	DefaultIdleTimeout   = 5 * time.Minute
	DefaultMaxAgeSeconds = 2520.0
)

var (
	ErrRedisClosed = errors.New("Redis is already closed")
	ErrKeyNotFound = errors.New("Key '%s' doesn't found")
)

type Config struct {
	// Network "tcp"
	Network string
	// Addr "127.0.01:6379"
	Addr string
	// Password string .If no password then no 'AUTH'. Default ""
	Password string
	// If Database is empty "" then no 'SELECT'. Default ""
	Database string
	// MaxIdle 0 no limit
	MaxIdle int
	// MaxActive 0 no limit
	MaxActive int
	// IdleTimeout 5 * time.Minute
	IdleTimeout time.Duration
	// Prefix "myprefix-for-this-website". Default ""
	Prefix string
	// MaxAgeSeconds how much long the redis should keep the session in seconds. Default 2520.0 (42minutes)
	MaxAgeSeconds int
}

type Service struct {
	Connected bool
	Config    *Config
	pool      *redis.Pool
}

func (r *Service) PingPong() (bool, error) {
	c := r.pool.Get()
	defer c.Close()
	msg, err := c.Do("PING")
	if err != nil || msg == nil {
		return false, err
	}
	return (msg == "PONG"), nil
}

func (r *Service) CloseConnection() error {
	if r.pool != nil {
		return r.pool.Close()
	}
	return ErrRedisClosed.Return()
}

// key string, value string, you can use utils.Serialize(&myobject{}) to convert an object to []byte
func (r *Service) Set(key string, value []byte, maxageseconds ...float64) (err error) { // map[interface{}]interface{}) (err error) {
	maxage := DefaultMaxAgeSeconds //42 minutes
	c := r.pool.Get()
	defer c.Close()
	if err = c.Err(); err != nil {
		return
	}
	if len(maxageseconds) > 0 {
		if max := maxageseconds[0]; max >= 0 {
			maxage = max
		}
	}
	_, err = c.Do("SETEX", r.Config.Prefix+key, maxage, value)
	return
}

// Get returns value, err by its key
// you can use utils.Deserialize((.Get("yourkey"),&theobject{})
//returns nil and a filled error if something wrong happens
func (r *Service) Get(key string) (interface{}, error) {
	c := r.pool.Get()
	defer c.Close()
	if err := c.Err(); err != nil {
		return nil, err
	}

	redisVal, err := c.Do("GET", r.Config.Prefix+key)

	if err != nil {
		return nil, err
	}
	if redisVal == nil {
		return nil, ErrKeyNotFound.Format(key)
	}
	return redisVal, nil
}

// Get returns value, err by its key
// you can use utils.Deserialize((.Get("yourkey"),&theobject{})
//returns nil and a filled error if something wrong happens
func (r *Service) GetBytes(key string) ([]byte, error) {
	c := r.pool.Get()
	defer c.Close()
	if err := c.Err(); err != nil {
		return nil, err
	}

	redisVal, err := c.Do("GET", r.Config.Prefix+key)

	if err != nil {
		return nil, err
	}
	if redisVal == nil {
		return nil, ErrKeyNotFound.Format(key)
	}

	return redis.Bytes(redisVal, err)
}

// GetString returns value, err by its key
// you can use utils.Deserialize((.GetString("yourkey"),&theobject{})
//returns empty string and a filled error if something wrong happens
func (r *Service) GetString(key string) (string, error) {
	redisVal, err := r.Get(key)
	if redisVal == nil {
		return "", ErrKeyNotFound.Format(key)
	}

	sVal, err := redis.String(redisVal, err)
	if err != nil {
		return "", err
	}
	return sVal, nil
}

// GetInt returns value, err by its key
// you can use utils.Deserialize((.GetInt("yourkey"),&theobject{})
//returns -1 int and a filled error if something wrong happens
func (r *Service) GetInt(key string) (int, error) {
	redisVal, err := r.Get(key)
	if redisVal == nil {
		return -1, ErrKeyNotFound.Format(key)
	}

	intVal, err := redis.Int(redisVal, err)
	if err != nil {
		return -1, err
	}
	return intVal, nil
}

// GetStringMap returns map[string]string, err by its key
//returns nil  and a filled error if something wrong happens
func (r *Service) GetStringMap(key string) (map[string]string, error) {
	redisVal, err := r.Get(key)
	if redisVal == nil {
		return nil, ErrKeyNotFound.Format(key)
	}

	_map, err := redis.StringMap(redisVal, err)
	if err != nil {
		return nil, err
	}
	return _map, nil
}

func (r *Service) GetAll(key string) (map[string]string, error) {
	c := r.pool.Get()
	defer c.Close()
	if err := c.Err(); err != nil {
		return nil, err
	}

	reply, err := c.Do("HGETALL", r.Config.Prefix+key)

	if err != nil {
		return nil, err
	}
	if reply == nil {
		return nil, ErrKeyNotFound.Format(key)
	}

	return redis.StringMap(reply, err)

}

func (r *Service) GetAllKeysByPrefix(prefix string) ([]string, error) {
	c := r.pool.Get()
	defer c.Close()
	if err := c.Err(); err != nil {
		return nil, err
	}

	reply, err := c.Do("KEYS", r.Config.Prefix+prefix)

	if err != nil {
		return nil, err
	}
	if reply == nil {
		return nil, ErrKeyNotFound.Format(prefix)
	}
	return redis.Strings(reply, err)

}

func (r *Service) Delete(key string) error {
	c := r.pool.Get()
	defer c.Close()
	if _, err := c.Do("DEL", r.Config.Prefix+key); err != nil {
		return err
	}
	return nil
}

func dial(network string, addr string, pass string) (redis.Conn, error) {
	if network == "" {
		network = DefaultNetwork
	}
	if addr == "" {
		addr = DefaultAddr
	}
	c, err := redis.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	if pass != "" {
		if _, err := c.Do("AUTH", pass); err != nil {
			c.Close()
			return nil, err
		}
	}
	return c, err
}

func (r *Service) Connect() {
	config := r.Config

	if config.IdleTimeout <= 0 {
		config.IdleTimeout = 5 * time.Minute
	}

	if config.Network == "" {
		config.Network = DefaultNetwork
	}

	if config.Addr == "" {
		config.Addr = DefaultAddr
	}

	if config.MaxAgeSeconds <= 0 {
		config.MaxAgeSeconds = DefaultMaxAgeSeconds //42 minutes, yes the 42 number :)
	}

	pool := &redis.Pool{IdleTimeout: config.IdleTimeout, MaxIdle: config.MaxIdle, MaxActive: config.MaxActive}
	pool.TestOnBorrow = func(c redis.Conn, t time.Time) error {
		_, err := c.Do("PING")
		return err
	}
	if config.Database != "" {
		pool.Dial = func() (redis.Conn, error) {
			c, err := dial(config.Network, config.Addr, config.Password)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("SELECT", config.Database); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		}
	} else {
		pool.Dial = func() (redis.Conn, error) {
			return dial(config.Network, config.Addr, config.Password)
		}
	}
	r.Connected = true
	r.pool = pool
}

func Empty() *Service {
	return NewFromPool(&redis.Pool{})
}

func New(config *Config) *Service {
	r := Empty()
	r.Config = config
	r.Connect()
	return r
}

func NewFromPool(pool *redis.Pool) *Service {
	return &Service{pool: pool, Config: &Config{}}
}

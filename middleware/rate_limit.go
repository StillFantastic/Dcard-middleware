package middleware

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"net/http"
	"dcard/redis_manager"
	"strconv"
	"strings"
	"time"
)

const WINDOW_SIZE = 60 * 60
const MAX_REQUEST_COUNT = 1000

type RateLimiter struct{}
type RequestRecord struct {
	RequestTimestamp int64 `json:"requestTimestamp"`
	RequestCount     int `json:"requestCount"`
}

func checkRate(remoteIP string) (*RequestRecord, error) {
	record := RequestRecord{}
	value, err := redis_manager.Get(remoteIP)
	if err != nil {
		// If this ip is not in the redis, create a new record for it
		if err == redis.ErrNil {
			record = RequestRecord{
				RequestTimestamp: time.Now().Unix(),
				RequestCount:     1,
			}
			data, err := json.Marshal(record)
			if err != nil {
					return nil, err
			}
			conn := redis_manager.Pool.Get()
			defer conn.Close()
			_, err = conn.Do("SETEX", remoteIP, WINDOW_SIZE, string(data))
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		// If record is found, check the number of requests made
		err = json.Unmarshal([]byte(value), &record)
		if err != nil {
			return nil, err
		}
		// Update the record
		record.RequestCount += 1
		data, err := json.Marshal(record)
		if err != nil {
			return nil, err
		}
		if record.RequestCount <= MAX_REQUEST_COUNT + 1 {
			now := time.Now().Unix()
			exp := record.RequestTimestamp + WINDOW_SIZE - now
			if exp < 0 {
				exp = 0
			}
			conn := redis_manager.Pool.Get()
			defer conn.Close()
			_, err = conn.Do("SETEX", remoteIP, exp, string(data))
			if err != nil {
				return nil, err
			}
		}
	}
	return &record, nil
}

func (*RateLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	remoteIP := strings.Split(r.RemoteAddr, ":")[0]
	record, err := checkRate(remoteIP)
	if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
	}

	now := time.Now().Unix()
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(record.RequestTimestamp + WINDOW_SIZE - now, 10))
	if record.RequestCount > MAX_REQUEST_COUNT {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusTooManyRequests)
		return
	} else {
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(MAX_REQUEST_COUNT - record.RequestCount))
	}
	next.ServeHTTP(w, r)
}

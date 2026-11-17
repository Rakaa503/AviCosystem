package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

const rateLimitScript = `
local current = redis.call("INCR", KEYS[1])

if current == 1 then
    redis.call("PEXPIRE", KEYS[1], ARGV[1])
end

local ttl = redis.call("PTTL", KEYS[1])

return {current, ttl}
`

func RateLimit(client *redis.Client, limit int, window time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		ip := c.IP()

		key := fmt.Sprintf("avicosystem:rate_limit:%s", ip)

		ctx, cancel := context.WithTimeout(
			c.Context(),
			3*time.Second,
		)
		defer cancel()

		result, err := client.Eval(
			ctx,
			rateLimitScript,
			[]string{key},
			window.Milliseconds(),
		).Slice()

		if err != nil {
			return fiber.NewError(
				fiber.StatusServiceUnavailable,
				"Rate limiting service unavailable",
			)
		}

		if len(result) != 2 {
			return fiber.NewError(
				fiber.StatusInternalServerError,
				"Invalid rate limit response",
			)
		}

		current, err := strconv.ParseInt(fmt.Sprint(result[0]), 10, 64)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError)
		}

		ttl, err := strconv.ParseInt(fmt.Sprint(result[1]), 10, 64)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError)
		}

		remaining := int64(limit) - current
		if remaining < 0 {
			remaining = 0
		}

		c.Set("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))

		if ttl < 0 {
			ttl = window.Milliseconds()
		}

		resetSeconds := (ttl + 999) / 1000
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetSeconds, 10))

		if current > int64(limit) {
			retryAfter := resetSeconds
			if retryAfter < 1 {
				retryAfter = 1
			}

			c.Set("Retry-After", strconv.FormatInt(retryAfter, 10))

			return fiber.NewError(
				fiber.StatusTooManyRequests,
				"Too many requests",
			)
		}

		return c.Next()
	}
}

package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	MaxRequests  int           // Maximum requests allowed
	Window       time.Duration // Time window
	KeyPrefix    string        // Redis key prefix
	IsProduction bool          // If true, fail closed on Redis errors (secure)
}

// RateLimitMiddleware creates a rate limiting middleware using Redis
func RateLimitMiddleware(redisClient *redis.Client, config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (IP address)
		clientIP := c.ClientIP()
		key := fmt.Sprintf("%s:%s", config.KeyPrefix, clientIP)

		ctx := context.Background()

		// Increment counter
		pipe := redisClient.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, config.Window)
		_, err := pipe.Exec(ctx)

		if err != nil {
			log.Printf("Rate limit Redis error for %s: %v", clientIP, err)

			// Security: Fail closed in production, fail open in development
			if config.IsProduction {
				// Production: Reject request to prevent rate limit bypass
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error":   "Service temporarily unavailable",
					"message": "Please try again in a moment",
				})
				c.Abort()
				return
			}

			// Development: Allow request for easier testing
			c.Next()
			return
		}

		// Get current count
		count := incr.Val()

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.MaxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, config.MaxRequests-int(count))))

		// Check if limit exceeded
		if count > int64(config.MaxRequests) {
			// Get TTL for Retry-After header
			ttl, _ := redisClient.TTL(ctx, key).Result()
			c.Header("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("Too many requests. Please try again in %d seconds.", int(ttl.Seconds())),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthRateLimitMiddleware creates rate limiting for authentication endpoints
// Relaxed for better user experience - allows multiple login attempts
func AuthRateLimitMiddleware(redisClient *redis.Client, isProduction bool) gin.HandlerFunc {
	return RateLimitMiddleware(redisClient, RateLimitConfig{
		MaxRequests:  20,               // 20 attempts (relaxed from 5)
		Window:       15 * time.Minute, // per 15 minutes
		KeyPrefix:    "ratelimit:auth",
		IsProduction: isProduction,
	})
}

// APIRateLimitMiddleware creates general rate limiting for API endpoints
// Relaxed for better performance - allows more frequent requests
func APIRateLimitMiddleware(redisClient *redis.Client, isProduction bool) gin.HandlerFunc {
	return RateLimitMiddleware(redisClient, RateLimitConfig{
		MaxRequests:  300,             // 300 requests (relaxed from 100)
		Window:       1 * time.Minute, // per minute
		KeyPrefix:    "ratelimit:api",
		IsProduction: isProduction,
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

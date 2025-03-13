package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/vultisig/vultisigner/config"
)

func (s *Server) statsdMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		duration := time.Since(start).Milliseconds()

		// Send metrics to statsd
		_ = s.sdClient.Incr("http.requests", []string{"path:" + c.Path()}, 1)
		_ = s.sdClient.Timing("http.response_time", time.Duration(duration)*time.Millisecond, []string{"path:" + c.Path()}, 1)
		_ = s.sdClient.Incr("http.status."+fmt.Sprint(c.Response().Status), []string{"path:" + c.Path(), "method:" + c.Request().Method}, 1)

		return err
	}
}

func (s *Server) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg, err := config.ReadConfig("config-verifier")
		if err != nil {
			s.logger.Fatal("Failed to read verifier config: ", err)
		}

		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Missing token"})
		}

		tokenStr := authHeader[len("Bearer "):]

		// parse and validate JWT
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return cfg.Server.Auth.JwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
		}

		// TODO: create endpoint to encode and issue this token
		claims, _ := token.Claims.(jwt.MapClaims)
		c.Set("id", claims["id"])

		return next(c)
	}
}

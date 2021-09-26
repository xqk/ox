package oecho

import (
	"fmt"
	"net/http"
	"ox/pkg/olog"
	"runtime"
	"time"

	"ox/pkg/metric"
	"ox/pkg/trace"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func extractAID(c echo.Context) string {
	return c.Request().Header.Get("AID")
}

// RecoverMiddleware ...
func recoverMiddleware(logger *olog.Logger, slowQueryThresholdInMilli int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (err error) {
			var beg = time.Now()
			var fields = make([]olog.Field, 0, 8)

			defer func() {
				fields = append(fields, zap.Float64("cost", time.Since(beg).Seconds()))
				if rec := recover(); rec != nil {
					switch rec := rec.(type) {
					case error:
						err = rec
					default:
						err = fmt.Errorf("%v", rec)
					}

					stack := make([]byte, 4096)
					length := runtime.Stack(stack, true)
					fields = append(fields, zap.ByteString("stack", stack[:length]))
				}
				fields = append(fields,
					zap.String("method", ctx.Request().Method),
					zap.Int("code", ctx.Response().Status),
					zap.String("host", ctx.Request().Host),
					zap.String("path", ctx.Request().URL.Path),
				)
				if slowQueryThresholdInMilli > 0 {
					if cost := int64(time.Since(beg)) / 1e6; cost > slowQueryThresholdInMilli {
						fields = append(fields, zap.Int64("slow", cost))
					}
				}
				if err != nil {
					fields = append(fields, zap.String("err", err.Error()))
					logger.Error("access", fields...)
					return
				}
				logger.Info("access", fields...)
			}()

			return next(ctx)
		}
	}
}

func metricServerInterceptor() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			beg := time.Now()
			err = next(c)
			method := c.Request().Method + "_" + c.Path()
			peer := c.RealIP()
			if aid := extractAID(c); aid != "" {
				peer += "?aid=" + aid
			}
			metric.ServerHandleHistogram.Observe(time.Since(beg).Seconds(), metric.TypeHTTP, method, peer)
			metric.ServerHandleCounter.Inc(metric.TypeHTTP, method, peer, http.StatusText(c.Response().Status))
			return err
		}
	}
}
func traceServerInterceptor() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			span, ctx := trace.StartSpanFromContext(
				c.Request().Context(),
				c.Request().Method+" "+c.Path(),
				trace.TagComponent("http"),
				trace.TagSpanKind("server"),
				trace.HeaderExtractor(c.Request().Header),
				trace.CustomTag("http.url", c.Path()),
				trace.CustomTag("http.method", c.Request().Method),
				trace.CustomTag("peer.ipv4", c.RealIP()),
			)
			c.SetRequest(c.Request().WithContext(ctx))
			defer span.Finish()
			return next(c)
		}
	}
}

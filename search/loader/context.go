package loader

import (
	"context"

	"github.com/OptLTD/library/search/source"
)

type loaderContextKey string

const (
	loaderBaseKey  loaderContextKey = "loader.base"  // string (path)
	loaderScopeKey loaderContextKey = "loader.scope" // map[string]any
	loaderLogIDKey loaderContextKey = "loader.logid" // string
	// loader extends, 用于加载父模型 schema
	loaderExtendsKey loaderContextKey = "loader.extends" // ExtendsLoader
	loaderVisitedKey loaderContextKey = "loader.visited" // map[string]bool
)

// ExtendsLoader 按 model 名加载父模型 schema（与运行时 GetSource 同源，不依赖 bundle 目录）。
type ExtendsLoader func(ctx context.Context, model string) (*source.Value, error)

// WithScope 设置 loader 的查询范围（用于 MYSQL/MONGO）
func WithScope(ctx context.Context, scope map[string]any) context.Context {
	return context.WithValue(ctx, loaderScopeKey, scope)
}

// GetScope 获取 loader 的查询范围
func GetScope(ctx context.Context) map[string]any {
	if ctx == nil {
		return nil
	}
	if val := ctx.Value(loaderScopeKey); val != nil {
		if scope, ok := val.(map[string]any); ok {
			return scope
		}
	}
	return nil
}

// WithBase 设置 loader 的 base path（用于 JSON/EMBED）
func WithBase(ctx context.Context, base string) context.Context {
	return context.WithValue(ctx, loaderBaseKey, base)
}

// GetBase 获取 loader 的 base path
func GetBase(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val := ctx.Value(loaderBaseKey); val != nil {
		if base, ok := val.(string); ok {
			return base
		}
	}
	return ""
}

// WithLogID 设置 LogID（统一日志使用）
func WithLogID(ctx context.Context, logID string) context.Context {
	return context.WithValue(ctx, loaderLogIDKey, logID)
}

func getExtendsVisited(ctx context.Context) map[string]bool {
	if ctx == nil {
		return nil
	}
	if val := ctx.Value(loaderVisitedKey); val != nil {
		if visited, ok := val.(map[string]bool); ok {
			return visited
		}
	}
	return nil
}

// markExtendsVisiting 标记正在加载的 extends 父模型，用于检测环。
func markExtendsVisiting(ctx context.Context, model string) (context.Context, bool) {
	if model == "" {
		return ctx, true
	}
	visited := getExtendsVisited(ctx)
	if visited != nil && visited[model] {
		return ctx, false
	}
	next := map[string]bool{}
	for key, val := range visited {
		next[key] = val
	}
	next[model] = true
	return context.WithValue(ctx, loaderVisitedKey, next), true
}

// WithExtendsLoader 注册 extends 父模型加载器（embed 场景由 entry.LoadEmbedSource 提供）。
func WithExtendsLoader(ctx context.Context, fn ExtendsLoader) context.Context {
	if fn == nil {
		return ctx
	}
	return context.WithValue(ctx, loaderExtendsKey, fn)
}

// GetExtendsLoader 获取 extends 父模型加载器
func GetExtendsLoader(ctx context.Context) ExtendsLoader {
	if ctx == nil {
		return nil
	}
	if val := ctx.Value(loaderExtendsKey); val != nil {
		if fn, ok := val.(ExtendsLoader); ok {
			return fn
		}
	}
	return nil
}

// GetLogID 获取 LogID
func GetLogID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val := ctx.Value(loaderLogIDKey); val != nil {
		if logID, ok := val.(string); ok {
			return logID
		}
	}
	return ""
}

package loader

import "context"

type loaderContextKey string

const (
	loaderBaseKey  loaderContextKey = "loader.base"  // string (path)
	loaderScopeKey loaderContextKey = "loader.scope" // map[string]any
	loaderLogIDKey loaderContextKey = "loader.logid" // string
)

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

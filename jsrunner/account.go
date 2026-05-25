package js

import "context"

type accountKey struct{}

// WithAccount attaches the current JS/storage caller identity to ctx (e.g. *request.Account from search package).
func WithAccount(ctx context.Context, account any) context.Context {
	if c, ok := ctx.(interface{ SetValue(key, value any) }); ok {
		c.SetValue(accountKey{}, account)
		return ctx
	}
	return context.WithValue(ctx, accountKey{}, account)
}

// AccountFromCtx returns the value set by WithAccount, or nil if unset.
func AccountFromCtx(ctx context.Context) any {
	if ctx == nil {
		return nil
	}
	return ctx.Value(accountKey{})
}

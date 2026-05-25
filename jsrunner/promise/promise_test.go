package promise

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grafana/sobek"
	js "jsrunner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testVM(t *testing.T) (js.VM, *sobek.Runtime, context.Context, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	vm := js.NewVM()
	return vm, vm.Runtime(), ctx, cancel
}

// flushVM lets the event loop run jobs enqueued by promise.New goroutines.
func flushVM(ctx context.Context, vm js.VM) error {
	return vm.Run(ctx, func() error { return nil })
}

func TestPromise(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		t.Run("resolve", func(t *testing.T) {
			vm, rt, ctx, cancel := testVM(t)
			defer cancel()

			var v sobek.Value
			err := vm.Run(ctx, func() error {
				v = New(rt, func(callback Callback) {
					callback(func() (any, error) {
						return "resolve", nil
					})
				})
				return nil
			})
			require.NoError(t, err)
			require.NoError(t, flushVM(ctx, vm))

			result, err := Result(v)
			require.NoError(t, err)
			assert.Equal(t, "resolve", result)
		})

		t.Run("reject", func(t *testing.T) {
			vm, rt, ctx, cancel := testVM(t)
			defer cancel()

			var v sobek.Value
			err := vm.Run(ctx, func() error {
				v = New(rt, func(callback Callback) {
					callback(func() (any, error) {
						return nil, errors.New("reject")
					})
				})
				return nil
			})
			require.NoError(t, err)
			require.NoError(t, flushVM(ctx, vm))

			_, err = Result(v)
			assert.ErrorContains(t, err, "reject")
		})

		t.Run("panic on async", func(t *testing.T) {
			vm, rt, ctx, cancel := testVM(t)
			defer cancel()

			assert.NotPanics(t, func() {
				var v sobek.Value
				err := vm.Run(ctx, func() error {
					v = New(rt, func(callback Callback) {
						panic("reject")
					})
					return nil
				})
				require.NoError(t, err)
				require.NoError(t, flushVM(ctx, vm))

				_, err = Result(v)
				assert.ErrorContains(t, err, "reject")
			})
		})

		t.Run("panic on callback", func(t *testing.T) {
			vm, rt, ctx, cancel := testVM(t)
			defer cancel()

			assert.NotPanics(t, func() {
				var v sobek.Value
				err := vm.Run(ctx, func() error {
					v = New(rt, func(callback Callback) {
						callback(func() (any, error) {
							panic("reject")
						})
					})
					return nil
				})
				require.NoError(t, err)
				require.NoError(t, flushVM(ctx, vm))

				_, err = Result(v)
				assert.ErrorContains(t, err, "reject")
			})
		})
	})

	t.Run("example", func(t *testing.T) {
		result := `{"foo":"bar"}`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(result))
		}))
		defer server.Close()

		fetch := func(call sobek.FunctionCall, rt *sobek.Runtime) sobek.Value {
			return New(rt, func(callback Callback) {
				res, err := http.Get(call.Argument(0).String())
				callback(func() (any, error) {
					if err != nil {
						return nil, err
					}
					defer res.Body.Close()
					data, err := io.ReadAll(res.Body)
					if err != nil {
						return nil, err
					}
					return string(data), nil
				})
			})
		}

		vm, _, ctx, cancel := testVM(t)
		defer cancel()

		var value sobek.Value
		err := vm.Run(ctx, func() error {
			_ = vm.Runtime().Set("fetch", fetch)
			var runErr error
			value, runErr = vm.Runtime().RunString(fmt.Sprintf(`fetch("%s")`, server.URL))
			return runErr
		})
		require.NoError(t, err)
		require.NoError(t, flushVM(ctx, vm))

		v, err := Result(value)
		require.NoError(t, err)
		assert.Equal(t, result, v)
	})
}

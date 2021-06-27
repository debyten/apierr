package apierr

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

const dummyTarget = "http://dummy.com"

func TestHandle(t *testing.T) {
	type args struct {
		err error
		w   *httptest.ResponseRecorder
	}
	tests := []struct {
		name string
		args args
		want bool
		verify func(w *httptest.ResponseRecorder) error
	}{
		{
			name: "not apierr",
			args: args{
				err: errors.New("not an apierr"),
				w:   httptest.NewRecorder(),
			},
			want: false,
		},
		{
			name: "simple apierr",
			args: args{
				err: New(errors.New("new err"), http.StatusBadRequest),
				w:   httptest.NewRecorder(),
			},
			want: true,
			verify: func(w *httptest.ResponseRecorder) error {
				if w.Code != http.StatusBadRequest {
					return fmt.Errorf("expected %d status code, got %d", http.StatusBadRequest, w.Code)
				}
				return nil
			},
		},
		{
			name: "simple apierr with extra",
			args: args{
				err: New(errors.New("new err"), http.StatusBadRequest, "my", "custom", "headers"),
				w:   httptest.NewRecorder(),
			},
			want: true,
			verify: func(w *httptest.ResponseRecorder) error {
				if w.Code != http.StatusBadRequest {
					return fmt.Errorf("expected %d status code, got %d", http.StatusBadRequest, w.Code)
				}
				result := w.Header().Get(ErrHeader)
				if result != "my,custom,headers" {
					return fmt.Errorf("expected ErrHeader 'my,custom,headers', got: %s", result)
				}
				return nil
			},
		},
		{
			name: "wrapped apierr",
			args: args{
				err: fmt.Errorf("wrapped: %w", New(errors.New("new err"), http.StatusBadRequest, "my", "custom", "headers")),
				w:   httptest.NewRecorder(),
			},
			want: true,
			verify: func(w *httptest.ResponseRecorder) error {
				if w.Code != http.StatusBadRequest {
					return fmt.Errorf("expected %d status code, got %d", http.StatusBadRequest, w.Code)
				}
				result := w.Header().Get(ErrHeader)
				if result != "my,custom,headers" {
					return fmt.Errorf("expected ErrHeader 'my,custom,headers', got: %s", result)
				}
				return nil
			},
		},
		{
			name: "apierr with custom headers",
			args: args{
				err: FromText("my err", http.StatusBadRequest).CustomHeader("X-Custom", "XXX"),
				w:   httptest.NewRecorder(),
			},
			want: true,
			verify: func(w *httptest.ResponseRecorder) error {
				if w.Code != http.StatusBadRequest {
					return fmt.Errorf("expected %d status code, got %d", http.StatusBadRequest, w.Code)
				}
				result := w.Header().Get("X-Custom")
				if result != "XXX" {
					return fmt.Errorf("expected header X-Custom = XXX got: %s", result)
				}
				return nil
			},
		},
		{
			name: "apierr with GOOD status code",
			args: args{
				err: FromText("no content", http.StatusNoContent),
				w:   httptest.NewRecorder(),
			},
			want: true,
			verify: func(w *httptest.ResponseRecorder) error {
				if w.Code != http.StatusNoContent {
					return fmt.Errorf("expected %d status code, got %d", http.StatusBadRequest, w.Code)
				}
				b, err := io.ReadAll(w.Result().Body)
				if err != nil {
					return err
				}
				if len(b) > 0 {
					return fmt.Errorf("unexpected body content %s", string(b))
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Handle(tt.args.err, tt.args.w); got != tt.want {
				t.Errorf("Handle() = %v, want %v", got, tt.want)
			}
			if tt.verify == nil {
				return
			}
			if err := tt.verify(tt.args.w); err != nil {
				t.Errorf("Verify() = %v", err)
			}
		})
	}
}

func TestHandleISE(t *testing.T) {
	type args struct {
		err error
		w   *httptest.ResponseRecorder
		r   *http.Request
	}
	tests := []struct {
		name string
		args args
		verify func(w *httptest.ResponseRecorder) error
	}{
		{
			name: "internal server error",
			args: args{
				err: errors.New("ISE"),
				w:   httptest.NewRecorder(),
				r:   httptest.NewRequest(http.MethodGet, dummyTarget, nil),
			},
			verify: func(w *httptest.ResponseRecorder) error {
				if w.Code != http.StatusInternalServerError {
					return fmt.Errorf("expected %d, got %d", http.StatusInternalServerError, w.Code)
				}
				return nil
			},
		},
		{
			name: "normal handle",
			args: args{
				err: FromText("err", http.StatusBadRequest),
				w:   httptest.NewRecorder(),
				r:   httptest.NewRequest(http.MethodGet, dummyTarget, nil),
			},
			verify: func(w *httptest.ResponseRecorder) error {
				if w.Code != http.StatusBadRequest {
					return fmt.Errorf("expected %d, got %d", http.StatusBadRequest, w.Code)
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleISE(tt.args.err, tt.args.w, tt.args.r)
			if err := tt.verify(tt.args.w); err != nil {
				t.Errorf("Verify() got: %v", err)
			}
		})
	}
}

func TestHandleISEDBNotFound(t *testing.T) {
	toReset := DefaultDBNotFoundHandler
	defer func() {
		DefaultDBNotFoundHandler = toReset
	}()
	DefaultDBNotFoundHandler = func(err error) bool {
		return true
	}
	rec := httptest.NewRecorder()
	HandleISE(errors.New(""), rec, httptest.NewRequest(http.MethodGet, dummyTarget, nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected not found, got %d", rec.Code)
	}
}

func TestHandleISEDecorator(t *testing.T) {
	const customerDecHeader = "X-Custom-Dec"
	const expectedValue = 1
	type xDecorator int
	const xDecoratorValue xDecorator = iota
	dummyCtx := context.WithValue(context.Background(), xDecoratorValue, 1)
	dec := func(w http.ResponseWriter, r *http.Request) {
		v, ok := r.Context().Value(xDecoratorValue).(int)
		if !ok {
			return
		}
		w.Header().Add(customerDecHeader, strconv.Itoa(v))
	}
	AddDecorator(dec)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, dummyTarget, nil)
	HandleISE(errors.New(""), rec, req.WithContext(dummyCtx))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal server error, got %d", rec.Code)
	}
	result := rec.Header().Get(customerDecHeader)
	if result != "1" {
		t.Fatalf("expected %s header equal to %d got %s", customerDecHeader, expectedValue, result)
	}
}
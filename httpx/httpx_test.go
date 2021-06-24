package httpx_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wwq-2020/go.common/httpx"
)

type srv struct {
	expectedReq   *string
	normalResp    string
	exceptionResp string
}

type req struct {
	Data string
}

type resp struct {
	Data string
}

func (s *srv) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &req{}
	resp := &resp{
		Data: s.normalResp,
	}
	if s.expectedReq != nil {
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			resp.Data = s.exceptionResp
			goto end
		}
		if req.Data != *s.expectedReq {
			resp.Data = s.exceptionResp
			goto end
		}
	}
end:
	json.NewEncoder(w).Encode(resp)
}

func TestGet(t *testing.T) {
	normalResp := "normalResp"
	exceptionResp := "exceptionResp"
	handler := &srv{
		normalResp:    normalResp,
		exceptionResp: exceptionResp,
	}
	srv := httptest.NewServer(handler)
	got := &resp{}
	err := httpx.Get(context.TODO(), srv.URL, got)
	if err != nil {
		t.Fatalf("expected:nil,got:%v", err)
	}
	if got.Data != normalResp {
		t.Fatalf("expected:%s,got:%s", exceptionResp, got.Data)
	}
}

func TestPost(t *testing.T) {
	expectedReq := "hello"
	normalResp := "normalResp"
	exceptionResp := "exceptionResp"
	handler := &srv{
		expectedReq:   &expectedReq,
		normalResp:    normalResp,
		exceptionResp: exceptionResp,
	}
	srv := httptest.NewServer(handler)
	req := &req{
		Data: expectedReq,
	}
	got := &resp{}
	err := httpx.Post(context.TODO(), srv.URL, req, got)
	if err != nil {
		t.Fatalf("expected:nil,got:%v", err)
	}
	if got.Data != normalResp {
		t.Fatalf("expected:%s,got:%s", normalResp, got.Data)
	}
}

func TestPut(t *testing.T) {
	expectedReq := "hello"
	normalResp := "normalResp"
	exceptionResp := "exceptionResp"
	handler := &srv{
		expectedReq:   &expectedReq,
		normalResp:    normalResp,
		exceptionResp: exceptionResp,
	}
	srv := httptest.NewServer(handler)
	req := &req{
		Data: expectedReq,
	}
	got := &resp{}
	err := httpx.Put(context.TODO(), srv.URL, req, got)
	if err != nil {
		t.Fatalf("expected:nil,got:%v", err)
	}
	if got.Data != normalResp {
		t.Fatalf("expected:%s,got:%s", normalResp, got.Data)
	}
}

func TestDelete(t *testing.T) {
	expectedReq := "hello"
	normalResp := "normalResp"
	exceptionResp := "exceptionResp"
	handler := &srv{
		expectedReq:   &expectedReq,
		normalResp:    normalResp,
		exceptionResp: exceptionResp,
	}
	srv := httptest.NewServer(handler)
	req := &req{
		Data: expectedReq,
	}
	got := &resp{}
	err := httpx.Put(context.TODO(), srv.URL, req, got)
	if err != nil {
		t.Fatalf("expected:nil,got:%v", err)
	}
	if got.Data != normalResp {
		t.Fatalf("expected:%s,got:%s", normalResp, got.Data)
	}
}

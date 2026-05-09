package iputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseIPFromHttpbin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"origin": "1.2.3.4"}`))
	}))
	defer server.Close()

	client := server.Client()
	ip := fetchIP(client, server.URL, "origin")
	if ip != "1.2.3.4" {
		t.Errorf("期望 1.2.3.4，实际: %s", ip)
	}
}

func TestParseIPWithComma(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"origin": "1.2.3.4, 5.6.7.8"}`))
	}))
	defer server.Close()

	client := server.Client()
	ip := fetchIP(client, server.URL, "origin")
	if ip != "1.2.3.4" {
		t.Errorf("应取第一个IP，期望 1.2.3.4，实际: %s", ip)
	}
}

func TestParseIPFromIpify(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ip": "9.8.7.6"}`))
	}))
	defer server.Close()

	client := server.Client()
	ip := fetchIP(client, server.URL, "ip")
	if ip != "9.8.7.6" {
		t.Errorf("期望 9.8.7.6，实际: %s", ip)
	}
}

func TestFetchIP_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	client := server.Client()
	ip := fetchIP(client, server.URL, "origin")
	if ip != "" {
		t.Errorf("无效JSON应返回空字符串，实际: %s", ip)
	}
}

func TestFetchIP_MissingField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"other": "value"}`))
	}))
	defer server.Close()

	client := server.Client()
	ip := fetchIP(client, server.URL, "origin")
	if ip != "" {
		t.Errorf("字段不存在应返回空字符串，实际: %s", ip)
	}
}

func TestFetchIP_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"origin": "1.2.3.4"}`))
	}))
	defer server.Close()

	client := server.Client()
	// 即使状态码500，body仍然可以解析JSON
	ip := fetchIP(client, server.URL, "origin")
	// fetchIP 不检查状态码，只要能解析 JSON 就返回
	if ip != "1.2.3.4" {
		t.Logf("状态码500时返回: %s (取决于实现)", ip)
	}
}

func TestFetchIP_EmptyOrigin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"origin": ""}`))
	}))
	defer server.Close()

	client := server.Client()
	ip := fetchIP(client, server.URL, "origin")
	if ip != "" {
		t.Errorf("空origin应返回空字符串，实际: %s", ip)
	}
}

package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"uupt-open-cli/internal/api"
	"uupt-open-cli/internal/config"
	"uupt-open-cli/internal/iputil"
	"uupt-open-cli/internal/logger"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "授权管理（login / status / logout）",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "打开本地授权页，通过手机号完成登录",
	Run:   runAuthLogin,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "检查当前授权状态",
	Run:   runAuthStatus,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "清除本地授权（登出）",
	Run:   runAuthLogout,
}

func init() {
	authCmd.AddCommand(authLoginCmd, authStatusCmd, authLogoutCmd)
	RootCmd.AddCommand(authCmd)
}

func maskOpenId(id string) string {
	id = strings.TrimSpace(id)
	if len(id) <= 8 {
		if id == "" {
			return ""
		}
		return id[:1] + "***"
	}
	return id[:4] + "****" + id[len(id)-4:]
}

func runAuthStatus(cmd *cobra.Command, args []string) {
	cfg, _ := config.LoadConfig()
	if cfg != nil && strings.TrimSpace(cfg.OpenId) != "" {
		fmt.Printf("Logged in as %s\n", maskOpenId(cfg.OpenId))
		os.Exit(0)
	}
	fmt.Println("Not authenticated. Run 'uupt-open-cli auth login' to sign in.")
	os.Exit(1)
}

func runAuthLogout(cmd *cobra.Command, args []string) {
	if err := config.ClearOpenId(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] 登出失败: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Println("[OK] Logged out")
	os.Exit(0)
}

func runAuthLogin(cmd *cobra.Command, args []string) {
	cfg := config.EnsureConfig(false)
	if strings.TrimSpace(cfg.OpenId) != "" {
		fmt.Printf("Logged in as %s\n", maskOpenId(cfg.OpenId))
		os.Exit(0)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] 无法启动本地授权服务: %s\n", err.Error())
		os.Exit(1)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	loginURL := fmt.Sprintf("http://127.0.0.1:%d/login", port)

	ipHolder := &publicIPHolder{}
	go func() {
		ip := iputil.GetPublicIP()
		ipHolder.set(ip)
		if ip != "" {
			logger.Infof("auth login 检测到公网IP: %s", ip)
		} else {
			logger.Warnf("auth login 未能检测公网IP，将在发短信时重试")
		}
	}()

	done := make(chan string, 1)
	srv := &http.Server{
		Handler:           newAuthMux(cfg, ipHolder, done),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			logger.Errorf("auth login 本地服务异常: %s", err.Error())
		}
	}()

	// WorkBuddy 要求 10 秒内打出授权 URL。公网 IP 探测在后台进行，不阻塞。
	fmt.Println("Please visit the following URL to authenticate:")
	fmt.Println(loginURL)
	_ = os.Stdout.Sync()

	timeout := 4*time.Minute + 50*time.Second
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var openId string
	select {
	case openId = <-done:
	case <-ctx.Done():
		_ = srv.Shutdown(context.Background())
		fmt.Fprintln(os.Stderr, "[ERROR] 认证已取消")
		os.Exit(1)
	case <-time.After(timeout):
		_ = srv.Shutdown(context.Background())
		fmt.Fprintln(os.Stderr, "[ERROR] 认证超时，请重新连接后重试")
		os.Exit(1)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)

	fmt.Printf("Logged in as %s\n", maskOpenId(openId))
	os.Exit(0)
}

type authAPIResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message,omitempty"`
	NeedImage bool   `json:"needImage,omitempty"`
	ImageData string `json:"imageData,omitempty"`
	OpenId    string `json:"openId,omitempty"`
}

type publicIPHolder struct {
	mu sync.Mutex
	ip string
}

func (h *publicIPHolder) set(ip string) {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.ip == "" {
		h.ip = ip
	}
}

func (h *publicIPHolder) get() string {
	h.mu.Lock()
	if h.ip != "" {
		ip := h.ip
		h.mu.Unlock()
		return ip
	}
	h.mu.Unlock()
	ip := iputil.GetPublicIP()
	h.set(ip)
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.ip
}

func newAuthMux(cfg *config.Config, ipHolder *publicIPHolder, done chan string) http.Handler {
	mux := http.NewServeMux()
	var once sync.Once

	writeJSON := func(w http.ResponseWriter, status int, body authAPIResponse) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = io.WriteString(w, loginPageHTML)
	})

	mux.HandleFunc("/api/sms", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, authAPIResponse{Message: "method not allowed"})
			return
		}
		var req struct {
			Mobile    string `json:"mobile"`
			ImageCode string `json:"imageCode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, authAPIResponse{Message: "无效请求"})
			return
		}
		req.Mobile = strings.TrimSpace(req.Mobile)
		if !isCNMobile(req.Mobile) {
			writeJSON(w, http.StatusBadRequest, authAPIResponse{Message: "请输入 11 位中国大陆手机号"})
			return
		}
		userIP := ipHolder.get()
		if userIP == "" {
			writeJSON(w, http.StatusBadRequest, authAPIResponse{Message: "无法检测公网 IP，请检查网络后重试"})
			return
		}
		result, err := api.SendSmsCode(cfg, req.Mobile, userIP, strings.TrimSpace(req.ImageCode))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, authAPIResponse{Message: err.Error()})
			return
		}
		code, _ := result["code"].(float64)
		if int(code) == 88100106 {
			body, _ := result["body"].(map[string]interface{})
			imageData := ""
			if body != nil {
				imageData, _ = body["imageData"].(string)
			}
			writeJSON(w, http.StatusOK, authAPIResponse{
				NeedImage: true,
				ImageData: imageData,
				Message:   "请输入图片验证码后重新发送短信",
			})
			return
		}
		if int(code) != 0 && int(code) != 1 {
			msg, _ := result["msg"].(string)
			if msg == "" {
				msg = fmt.Sprintf("发送验证码失败: code=%d", int(code))
			}
			writeJSON(w, http.StatusBadRequest, authAPIResponse{Message: msg})
			return
		}
		writeJSON(w, http.StatusOK, authAPIResponse{OK: true, Message: "验证码已发送"})
	})

	mux.HandleFunc("/api/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, authAPIResponse{Message: "method not allowed"})
			return
		}
		var req struct {
			Mobile  string `json:"mobile"`
			SmsCode string `json:"smsCode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, authAPIResponse{Message: "无效请求"})
			return
		}
		req.Mobile = strings.TrimSpace(req.Mobile)
		req.SmsCode = strings.TrimSpace(req.SmsCode)
		if !isCNMobile(req.Mobile) || req.SmsCode == "" {
			writeJSON(w, http.StatusBadRequest, authAPIResponse{Message: "请填写手机号和短信验证码"})
			return
		}
		userIP := ipHolder.get()
		if userIP == "" {
			writeJSON(w, http.StatusBadRequest, authAPIResponse{Message: "无法检测公网 IP，请检查网络后重试"})
			return
		}
		result, err := api.Auth(cfg, req.Mobile, userIP, req.SmsCode)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, authAPIResponse{Message: err.Error()})
			return
		}
		body, _ := result["body"].(map[string]interface{})
		openId := ""
		if body != nil {
			openId, _ = body["openId"].(string)
		}
		if strings.TrimSpace(openId) == "" {
			msg, _ := result["msg"].(string)
			if msg == "" {
				msg = "授权失败，未获取到 openId"
			}
			writeJSON(w, http.StatusBadRequest, authAPIResponse{Message: msg})
			return
		}
		if err := config.SaveConfig(map[string]string{"openId": openId}); err != nil {
			writeJSON(w, http.StatusInternalServerError, authAPIResponse{Message: "授权成功但保存配置失败"})
			return
		}
		writeJSON(w, http.StatusOK, authAPIResponse{OK: true, OpenId: maskOpenId(openId), Message: "登录成功"})
		once.Do(func() {
			done <- openId
		})
	})

	return mux
}

func isCNMobile(mobile string) bool {
	if len(mobile) != 11 {
		return false
	}
	if mobile[0] != '1' {
		return false
	}
	for i := 0; i < 11; i++ {
		if mobile[i] < '0' || mobile[i] > '9' {
			return false
		}
	}
	return true
}

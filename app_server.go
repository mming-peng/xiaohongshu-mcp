package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/browser"
)

// AppServer 应用服务器结构体，封装所有服务和处理器
type AppServer struct {
	xiaohongshuService *XiaohongshuService
	mcpServer          *mcp.Server
	router             *gin.Engine
	httpServer         *http.Server
	openBrowsers       []*browser.Browser // 追踪通过 open_homepage 打开的浏览器
	mu                 sync.Mutex         // 保护 openBrowsers 的并发访问
}

// NewAppServer 创建新的应用服务器实例
func NewAppServer(xiaohongshuService *XiaohongshuService) *AppServer {
	appServer := &AppServer{
		xiaohongshuService: xiaohongshuService,
	}

	// 初始化 MCP Server（需要在创建 appServer 之后，因为工具注册需要访问 appServer）
	appServer.mcpServer = InitMCPServer(appServer)

	return appServer
}

// Start 启动服务器
func (s *AppServer) Start(port string) error {
	s.router = setupRoutes(s)

	s.httpServer = &http.Server{
		Addr:    port,
		Handler: s.router,
	}

	// 启动服务器的 goroutine
	go func() {
		logrus.Infof("启动 HTTP 服务器: %s", port)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Errorf("服务器启动失败: %v", err)
			os.Exit(1)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Infof("正在关闭服务器...")

	// 关闭所有打开的浏览器
	s.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		logrus.Warnf("等待连接关闭超时，强制退出: %v", err)
	} else {
		logrus.Infof("服务器已优雅关闭")
	}

	return nil
}

// trackBrowser 注册一个新打开的浏览器实例
func (s *AppServer) trackBrowser(b *browser.Browser) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.openBrowsers = append(s.openBrowsers, b)
	logrus.Infof("已追踪新浏览器实例，当前总数: %d", len(s.openBrowsers))
}

// Shutdown 关闭所有打开的浏览器
func (s *AppServer) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.openBrowsers) > 0 {
		logrus.Infof("正在关闭 %d 个打开的浏览器...", len(s.openBrowsers))
		for i, b := range s.openBrowsers {
			if b != nil {
				b.ForceClose() // 使用 ForceClose 真正关闭浏览器进程
				logrus.Debugf("已关闭浏览器 %d/%d", i+1, len(s.openBrowsers))
			}
		}
		s.openBrowsers = nil
		logrus.Info("所有浏览器已关闭")
	}
}

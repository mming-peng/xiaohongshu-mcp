package browser

import (
	"encoding/json"
	"path/filepath"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/cookies"
)

type browserConfig struct {
	binPath     string
	userDataDir string
}

type Option func(*browserConfig)

func WithBinPath(binPath string) Option {
	return func(c *browserConfig) {
		c.binPath = binPath
	}
}

func WithUserDataDir(userDataDir string) Option {
	return func(c *browserConfig) {
		c.userDataDir = userDataDir
	}
}

// Browser 封装 rod.Browser 以保持接口兼容
type Browser struct {
	browser  *rod.Browser
	launcher *launcher.Launcher
}

func (b *Browser) NewPage() *rod.Page {
	return stealth.MustPage(b.browser)
}

func (b *Browser) Close() {
	// 在复用模式下，不关闭浏览器实例
	// 只释放当前连接，浏览器继续运行以供后续复用
	// 注意：不调用 browser.MustClose() 和 launcher.Cleanup()
	logrus.Debug("释放浏览器连接（浏览器实例继续运行以供复用）")
}

func NewBrowser(headless bool, options ...Option) *Browser {
	cfg := &browserConfig{
		userDataDir: filepath.Join(".", ".browser-data"), // 默认用户数据目录
	}
	for _, opt := range options {
		opt(cfg)
	}

	// 创建 launcher
	l := launcher.New().
		Headless(headless).
		Set("--no-sandbox").
		Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36").
		Set("--remote-debugging-port", "9222"). // 固定端口，允许复用浏览器实例
		Leakless(false)                         // 不使用 leakless 模式，允许浏览器保持运行

	// 设置用户数据目录（关键：保持草稿持久化）
	if cfg.userDataDir != "" {
		l = l.UserDataDir(cfg.userDataDir)
		logrus.Infof("浏览器使用用户数据目录: %s (草稿将持久化保存)", cfg.userDataDir)
	}

	// 设置自定义 Chrome 路径
	if cfg.binPath != "" {
		l = l.Bin(cfg.binPath)
	}

	var browser *rod.Browser
	var url string

	// 尝试连接已存在的浏览器实例
	url, err := l.Launch()
	if err != nil {
		// 如果启动失败，尝试连接到已有的实例
		logrus.Infof("启动新浏览器失败，尝试连接已有实例: %v", err)
		url = "ws://127.0.0.1:9222"
	}

	browser = rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		logrus.Fatalf("连接浏览器失败: %v", err)
	}

	// 加载 cookies
	cookiePath := cookies.GetCookiesFilePath()
	cookieLoader := cookies.NewLoadCookie(cookiePath)

	if data, err := cookieLoader.LoadCookies(); err == nil {
		var cookieList []*proto.NetworkCookie
		if err := json.Unmarshal(data, &cookieList); err == nil {
			browser.MustSetCookies(cookieList...)
			logrus.Debugf("loaded cookies from file successfully")
		}
	} else {
		logrus.Warnf("failed to load cookies: %v", err)
	}

	return &Browser{
		browser:  browser,
		launcher: l,
	}
}

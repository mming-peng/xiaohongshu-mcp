package draft

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Draft 草稿结构
type Draft struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Images     []string  `json:"images"`
	Tags       []string  `json:"tags"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	SavedToXHS bool      `json:"saved_to_xhs"` // 是否已保存到小红书
}

// Manager 草稿管理器
type Manager struct {
	draftsDir string
}

// NewManager 创建草稿管理器
func NewManager(draftsDir string) (*Manager, error) {
	// 确保草稿目录存在
	if err := os.MkdirAll(draftsDir, 0755); err != nil {
		return nil, fmt.Errorf("创建草稿目录失败: %w", err)
	}

	return &Manager{
		draftsDir: draftsDir,
	}, nil
}

// Save 保存草稿到本地
func (m *Manager) Save(draft *Draft) error {
	// 如果没有ID，生成新ID
	if draft.ID == "" {
		draft.ID = uuid.New().String()
		draft.CreatedAt = time.Now()
	}
	draft.UpdatedAt = time.Now()

	// 序列化为JSON
	data, err := json.MarshalIndent(draft, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化草稿失败: %w", err)
	}

	// 保存到文件
	filename := filepath.Join(m.draftsDir, draft.ID+".json")
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("保存草稿文件失败: %w", err)
	}

	logrus.Infof("草稿已保存到本地: %s", filename)
	return nil
}

// List 列出所有草稿
func (m *Manager) List() ([]*Draft, error) {
	files, err := os.ReadDir(m.draftsDir)
	if err != nil {
		return nil, fmt.Errorf("读取草稿目录失败: %w", err)
	}

	var drafts []*Draft
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		draftID := file.Name()[:len(file.Name())-5] // 移除.json后缀
		draft, err := m.Get(draftID)
		if err != nil {
			logrus.Warnf("读取草稿失败: %s, %v", file.Name(), err)
			continue
		}

		drafts = append(drafts, draft)
	}

	// 按更新时间倒序排列
	sort.Slice(drafts, func(i, j int) bool {
		return drafts[i].UpdatedAt.After(drafts[j].UpdatedAt)
	})

	return drafts, nil
}

// Get 获取指定ID的草稿
func (m *Manager) Get(id string) (*Draft, error) {
	filename := filepath.Join(m.draftsDir, id+".json")

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("草稿不存在: %s", id)
		}
		return nil, fmt.Errorf("读取草稿文件失败: %w", err)
	}

	var draft Draft
	if err := json.Unmarshal(data, &draft); err != nil {
		return nil, fmt.Errorf("解析草稿数据失败: %w", err)
	}

	return &draft, nil
}

// Delete 删除指定ID的草稿
func (m *Manager) Delete(id string) error {
	filename := filepath.Join(m.draftsDir, id+".json")

	if err := os.Remove(filename); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("草稿不存在: %s", id)
		}
		return fmt.Errorf("删除草稿文件失败: %w", err)
	}

	logrus.Infof("草稿已删除: %s", id)
	return nil
}

// GetDraftsDir 获取草稿目录路径
func (m *Manager) GetDraftsDir() string {
	return m.draftsDir
}

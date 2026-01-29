package backup

import (
	"diarygo/internal/config"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Manager 管理数据库备份
type Manager struct {
	DBPath   string        // 原始数据库文件路径
	Dir      string        // 备份目录
	Interval time.Duration // 备份间隔，0表示不备份
	Keep     int           // 保留最近N个备份
	stop     chan struct{}
}

// NewManager 创建备份管理器
func NewManager(dbPath, dir string, interval time.Duration, keep int) *Manager {
	return &Manager{
		DBPath:   dbPath,
		Dir:      dir,
		Interval: interval,
		Keep:     keep,
		stop:     make(chan struct{}),
	}
}

// Run 启动定时备份
func (m *Manager) Run() {
	if m.Interval <= 0 {
		return
	}

	ticker := time.NewTicker(m.Interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				_ = m.CheckAndBackup()
			case <-m.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop 停止定时备份
func (m *Manager) Stop() {
	close(m.stop)
}

// CheckAndBackup 检查上次备份时间并创建备份
func (m *Manager) CheckAndBackup() error {
	if m.Interval <= 0 {
		return nil
	}

	last, err := m.lastBackupTime()
	if err != nil {
		// 没有备份文件就直接创建
		return m.createAndCleanup()
	}

	// 距离上次备份太近，不创建
	if time.Since(last) < m.Interval {
		return m.cleanup()
	}

	return m.createAndCleanup()
}

// lastBackupTime 获取最近一次备份时间
func (m *Manager) lastBackupTime() (time.Time, error) {
	files, err := os.ReadDir(m.Dir)
	if err != nil {
		return time.Time{}, err
	}

	var latest time.Time
	found := false
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name() // 假设格式 diary_20260129_150405.db
		if !strings.HasPrefix(name, "diary_") || !strings.HasSuffix(name, ".db") {
			continue
		}
		tsStr := strings.TrimSuffix(strings.TrimPrefix(name, "diary_"), ".db")
		t, err := time.Parse("20060102_150405", tsStr)
		if err != nil {
			continue
		}
		if !found || t.After(latest) {
			latest = t
			found = true
		}
	}
	if !found {
		return time.Time{}, errors.New("no backup found")
	}
	return latest, nil
}

// createAndCleanup 创建备份并清理旧文件
func (m *Manager) createAndCleanup() error {
	now := time.Now()
	backupFile := fmt.Sprintf("%s/diary_%s.db", m.Dir, now.Format("20060102_150405"))

	// 保证目录存在
	if err := os.MkdirAll(m.Dir, 0755); err != nil {
		return err
	}

	src, err := os.Open(m.DBPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(backupFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return m.cleanup()
}

// cleanup 删除多余的旧备份
func (m *Manager) cleanup() error {
	files, err := os.ReadDir(m.Dir)
	if err != nil {
		return err
	}

	var backups []os.DirEntry
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if strings.HasPrefix(name, "diary_") && strings.HasSuffix(name, ".db") {
			backups = append(backups, f)
		}
	}

	if len(backups) <= m.Keep {
		return nil
	}

	// 按文件名升序排序（旧的在前）
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Name() < backups[j].Name()
	})

	for i := 0; i < len(backups)-m.Keep; i++ {
		_ = os.Remove(filepath.Join(m.Dir, backups[i].Name()))
	}

	return nil
}

func StartBackup(cfg *config.Repository) *Manager {
	dbPath := cfg.Get("global", "db_name")
	if dbPath == "" {
		dbPath = "data/diary.db"
	}

	interval, _ := time.ParseDuration(
		cfg.Get("backup", "interval"),
	)
	if interval <= 0 {
		return nil
	}
	keep := cfg.GetInt("backup", "keep", 7)

	backupDir := cfg.Get("backup", "dir")
	if backupDir == "" {
		return nil
	}

	m := NewManager(
		dbPath,
		backupDir,
		interval,
		keep,
	)

	m.CheckAndBackup() // 启动立即检查
	m.Run()
	return m
}

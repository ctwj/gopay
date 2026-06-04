package config

import (
	"os"
	"path/filepath"
	"testing"
)

// 测试 LoadConfigFile 正常加载
func TestLoadConfigFile(t *testing.T) {
	// 创建临时配置文件
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "test.env")
	content := `# 测试配置文件
DB=host=127.0.0.1 port=5432 user=gopay password=secret dbname=gopay sslmode=disable
HOST=0.0.0.0
PORT=9090

# 带引号的值
EMPTY_LINE_ABOVE=yes
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("创建测试配置文件失败: %v", err)
	}

	// 重置全局状态
	ConfigFileValues = nil

	err := LoadConfigFile(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfigFile 失败: %v", err)
	}

	// 验证加载的值
	tests := []struct {
		key      string
		expected string
	}{
		{"DB", "host=127.0.0.1 port=5432 user=gopay password=secret dbname=gopay sslmode=disable"},
		{"HOST", "0.0.0.0"},
		{"PORT", "9090"},
		{"EMPTY_LINE_ABOVE", "yes"},
	}

	for _, tc := range tests {
		got, ok := ConfigFileValues[tc.key]
		if !ok {
			t.Errorf("配置键 %q 未找到", tc.key)
			continue
		}
		if got != tc.expected {
			t.Errorf("配置键 %q = %q, 期望 %q", tc.key, got, tc.expected)
		}
	}

	if len(ConfigFileValues) != 4 {
		t.Errorf("加载了 %d 个配置项, 期望 4 个", len(ConfigFileValues))
	}
}

// 测试 LoadConfigFile 空路径
func TestLoadConfigFileEmpty(t *testing.T) {
	ConfigFileValues = nil
	err := LoadConfigFile("")
	if err != nil {
		t.Errorf("空路径应该返回 nil, 但返回: %v", err)
	}
	if ConfigFileValues != nil {
		t.Error("空路径不应加载任何配置")
	}
}

// 测试 LoadConfigFile 不存在的文件
func TestLoadConfigFileNotFound(t *testing.T) {
	ConfigFileValues = nil
	err := LoadConfigFile("/nonexistent/path/config.env")
	if err == nil {
		t.Error("不存在的文件应该返回错误")
	}
}

// 测试配置文件中的引号去除
func TestLoadConfigFileQuotes(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "quoted.env")
	content := `DB="host=localhost dbname=test"
HOST='192.168.1.1'
PORT=8080
`
	os.WriteFile(cfgPath, []byte(content), 0644)
	ConfigFileValues = nil

	LoadConfigFile(cfgPath)

	if v := ConfigFileValues["DB"]; v != "host=localhost dbname=test" {
		t.Errorf("双引号未去除: got %q", v)
	}
	if v := ConfigFileValues["HOST"]; v != "192.168.1.1" {
		t.Errorf("单引号未去除: got %q", v)
	}
}

// 测试 GetConfigValue 优先级
func TestGetConfigValuePriority(t *testing.T) {
	ConfigFileValues = map[string]string{
		"DB":   "from_config_file",
		"HOST": "config_host",
		"PORT": "9090",
	}

	tests := []struct {
		name     string
		flagVal  string
		key      string
		defVal   string
		expected string
	}{
		{"命令行优先", "from_flag", "DB", "default", "from_flag"},
		{"配置文件其次", "", "DB", "default", "from_config_file"},
		{"默认值兜底", "", "NONEXIST", "default_val", "default_val"},
		{"空格flag不算", "  ", "HOST", "0.0.0.0", "config_host"},
		{"配置文件PORT", "", "PORT", "8080", "9090"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GetConfigValue(tc.flagVal, tc.key, tc.defVal)
			if got != tc.expected {
				t.Errorf("GetConfigValue(%q, %q, %q) = %q, 期望 %q",
					tc.flagVal, tc.key, tc.defVal, got, tc.expected)
			}
		})
	}
}

// 测试 FindConfigFile 指定路径
func TestFindConfigFileExplicit(t *testing.T) {
	// 绝对路径应转为绝对路径（Windows 上可能改变盘符）
	abs := FindConfigFile("/etc/gopay/myconfig.env")
	if !filepath.IsAbs(abs) {
		t.Errorf("指定路径应返回绝对路径, got %q", abs)
	}

	// 相对路径应转为绝对路径
	rel := FindConfigFile("myconfig.env")
	if !filepath.IsAbs(rel) {
		t.Errorf("相对路径应转为绝对路径, got %q", rel)
	}
}

// 测试 FindConfigFile 自动查找
func TestFindConfigFileAutoDetect(t *testing.T) {
	// 在当前目录创建 gopay.env
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "gopay.env")
	os.WriteFile(cfgPath, []byte("DB=test\n"), 0644)

	// 保存当前目录并切换
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	found := FindConfigFile("")
	if found == "" {
		t.Error("应该自动找到 gopay.env")
	}
	if found != cfgPath {
		t.Errorf("找到的路径 %q 不匹配 %q", found, cfgPath)
	}
}

// 测试 FindConfigFile 找不到
func TestFindConfigFileNotFound(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	found := FindConfigFile("")
	if found != "" {
		t.Errorf("空目录不应找到配置文件, got %q", found)
	}
}

// 测试 fileExists
func TestFileExists(t *testing.T) {
	dir := t.TempDir()

	// 存在的文件
	existFile := filepath.Join(dir, "exists.txt")
	os.WriteFile(existFile, []byte("test"), 0644)
	if !fileExists(existFile) {
		t.Error("存在的文件应返回 true")
	}

	// 不存在的文件
	if fileExists(filepath.Join(dir, "nope.txt")) {
		t.Error("不存在的文件应返回 false")
	}

	// 目录不算文件
	if fileExists(dir) {
		t.Error("目录不应算作文件")
	}
}

// 测试注释和空行被正确跳过
func TestLoadConfigFileComments(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "comments.env")
	content := `# 这是注释
   # 带缩进的注释

KEY1=value1
# 中间的注释
KEY2=value2
`
	os.WriteFile(cfgPath, []byte(content), 0644)
	ConfigFileValues = nil

	LoadConfigFile(cfgPath)

	if len(ConfigFileValues) != 2 {
		t.Errorf("应加载 2 个配置项, got %d: %v", len(ConfigFileValues), ConfigFileValues)
	}
	if ConfigFileValues["KEY1"] != "value1" {
		t.Errorf("KEY1 = %q, 期望 value1", ConfigFileValues["KEY1"])
	}
}

// 测试等号在值中的情况
func TestLoadConfigFileEqualsInValue(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "equals.env")
	content := `DB=host=localhost password=abc=123 dbname=test
`
	os.WriteFile(cfgPath, []byte(content), 0644)
	ConfigFileValues = nil

	LoadConfigFile(cfgPath)

	if v := ConfigFileValues["DB"]; v != "host=localhost password=abc=123 dbname=test" {
		t.Errorf("值中的等号应保留, got %q", v)
	}
}

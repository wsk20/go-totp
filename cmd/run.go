// Package cmd
// Author: wsk20
// Created on: 2025-10-16 11:05:03
package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/wsk20/go-totp/pkg/totp"
)

// ANSI 颜色码
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
	Bold   = "\033[1m"
)

// 数据文件
var accountFile = os.ExpandEnv("$HOME/.totp_accounts.json")

type OTPConfig struct {
	Label     string         `json:"label"`
	Secret    string         `json:"secret"`
	Algorithm totp.Algorithm `json:"algorithm"`
	Period    int64          `json:"period"`
	Digits    int            `json:"digits"`
	Issuer    string         `json:"issuer"`
}

// 工具函数
func clearScreen() { fmt.Print("\033[H\033[2J") }
func beep()        { fmt.Print("\a") }

func progressBar(total, left float64) string {
	const barWidth = 20
	ratio := 1 - (left / total)
	filled := int(ratio * barWidth)
	if filled > barWidth {
		filled = barWidth
	}
	color := Green
	if left <= total*0.25 {
		color = Red
	} else if left <= total*0.5 {
		color = Yellow
	}
	return fmt.Sprintf("%s%s%s%s", color, strings.Repeat("█", filled), strings.Repeat("░", barWidth-filled), Reset)
}

// 解析 otpauth:// URI
func parseOtpauthURL(uri string) (*OTPConfig, error) {
	if !strings.HasPrefix(uri, "otpauth://") {
		return nil, fmt.Errorf("不是有效 otpauth:// URI")
	}
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if u.Host != "totp" {
		return nil, fmt.Errorf("不支持的类型: %s (仅支持 totp)", u.Host)
	}
	label := strings.TrimPrefix(u.Path, "/")
	q := u.Query()
	secret := q.Get("secret")
	if secret == "" {
		return nil, fmt.Errorf("URI 中缺少 secret")
	}
	algo := strings.ToUpper(q.Get("algorithm"))
	if algo == "" {
		algo = "SHA1"
	}
	period := int64(30)
	if p := q.Get("period"); p != "" {
		fmt.Sscanf(p, "%d", &period)
	}
	digits := 6
	if d := q.Get("digits"); d != "" {
		fmt.Sscanf(d, "%d", &digits)
	}
	issuer := q.Get("issuer")
	return &OTPConfig{
		Label:     label,
		Secret:    secret,
		Algorithm: totp.Algorithm(algo),
		Period:    period,
		Digits:    digits,
		Issuer:    issuer,
	}, nil
}

// 去重函数
func uniqueAccounts(accounts []OTPConfig) []OTPConfig {
	seen := make(map[string]bool)
	var result []OTPConfig
	for _, a := range accounts {
		if !seen[a.Label] {
			seen[a.Label] = true
			result = append(result, a)
		}
	}
	return result
}

// 本地账户操作
func loadAccounts() ([]OTPConfig, error) {
	if _, err := os.Stat(accountFile); os.IsNotExist(err) {
		// 文件不存在，创建空文件
		emptyData := []byte("[]")
		if err := os.WriteFile(accountFile, emptyData, 0644); err != nil {
			return nil, fmt.Errorf("创建账户文件失败: %v", err)
		}
		return []OTPConfig{}, nil
	}

	data, err := os.ReadFile(accountFile)
	if err != nil {
		return nil, err
	}

	var accounts []OTPConfig
	if err := json.Unmarshal(data, &accounts); err != nil {
		return nil, err
	}
	return uniqueAccounts(accounts), nil
}

func saveAccounts(accounts []OTPConfig) error {
	accounts = uniqueAccounts(accounts)
	data, _ := json.MarshalIndent(accounts, "", "  ")
	return os.WriteFile(accountFile, data, 0644)
}

func removeAccount(accounts []OTPConfig, label string) ([]OTPConfig, bool) {
	for i, a := range accounts {
		if a.Label == label {
			return append(accounts[:i], accounts[i+1:]...), true
		}
	}
	return accounts, false
}

// 显示 TOTP（无闪烁版本）
func displayAccounts(accounts []OTPConfig, firstDraw bool) {
	if firstDraw {
		// 第一次完整绘制所有静态信息
		fmt.Print("\033[H\033[2J")
		fmt.Println(Bold + Cyan + "🔐 多账户动态 TOTP 管理器" + Reset)
		fmt.Println(strings.Repeat("=", 40))
		for _, cfg := range accounts {
			if cfg.Issuer != "" {
				fmt.Printf("服务提供者: %s\n", cfg.Issuer)
			}
			fmt.Printf("账户: %s\n", cfg.Label)
			fmt.Printf("算法: %s | 步长: %ds\n", cfg.Algorithm, cfg.Period)
			fmt.Printf("验证码: \n")
			fmt.Printf("剩余时间: \n")
			fmt.Println(strings.Repeat("-", 40))
		}
		fmt.Println("按 Ctrl+C 退出")
		return
	}

	// 移动光标到标题下方（回顶部，不清屏）
	// 跳过标题两行 + 分隔线一行
	fmt.Printf("\033[%d;0H", 3) // 移动到第3行

	now := time.Now()

	for i, cfg := range accounts {
		code, start, end, err := totp.GenerateCurrentTOTP(cfg.Secret, cfg.Algorithm)
		if err != nil {
			fmt.Printf("%s❌ 生成失败: %v%s\n", Red, err, Reset)
			continue
		}

		total := end.Sub(start).Seconds()
		left := int(end.Sub(now).Seconds())
		if left < 0 {
			left = 0
		}
		if left <= 5 {
			beep()
		}

		// 计算当前账户在屏幕上的起始行
		// 每个账户块为 6 行（含分隔线）
		startLine := 3 + i*6
		// 移动到对应账户的“验证码”那一行
		fmt.Printf("\033[%d;0H", startLine+3)
		fmt.Printf("验证码: %s%s%s   \n", Green, code, Reset)

		// 下一行更新剩余时间
		fmt.Printf("剩余时间: %2d 秒 [%s]   \n", left, progressBar(total, float64(left)))
	}
}

// Run 主程序
func Run() {
	addURI := flag.String("add", "", "添加账户 otpauth:// URI")
	removeLabel := flag.String("remove", "", "删除账户，通过 label")
	list := flag.Bool("list", false, "列出所有账户")
	verifyCode := flag.String("verify", "", "验证输入验证码")
	accountLabel := flag.String("account", "", "只显示或验证指定账户, 可逗号分隔")
	addUser := flag.String("add-user", "", "添加账户用户名")
	addKey := flag.String("add-key", "", "添加账户密钥")
	addIssuer := flag.String("add-issuer", "", "服务提供者 / 平台名称")
	addAlgo := flag.String("add-algo", "SHA1", "哈希算法: SHA1/SHA256/SHA512")
	addPeriod := flag.Int64("add-period", 30, "时间步长 (秒)")
	addDigits := flag.Int("add-digits", 6, "验证码位数")

	flag.Parse()

	accounts, err := loadAccounts()
	if err != nil {
		log.Fatalf("读取账户失败: %v", err)
	}

	// 添加账户
	if *addURI != "" {
		cfg, err := parseOtpauthURL(*addURI)
		if err != nil {
			log.Fatalf("解析 URI 失败: %v", err)
		}

		// 检查重复
		exists := false
		for i, a := range accounts {
			if a.Label == cfg.Label {
				exists = true
				accounts[i] = *cfg
				break
			}
		}
		if !exists {
			accounts = append(accounts, *cfg)
			fmt.Printf("✅ 添加成功: %s\n", cfg.Label)
		} else {
			fmt.Printf("⚠️ 已存在相同账户，已更新: %s\n", cfg.Label)
		}

		if err := saveAccounts(accounts); err != nil {
			log.Fatalf("保存账户失败: %v", err)
		}
		return
	}

	// 删除账户
	if *removeLabel != "" {
		newAccs, ok := removeAccount(accounts, *removeLabel)
		if !ok {
			log.Fatalf("账户不存在: %s", *removeLabel)
		}
		accounts = newAccs
		if err := saveAccounts(accounts); err != nil {
			log.Fatalf("保存账户失败: %v", err)
		}
		fmt.Printf("✅ 删除成功: %s\n", *removeLabel)
		return
	}

	// 列出账户
	if *list {
		fmt.Println("已保存账户列表:")
		for _, a := range accounts {
			fmt.Printf("- %s (%s) [%s]\n", a.Label, a.Issuer, a.Algorithm)
		}
		return
	}

	// 过滤指定账户 (支持逗号)
	var selectedAccounts []OTPConfig
	if *accountLabel != "" {
		labels := strings.Split(*accountLabel, ",")
		labelMap := make(map[string]bool)
		for _, l := range labels {
			labelMap[strings.TrimSpace(l)] = true
		}
		for _, a := range accounts {
			if labelMap[a.Label] {
				selectedAccounts = append(selectedAccounts, a)
				delete(labelMap, a.Label)
			}
		}
		if len(labelMap) > 0 {
			missing := []string{}
			for l := range labelMap {
				missing = append(missing, l)
			}
			log.Fatalf("❌ 未找到账户: %s", strings.Join(missing, ", "))
		}
	} else {
		selectedAccounts = accounts
	}

	// 通过用户名 + 密钥直接添加
	if *addUser != "" && *addKey != "" {
		cfg := &OTPConfig{
			Label:     *addUser,
			Secret:    *addKey,
			Issuer:    *addIssuer,
			Algorithm: totp.Algorithm(strings.ToUpper(*addAlgo)),
			Period:    *addPeriod,
			Digits:    *addDigits,
		}

		// 检查重复
		exists := false
		for i, a := range accounts {
			if a.Label == cfg.Label {
				exists = true
				accounts[i] = *cfg
				break
			}
		}
		if !exists {
			accounts = append(accounts, *cfg)
			fmt.Printf("✅ 添加成功: %s\n", cfg.Label)
		} else {
			fmt.Printf("⚠️ 已存在相同账户，已更新: %s\n", cfg.Label)
		}

		if err := saveAccounts(accounts); err != nil {
			log.Fatalf("保存账户失败: %v", err)
		}
		return
	}

	// 验证验证码
	if *verifyCode != "" {
		if len(selectedAccounts) == 0 {
			log.Fatal("❌ 没有指定账户可验证")
		}
		valid := totp.ValidateTOTP(selectedAccounts[0].Secret, *verifyCode, selectedAccounts[0].Period, 1, selectedAccounts[0].Algorithm)
		if valid {
			fmt.Printf("%s✅ 验证成功 (%s)%s\n", Green, selectedAccounts[0].Label, Reset)
		} else {
			fmt.Printf("%s❌ 验证失败 (%s)%s\n", Red, selectedAccounts[0].Label, Reset)
		}
		return
	}

	// 动态显示
	if len(selectedAccounts) == 0 {
		fmt.Println("❌ 当前没有任何账户，请使用 --add 添加账户")
		return
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// 隐藏光标
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h") // 程序退出时恢复光标

	displayAccounts(selectedAccounts, true) // 首次完整绘制
	for {
		select {
		case <-ticker.C:
			displayAccounts(selectedAccounts, false) // 仅局部更新
		case <-sigChan:
			fmt.Print("\033[?25h") // 显示光标
			fmt.Print("\r\033[2K") // 清空当前行
			fmt.Println("👋 已退出。")
			return
		}
	}
}

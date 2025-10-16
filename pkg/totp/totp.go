// Package totp
// Author: wsk20
// Created on: 2025-10-16 10:58:00
package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"hash"
	"strings"
	"sync"
	"time"
)

//
// 支持：
// - Base32 自动补齐、容错大小写
// - 支持 SHA1 / SHA256 / SHA512 算法
// - 可配置时间漂移容忍度
// - 返回有效期范围 (用于 CLI 展示)
// - 缓存 Base32 解码结果，提高性能
//

// Algorithm 表示哈希算法类型
type Algorithm string

const (
	SHA1   Algorithm = "SHA1"   // 默认 SHA1
	SHA256 Algorithm = "SHA256" // SHA256
	SHA512 Algorithm = "SHA512" // SHA512
)

// DefaultStep 默认时间步长（秒），TOTP 通常为 30 秒
const DefaultStep int64 = 30

// 缓存解码的 Base32 密钥（提高频繁调用性能）
var (
	cacheMu    sync.RWMutex
	cachedKey  []byte
	cachedText string
)

// decodeBase32Secret 安全解码 Base32 密钥
// 功能：
// - 自动将小写转大写
// - 去掉空格
// - 自动补齐 Base32 = 号
// - 支持缓存，提高性能
func decodeBase32Secret(secret string) ([]byte, error) {
	// 转大写并去掉空格
	secret = strings.ToUpper(strings.ReplaceAll(secret, " ", ""))

	// 补齐 Base32 长度为 8 的倍数
	if mod := len(secret) % 8; mod != 0 {
		secret += strings.Repeat("=", 8-mod)
	}

	// 读取缓存
	cacheMu.RLock()
	if secret == cachedText && cachedKey != nil {
		key := make([]byte, len(cachedKey))
		copy(key, cachedKey)
		cacheMu.RUnlock()
		return key, nil
	}
	cacheMu.RUnlock()

	// Base32 解码
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		// 尝试不带 Padding 的解码
		key, err = base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
		if err != nil {
			return nil, fmt.Errorf("[TOTP] Base32解码失败: %w", err)
		}
	}

	// 写入缓存
	cacheMu.Lock()
	cachedText = secret
	cachedKey = key
	cacheMu.Unlock()

	return key, nil
}

// getHMACFunc 返回对应算法的哈希函数，用于生成 HMAC
func getHMACFunc(algo Algorithm) func() hash.Hash {
	switch algo {
	case SHA256:
		return sha256.New
	case SHA512:
		return sha512.New
	default: // 默认使用 SHA1
		return sha1.New
	}
}

// GenerateTOTP 生成当前时间的一次性密码（TOTP）
// 参数说明：
// - secret: Base32 编码的密钥
// - timestep: 时间步长（秒）
// - algo: 哈希算法（SHA1/SHA256/SHA512）
// 返回 6 位字符串验证码
func GenerateTOTP(secret string, timestep int64, algo Algorithm) (string, error) {
	return GenerateTOTPWithTime(secret, timestep, time.Now(), algo)
}

// GenerateTOTPWithTime 生成指定时间点的 TOTP
// 支持 SHA1/SHA256/SHA512
func GenerateTOTPWithTime(secret string, timestep int64, t time.Time, algo Algorithm) (string, error) {
	// 解码 Base32 密钥
	key, err := decodeBase32Secret(secret)
	if err != nil {
		return "", err
	}

	// 计算时间计数器（Unix 时间 / timestep）
	counter := t.Unix() / timestep
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(counter)) //  // 转成 8 字节

	// 生成 HMAC
	h := hmac.New(getHMACFunc(algo), key)
	h.Write(buf[:])
	sum := h.Sum(nil)

	// 动态截取（Dynamic Truncation）
	offset := sum[len(sum)-1] & 0x0F
	binCode := (uint32(sum[offset])&0x7F)<<24 |
		(uint32(sum[offset+1])&0xFF)<<16 |
		(uint32(sum[offset+2])&0xFF)<<8 |
		(uint32(sum[offset+3]) & 0xFF)

	// 对 10^6 取余，得到 6 位验证码
	code := binCode % 1000000
	return fmt.Sprintf("%06d", code), nil
}

// ValidateTOTP 验证用户输入的验证码是否正确
// 参数说明：
// - secret: Base32 密钥
// - code: 用户输入的验证码
// - timestep: 时间步长
// - window: 前后允许的时间步数（容忍时间漂移）
// - algo: 哈希算法
func ValidateTOTP(secret, code string, timestep int64, window int, algo Algorithm) bool {
	for i := -window; i <= window; i++ {
		validCode, err := GenerateTOTPWithTime(secret, timestep, time.Now().Add(time.Duration(i)*time.Duration(timestep)*time.Second), algo)
		if err == nil && validCode == code {
			return true
		}
	}
	return false
}

// GenerateCurrentTOTP 生成当前时刻的验证码，并返回有效时间范围
// 返回值：
// - code: 当前验证码
// - start: 当前验证码有效开始时间
// - end: 当前验证码有效结束时间
func GenerateCurrentTOTP(secret string, algo Algorithm) (code string, start, end time.Time, err error) {
	code, err = GenerateTOTP(secret, DefaultStep, algo)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	now := time.Now()
	start = time.Unix((now.Unix()/DefaultStep)*DefaultStep, 0)
	end = start.Add(time.Duration(DefaultStep) * time.Second)
	return code, start, end, nil
}

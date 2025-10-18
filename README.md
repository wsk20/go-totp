# 多账户 TOTP 管理器

[English Version](README_en.md)

一个基于 Go 的 **多账户 TOTP（时间一次性密码）管理器**，支持：

- 从 `otpauth://` URI 添加账户  
- 手动添加账户（用户名 + 密钥）  
- 删除、列出账户  
- 验证输入的验证码  
- 动态显示多个账户的 TOTP 值及倒计时  
- 支持多种算法（SHA1/SHA256/SHA512）和可配置步长及位数  
- 跨平台，本地保存账户信息到 `~/.totp_accounts.json`  

---

## 安装

```bash
go install github.com/wsk20/go-totp@latest
````

安装完成后，可在命令行直接使用：

```bash
go-totp
```

---

## 使用示例

### 1. 添加账户（URI 方式）

```bash
go-totp --add "otpauth://totp/label?secret=ABC123&issuer=Example&algorithm=SHA1&period=30&digits=6"
```

输出示例：

```
✅ 添加成功: label
```

### 2. 添加账户（手动方式）

```bash
go-totp --add-user alice --add-key ABC123 --add-issuer Example --add-algo SHA1 --add-period 30 --add-digits 6
```

### 3. 删除账户

```bash
go-totp --remove alice
```

### 4. 列出所有账户

```bash
go-totp --list
```

输出示例：

```
- alice (Example) [SHA1]
- bob (Google) [SHA1]
```

### 5. 仅显示或验证指定账户

```bash
go-totp --account alice
```

```bash
go-totp --account alice --verify 123456
```

### 6. 运行动态显示 TOTP

```bash
go-totp
```

* 支持多个账户同时显示
* 实时倒计时，快到期时会提示 `beep`
* 支持 Ctrl+C 退出

---

## 动态显示示意

程序运行后，会在终端显示类似如下：

```
🔐 多账户动态 TOTP 管理器
========================================
服务提供者: Example
账户: alice
算法: SHA1 | 步长: 30s
验证码: 123456
剩余时间: 25 秒 [████████░░░░░░░░░░]

服务提供者: Google
账户: bob
算法: SHA1 | 步长: 30s
验证码: 654321
剩余时间: 12 秒 [█████████████░░░░░]
```

* 每秒更新验证码和剩余时间
* 剩余时间 <= 5 秒时会发出提示音 `beep`
* 支持任意数量账户，自动排列

---

## 参数说明

| 参数             | 说明                                |
| -------------- | --------------------------------- |
| `--add`        | 添加账户 URI（otpauth://totp/...）      |
| `--remove`     | 删除账户，通过 label                     |
| `--list`       | 列出所有账户                            |
| `--verify`     | 验证输入验证码                           |
| `--account`    | 指定账户，可逗号分隔                        |
| `--add-user`   | 添加账户用户名（手动方式）                     |
| `--add-key`    | 添加账户密钥（手动方式）                      |
| `--add-issuer` | 服务提供者/平台名称                        |
| `--add-algo`   | 哈希算法: SHA1/SHA256/SHA512（默认 SHA1） |
| `--add-period` | 时间步长（秒，默认 30）                     |
| `--add-digits` | 验证码位数（默认 6）                       |

---

## 文件存储

账户信息存储在用户主目录下：

```
~/.totp_accounts.json
```

* 自动去重
* JSON 格式，方便手动备份或迁移

---

## ANSI 颜色显示

* ✅ 成功：绿色
* ⚠️ 警告：黄色
* ❌ 错误：红色
* 动态倒计时显示彩色进度条

---

## 注意事项

* 仅支持 TOTP（不支持 HOTP）
* Ctrl+C 退出后会恢复光标并清屏
* 支持 SHA1/SHA256/SHA512 算法

---

## License

MIT License

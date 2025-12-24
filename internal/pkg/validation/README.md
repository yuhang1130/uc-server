# Validation Package

这个包提供了验证相关的功能，包括密码强度验证和验证错误翻译。

## 功能

### 1. 密码强度验证

提供了多种密码验证功能：

- `ValidatePasswordStrength(password string)` - 使用默认策略验证密码
- `ValidatePasswordWithPolicy(password string, policy PasswordPolicy)` - 使用自定义策略验证
- `ValidatePasswordWithCommonChecks(password string)` - 综合验证（包括弱密码检查）

### 2. 验证错误翻译

将 `go-playground/validator` 的验证错误翻译为友好的中文消息。

#### 使用方法

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/yuhang1130/gin-server/internal/pkg/validation"
)

func Handler(c *gin.Context) {
    var req SomeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // 翻译验证错误
        friendlyMsg := validation.TranslateValidationError(err)
        c.JSON(400, gin.H{"error": friendlyMsg})
        return
    }
}
```

#### 支持的验证标签翻译

| 标签 | 中文提示示例 |
|------|-------------|
| `required` | 用户名不能为空 |
| `email` | 邮箱格式不正确 |
| `min` | 密码长度不能少于8个字符 |
| `max` | 用户名长度不能超过32个字符 |
| `alphanum` | 用户名只能包含字母和数字 |
| `numeric` | 字段必须是数字 |
| `url` | 字段必须是有效的URL |
| `oneof` | 字段必须是以下值之一: ... |

#### 字段名翻译

以下字段会自动翻译为中文：

- `Username` → 用户名
- `Email` → 邮箱
- `Password` → 密码
- `OldPassword` → 原密码
- `NewPassword` → 新密码
- `Token` → 令牌
- `RefreshToken` → 刷新令牌

如需添加更多字段翻译，请在 `validator_error.go` 的 `translateFieldName()` 函数中添加。

#### 错误信息对比

**修改前：**
```
Key: 'RegisterRequest.Username' Error:Field validation for 'Username' failed on the 'required' tag
```

**修改后：**
```
用户名不能为空
```

## 示例

### 完整示例

```go
package handler

import (
    "github.com/gin-gonic/gin"
    "github.com/yuhang1130/gin-server/internal/dto"
    "github.com/yuhang1130/gin-server/internal/pkg/validation"
    "github.com/yuhang1130/gin-server/pkg/response"
)

func Register(c *gin.Context) {
    var req dto.RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // 自动翻译验证错误
        response.BadRequest(c, validation.TranslateValidationError(err))
        return
    }

    // 业务逻辑...
}
```

### 多个错误的情况

如果有多个字段验证失败，错误信息会用分号分隔：

```
用户名不能为空; 邮箱格式不正确; 密码长度不能少于8个字符
```

### 特殊错误处理

- **EOF错误**（空请求体）: `"请求体不能为空"`
- **其他非验证错误**: 返回原始错误信息

## 测试

运行测试：

```bash
go test -v ./internal/pkg/validation/...
```

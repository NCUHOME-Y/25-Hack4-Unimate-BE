# 🚀 知序Unimate 后端服务

基于 Gin + GORM + MySQL + WebSocket 的轻量级习惯养成 & 学习打卡社区后端。
核心设计：Flag 即帖子——用户创建的公开 Flag 自动成为论坛内容，零额外成本。

---

## 📦 技术栈

- Web 框架：Gin 1.9
- ORM：GORM v2
- 数据库：MySQL 8.0+
- 实时通信：Gorilla WebSocket
- 日志：Logrus（落盘+控制台）
- 配置：godotenv（.env）
- 鉴权：JWT（HS256）

---

## 🗂️ 项目结构

.
├── main.go  // 入口：注册路由、初始化定时任务

├── internal

│   ├── app

│   │   ├── handler/           // 路由组（用户/Flag/帖子/聊天/封神榜/学习/成就）

│   │   ├── service/           // 业务逻辑（含 WebSocket Hub、AI 计划生成）

│   │   ├── repository/        // DAO 层（GORM 封装）

│   │   └── model/             // 实体 & 表定义

├── util/

│   ├── logger.go              // Logrus 封装

│   └── jwt.go                 // JWT 生成/解析/刷新

├── .env                       // 数据库、JWT_SECRET、APIKEY

├── scripts/

│   └── unimate.sql            // 初始化 SQL（含索引、外键）

├── go.mod

├── Dockerfile                 // 多阶段构建

└── README.md

---

## ⚙️ 快速开始

1. 克隆 & 依赖
   git clone https://github.com/NCUHOME-Y/25-Hack4-Unimate-BE.git
   cd 25-Hack4-Unimate-BE
   go mod tidy
2. 数据库
   mysql -u root -p < scripts/unimate.sql
   库名 unimate，字符集 utf8mb4
3. **数据迁移（仅旧版本升级需要）**
   ⚠️ **新安装请跳过此步**！仅当满足以下条件时才执行：

- 你是从旧版本升级（非全新安装）
- 数据库表 `users` 中仍存在旧列：`is_remind`、`remind_hour`、`remind_min`

执行迁移命令：

```bash
mysql -u root -p unimate < backend/cmd/tools/migrate_reminder_fields/migrate.sql
```

该迁移会将旧用户的 `is_remind/remind_hour/remind_min` 字段数据复制到新字段 `is_study_remind/study_remind_hour/study_remind_min`。

验证迁移结果（可选）：

```bash
mysql -u root -p unimate < backend/cmd/tools/migrate_reminder_fields/verify.sql
```

**如果已手动删除旧列，或从未有过旧列，则无需也无法执行该迁移**。

4. 环境变量
   cp .env.example .env
   必须项
   DB_DSN="user:pass@tcp(127.0.0.1:3306)/unimate?charset=utf8mb4&parseTime=True&loc=Local"
   JWT_SECRET="32位随机字符串"
   APIKEY="阿里云百炼 API Key"   # AI 计划生成用（模型：qwen3.5-flash）
   DASHSCOPE_REGION="cn"        # 可选：cn / intl / auto，默认国内优先

# DASHSCOPE_BASE_URL="https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"  # 可选：显式指定端点

通用 AI 配置（仅改环境变量即可切平台/模型，无需重建镜像）
AI_PROVIDER="dashscope"      # dashscope | openrouter | zhipu | openai
AI_API_KEY=""                # 可选，优先级高于 APIKEY
AI_MODELS="qwen3.5-flash,qwen-plus"   # 逗号分隔，按顺序回退
AI_BASE_URL=""               # 可选，显式指定主端点
AI_ENDPOINTS=""              # 可选，多个端点逗号分隔（优先级最高）
AI_TIMEOUT_SECONDS="10"      # 单次请求超时秒数
AI_RETRY_PER_ENDPOINT="2"    # 每个端点重试次数
AI_ALLOW_CROSS_REGION_FALLBACK="true" # 是否允许跨区域端点回退

平台示例

- 百炼（默认）:
  AI_PROVIDER=dashscope
  AI_MODELS=qwen3.5-flash,qwen-plus
  DASHSCOPE_REGION=cn
- OpenRouter:
  AI_PROVIDER=openrouter
  AI_API_KEY=<openrouter_key>
  AI_MODELS=qwen/qwen3-next-80b-a3b-instruct:free,openrouter/auto
- 智谱:
  AI_PROVIDER=zhipu
  AI_API_KEY=<zhipu_key>
  AI_MODELS=glm-4.7-flash,glm-4-flash-250414

5. 运行
   go run main.go
   → 监听 0.0.0.0:8080

---

## 🔑 统一规范

- 鉴权：Authorization: Bearer `<JWT>`（登录/注册除外）
- 成功格式：{"success":true, "data": ...}
- 错误格式：{"success":false, "message":"..."}
- 时间：UTC，格式 2006-01-02T15:04:05Z
- 分页：page=1&limit=20，默认 page=1, limit=20

---

## 🌟 核心业务规则

1. Flag 即帖子is_hiden=false 的 Flag 自动出现在论坛；点赞/评论直接写 flags 表。
2. 每天只能打卡一次数据库层 UNIQUE(user_id, date) 兜底；支持主动打卡 & 学习≥30 min 被动打卡。
3. 学习时长前端计时，后端只接收分钟单位；每日首次≥30 min 自动触发被动打卡。
4. 封神榜按 user.count（积分）实时降序。
5. 成就
   注册即初始化 5 个默认成就；后端定时检测并自动解锁。

---

## 📖 接口速览（已上线 30+）

| 模块      | 方法   | 路径                     | 功能                                                    |
| --------- | ------ | ------------------------ | ------------------------------------------------------- |
| 认证      | POST   | /api/register            | 注册                                                    |
|           | POST   | /api/login               | 登录                                                    |
|           | GET    | /api/getUser             | 当前用户信息                                            |
|           | PUT    | /updatePassword          | 修改密码                                                |
|           | PUT    | /updateUsername          | 重命名                                                  |
| Flag      | POST   | /api/addFlag             | 创建任务                                                |
|           | GET    | /api/getUserFlags        | 我的全部 Flag                                           |
|           | PUT    | /api/doneFlag            | 记一次进度                                              |
|           | PUT    | /api/finshDoneFlag       | 直接标记完成                                            |
|           | DELETE | /api/deleteFlag          | 删除                                                    |
|           | PUT    | /api/updateFlagHide      | 同步/取消同步到论坛                                     |
| 论坛      | POST   | /api/postUserPost        | 发普通帖子                                              |
|           | DELETE | /api/deleteUserPost      | 删帖                                                    |
|           | POST   | /api/commentOnPost       | 评论（支持 Flag/Post）                                  |
|           | DELETE | /api/deleteComment       | 删评论                                                  |
|           | GET    | /api/getAllPosts         | 全部帖子（Flag+Post）                                   |
| 打卡      | PUT    | /api/updateDaka          | 主动打卡                                                |
|           | GET    | /api/getDakaRecords      | 本月打卡记录                                            |
| 学习      | POST   | /api/addLearnTime        | 提交时长                                                |
|           | GET    | /api/getLearnTime        | 最近 30 条                                              |
| 排行      | GET    | /api/ranking             | Top20                                                   |
| 成就      | GET    | /api/getUserAchievement  | 已解锁成就                                              |
| AI        | POST   | /api/ai/generate-plan    | 生成学习计划（阿里云百炼：qwen3.5-flash，默认国内优先） |
| WebSocket | GET    | /ws/chat?token=`<JWT>` | 群聊                                                    |

完整文档 & 示例请求 → docs/api.md

---

## 🧪 WebSocket 快速测试

wscat -c "ws://localhost:8080/ws/chat?token=`<JWT>`
收到 { "type":"welcome","data":{"online_count":3} }
复制

---

## 🚢 部署

### 二进制

CGO_ENABLED=0 GOOS=linux go build -o unimate
./unimate

### Docker

docker build -t unimate .
docker run -d -p 8080:8080 --env-file .env unimate

官方镜像
ghcr.io/ncuhome-y/unimate-backend:latest

---

## 鸣谢

1.感谢Hackweek第四组的所有成员

2.感谢Github开源包的作者们
--------------------------

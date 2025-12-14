# 一言功能部署说明

## 功能概述
- 每天早上8点自动发布一条来自"一言"网站的励志内容
- 账户名：一言
- 头像：avatar3（对应数据库中head_show=3）
- API来源：https://v1.hitokoto.cn

## 部署步骤

### 1. 在服务器上更新Docker容器

SSH登录到服务器后，执行以下命令：

```bash
# 拉取最新镜像
sudo docker pull llffkk/unimate:latest

# 停止并删除旧容器
sudo docker stop unimate-backend
sudo docker rm unimate-backend

# 启动新容器
sudo docker run -d \
  --name unimate-backend \
  --network host \
  -e PORT=8080 \
  -e GIN_MODE=release \
  -e DB_DSN='root:UnimateMysql@2025@tcp(127.0.0.1:3306)/unimate?charset=utf8mb4&parseTime=True&loc=Local' \
  -e APIKEY='sk-mjvyhztfgnlnzxtfvdjzvmhzakzygvqczwmmxdvpvbfqmfjl' \
  llffkk/unimate:latest

# 检查容器状态
sudo docker ps | grep unimate-backend

# 查看日志确认一言功能已启动
sudo docker logs --tail 100 unimate-backend
```

### 2. 验证一言功能

查看日志应该能看到类似以下输出：
```
✅ 一言系统账户创建成功 (ID: 9999)
✅ 一言定时任务已启动，每天早上8点发布
⏰ 下次一言发布时间: 2025-12-16 08:00:00 (还有 XX.X 小时)
```

### 3. 手动测试一言发布（可选）

如果想立即测试功能，可以调用测试接口：

```bash
curl -X POST http://139.199.157.76:8080/api/triggerHitokoto
```

成功响应：
```json
{
  "success": true,
  "message": "一言发布成功"
}
```

然后在前端翰林院页面查看是否有"一言"账户发布的帖子。

## 技术实现

### 后端代码结构

1. **hitokoto_service.go** - 一言定时任务核心服务
   - `StartHitokotoScheduler()` - 启动定时任务
   - `fetchHitokoto()` - 调用一言API获取内容
   - `ensureHitokotoUser()` - 确保系统账户存在
   - `postHitokoto()` - 发布一言到翰林院

2. **hitokoto_handler.go** - 测试接口
   - `TriggerHitokoto()` - 手动触发一言发布（用于测试）

3. **main.go** - 启动入口
   - 在服务启动时调用 `service.StartHitokotoScheduler()`

### 系统账户信息

- **User ID**: 9999
- **用户名**: 一言
- **邮箱**: hitokoto@system.local
- **头像**: head_show = 3 (avatar3)
- **状态**: "一言·每日智慧"

### 定时任务逻辑

1. 服务启动后立即检查并创建一言系统账户（如果不存在）
2. 计算下次8点的时间
   - 如果当前时间在8点之前，等待到今天8点
   - 如果当前时间在8点之后，等待到明天8点
3. 到达8点后执行以下操作：
   - 调用一言API获取内容
   - 格式化内容（包含来源和作者信息）
   - 以"一言"账户身份发布帖子到翰林院
   - 等待1分钟后继续循环

### 帖子格式

一言帖子的标题固定为"每日一言"，内容格式根据API返回的信息自动调整：

- 有作者和来源：`{内容}\n\n—— {作者}《{来源}》`
- 仅有来源：`{内容}\n\n—— 《{来源}》`
- 仅有作者：`{内容}\n\n—— {作者}`
- 无来源无作者：`{内容}`

## 前端适配

前端已自动适配：
- 头像系统会自动将 head_show=3 映射到 avatar3
- 翰林院页面会正常显示一言账户发布的帖子
- 用户可以点赞、评论一言的帖子

## 故障排查

### 1. 检查容器是否正常运行
```bash
sudo docker ps | grep unimate-backend
```

### 2. 查看容器日志
```bash
sudo docker logs -f unimate-backend
```

### 3. 检查一言账户是否创建成功
```bash
# 在服务器上连接MySQL
mysql -u root -p
use unimate;
SELECT id, name, email, head_show, status FROM users WHERE id = 9999;
```

### 4. 手动测试一言API
```bash
curl https://v1.hitokoto.cn
```

### 5. 检查帖子是否发布成功
```bash
# 查询一言发布的帖子
mysql -u root -p
use unimate;
SELECT id, title, content, created_at FROM posts WHERE user_id = 9999 ORDER BY created_at DESC LIMIT 5;
```

## 注意事项

1. 服务器时间必须正确设置，否则定时任务可能不准时
2. 一言API需要外网访问，确保服务器可以访问 v1.hitokoto.cn
3. 如果数据库中已存在ID=9999的用户，系统会自动跳过创建
4. 每次服务重启都会重新计算下次发布时间，不会丢失定时任务

## 更新日志

- 2025-12-15: 初始版本，实现每日8点自动发布一言功能

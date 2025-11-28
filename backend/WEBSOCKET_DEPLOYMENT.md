# WebSocket 部署问题解决方案

## 问题描述
前端部署到生产环境后，WebSocket 连接失败，错误信息：
```
WebSocket连接失败: wss://unimate.ncuos.com/ws/chat?...
```

## 根本原因
生产环境使用 Nginx 作为反向代理，但 **Nginx 没有正确配置 WebSocket 支持**。

WebSocket 协议需要特殊的 HTTP 头来升级连接：
- `Upgrade: websocket`
- `Connection: Upgrade`

如果 Nginx 配置中缺少这些头部设置，WebSocket 握手会失败。

## 解决步骤

### 1. 在服务器上配置 Nginx

将 `nginx.conf` 文件上传到服务器，放置到适当位置：

**方法 A：使用 Nginx 站点配置**
```bash
# 复制配置文件到 Nginx 站点目录
sudo cp nginx.conf /etc/nginx/sites-available/unimate

# 创建符号链接
sudo ln -s /etc/nginx/sites-available/unimate /etc/nginx/sites-enabled/

# 删除默认配置（如果存在）
sudo rm /etc/nginx/sites-enabled/default
```

**方法 B：直接修改主配置**
```bash
# 备份原配置
sudo cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.backup

# 将配置添加到 http 块中
sudo nano /etc/nginx/nginx.conf
```

### 2. 修改配置文件中的路径

编辑 Nginx 配置，替换以下占位符：

```nginx
# SSL 证书路径（必须修改）
ssl_certificate /path/to/your/fullchain.pem;
ssl_certificate_key /path/to/your/privkey.pem;

# 前端构建文件路径（如果 Nginx 托管前端）
root /path/to/frontend/dist;
```

**获取 SSL 证书：**

如果还没有 SSL 证书，可以使用 Let's Encrypt 免费获取：
```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d unimate.ncuos.com
```

### 3. 测试并重启 Nginx

```bash
# 测试配置语法
sudo nginx -t

# 如果测试通过，重新加载 Nginx
sudo systemctl reload nginx

# 或重启 Nginx
sudo systemctl restart nginx
```

### 4. 检查防火墙设置

确保服务器防火墙允许 80 和 443 端口：

```bash
# UFW 防火墙
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw reload

# 或 firewalld
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

### 5. 验证 Go 后端正在运行

确保 Go 后端服务在 8080 端口运行：

```bash
# 检查端口占用
sudo netstat -tulnp | grep 8080

# 或使用 ss
sudo ss -tulnp | grep 8080

# 检查 Docker 容器状态（如果使用 Docker）
docker ps | grep unimate-backend
```

## WebSocket 配置关键点

Nginx 配置中 WebSocket 的核心部分：

```nginx
location /ws/ {
    proxy_pass http://127.0.0.1:8080;
    
    # ⭐ 关键：WebSocket 协议升级
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    
    # ⭐ 关键：长连接超时设置
    proxy_connect_timeout 7d;
    proxy_send_timeout 7d;
    proxy_read_timeout 7d;
    
    # ⭐ 关键：禁用缓冲
    proxy_buffering off;
}
```

## 调试方法

### 1. 查看 Nginx 错误日志
```bash
sudo tail -f /var/log/nginx/unimate_error.log
```

### 2. 查看 Go 后端日志
```bash
# 如果使用 Docker
docker logs -f unimate-backend

# 如果直接运行
tail -f /path/to/backend/logs/app.log
```

### 3. 使用浏览器开发者工具
- 打开浏览器开发者工具（F12）
- 切换到 Network 标签
- 筛选 WS（WebSocket）
- 查看连接状态和错误信息

### 4. 测试 WebSocket 连接
使用在线工具测试 WebSocket：
```
wss://unimate.ncuos.com/ws/chat?token=YOUR_TOKEN
```

推荐工具：
- https://www.websocket.org/echo.html
- https://www.piesocket.com/websocket-tester

## 常见问题

### Q1: 502 Bad Gateway 错误
**原因：** 后端服务未运行或端口不匹配

**解决：**
```bash
# 检查后端是否运行
docker ps | grep backend

# 检查端口
sudo netstat -tulnp | grep 8080
```

### Q2: SSL 证书错误
**原因：** 证书路径错误或证书过期

**解决：**
```bash
# 检查证书有效期
sudo certbot certificates

# 更新证书
sudo certbot renew
```

### Q3: WebSocket 连接后立即断开
**原因：** 超时设置太短或缺少心跳包

**解决：**
- 增加 `proxy_read_timeout` 时长
- 在前端或后端实现 WebSocket 心跳机制

### Q4: CORS 错误
**原因：** 跨域配置问题

**解决：** 在 Go 后端添加 CORS 中间件（通常已配置）

## 验证成功的标志

✅ 浏览器控制台显示：
```
✅ 私聊WebSocket连接已建立
```

✅ Network 标签显示：
```
Status: 101 Switching Protocols
```

✅ Nginx 日志无错误

✅ 消息可以正常收发

## 需要帮助？

如果按照以上步骤仍然无法解决，请提供：
1. Nginx 错误日志
2. Go 后端日志
3. 浏览器控制台完整错误信息
4. 服务器操作系统和 Nginx 版本

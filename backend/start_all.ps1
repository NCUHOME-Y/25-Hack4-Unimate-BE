# 启动 User 服务
Start-Process -FilePath "go" -ArgumentList "run app/user/api/user.go -f app/user/api/etc/user-api.yaml" -NoNewWindow
Write-Host "User Service started on port 8888"

# 启动 Goal 服务
Start-Process -FilePath "go" -ArgumentList "run app/goal/api/goal.go -f app/goal/api/etc/goal-api.yaml" -NoNewWindow
Write-Host "Goal Service started on port 8889"

# 启动 Social 服务
Start-Process -FilePath "go" -ArgumentList "run app/social/api/social.go -f app/social/api/etc/social-api.yaml" -NoNewWindow
Write-Host "Social Service started on port 8890"

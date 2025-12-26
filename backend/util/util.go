package utils

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-mail/mail/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var (
	jwtSecret        []byte
	jwtSecretOnce    sync.Once
	jwtSecretLoadErr error
)

var logger = logrus.New()

func loadJWTSecret() ([]byte, error) {
	jwtSecretOnce.Do(func() {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			jwtSecretLoadErr = errors.New("JWT_SECRET 环境变量未设置")
			return
		}
		jwtSecret = []byte(secret)
	})

	if jwtSecretLoadErr != nil {
		return nil, jwtSecretLoadErr
	}
	return jwtSecret, nil
}

func RequireJWTSecret() {
	if _, err := loadJWTSecret(); err != nil {
		log.Fatal("❌ 安全错误：JWT_SECRET 环境变量未设置！请在 .env 文件中设置")
	}
}

func init() {
	file, err := os.OpenFile("Unimate.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logger.Fatal("无法打开日志文件: ", err)
	}

	// 设置同时输出到控制台和文件
	multiWriter := io.MultiWriter(os.Stdout, file)
	logger.SetOutput(multiWriter)

	// 设置日志格式为JSON，便于后续分析
	logger.SetFormatter(&logrus.JSONFormatter{})
}

// Claims 结构体
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT Token
func GenerateToken(userID uint, username, email string) (string, error) {
	secret, err := loadJWTSecret()
	if err != nil {
		return "", err
	}

	now := time.Now()
	expireTime := now.Add(7 * 24 * time.Hour) // 🔒 Token 有效期 7 天（安全加固）

	claims := Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "Unimate_app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseToken 解析 JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	secret, err := loadJWTSecret()
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken 刷新 Token
func RefreshToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	return GenerateToken(claims.UserID, claims.Username, claims.Email)
}

func HashPassword(password string) (string, error) {
	passwordBytes := []byte(password)
	HashPassword, err := bcrypt.GenerateFromPassword(passwordBytes, 12)
	if err != nil {
		log.Printf("Hash password defeat")
		return "", err
	}
	return string(HashPassword), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func LogInfo(message string, fields logrus.Fields) {
	logger.WithFields(fields).Info(message)
}
func LogWarn(message string, fields logrus.Fields) {
	logger.WithFields(fields).Warn(message)
}
func LogError(message string, fields logrus.Fields) {
	logger.WithFields(fields).Error(message)
}
func LogDebug(message string, fields logrus.Fields) {
	logger.WithFields(fields).Debug(message)
}

// 发送邮箱?
func SentEmail(to, subject, body string) error {
	// 验证环境变量
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := os.Getenv("SMTP_FROM")

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" || smtpFrom == "" {
		return fmt.Errorf("SMTP配置不完整")
	}

	m := mail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	port, err := strconv.Atoi(smtpPort)
	if err != nil {
		return fmt.Errorf("SMTP端口配置错误")
	}

	d := mail.NewDialer(smtpHost, port, smtpUser, smtpPass)
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

// 生成验证码
func GenerateCode() string {
	var num uint32
	binary.Read(rand.Reader, binary.BigEndian, &num)
	return fmt.Sprintf("%06d", num%1_000_000)
}

// GetAvatarPath 根据用户的 HeadShow 字段获取头像路径
// headShow: 用户选择的头像编号（1-36）
// 返回: 头像的API路径，用于前端访问
func GetAvatarPath(headShow int) string {
	if headShow > 0 && headShow <= 36 {
		return fmt.Sprintf("/api/avatar/%d", headShow)
	}
	// 默认返回第一个头像
	return "/api/avatar/1"
}

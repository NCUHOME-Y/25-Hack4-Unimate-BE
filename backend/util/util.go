package utils

import (
	"errors"
	"io"
	"log"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// JWT 密钥（生产环境应从环境变量或配置文件中读取）
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

var logger = logrus.New()

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
	now := time.Now()
	expireTime := now.Add(24 * time.Hour) // Token 有效期 24 小时

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
	return token.SignedString(jwtSecret)
}

// ParseToken 解析 JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
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
func LogError(message string, fields logrus.Fields) {
	logger.WithFields(fields).Error(message)
}
func LogDebug(message string, fields logrus.Fields) {
	logger.WithFields(fields).Debug(message)
}

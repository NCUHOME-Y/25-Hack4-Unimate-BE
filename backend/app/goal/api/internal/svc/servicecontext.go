// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"log"

	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/app/goal/api/internal/config"
	"github.com/NCUHOME-Y/25-Hack4-Unimate-BE/internal/app/repository"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.DataSource), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the global DB variable in the original repository package
	repository.DB = db

	return &ServiceContext{
		Config: c,
		DB:     db,
	}
}

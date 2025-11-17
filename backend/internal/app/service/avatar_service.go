package service

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ServeAvatar 返回用户头像图片，路径按优先级查找：
// 1. ../frontend/src/assets/head/<file>
// 2. ./assets/head/<file>
// 3. 可执行文件同级的 assets/head/<file>
func ServeAvatar() gin.HandlerFunc {
	// 头像文件名列表（与前端/工具中使用的顺序一致，共32个头像）
	avatarFiles := []string{
		"screenshot_20251114_131601.png",
		"screenshot_20251114_131629.png",
		"screenshot_20251114_131937.png",
		"screenshot_20251114_131951.png",
		"screenshot_20251114_132014.png",
		"screenshot_20251114_133459.png",
		"微信图片_20251115203432_32_227.jpg",
		"微信图片_20251115203433_33_227.jpg",
		"微信图片_20251115203434_34_227.jpg",
		"微信图片_20251115203434_35_227.jpg",
		"微信图片_20251115203435_36_227.jpg",
		"微信图片_20251115203436_37_227.jpg",
		"微信图片_20251116131024_45_227.jpg",
		"微信图片_20251116131024_46_227.jpg",
		"微信图片_20251116131025_47_227.jpg",
		"微信图片_20251116131026_48_227.jpg",
		"微信图片_20251116131027_49_227.jpg",
		"微信图片_20251116131028_50_227.jpg",
		"微信图片_20251116131029_51_227.jpg",
		"微信图片_20251116131030_52_227.jpg",
		"微信图片_20251116131031_53_227.jpg",
		"微信图片_20251117235910_62_227.jpg",
		"微信图片_20251117235910_63_227.jpg",
		"微信图片_20251117235911_64_227.jpg",
		"微信图片_20251117235912_65_227.jpg",
		"微信图片_20251117235913_66_227.jpg",
		"微信图片_20251117235914_67_227.jpg",
		"微信图片_20251117235915_68_227.jpg",
		"微信图片_20251117235916_69_227.jpg",
		"微信图片_20251117235917_71_227.jpg",
		"微信图片_20251118000147_72_227.jpg",
		"微信图片_20251118000148_74_227.jpg",
	}

	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 || id > len(avatarFiles) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid avatar id"})
			return
		}

		filename := avatarFiles[id-1]

		// 候选路径列表
		candidates := []string{
			filepath.Join("..", "frontend", "src", "assets", "head", filename),
			filepath.Join(".", "assets", "head", filename),
		}

		// 可执行文件同级 assets/head
		if execPath, err := os.Executable(); err == nil {
			candidates = append(candidates, filepath.Join(filepath.Dir(execPath), "assets", "head", filename))
		}

		for _, p := range candidates {
			if _, err := os.Stat(p); err == nil {
				// 找到文件，返回
				c.File(p)
				return
			}
		}

		// 如果都没找到，返回 404
		c.JSON(http.StatusNotFound, gin.H{"error": "avatar not found"})
	}
}

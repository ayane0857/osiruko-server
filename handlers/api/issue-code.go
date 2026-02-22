package api

import (
	"crypto/rand"
	"math/big"
	"osiruko/models"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(n int) (string, error) {
	result := make([]byte, n)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		result[i] = letters[num.Int64()]
	}
	return string(result), nil
}

func IssueCodeHandler(c *cache.Cache, db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		minecraftUUID := ctx.Query("minecraft-uuid")
		if minecraftUUID == "" {
			ctx.JSON(400, gin.H{
				"error": "Missing minecraft-uuid parameter",
			})
			return
		}
		if !isValidUUID(minecraftUUID) {
			ctx.JSON(400, gin.H{
				"error": "Invalid minecraft-uuid format",
			})
			return
		}
		result := db.Where("minecraft_uuid = ?", minecraftUUID).First(&models.Users{})
		if result.Error == nil {
			ctx.JSON(400, gin.H{
				"error": "User already exists",
			})
			return
		}
		verifyCode, err := randomString(6)
		if err != nil {
			ctx.JSON(500, gin.H{
				"error": "Failed to generate code",
			})
			return
		}
		c.Set(verifyCode, minecraftUUID, cache.DefaultExpiration)
		ctx.JSON(200, gin.H{
			"verify_code": verifyCode,
		})
	}
}
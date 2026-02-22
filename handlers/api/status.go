package api

import (
	"osiruko/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func StatusHandler(db *gorm.DB) gin.HandlerFunc {
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
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				ctx.JSON(202, gin.H{
					"bool": false,
				})
			} else {
				ctx.JSON(500, gin.H{
					"error": "Database error",
				})
			}
			return
		}
		ctx.JSON(200, gin.H{
			"bool": true,
		})
	}
}
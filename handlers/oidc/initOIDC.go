package oidc

import (
	"context"
	"log"
	"net/http"
	"os"

	"osiruko/models"

	"crypto/rand"
	"encoding/base64"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

var (
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
	domain       string
)

type UserClaims struct {
	Email         string `json:"email"`
	Name          string `json:"name"`
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func InitOIDC() {
	ctx := context.Background()
	// 例: GoogleのOIDCプロバイダを使用する場合
	provider, err := oidc.NewProvider(ctx, "https://auth.uniproject.jp")
	if err != nil {
		log.Fatal(err)
	}
	godotenv.Load()
	client_id := os.Getenv("client_id")
	log.Printf("DEBUG: client_id is '%s'\n", client_id)
	client_secret := os.Getenv("client_secret")
	domain = os.Getenv("redirect_url")

	oauth2Config = oauth2.Config{
		ClientID:     client_id,
		ClientSecret: client_secret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  domain,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier = provider.Verifier(&oidc.Config{ClientID: oauth2Config.ClientID})
}

func HandleAuthLogin(c *cache.Cache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		verifyCode:= ctx.Query("verify_code")
		if verifyCode == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing verify_code parameter",
			})
			return
		}
		state := generateState()

		ctx.SetCookie("oauth_state", state, 600, "/", "", false, true)
		minecraftUUID, found := c.Get(verifyCode)
		if !found {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid verify_code",
			})
			return
		}
		c.Set(state, minecraftUUID, cache.DefaultExpiration)

		ctx.Redirect(http.StatusFound, oauth2Config.AuthCodeURL(state))
	}
}

func HandleCallback(c *cache.Cache, db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		queryState := ctx.Query("state")

		// Cookie の state
		cookieState, err := ctx.Cookie("oauth_state")
		if err != nil || queryState != cookieState {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid state",
			})
			return
		}
		bgCtx := context.Background()

		oauth2Token, err := oauth2Config.Exchange(bgCtx, ctx.Query("code"))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
			return
		}

		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "No id_token field in oauth2 token"})
			return
		}

		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify ID Token"})
			return
		}

		oidcUserID := idToken.Subject
		_ = oidcUserID

		var claims UserClaims

		// 2. IDトークンからクレーム（ユーザー情報）を抽出
		if err := idToken.Claims(&claims); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse claims"})
			log.Printf("クレームのパースに失敗しました: %v", err)
			return
		}

		minecraftUUIDRaw, found := c.Get(queryState)
		if !found {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "state expired or invalid"})
			return
		}

		minecraftUUID, ok := minecraftUUIDRaw.(string)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid cache type"})
			return
		}

		parsedUUID, err := uuid.Parse(minecraftUUID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
			return
		}

		if err := db.Create(&models.Users{
			MinecraftUUID: parsedUUID,
			Name:          claims.Name,
			Email:         claims.Email,
		}).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "認証が完了しました！Minecraftに戻ってください。"})
	}
}
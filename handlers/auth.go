package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

type SignUpRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	Recaptcha string `json:"recaptcha" binding:"required"`
	Version   string `json:"version" binding:"required"` // "v2" or "v3"
}

type RecaptchaResponse struct {
	Success     bool     `json:"success"`
	Score       float64  `json:"score"`  // v3
	Action      string   `json:"action"` // v3
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}

func SignUp(c *gin.Context) {
	var req SignUpRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isValid := verifyRecaptcha(req.Recaptcha, req.Version)
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid captcha"})
		return
	}

	// TODO: Save user to DB (omitted)
	c.JSON(http.StatusOK, gin.H{"message": "Signup successful"})
}

func verifyRecaptcha(token string, version string) bool {
	var secret string
	switch version {
	case "v2":
		// secret = os.Getenv("RECAPTCHA_V2_SECRET")
		secret = "6LfgLZsrAAAAAOO4jnZa_UAM-BOCJ2KKz0RK6nOy"
	case "v3":
		// secret = os.Getenv("RECAPTCHA_V3_SECRET")
		secret = "6LdqKJsrAAAAAJTWQioCCbkwKcpydIiKBKnAGh_P"
	default:
		return false
	}

	client := resty.New()
	resp, err := client.R().
		SetFormData(map[string]string{
			"secret":   secret,
			"response": token,
		}).
		SetResult(&RecaptchaResponse{}).
		Post("https://www.google.com/recaptcha/api/siteverify")

	if err != nil {
		return false
	}

	result := resp.Result().(*RecaptchaResponse)

	if version == "v3" {
		// Optional: You can check score threshold (0.5 - 1.0)
		return result.Success && result.Score >= 0.5
	}

	return result.Success
}

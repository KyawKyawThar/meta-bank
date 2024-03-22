package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type userRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=7"`
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name" binding:"required"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active" binding:"required"`
}

type userResponse struct {
	Username          string `json:"username"`
	Password          string `json:"password"`
	Email             string `json:"email"`
	FullName          string `json:"full_name"`
	Role              string `json:"role"`
	IsActive          bool   `json:"is_active"`
	PasswordChangedAt string `json:"password_changed_at"`
	CreatedAt         string `json:"created_at"`
}

func (s *Server) createUser(ctx *gin.Context) {

	var req userRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
	}

	//arg := db.CreateUserParams{
	//	Username: req.Username,
	//	Password: req.Password,
	//	Email:    req.Email,
	//	FullName: req.FullName,
	//	Role:     req.Role,
	//	IsActive: req.IsActive,
	//}

	fmt.Println("createUserCalled")

}

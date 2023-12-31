package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/mjaliz/gotracktime/internal/constants"
	"github.com/mjaliz/gotracktime/internal/models"
	"github.com/mjaliz/gotracktime/internal/utils"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strings"
	"time"
)

func (repo *DBRepo) SignUp(c *gin.Context) {
	var userInput models.SignUpInput
	if err := c.ShouldBindJSON(&userInput); err != nil {
		validationErrs := utils.ParseValidationError(err)
		utils.FailedResponse(c, http.StatusBadRequest, validationErrs, "")
		return
	}
	if userInput.Password != userInput.PasswordConfirm {
		utils.FailedResponse(c, http.StatusBadRequest, nil, "password and password confirm didn't match")
		return
	}
	hashedPassword, err := utils.HashPassword(userInput.Password)
	if err != nil {
		utils.FailedResponse(c, http.StatusInternalServerError, nil, "")
		return
	}
	userInput.Password = hashedPassword
	userDB, err := repo.DB.InsertUser(userInput)
	if err != nil {
		log.Println(err)
		if strings.Contains(err.Error(), "duplicate key value") {
			utils.FailedResponse(c, http.StatusBadRequest, nil, "email already exists")
		}
		return
	}
	utils.SuccessResponse(c, http.StatusCreated, userDB.FilterUserResponse(), "")
}

func (repo *DBRepo) SignIn(c *gin.Context) {
	var userInput models.SignInInput
	if err := c.ShouldBindJSON(&userInput); err != nil {
		validationErrs := utils.ParseValidationError(err)
		utils.FailedResponse(c, http.StatusBadRequest, validationErrs, "")
		return
	}
	userDB, err := repo.DB.FindUserByEmail(userInput)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.FailedResponse(c, http.StatusUnauthorized, nil, "")
			return
		}
		utils.FailedResponse(c, http.StatusInternalServerError, nil, "")
		return
	}
	if err = utils.ComparePassword(userDB.Password, userInput.Password); err != nil {
		utils.FailedResponse(c, http.StatusUnauthorized, nil, "")
		return
	}
	expiredAt := time.Now().UTC().Add(constants.JWTExpireDuration)
	accessToken, err := utils.GenerateJWT(&userDB, expiredAt)
	if err != nil {
		utils.FailedResponse(c, http.StatusInternalServerError, nil, "")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, models.SignInOutput{AccessToken: accessToken}, "")
}

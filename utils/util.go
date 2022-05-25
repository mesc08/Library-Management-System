package utils

import (
	"errors"
	"project3/config"
	"project3/models"
	"regexp"
	"time"
	"unicode"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator"
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

func ValidateUsersRequestBody(user models.User) error {
	validate := validator.New()
	err := validate.Struct(user)
	if err != nil {
		return err
	}
	resemail := isValidateEmail(user.Email)
	if !resemail {
		err := errors.New("emailid error: invalid email address")
		return err
	}
	respwd := isValidPassword(user.Password)
	if !respwd {
		err := errors.New("password error: invalid password entered")
		return err
	}
	if user.ConfirmPassword != "" {
		respwd := isValidPassword(user.Password)
		if !respwd {
			err := errors.New("password error: invalid password entered")
			return err
		}
	}
	return nil
}

func isValidPassword(password string) bool {
	var upp, low, num, sym bool
	var tot uint8
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			upp = true
			tot++
		case unicode.IsLower(char):
			low = true
			tot++
		case unicode.IsNumber(char):
			num = true
			tot++
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			sym = true
			tot++
		default:
			return false
		}
	}
	if !upp || !low || !num || !sym || tot < 8 {
		return false
	}
	return true
}

func isValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(email)
}

func GenerateID() string {
	guid := xid.New()
	return guid.String()
}

func GenerateJWT(email string) (string, error) {
	var mySigningkey = []byte(config.ViperConfig.SecretKey)
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = email
	claims["authorized"] = true
	claims["exp"] = time.Now().Add(time.Minute * 60).Unix()

	tokenString, err := token.SignedString(mySigningkey)

	if err != nil {
		return "", err
	}
	return tokenString, nil

}

func GeneratehashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidateBookRequestBody(book models.Book) error {
	validate := validator.New()
	err := validate.Struct(book)

	return err
}

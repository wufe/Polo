package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/services"
	"github.com/wufe/polo/pkg/utils"
)

type SignInCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignUpCredentials struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type Token struct {
	Role        string `json:"role"`
	Email       string `json:"email"`
	TokenString string `json:"token"`
}

func (h *Handler) injectAuthEndpoints(router *httprouter.Router, authentication services.AuthenticationService) {
	router.POST("/_polo_/signin", h.getSignInHandler(authentication))
	router.POST("/_polo_/signup", h.withAuth(models.UserRoleAdmin, h.getSignUpHandler(authentication)))
}

func (h *Handler) getSignInHandler(authentication services.AuthenticationService) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

		write := h.r.Write(rw)

		var authdetails SignInCredentials
		err := json.NewDecoder(r.Body).Decode(&authdetails)
		if err != nil {
			write(h.r.BadRequest(fmt.Sprintf("Error in reading body")))
			return
		}

		user, err := authentication.GetUserByCredentials(authdetails.Email, authdetails.Password)

		if err != nil {
			write(h.r.BadRequest(fmt.Sprintf("Wrong credentials")))
			return
		}

		validToken, err := utils.GenerateJWT(user.Email, string(user.Role))
		if err != nil {
			write(h.r.BadRequest(fmt.Sprintf("Error generating JWT")))
			return
		}

		var token Token
		token.Email = user.Email
		token.Role = string(user.Role)
		token.TokenString = validToken

		write(h.r.Ok(token))
	}
}

func (h *Handler) getSignUpHandler(authentication services.AuthenticationService) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

		write := h.r.Write(rw)

		var credentials SignUpCredentials
		err := json.NewDecoder(r.Body).Decode(&credentials)

		if err != nil {
			write(h.r.BadRequest(fmt.Sprintf("Error in reading body")))
			return
		}

		err = authentication.AddUser(credentials.Name, credentials.Email, credentials.Password, credentials.Role)

		if err != nil {
			write(h.r.BadRequest(err.Error()))
			return
		}

		write(h.r.Ok(nil))

	}
}

func (h *Handler) withAuth(role models.UserRole, handler httprouter.Handle) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

		write := h.r.Write(rw)

		if r.Header["Token"] == nil {
			write(h.r.Unauthorized(nil))
			return
		}

		token, err := utils.ParseJWT(r.Header["Token"][0])

		if err != nil {
			write(h.r.Unauthorized(fmt.Errorf("Expired token")))
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if claims["role"] == string(role) {
				handler(rw, r, p)
				return
			}
		}

		write(h.r.Unauthorized(nil))
	}
}

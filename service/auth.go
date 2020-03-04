package service

import (
	"net/http"

	"github.com/calvinchengx/gin-go-pg/apperr"
	"github.com/calvinchengx/gin-go-pg/repository/auth"
	"github.com/calvinchengx/gin-go-pg/request"
	"github.com/gin-gonic/gin"
)

// AuthRouter creates new auth http service
func AuthRouter(svc *auth.Service, r *gin.Engine) {
	a := Auth{svc}
	r.POST("/mobile", a.mobile) // mobile: passwordless authentication which handles both the signup scenario and the login scenario
	r.POST("/signup", a.signup) // email: creates user object
	r.POST("/login", a.login)
	r.GET("/refresh/:token", a.refresh)
	r.GET("/verification/:token", a.verify)  // email: on verification token submission, mark user as verified and return jwt
	r.POST("/mobile/verify", a.mobileVerify) // mobile: on sms code submission, either mark user as verified and return jwt, or update last_login and return jwt
}

// Auth represents auth http service
type Auth struct {
	svc *auth.Service
}

func (a *Auth) login(c *gin.Context) {
	cred, err := request.Login(c)
	if err != nil {
		return
	}
	r, err := a.svc.Authenticate(c, cred.Email, cred.Password)
	if err != nil {
		apperr.Response(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func (a *Auth) refresh(c *gin.Context) {
	r, err := a.svc.Refresh(c, c.Param("token"))
	if err != nil {
		apperr.Response(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func (a *Auth) signup(c *gin.Context) {
	e, err := request.AccountSignup(c)
	if err != nil {
		apperr.Response(c, err)
		return
	}
	err = a.svc.Signup(c, e)
	if err != nil {
		apperr.Response(c, err)
		return
	}
	c.Status(http.StatusCreated)
}

func (a *Auth) verify(c *gin.Context) {
	token := c.Param("token")
	err := a.svc.Verify(c, token)
	if err != nil {
		apperr.Response(c, err)
		return
	}
	c.Status(http.StatusOK)
}

// mobile handles a passwordless mobile signup/login
// if user with country_code and mobile already exists, simply return 200
// if user does not exist yet, we attempt to create the new user object, on success 201, otherwise 500
// the client should call /mobile/verify next, if it receives 201 (newly created user object) or 200 (success, and user was previously created)
// we can use this status code in the client to prepare our request object with Signup attribute as true (201) or false (200)
func (a *Auth) mobile(c *gin.Context) {
	m, err := request.AccountSignupMobile(c)
	if err != nil {
		apperr.Response(c, err)
		return
	}
	err = a.svc.SignupMobile(c, m)
	if err != nil {
		apperr.Response(c, err)
		return
	}
	c.Status(http.StatusCreated)
}

// mobileVerify handles the next API call after the previous client call to /mobile
// we mark user verified AND return jwt
func (a *Auth) mobileVerify(c *gin.Context) {
	m, err := request.AccountVerifyMobile(c)
	if err != nil {
		apperr.Response(c, err)
		return
	}
	err = a.svc.VerifyMobile(c, m.CountryCode, m.Mobile, m.Code)
	if err != nil {
		apperr.Response(c, err)
		return
	}
	if m.Signup {
		// mark our user as verified and generate jwt
	} else {
		// update our user's last_login attribute and generate jwt
	}
	c.Status(http.StatusOK)
}

package web

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"webook/internal/domain"
	"webook/internal/service"
)

type UserHandler struct {
	svc         *service.UserService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	// 信息校验：正则表达式
	const (
		emailRegexPattern    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None) // 预编译
	return &UserHandler{
		svc:         svc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	s := server.Group("/users")
	s.POST("/signup", u.SignUp) // 注册
	// s.POST("/login", u.Login)   // 登录
	s.POST("/login", u.LoginJWT) // 通过JWT登录
	s.POST("/edit", u.Edit)      // 编辑
	// s.GET("/profile", u.Profile)
	s.GET("/profile", u.ProfileJWT)

}

// SignUp 注册
func (u *UserHandler) SignUp(c *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		ConfirmPassword string `json:"confirmPassword"`
		Password        string `json:"password"`
	}

	var req SignUpReq
	if err := c.Bind(&req); err != nil {
		_ = fmt.Errorf("sign up fail. %w\n", err)
		return
	}

	// 信息校验：正则表达式
	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		c.String(http.StatusOK, "System error.") // 不要暴露过多的内部细节
		return
	}
	if !ok {
		c.String(http.StatusOK, "The email format is incorrect.")
		return
	}

	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		c.String(http.StatusOK, "system error.") // 不要暴露过多的内部细节
		return
	}
	if !ok {
		c.String(http.StatusOK, "The password must be longer than 8 characters and include both numbers and special symbols.")
		return
	}

	// 调用一下svc的方法
	if err = u.svc.SignUp(c, domain.User{
		Email:    req.Email,
		Password: req.Password,
	}); err != nil {
		c.String(http.StatusOK, "system error.")
	}

	if errors.Is(err, service.ErrUserDuplicateEmail) {
		c.String(http.StatusOK, "email conflict, please change another email.")
		return
	}
	if err != nil {
		c.String(http.StatusOK, "system error, fail to sign up.")
		return
	}

	if req.ConfirmPassword != req.Password {
		c.String(http.StatusOK, "passwords do not match.")
		return
	}

	// 注册成功
	c.String(http.StatusOK, "sign up success!")
}

// Login 登录
func (u *UserHandler) Login(c *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := c.Bind(&req); err != nil {
		_ = fmt.Errorf("login fail. %w\n", err)
		return
	}

	user, err := u.svc.Login(c, req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		c.String(http.StatusOK, "user or password error.")
		return
	}
	if err != nil {
		c.String(http.StatusOK, "system error,"+err.Error())
		return
	}

	// 在此登录成功了 设置session里的值 步骤2
	sess := sessions.Default(c)
	sess.Set("userId", user.Id)
	sess.Options(sessions.Options{
		// Secure:   true,
		// HttpOnly: true,
		MaxAge: 30 * 60, // 登录有效期30分钟
	})
	sess.Save()

	// 登录成功
	c.String(http.StatusOK, "login success!")

	return
}

// LoginJWT 登录
func (u *UserHandler) LoginJWT(c *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := c.Bind(&req); err != nil {
		_ = fmt.Errorf("login fail. %w\n", err)
		return
	}

	user, err := u.svc.Login(c, req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		c.String(http.StatusOK, "user or password error.")
		return
	}
	if err != nil {
		c.String(http.StatusOK, "system error,"+err.Error())
		return
	}

	// 步骤2 在此设置JWT登录态 生成一个JWT token
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		Uid: user.Id,
		// UserAgent: c.GetHeader("User-Agent"),
		UserAgent: c.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("Cb3cErlIjTEzfHwr6uhsMZ8On5s5EMPK"))
	if err != nil {
		c.String(http.StatusInternalServerError, "system error,"+err.Error())
		return
	}
	c.Header("x-jwt-token", tokenStr)
	fmt.Printf("tokenStr-------->%v\n", tokenStr)
	fmt.Printf("user--------->%v\n", user)

	// 登录成功
	c.String(http.StatusOK, "login success!")

	return
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64 // Uid: 额外加自己的数据在token里
	UserAgent string
}

func (u *UserHandler) LogOut(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Options(sessions.Options{
		// Secure:   true,   // 只在生产环境中设置这两个
		// HttpOnly: true,
		MaxAge: -1, // 把cookie删掉, 即退出登录了
	})
	sess.Save()
	c.String(http.StatusOK, "log out success!")
}

func (u *UserHandler) Edit(c *gin.Context) {
	type editReq struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		Bio      string `json:"bio"`
	}
	var req editReq
	if err := c.Bind(&req); err != nil {
		_ = fmt.Errorf("edit fail. %w\n", err)
		return
	}

	// 设置一些限制
	const MaxNickNameLength = 30
	const MaxBioLength = 300
	if len(req.Nickname) > MaxNickNameLength {
		c.String(http.StatusOK, "nickname must be less than 30 characters.")
		return
	}
	if len(req.Bio) > MaxBioLength {
		c.String(http.StatusOK, "bio must be less than 300 characters.")
		return
	}
	birthdayRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`, regexp.None)
	if ok, _ := birthdayRegex.MatchString(req.Birthday); !ok {
		c.String(http.StatusOK, "birthday must be in the format YYYY-MM-DD.")
		return
	}

	uid := sessions.Default(c).Get("userId")
	err := u.svc.Edit(c, uid, domain.User{
		NickName: req.Nickname,
		Birthday: req.Birthday,
		Bio:      req.Bio,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to update profile.")
		return
	}

	editJson := editReq{
		Nickname: req.Nickname,
		Birthday: req.Birthday,
		Bio:      req.Bio,
	}

	c.JSON(http.StatusOK, gin.H{"Edit your profile in here...": editJson})
}

func (u *UserHandler) Profile(c *gin.Context) {
	uid := sessions.Default(c).Get("userId")
	uu, err := u.svc.Profile(c, uid)
	if errors.Is(err, service.ErrRecordNotFound) {
		c.String(http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to retrieve profile")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"nickname": uu.NickName,
		"birthday": uu.Birthday,
		"bio":      uu.Bio})
}

func (u *UserHandler) ProfileJWT(c *gin.Context) {
	cl, _ := c.Get("claims")
	claims, ok := cl.(*UserClaims) // 类型断言
	if !ok {
		c.String(http.StatusOK, "system error")
	}
	fmt.Printf("claims.Uid-------->%v\n", claims.Uid)
	c.String(http.StatusOK, "hi, here is profile.")
}

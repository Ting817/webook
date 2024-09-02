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

const biz = "login"

// 写法1 确保 UserHandler 上实现了 handler 接口
var _ handler = &UserHandler{}

// 写法2 这个更优雅
var _ handler = (*UserHandler)(nil)

type UserHandler struct {
	svc         service.UserService
	codeSvc     service.CodeService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService) *UserHandler {
	// 信息校验：正则表达式
	const (
		emailRegexPattern    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None) // 预编译
	return &UserHandler{
		svc:         svc,
		codeSvc:     codeSvc,
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
	s.POST("/login_sms/code/send", u.SendLoginSMSCode)
	s.POST("/login_sms", u.LoginSMS)
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

	if req.ConfirmPassword != req.Password {
		c.String(http.StatusOK, "passwords do not match.")
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
	err = u.svc.SignUp(c, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	if err == service.ErrUserDuplicate {
		c.String(http.StatusOK, "email conflict.")
		return
	}

	if err != nil {
		c.String(http.StatusOK, "system error.")
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

	if err = u.setJWTToken(c, user.Id); err != nil {
		c.String(http.StatusOK, "system error,"+err.Error())
		return
	}

	// 登录成功
	c.String(http.StatusOK, "login success!")
	return
}

func (u *UserHandler) setJWTToken(c *gin.Context, uid int64) error {
	// 步骤2 在此设置JWT登录态 生成一个JWT token
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		Uid: uid,
		// UserAgent: c.GetHeader("User-Agent"),
		UserAgent: c.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("Cb3cErlIjTEzfHwr6uhsMZ8On5s5EMPK"))
	if err != nil {
		c.String(http.StatusInternalServerError, "system error,"+err.Error())
		return err
	}
	c.Header("x-jwt-token", tokenStr)
	return nil
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
	// 注意，其它字段，尤其是密码、邮箱和手机, 修改都要通过别的手段, 邮箱和手机都要验证
	type Req struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		Bio      string `json:"bio"`
	}
	var req Req
	if err := c.Bind(&req); err != nil {
		_ = fmt.Errorf("edit fail. %w\n", err)
		return
	}

	// 设置一些限制
	const MaxNickNameLength = 30
	const MaxBioLength = 300
	if req.Nickname == "" {
		c.JSON(http.StatusOK, Result{
			Code: 4, Msg: "nickname can be empty.",
		})
	}
	if len(req.Nickname) > MaxNickNameLength {
		c.JSON(http.StatusOK, Result{
			Code: 4, Msg: "nickname must be less than 30 characters.",
		})
		return
	}
	if len(req.Bio) > MaxBioLength {
		c.JSON(http.StatusOK, Result{
			Code: 4, Msg: "bio must be less than 300 characters.",
		})
		return
	}
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Code: 4, Msg: "birthday must be in the format YYYY-MM-DD.",
		})
		return
	}
	uc := c.MustGet("user").(UserClaims)
	err = u.svc.UpdateNonSensitiveInfo(c, domain.User{
		Id:       uc.Uid,
		NickName: req.Nickname,
		Birthday: birthday,
		Bio:      req.Bio,
	})
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Code: 4, Msg: "Failed to update profile.",
		})
		return
	}

	c.JSON(http.StatusOK, Result{Msg: "ok!"})
}

func (u *UserHandler) Profile(c *gin.Context) {
	uid := sessions.Default(c).Get("userId").(int64)
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
	// 重新控制 profile 防止密码泄露
	type Profile struct {
		Email    string
		Phone    string
		Nickname string
		Birthday string
		Bio      string
	}
	uc := c.MustGet("user").(UserClaims)
	ucId, err := u.svc.Profile(c, uc.Uid) // 类型断言
	if err != nil {
		c.String(http.StatusOK, "system error"+err.Error())
	}
	c.JSON(http.StatusOK, Profile{
		Email:    ucId.Email,
		Phone:    ucId.Phone,
		Nickname: ucId.NickName,
		Birthday: ucId.Birthday.Format(time.DateOnly),
		Bio:      ucId.Bio,
	})
}

func (u *UserHandler) SendLoginSMSCode(c *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := c.Bind(&req); err != nil {
		return
	}
	// 是不是合法的手机号 考虑正则表达式
	if req.Phone == "" {
		c.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "phone input error",
		})
	}
	err := u.codeSvc.Send(c, biz, req.Phone)
	switch err {
	case nil:
		c.JSON(http.StatusOK, Result{
			Msg: "send code success!",
		})
	case service.ErrCodeSendTooMany:
		c.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "code send too many, please try it again later",
		})
	default:
		c.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "system error",
		})
	}
}

func (u *UserHandler) LoginSMS(c *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := c.Bind(&req); err != nil {
		return
	}
	ok, err := u.codeSvc.Verify(c, biz, req.Phone, req.Code)
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "system error!" + err.Error(),
		})
		return
	}
	if !ok {
		c.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "code error!",
		})
		return
	}

	user, err := u.svc.FindOrCreate(c, req.Phone)
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "system error!" + err.Error(),
		})
		return
	}

	if err = u.setJWTToken(c, user.Id); err != nil {
		c.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "set jwt token error!" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Result{
		Msg: "code verify success!",
	})
}

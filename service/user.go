package service

import (
	"context"
	"errors"
	"mime/multipart"
	"strings"
	"sync"
	"time"

	logging "github.com/sirupsen/logrus"
	"gopkg.in/mail.v2"

	"mall/conf"
	"mall/consts"
	"mall/pkg/e"
	util "mall/pkg/utils"
	"mall/repository/db/dao"
	"mall/repository/db/model"
	"mall/types"
)

var UserSrvIns *UserSrv
var UserSrvOnce sync.Once

type UserSrv struct {
}

func GetUserSrv() *UserSrv {
	UserSrvOnce.Do(func() {
		UserSrvIns = &UserSrv{}
	})
	return UserSrvIns
}

func (s *UserSrv) Register(ctx context.Context, req *types.UserServiceReq) (types.Response, error) {
	var user *model.User
	code := e.SUCCESS
	if req.Key == "" || len(req.Key) != 16 {
		code = e.ERROR
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Data:   "密钥长度不足",
		}, errors.New("迷药cha")
	}
	util.Encrypt.SetKey(req.Key)
	userDao := dao.NewUserDao(ctx)
	_, exist, err := userDao.ExistOrNotByUserName(req.UserName)
	if err != nil {
		code = e.ErrorDatabase
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}, err
	}
	if exist {
		code = e.ErrorExistUser
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}, errors.New("已经存在了")
	}
	user = &model.User{
		NickName: req.NickName,
		UserName: req.UserName,
		Status:   model.Active,
		Money:    util.Encrypt.AesEncoding("10000"), // 初始金额
	}
	// 加密密码
	if err = user.SetPassword(req.Password); err != nil {
		logging.Info(err)
		code = e.ErrorFailEncryption
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}, err
	}
	if conf.UploadModel == consts.UploadModelOss {
		user.Avatar = "http://q1.qlogo.cn/g?b=qq&nk=294350394&s=640"
	} else {
		user.Avatar = "avatar.JPG"
	}
	// 创建用户
	err = userDao.CreateUser(user)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}, err
	}
	return types.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}, nil
}

// Login 用户登陆函数
func (s *UserSrv) Login(ctx context.Context, req *types.UserServiceReq) (types.Response, error) {
	var user *model.User
	code := e.SUCCESS
	userDao := dao.NewUserDao(ctx)
	user, exist, err := userDao.ExistOrNotByUserName(req.UserName)
	if !exist { // 如果查询不到，返回相应的错误
		logging.Info(err)
		code = e.ErrorUserNotFound
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}, errors.New("不存在")
	}
	if user.CheckPassword(req.Password) == false {
		code = e.ErrorNotCompare
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}, errors.New("加密失败了")
	}
	token, err := util.GenerateToken(user.ID, req.UserName, 0)
	if err != nil {
		logging.Info(err)
		code = e.ErrorAuthToken
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}, err
	}
	return types.Response{
		Status: code,
		Data:   types.TokenData{User: types.BuildUser(user), Token: token},
		Msg:    e.GetMsg(code),
	}, nil
}

// Update 用户修改信息
func (s *UserSrv) Update(ctx context.Context, uId uint, req *types.UserServiceReq) (types.Response, error) {
	var user *model.User
	var err error
	code := e.SUCCESS
	// 找到用户
	userDao := dao.NewUserDao(ctx)
	user, err = userDao.GetUserById(uId)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}, err
	}
	if req.NickName != "" {
		user.NickName = req.NickName
	}

	err = userDao.UpdateUserById(uId, user)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}, err
	}

	return types.Response{
		Status: code,
		Data:   types.BuildUser(user),
		Msg:    e.GetMsg(code),
	}, nil
}

func (s *UserSrv) Post(ctx context.Context, uId uint, file multipart.File, fileSize int64, req *types.UserServiceReq) (types.Response, error) {
	code := e.SUCCESS
	var user *model.User
	var err error

	userDao := dao.NewUserDao(ctx)
	user, err = userDao.GetUserById(uId)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}, err
	}
	var path string
	if conf.UploadModel == consts.UploadModelLocal { // 兼容两种存储方式
		path, err = util.UploadAvatarToLocalStatic(file, uId, user.UserName)
	} else {
		path, err = util.UploadToQiNiu(file, fileSize)
	}
	if err != nil {
		code = e.ErrorUploadFile
		return types.Response{
			Status: code,
			Data:   e.GetMsg(code),
			Error:  path,
		}, err
	}

	user.Avatar = path
	err = userDao.UpdateUserById(uId, user)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}, err
	}
	return types.Response{
		Status: code,
		Data:   types.BuildUser(user),
		Msg:    e.GetMsg(code),
	}, nil
}

// Send 发送邮件
func (s *UserSrv) Send(ctx context.Context, id uint, req *types.SendEmailServiceReq) (types.Response, error) {
	code := e.SUCCESS
	var address string
	var notice *model.Notice

	token, err := util.GenerateEmailToken(id, req.OperationType, req.Email, req.Password)
	if err != nil {
		logging.Info(err)
		code = e.ErrorAuthToken
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}, err
	}

	noticeDao := dao.NewNoticeDao(ctx)
	notice, err = noticeDao.GetNoticeById(req.OperationType)
	if err != nil {
		logging.Info(err)
		code = e.ErrorDatabase
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}, err
	}
	address = conf.ValidEmail + token
	mailStr := notice.Text
	mailText := strings.Replace(mailStr, "Email", address, -1)
	m := mail.NewMessage()
	m.SetHeader("From", conf.SmtpEmail)
	m.SetHeader("To", req.Email)
	m.SetHeader("Subject", "FanOne")
	m.SetBody("text/html", mailText)
	d := mail.NewDialer(conf.SmtpHost, 465, conf.SmtpEmail, conf.SmtpPass)
	d.StartTLSPolicy = mail.MandatoryStartTLS
	if err := d.DialAndSend(m); err != nil {
		logging.Info(err)
		code = e.ErrorSendEmail
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}, err
	}
	return types.Response{
		Status: code,
		Msg:    e.GetMsg(code),
	}, nil
}

// Valid 验证内容
func (s *UserSrv) Valid(ctx context.Context, token string, req *types.ValidEmailServiceReq) (types.Response, error) {
	var userID uint
	var email string
	var password string
	var operationType uint
	code := e.SUCCESS

	// 验证token
	if token == "" {
		code = e.InvalidParams
	} else {
		claims, err := util.ParseEmailToken(token)
		if err != nil {
			logging.Info(err)
			code = e.ErrorAuthCheckTokenFail
		} else if time.Now().Unix() > claims.ExpiresAt {
			code = e.ErrorAuthCheckTokenTimeout
		} else {
			userID = claims.UserID
			email = claims.Email
			password = claims.Password
			operationType = claims.OperationType
		}
	}
	if code != e.SUCCESS {
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}, errors.New("操作失败")
	}

	// 获取该用户信息
	userDao := dao.NewUserDao(ctx)
	user, err := userDao.GetUserById(userID)
	if err != nil {
		code = e.ErrorDatabase
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}, err
	}
	if operationType == 1 {
		// 1:绑定邮箱
		user.Email = email
	} else if operationType == 2 {
		// 2：解绑邮箱
		user.Email = ""
	} else if operationType == 3 {
		// 3：修改密码
		err = user.SetPassword(password)
		if err != nil {
			code = e.ErrorDatabase
			return types.Response{
				Status: code,
				Msg:    e.GetMsg(code),
			}, err
		}
	}
	err = userDao.UpdateUserById(userID, user)
	if err != nil {
		code = e.ErrorDatabase
		return types.Response{
			Status: code,
			Msg:    e.GetMsg(code),
		}, err
	}
	// 成功则返回用户的信息
	return types.Response{
		Status: code,
		Msg:    e.GetMsg(code),
		Data:   types.BuildUser(user),
	}, err
}

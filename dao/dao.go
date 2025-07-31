package dao

import (
	"github.com/daodao97/xgo/xdb"
	_ "github.com/go-sql-driver/mysql"
)

var UserModel xdb.Model
var UserBalanceModel xdb.Model
var ProjectUserBalanceModel xdb.Model
var ProjectApiTokenModel xdb.Model

type Options struct {
	UserModel xdb.Model
}

type Option func(*Options)

func WithUserModel(m xdb.Model) Option {
	return func(o *Options) {
		o.UserModel = m
	}
}

func Init(opts ...Option) error {
	options := &Options{
		UserModel: xdb.New("project_user"),
	}
	for _, opt := range opts {
		opt(options)
	}
	UserModel = xdb.New("project_user")
	if options.UserModel != nil {
		UserModel = options.UserModel
	}
	UserBalanceModel = xdb.New("project_user_balance")
	ProjectUserBalanceModel = xdb.New("project_user_balance")
	ProjectApiTokenModel = xdb.New("project_api_token")
	return nil
}

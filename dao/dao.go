package dao

import (
	"github.com/daodao97/goreact/base/login"

	"github.com/daodao97/xgo/xdb"
	_ "github.com/go-sql-driver/mysql"
)

var UserModel xdb.Model
var UserBalanceModel xdb.Model
var ProjectWanxModel xdb.Model
var ProjectUserBalanceModel xdb.Model
var ProjectApiTokenModel xdb.Model

func Init() error {
	UserModel = xdb.New("project_user")
	UserBalanceModel = xdb.New("project_user_balance")
	ProjectUserBalanceModel = xdb.New("project_user_balance")
	login.SetUserMoel(UserModel)
	ProjectApiTokenModel = xdb.New("project_api_token")
	return nil
}

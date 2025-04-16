package dao

import (
	"github.com/daodao97/xgo/xdb"
)

func GetUser(id string) ([]xdb.Record, error) {
	return UserModel.Selects()
}

func GetUserById(id int) (xdb.Record, error) {
	return UserModel.First(xdb.WhereEq("id", id))
}

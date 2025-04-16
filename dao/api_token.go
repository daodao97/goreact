package dao

import (
	"github.com/daodao97/xgo/xdb"
)

func GetApiToken(uid int64) (xdb.Record, error) {
	apiToken, err := ProjectApiTokenModel.First(xdb.WhereEq("uid", uid))
	if err != nil {
		return nil, err
	}
	return apiToken, nil
}

func GetApiTokenByToken(token string) (xdb.Record, error) {
	apiToken, err := ProjectApiTokenModel.First(xdb.WhereEq("token", token))
	if err != nil {
		return nil, err
	}
	return apiToken, nil
}

func CreateApiToken(uid int64, token string) error {
	_, err := ProjectApiTokenModel.Insert(xdb.Record{
		"uid":    uid,
		"token":  token,
		"status": 1,
	})
	if err != nil {
		return err
	}
	return nil
}

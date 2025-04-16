package dao

import (
	"context"
	"fmt"

	"github.com/daodao97/xgo/xdb"
)

func GetUserBalance(ctx context.Context, userID string) (int, error) {
	balance, err := UserBalanceModel.Ctx(ctx).First(xdb.WhereEq("uid", userID))
	fmt.Println("balance-debug", balance.GetInt("balance"), err)
	if err != nil {
		return 0, err
	}
	return balance.GetInt("balance"), nil
}

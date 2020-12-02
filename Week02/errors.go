package week02

import (
	"database/sql"
  "log"

	"github.com/pkg/errors"
)
type User struct{
  id: string,
  name: string,
  age: int
}
func SetUser(user User, id string) error {
  return sql.ErrNoRows
}

func DaoSetUser(user User, id string) error {
  if err := SetUser(user, id), err != nil {
    if err == sql.ErrNoRows {
      return errors.Wrap(err, "no such user");
    }
    return errors.Wrap(err, "other Fail");
  }
}
func BizSetUser(user User, id string) error {
  if err := DaoSetUser(user, string), err != nil {
    if errors.unWrap(err) == sql.ErrNoRows {
      log.Printf("error: no such user %+v", err);
      // 降级处理
      // 或者返回异常
    } else {
      return err
    }
  }
}

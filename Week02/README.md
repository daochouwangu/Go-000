# 第二周

作业：1. dao 层中当遇到一个 sql.ErrNoRows 的时候，是否应该 Wrap 这个 error，抛给上层。为什么?

答 应该。
有两种情况，一种是我的sql 出现norows是正常情况，这种时候应该由上层决定怎么处理，如果这时候直接处理异常（打日志） 可能会污染日志
第二种是异常情况，这种时候也应该由上层处理，因为只有上层的context 的价值是有用的，并且可以做降级处理。

fake代码：


```golang
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
```

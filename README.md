# go-message
A wrapper for message.net.tw's API written in Go

## Example

```go
package main

import (
	"fmt"

	"github.com/fshiori/go-message"
)

func main() {
	msg := message.NewMessage("account", "password")
  msg.Send(0, "20170808000000", []string{"0912345678"}, "測試１２３４５ＡＢＣ")
	queryLog, _ := msg.QueryLog(0, []string{}, "20170804143000", "", 0, 0, []string{})
	fmt.Println(queryLog)
}
```

package closelogger

import (
	"example.com/btclient/pkg/logutil"
	"fmt"
	"io"
)

func CloseOrLog(c io.Closer, resource string) {
	if err := c.Close(); err != nil {
		logutil.Printf(fmt.Sprintf("Error closing resource %s: %+v\n", resource, err))
	}
}

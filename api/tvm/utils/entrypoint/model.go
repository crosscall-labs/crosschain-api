package entrypoint

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type EntrypointMessage struct {
	Destination *address.Address
	Body        *cell.Cell
}

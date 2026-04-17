package bootstrap

import (
	"github.com/muhammadfarrasfajri/filantropi/database"
)

func InitDatabase() {
	database.ConnectMySQL()
}

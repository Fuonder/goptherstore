package postrge

import "github.com/Fuonder/goptherstore.git/internal/storage"

type PsqlStorage struct {
	conn *storage.DBConnection
}

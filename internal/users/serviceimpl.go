package users

import "context"

type UService struct {
	conn DatabaseUsers
}

func NewUService(conn DatabaseUsers) *UService {
	return &UService{conn: conn}
}

func (s *UService) GetUID(ctx context.Context, login string) (int, error) {
	UID, err := s.conn.GetUIDByUsername(ctx, login)
	if err != nil {
		return 0, err
	}
	return UID, nil
}

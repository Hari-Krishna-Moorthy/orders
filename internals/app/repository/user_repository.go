package repository

import (
	"sync"

	"github.com/Hari-Krishna-Moorthy/orders/internals/app/models"
	"github.com/Hari-Krishna-Moorthy/orders/internals/app/types"
)

type UserRepository struct {
	mu      sync.RWMutex
	auto    uint
	userMap map[string]*models.User
}

var (
	userRepo *UserRepository
	once     sync.Once
)

func NewUserRepository() *UserRepository {
	once.Do(func() { userRepo = &UserRepository{userMap: map[string]*models.User{}} })
	return userRepo
}

func GetUserRepository() *UserRepository { return userRepo }

func (r *UserRepository) Create(u *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.userMap[u.Email]; ok {
		return types.ALREADY_EXISTS_ERROR
	}
	r.auto++
	u.ID = r.auto
	r.userMap[u.Email] = u
	return nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.userMap[email]
	if !ok {
		return nil, types.NOT_FOUND_ERROR
	}
	return u, nil
}

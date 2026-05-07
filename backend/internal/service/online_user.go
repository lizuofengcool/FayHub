package service

import (
	"context"
	"fmt"
	"time"

	"fayhub/internal/model"
	"fayhub/pkg/redisclient"
	"fayhub/pkg/tokenblacklist"
)

const onlineUserPrefix = "online:user:"
const onlineUserTTL = 15 * time.Minute

type OnlineUserService struct{}

func (s *OnlineUserService) RecordActivity(ctx context.Context, user model.OnlineUser) error {
	key := fmt.Sprintf("%s%d", onlineUserPrefix, user.UserID)
	user.LastSeen = time.Now()
	return redisclient.Set(ctx, key, user, onlineUserTTL)
}

func (s *OnlineUserService) GetOnlineUsers(ctx context.Context) ([]model.OnlineUser, error) {
	users := make([]model.OnlineUser, 0)

	if !redisclient.IsEnabled() {
		return users, nil
	}

	rawClient := redisclient.GetRawClient()
	if rawClient == nil {
		return users, nil
	}

	keys, err := rawClient.Keys(ctx, onlineUserPrefix+"*").Result()
	if err != nil {
		return users, nil
	}

	for _, key := range keys {
		var user model.OnlineUser
		if err := redisclient.Get(ctx, key, &user); err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

func (s *OnlineUserService) ForceLogout(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("%s%d", onlineUserPrefix, userID)

	var user model.OnlineUser
	if err := redisclient.Get(ctx, key, &user); err != nil {
		return fmt.Errorf("用户不在线")
	}

	redisclient.Del(ctx, key)

	return nil
}

func (s *OnlineUserService) ForceLogoutWithToken(ctx context.Context, userID int64, tokenString string) error {
	key := fmt.Sprintf("%s%d", onlineUserPrefix, userID)

	var user model.OnlineUser
	if err := redisclient.Get(ctx, key, &user); err != nil {
		return fmt.Errorf("用户不在线")
	}

	redisclient.Del(ctx, key)

	if tokenString != "" {
		expiresAt := time.Now().Add(24 * time.Hour)
		return tokenblacklist.Add(ctx, tokenString, expiresAt)
	}

	return nil
}

func (s *OnlineUserService) GetOnlineCount(ctx context.Context) (int64, error) {
	if !redisclient.IsEnabled() {
		return 0, nil
	}

	rawClient := redisclient.GetRawClient()
	if rawClient == nil {
		return 0, nil
	}

	keys, err := rawClient.Keys(ctx, onlineUserPrefix+"*").Result()
	if err != nil {
		return 0, nil
	}

	return int64(len(keys)), nil
}

func (s *OnlineUserService) IsOnline(ctx context.Context, userID int64) bool {
	key := fmt.Sprintf("%s%d", onlineUserPrefix, userID)
	var user model.OnlineUser
	return redisclient.Get(ctx, key, &user) == nil
}

func (s *OnlineUserService) RemoveUser(ctx context.Context, userID int64) {
	key := fmt.Sprintf("%s%d", onlineUserPrefix, userID)
	redisclient.Del(ctx, key)
}

func (s *OnlineUserService) RecordLogin(ctx context.Context, user model.OnlineUser) error {
	key := fmt.Sprintf("%s%d", onlineUserPrefix, user.UserID)
	user.LoginAt = time.Now()
	user.LastSeen = time.Now()
	return redisclient.Set(ctx, key, user, onlineUserTTL)
}

package explorer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	api "github.com/usememos/memos/api/v1"
	"github.com/usememos/memos/common/log"
	"github.com/usememos/memos/store"
	"go.uber.org/zap"
)

type Explorer struct {
	Store *store.Store
}

func NewExplorer(store *store.Store) *Explorer {
	return &Explorer{
		Store: store,
	}
}

func (e *Explorer) Run(ctx context.Context) {
	log.Info("running explorer in background every 10 minutes")

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("stop explorer graceful.")
			return
		case <-ticker.C:
		}

		err := e.SyncAllExternalUsers(ctx)
		if err != nil {
			log.Error("fail to explore external user", zap.Error(err))
		}
	}
}

var (
	userRole = store.RoleExternal
	findUser = &store.FindUser{Role: &userRole}
)

func (e *Explorer) SyncAllExternalUsers(ctx context.Context) error {
	users, err := e.Store.ListUsers(ctx, findUser)
	if err != nil {
		return fmt.Errorf("fail to fetch external users list: %s", err)
	}

	for _, user := range users {
		err := e.syncExternalUser(ctx, user)
		if err != nil {
			return fmt.Errorf("fail to sync user ID=%d %w", user.ID, err)
		}
	}

	return nil
}

func (r *Explorer) syncExternalUser(ctx context.Context, user *store.User) error {
	u, err := url.Parse(user.Username)
	if err != nil {
		return fmt.Errorf("fail to parse external user address %w", err)
	}

	q := url.Values{
		"creatorUsername": {strings.TrimPrefix(u.Path, "/u/")},
		"rowStatus":       {store.Normal.String()},
		"limit":           {"2"},
	}
	u.Path = "/api/v1/memo"
	u.RawQuery = q.Encode()

	httpClient := http.Client{Timeout: time.Second}
	resp, err := httpClient.Get(u.String())
	if err != nil {
		return fmt.Errorf("fail to request exteranl user memo %w", err)
	}

	var respInfo []*api.Memo
	err = json.NewDecoder(resp.Body).Decode(&respInfo)
	if err != nil {
		return fmt.Errorf("fail to parse external user memos %w", err)
	}

	for _, m := range respInfo {
		create := store.Memo{
			CreatorID:  user.ID,
			CreatedTs:  m.CreatedTs,
			UpdatedTs:  m.UpdatedTs,
			Content:    m.Content,
			Visibility: store.Protected,
		}
		_, err := r.Store.CreateMemo(ctx, &create)
		if err != nil {
			return fmt.Errorf("fail to save memo for external user %w", err)
		}
	}

	return nil
}

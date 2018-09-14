// Copyright (c) 2018-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package store

import (
	"context"

	"github.com/mattermost/mattermost-server/model"
)

func (s *RedisSupplier) GroupSave(ctx context.Context, group *model.Group, hints ...LayeredStoreHint) *LayeredStoreSupplierResult {
	// TODO: Redis caching.
	return s.Next().GroupSave(ctx, group, hints...)
}

func (s *RedisSupplier) GroupGet(ctx context.Context, groupId string, hints ...LayeredStoreHint) *LayeredStoreSupplierResult {
	// TODO: Redis caching.
	return s.Next().GroupGet(ctx, groupId, hints...)
}

func (s *RedisSupplier) GroupGetAllPage(ctx context.Context, offset int, limit int, hints ...LayeredStoreHint) *LayeredStoreSupplierResult {
	// TODO: Redis caching.
	return s.Next().GroupGetAllPage(ctx, offset, limit, hints...)
}

func (s *RedisSupplier) GroupDelete(ctx context.Context, groupId string, hints ...LayeredStoreHint) *LayeredStoreSupplierResult {
	// TODO: Redis caching.
	return s.Next().GroupDelete(ctx, groupId, hints...)
}

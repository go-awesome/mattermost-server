// Copyright (c) 2018-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

type GroupMember struct {
	GroupId  string `json:"group_id"`
	UserId   string `json:"user_id"`
	CreateAt int64  `json:"create_at"`
	DeleteAt int64  `json:"delete_at"`
}

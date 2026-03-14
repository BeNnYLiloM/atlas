package service

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidProfile     = errors.New("invalid profile data")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrWorkspaceNotFound  = errors.New("workspace not found")
	ErrChannelNotFound    = errors.New("channel not found")
	ErrMessageNotFound    = errors.New("message not found")
	ErrTaskNotFound       = errors.New("task not found")
	ErrNotMember          = errors.New("user is not a member of workspace")
	ErrNotFound           = errors.New("not found")
	ErrSlowmode           = errors.New("slowmode: wait before sending next message")
	ErrProjectNotFound    = errors.New("project not found")
	ErrNotProjectMember   = errors.New("user is not a member of this project")
	ErrProjectArchived    = errors.New("project is archived")
	ErrLastLead           = errors.New("cannot remove the last lead from a project")
	ErrDMSelf             = errors.New("cannot create DM with yourself")
	ErrUserDeactivated    = errors.New("target user is deactivated")
)

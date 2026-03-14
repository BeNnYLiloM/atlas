export type UserStatus = 'online' | 'away' | 'dnd' | 'offline'

// User types
export interface User {
  id: string
  email: string
  display_name: string
  avatar_url: string | null
  status: UserStatus
  custom_status: string | null
  last_seen: string | null
  created_at: string
}

export interface UserCreate {
  email: string
  password: string
  display_name: string
}

export interface UserLogin {
  email: string
  password: string
}

export interface UserUpdate {
  display_name?: string
  avatar_url?: string
}

// Workspace types
export interface Workspace {
  id: string
  name: string
  owner_id: string
  description: string | null
  icon_url: string | null
  created_at: string
}

export interface WorkspaceMember {
  workspace_id: string
  user_id: string
  role: WorkspaceSystemRole
  display_name: string
  avatar_url: string | null
  nickname: string | null
  custom_roles: WorkspaceRole[]
}

export type WorkspaceSystemRole = 'owner' | 'admin' | 'member'

export interface WorkspaceCreate {
  name: string
}

export interface WorkspaceUpdate {
  name?: string
  description?: string | null
  icon_url?: string | null
}

export interface WorkspaceMemberUpdate {
  role?: WorkspaceSystemRole
  nickname?: string | null
}

// Channel types
export interface Channel {
  id: string
  workspace_id: string
  name: string
  type: ChannelType
  is_private: boolean
  topic: string | null
  slowmode_seconds: number
  position: number
  category_id: string | null
  project_id: string | null
  created_at: string
}

export interface ChannelCategory {
  id: string
  workspace_id: string
  name: string
  position: number
  is_private: boolean
  project_id: string | null
  created_at: string
}

export interface ChannelCategoryCreate {
  name: string
  is_private?: boolean
  project_id?: string | null
}

export interface ChannelCategoryUpdate {
  name?: string
  position?: number
  is_private?: boolean
}

export interface CategoryPermissions {
  roles: ChannelAllowedRole[]
  users: ChannelAllowedUser[]
}

export interface ChannelAllowedRole {
  channel_id: string
  role_id: string
  role_name: string
  role_color: string
}

export interface ChannelAllowedUser {
  channel_id: string
  user_id: string
  display_name: string
  avatar_url: string | null
}

export interface ChannelWithUnread extends Channel {
  unread_count: number
  mention_count: number
  notification_level: NotificationLevel
}

export type ChannelType = 'text' | 'voice'

export type NotificationLevel = 'all' | 'mentions' | 'nothing'

export interface ChannelCreate {
  workspace_id: string
  name: string
  type: ChannelType
  is_private: boolean
  category_id?: string | null
  project_id?: string | null
}

export interface ChannelUpdate {
  name?: string
  topic?: string | null
  is_private?: boolean
  slowmode_seconds?: number
}

export interface ChannelMemberInfo {
  user_id: string
  channel_id: string
  display_name: string
  avatar_url: string | null
}

// --- Workspace Roles ---

export interface RolePermissions {
  manage_workspace: boolean
  manage_roles: boolean
  manage_channels: boolean
  manage_members: boolean
  view_audit_log: boolean
  send_messages: boolean
  attach_files: boolean
  mention_everyone: boolean
  manage_messages: boolean
  view_channels: boolean
  create_projects: boolean
  view_all_projects: boolean
}

export interface WorkspaceRole {
  id: string
  workspace_id: string
  name: string
  color: string
  position: number
  is_system: boolean
  permissions: RolePermissions
  created_at: string
  member_count?: number
}

export interface WorkspaceRoleCreate {
  name: string
  color?: string
  permissions: RolePermissions
}

export interface WorkspaceRoleUpdate {
  name?: string
  color?: string
  permissions?: RolePermissions
}

export function defaultPermissions(): RolePermissions {
  return {
    manage_workspace: false,
    manage_roles: false,
    manage_channels: false,
    manage_members: false,
    view_audit_log: false,
    send_messages: true,
    attach_files: true,
    mention_everyone: false,
    manage_messages: false,
    view_channels: true,
    create_projects: false,
    view_all_projects: false,
  }
}

// --- Channel Permissions ---

export interface ChannelPermissions {
  roles: ChannelAllowedRole[]
  users: ChannelAllowedUser[]
}

// Message types
export interface Message {
  id: string
  channel_id: string
  user_id: string
  content: string
  parent_id: string | null
  created_at: string
  updated_at: string | null
  user?: User
  thread_replies_count?: number
  thread_unread_count?: number
}

export interface MessageCreate {
  channel_id: string
  content: string
  parent_id?: string
}

export interface MessageUpdate {
  content: string
}

// Auth types
export interface AuthTokens {
  access_token: string
  expires_at: number
}

export interface AuthResponse {
  user: User
  tokens: AuthTokens
}

export interface RefreshResponse {
  tokens: AuthTokens
}

// API response wrapper
export interface ApiResponse<T> {
  data: T
  message?: string
}

// WebSocket event types
export interface WSMessage {
  type: 'message' | 'typing' | 'presence' | 'channel_update'
  payload: unknown
}

export interface TypingEvent {
  channel_id: string
  user_id: string
  is_typing: boolean
}

export interface PresenceEvent {
  user_id: string
  status: 'online' | 'offline' | 'away'
}

// --- Project types ---

export interface Project {
  id: string
  workspace_id: string
  name: string
  description: string | null
  icon_url: string | null
  is_archived: boolean
  created_at: string
}

export interface ProjectMember {
  project_id: string
  user_id: string
  is_lead: boolean
  joined_at: string
  display_name: string
  avatar_url: string | null
}

export interface ProjectCreate {
  name: string
  description?: string | null
  icon_url?: string | null
}

export interface ProjectUpdate {
  name?: string
  description?: string | null
  icon_url?: string | null
  clear_icon_url?: boolean
}

export interface ProjectMemberAdd {
  user_id: string
}

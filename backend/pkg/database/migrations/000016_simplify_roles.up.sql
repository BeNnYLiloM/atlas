-- Migration: 000016_simplify_roles
-- Description: Упрощение системных ролей до owner/admin/member

UPDATE workspace_members SET role = 'member' WHERE role IN ('moderator', 'guest');

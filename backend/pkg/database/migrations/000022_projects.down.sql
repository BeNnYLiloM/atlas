-- Migration: 000022_projects (down)
ALTER TABLE channel_categories DROP COLUMN IF EXISTS project_id;
ALTER TABLE channels DROP COLUMN IF EXISTS project_id;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;

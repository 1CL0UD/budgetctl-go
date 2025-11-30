-- Modify "users" table
ALTER TABLE "public"."users" ADD COLUMN "name" text NULL, ADD COLUMN "avatar_url" text NULL, ADD COLUMN "preferences" jsonb NOT NULL DEFAULT '{}';

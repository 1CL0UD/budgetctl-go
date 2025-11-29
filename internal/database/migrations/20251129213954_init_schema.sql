-- Create "users" table
CREATE TABLE "public"."users" (
  "id" bigserial NOT NULL,
  "email" text NOT NULL,
  "password_hash" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id")
);
-- Create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX "users_email_key" ON "public"."users" ("email");
-- Create "transactions" table
CREATE TABLE "public"."transactions" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "amount" numeric(10,2) NOT NULL,
  "description" text NOT NULL,
  "category" text NOT NULL,
  "date" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_transactions_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_transactions_user" to table: "transactions"
CREATE INDEX "idx_transactions_user" ON "public"."transactions" ("user_id");

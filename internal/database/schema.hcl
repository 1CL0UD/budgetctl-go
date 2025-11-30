schema "public" {}

// 1. Users Table (For Auth)
table "users" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "email" {
    null = false
    type = text
  }
  column "name" {
    null = true
    type = text
  }
  column "avatar_url" {
    null = true
    type = text
  }
  column "preferences" {
    null    = false
    type    = jsonb
    default = sql("'{}'::jsonb")
  }
  column "password_hash" {
    null = false
    type = text
  }
  column "created_at" {
    null = false
    type = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.id]
  }

  index "users_email_key" {
    unique = true
    columns = [column.email]
  }
}

// 2. Transactions Table
table "transactions" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "amount" {
    null = false
    type = numeric(10, 2)
  }
  column "description" {
    null = false
    type = text
  }
  column "category" {
    null = false
    type = text
  }
  column "type" {
    null = false
    type = text
    default = sql("'expense'")
  }
  column "currency" {
    null = false
    type = text
    default = sql("'USD'")
  }
  column "status" {
    null = false
    type = text
    default = sql("'pending'")
  }
  column "account" {
    null = false
    type = text
    default = sql("''")
  }
  column "tags" {
    null    = false
    type    = sql("text[]")
    default = sql("'{}'::text[]")
  }
  column "notes" {
    null = true
    type = text
  }
  column "has_receipt" {
    null    = false
    type    = boolean
    default = sql("false")
  }
  column "receipt_url" {
    null = true
    type = text
  }
  column "date" {
    null = false
    type = timestamptz
    default = sql("now()")
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.id]
  }

  foreign_key "fk_transactions_user" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete   = CASCADE
  }

  index "idx_transactions_user" {
    columns = [column.user_id]
  }
}

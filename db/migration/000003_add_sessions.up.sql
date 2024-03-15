CREATE TABLE "sessions" (
  "id" uuid PRIMARY KEY NOT NULL,
  "username" varchar NOT NULL,
  "is_blocked" boolean NOT NULL DEFAULT false,
  "client_ip" varchar,
  "user_agent" varchar,
  "refresh_token" varchar NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);


ALTER TABLE "sessions" ADD FOREIGN KEY ("username") REFERENCES "users" ("username");


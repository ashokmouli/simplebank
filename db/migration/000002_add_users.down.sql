ALTER TABLE IF exists "accounts" DROP CONSTRAINT IF exists "owner_currency_key";
ALTER TABLE IF exists "accounts" DROP CONSTRAINT IF exists "accounts_owner_fkey";
DROP TABLE IF exists "users";
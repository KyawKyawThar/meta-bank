ALTER TABLE IF EXISTS entries DROP CONSTRAINT IF EXISTS entries_account_id_fkey;
ALTER TABLE IF EXISTS transfers DROP CONSTRAINT IF EXISTS transfers_from_account_id_fkey;
ALTER TABLE IF EXISTS transfers DROP CONSTRAINT IF EXISTS transfers_to_account_id_fkey;

DROP TABLE IF EXISTS entries;
DROP TABLE IF EXISTS transfers;




-- User requested permanent removal of the travel guide. Restore a pre-migration backup to recover it.
-- No CASCADE: unexpected dependencies must fail the transaction instead of deleting unrelated data.
DROP TABLE post_tags;
DROP TABLE posts;
DROP TABLE post_categories;
DROP TABLE tags;

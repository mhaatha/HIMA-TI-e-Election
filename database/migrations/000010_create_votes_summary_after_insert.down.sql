DROP TRIGGER IF EXISTS after_insert_votes_summary_trigger ON candidates;

DROP FUNCTION IF EXISTS create_votes_summary_after_insert;

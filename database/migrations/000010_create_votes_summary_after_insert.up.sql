CREATE OR REPLACE FUNCTION create_votes_summary_after_insert()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO votes_summary (candidate_id, total)
    VALUES (NEW.id, 0);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER after_insert_votes_summary_trigger
AFTER INSERT
ON candidates
FOR EACH ROW
        EXECUTE FUNCTION create_votes_summary_after_insert();

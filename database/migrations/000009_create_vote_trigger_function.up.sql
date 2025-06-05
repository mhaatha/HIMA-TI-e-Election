CREATE OR REPLACE FUNCTION count_total_votes_after_insert()
RETURNS TRIGGER AS $$
DECLARE
        candidate_count INT;
        total_all_votes INT;
        json_result TEXT;
BEGIN
        SELECT COUNT(*) INTO candidate_count
        FROM votes_summary 
        WHERE candidate_id = NEW.candidate_id;

        IF candidate_count > 0 THEN
                UPDATE votes_summary
                SET total = total + 1
                WHERE candidate_id = NEW.candidate_id;
        ELSE
                INSERT INTO votes_summary (candidate_id, total) VALUES (NEW.candidate_id, 1);
        END IF;

        SELECT SUM(vs.total) INTO total_all_votes
        FROM votes_summary vs
        JOIN candidates c ON c.id = vs.candidate_id
        WHERE EXTRACT(YEAR FROM c.created_at) = EXTRACT(YEAR FROM NOW());

        SELECT json_agg(
                json_build_object(
                        'candidate_id', vs.candidate_id,
                        'total_votes', vs.total,
                        'percentage', ROUND((vs.total::numeric / total_all_votes) * 100, 2)
                )
        )::TEXT INTO json_result
        FROM votes_summary vs
        JOIN candidates c ON c.id = vs.candidate_id
        WHERE EXTRACT(YEAR FROM c.created_at) = EXTRACT(YEAR FROM NOW());

        PERFORM pg_notify('votes_channel', json_result);

        RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER after_insert_votes_trigger
AFTER INSERT
ON votes
FOR EACH ROW
        EXECUTE FUNCTION count_total_votes_after_insert();

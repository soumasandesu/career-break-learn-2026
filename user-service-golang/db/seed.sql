-- Insert initial users
INSERT INTO users (id, name, last_seen) VALUES
    ('1', 'Alice', '2024-06-01 10:00:00'::timestamp),
    ('2', 'Bob', '2024-06-01 11:00:00'::timestamp),
    ('3', 'Charlie', '2024-06-01 12:00:00'::timestamp)
ON CONFLICT (id) DO NOTHING;

-- Insert initial user activity
INSERT INTO user_activities (feed_id, action_text_template) VALUES
    ('feed1', '{subject} commented on {object} post.')
ON CONFLICT (feed_id) DO NOTHING;

-- Insert subject referring (Alice is the subject)
INSERT INTO user_activity_subject_referring (feed_id, referring_type, referring_id, user_id) VALUES
    ('feed1', 'USER', '1', '1')
ON CONFLICT (feed_id, referring_id) DO NOTHING;

-- Insert object referring (Bob's post is the object)
INSERT INTO user_activity_object_referring (feed_id, referring_type, referring_id, user_id) VALUES
    ('feed1', 'POST', '1024', '2')
ON CONFLICT (feed_id, referring_id) DO NOTHING;

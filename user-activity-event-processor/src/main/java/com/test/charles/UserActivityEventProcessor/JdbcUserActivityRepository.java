package com.test.charles.UserActivityEventProcessor;

import com.test.charles.shared.models.UserActivity;
import com.test.charles.shared.models.UserActivityReferring;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.jdbc.core.namedparam.MapSqlParameterSource;
import org.springframework.jdbc.core.namedparam.NamedParameterJdbcTemplate;
import org.springframework.stereotype.Repository;

import java.sql.Timestamp;
import java.time.Instant;
import java.util.HashMap;
import java.util.Map;

@Repository
@RequiredArgsConstructor(onConstructor_ = @Autowired)
public class JdbcUserActivityRepository {

    private final JdbcTemplate jdbcTemplate;
    private final NamedParameterJdbcTemplate namedParameterJdbcTemplate;

    public void upsertUserActivity(final UserActivity activity) {
        // Upsert into user_activities
        final String upsertActivity = "INSERT INTO user_activities (feed_id, action_text_template, created_at) " +
                "VALUES (?, ?, ?) " +
                "ON CONFLICT (feed_id) DO UPDATE SET action_text_template = EXCLUDED.action_text_template";
        jdbcTemplate.update(upsertActivity, activity.getFeedId(), activity.getActionTextTemplate(), Timestamp.from(Instant.now()));

        // Delete existing referring rows for this feed_id and re-insert (simple approach)
        jdbcTemplate.update("DELETE FROM user_activity_subject_referring WHERE feed_id = ?", activity.getFeedId());
        jdbcTemplate.update("DELETE FROM user_activity_object_referring WHERE feed_id = ?", activity.getFeedId());

        // Insert subject referring
        final String insertSubject = "INSERT INTO user_activity_subject_referring (feed_id, referring_type, referring_id, user_id) VALUES (:feedId, :refType, :refId, :userId) ON CONFLICT (feed_id, referring_id) DO NOTHING";
        for (UserActivityReferring r : activity.getSubjectReferringList()) {
            Map<String, Object> params = new HashMap<>();
            params.put("feedId", activity.getFeedId());
            params.put("refType", r.getType().name());
            params.put("refId", r.getId());
            String userIdVal = r.getUserId().isEmpty() ? null : r.getUserId();
            params.put("userId", userIdVal);
            namedParameterJdbcTemplate.update(insertSubject, new MapSqlParameterSource(params));
        }

        // Insert object referring
        final String insertObject = "INSERT INTO user_activity_object_referring (feed_id, referring_type, referring_id, user_id) VALUES (:feedId, :refType, :refId, :userId) ON CONFLICT (feed_id, referring_id) DO NOTHING";
        for (UserActivityReferring r : activity.getObjectReferringList()) {
            Map<String, Object> params = new HashMap<>();
            params.put("feedId", activity.getFeedId());
            params.put("refType", r.getType().name());
            params.put("refId", r.getId());
            String userIdVal = r.getUserId().isEmpty() ? null : r.getUserId();
            params.put("userId", userIdVal);
            namedParameterJdbcTemplate.update(insertObject, new MapSqlParameterSource(params));
        }
    }
}

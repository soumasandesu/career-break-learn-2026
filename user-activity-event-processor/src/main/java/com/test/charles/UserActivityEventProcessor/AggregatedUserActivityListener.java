package com.test.charles.UserActivityEventProcessor;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.test.charles.shared.models.UserActivity;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor(onConstructor_ = @Autowired)
@Slf4j
public class AggregatedUserActivityListener {

    private final ObjectMapper objectMapper = new ObjectMapper();
    private final JdbcUserActivityRepository repository;

    @KafkaListener(topics = "aggregated-user-activities", groupId = "user-activity-group")
    public void listen(String message) {
        try {
            final UserActivity activity = objectMapper.readValue(message, UserActivity.class);
            log.info("Received aggregated user activity feedId={}", activity.getFeedId());
            repository.upsertUserActivity(activity);
        } catch (Exception e) {
            log.error("Failed to process aggregated activity message", e);
        }
    }
}

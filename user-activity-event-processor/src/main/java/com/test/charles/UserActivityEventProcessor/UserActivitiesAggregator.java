package com.test.charles.UserActivityEventProcessor;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.test.charles.UserActivityEventProcessor.config.JsonSerde;
import com.test.charles.shared.models.UserActivity;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.apache.kafka.common.serialization.Serdes;
import org.apache.kafka.streams.KeyValue;
import org.apache.kafka.streams.StreamsBuilder;
import org.apache.kafka.streams.kstream.*;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.time.Duration;
import java.util.List;

@Slf4j
@Service
@RequiredArgsConstructor(onConstructor_ = @Autowired)
public class UserActivitiesAggregator {

	private final ObjectMapper objectMapper = new ObjectMapper();

	@Value("${spring.kafka.streams.input-topic:user-activities}")
	private final String inputTopic;

	@Value("${spring.kafka.streams.output-topic:aggregated-user-activities}")
	private final String outputTopic;

	@Autowired
	public void buildPipeline(final StreamsBuilder streamsBuilder) {
		final KStream<String, String> sourceStream = streamsBuilder.stream(
				inputTopic,
				Consumed.with(Serdes.String(), Serdes.String())
		);

		// Parse JSON to UserActivity
		final KStream<String, UserActivity> activityStream = sourceStream
				.mapValues((_, jsonValue) -> {
					try {
						return objectMapper.readValue(jsonValue, UserActivity.class);
					} catch (Exception e) {
						log.error("Failed to parse UserActivity from JSON: {}", jsonValue, e);
						return null;
					}
				})
				.filter((_, value) -> value != null);

		// Group by a similarity key (e.g., same participants or action text)
		// Using participants as the grouping key for similarity
		final KGroupedStream<String, UserActivity> groupedStream = activityStream
				.map((_, activity) -> {
					final String similarityKey = generateSimilarityKey(activity);
					return KeyValue.pair(similarityKey, activity);
				})
				.groupByKey(Grouped.with(Serdes.String(), new JsonSerde<>(UserActivity.class)));

		// Aggregate similar messages within a time window
		final TimeWindowedKStream<String, UserActivity> windowedStream = groupedStream
				.windowedBy(TimeWindows.ofSizeWithNoGrace(Duration.ofMinutes(5)));

		// Aggregate into List, then convert to JSON string
		// Note: We use default serde for the aggregate value and handle serialization in mapValues
		final KTable<Windowed<String>, String> aggregatedTable = windowedStream
				.reduce(new Reducer<UserActivity>() {
					@Override
					public UserActivity apply(UserActivity a, UserActivity b) {
						// both object list of a & b should be same
						return UserActivity.newBuilder()
								.addAllObjectReferring(a.getObjectReferringList())
								.addAllSubjectReferring(a.getSubjectReferringList())
								.addAllSubjectReferring(b.getSubjectReferringList())
								.setActionTextTemplate(a.getActionTextTemplate())
								.build();
					}
				})
				.mapValues(this::serializeActivity);

		// Convert back to stream and output
		aggregatedTable
				.toStream()
				.map((windowedKey, activitiesJson) -> {
					final String outputKey = windowedKey.key();
					return KeyValue.pair(outputKey, activitiesJson);
				})
				.to(
						outputTopic,
						Produced.with(Serdes.String(), Serdes.String())
				);

		log.info("Kafka Streams pipeline configured: {} -> {}", inputTopic, outputTopic);
	}

	private String generateSimilarityKey(final UserActivity activity) {
		// Generate a key based on participants and action text
		// Activities with same participants and similar action text are considered similar
		final StringBuilder keyBuilder = new StringBuilder();
		
		// Sort participant IDs for consistent grouping
//		final List<String> subjectParticipantIds = activity.getSubjectReferringList()
//				.stream()
//				.map(participant -> "%s#%s".formatted(
//						participant.getType(),
//						participant.getId()
//				))
//				.sorted()
//				.toList();
//		keyBuilder.append(String.join("-", subjectParticipantIds));
//		keyBuilder.append(":");
		final List<String> objectParticipantIds = activity.getSubjectReferringList()
				.stream()
				.map(participant -> "%s#%s".formatted(
						participant.getType(),
						participant.getId()
				))
				.sorted()
				.toList();
		keyBuilder.append(String.join("-", objectParticipantIds));
		keyBuilder.append(":");
		keyBuilder.append(activity.getActionTextTemplate());
		
		return keyBuilder.toString();
	}

	private String serializeActivity(final UserActivity activities) {
		try {
			return objectMapper.writeValueAsString(activities);
		} catch (Exception e) {
			log.error("Failed to serialize activities to JSON", e);
			return "[]";
		}
	}
}

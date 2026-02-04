package com.test.charles.shared.models;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.List;
import lombok.AllArgsConstructor;
import lombok.Getter;

@Getter
@AllArgsConstructor
public class UserActivity {
	@JsonProperty("feedId")
	private final String id;

	@JsonProperty("participants")
	private final List<UserActivityParticipant> participants;

	@JsonProperty("actionText")
	private final String actionText;
}

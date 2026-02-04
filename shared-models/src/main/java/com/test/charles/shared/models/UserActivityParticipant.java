package com.test.charles.shared.models;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.AllArgsConstructor;
import lombok.Getter;

@Getter
@AllArgsConstructor
public class UserActivityParticipant {
	@JsonProperty("type")
	private final UserType type;

	@JsonProperty("id")
	private final String id;
}

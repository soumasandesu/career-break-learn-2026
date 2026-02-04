package com.test.charles.shared.models;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.AllArgsConstructor;
import lombok.Getter;

@Getter
@AllArgsConstructor
public class User {
	@JsonProperty("id")
	private final String id;

	@JsonProperty("name")
	private final String name;

	@JsonProperty("lastSeen")
	private final String lastSeen;
}

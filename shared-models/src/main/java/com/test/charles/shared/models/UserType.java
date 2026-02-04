package com.test.charles.shared.models;

public enum UserType {
	USER("user");

	private final String value;

	UserType(final String value) {
		this.value = value;
	}

	public String getValue() {
		return value;
	}
}

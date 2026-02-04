package com.test.charles.UserActivityEventProcessor.config;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.kafka.common.errors.SerializationException;
import org.apache.kafka.common.serialization.Deserializer;
import org.apache.kafka.common.serialization.Serde;
import org.apache.kafka.common.serialization.Serializer;

import java.io.IOException;
import java.util.Map;

public class JsonSerde<T> implements Serde<T> {
	private final ObjectMapper objectMapper;
	private final Class<T> clazz;

	public JsonSerde(final Class<T> clazz) {
		this.objectMapper = new ObjectMapper();
		this.clazz = clazz;
	}

	@Override
	public Serializer<T> serializer() {
		return new JsonSerializer<>(objectMapper);
	}

	@Override
	public Deserializer<T> deserializer() {
		return new JsonDeserializer<>(objectMapper, clazz);
	}

	@Override
	public void configure(final Map<String, ?> configs, final boolean isKey) {
		// No additional configuration needed
	}

	@Override
	public void close() {
		// No resources to close
	}

	private static class JsonSerializer<T> implements Serializer<T> {
		private final ObjectMapper objectMapper;

		public JsonSerializer(final ObjectMapper objectMapper) {
			this.objectMapper = objectMapper;
		}

		@Override
		public byte[] serialize(final String topic, final T data) {
			if (data == null) {
				return null;
			}
			try {
				return objectMapper.writeValueAsBytes(data);
			} catch (Exception e) {
				throw new SerializationException("Error serializing JSON message", e);
			}
		}
	}

	private static class JsonDeserializer<T> implements Deserializer<T> {
		private final ObjectMapper objectMapper;
		private final Class<T> clazz;
		private final TypeReference<T> typeReference;

		public JsonDeserializer(final ObjectMapper objectMapper, final Class<T> clazz) {
			this.objectMapper = objectMapper;
			this.clazz = clazz;
			this.typeReference = null;
		}

		@Override
		public T deserialize(final String topic, final byte[] data) {
			if (data == null) {
				return null;
			}
			try {
				if (typeReference != null) {
					return objectMapper.readValue(data, typeReference);
				}
				return objectMapper.readValue(data, clazz);
			} catch (IOException e) {
				throw new SerializationException("Error deserializing JSON message", e);
			}
		}
	}
}

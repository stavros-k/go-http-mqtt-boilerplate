package generate

import (
	"testing"
)

func TestNoopCollector(t *testing.T) {
	t.Parallel()

	collector := &NoopCollector{}

	// Test RegisterRoute
	route := &RouteInfo{
		OperationID: "test",
		Method:      "GET",
		Path:        "/test",
	}
	if err := collector.RegisterRoute(route); err != nil {
		t.Errorf("NoopCollector.RegisterRoute() should not return error, got: %v", err)
	}

	// Test RegisterMQTTPublication
	pub := &MQTTPublicationInfo{
		OperationID: "test_pub",
		Topic:       "test/topic",
	}
	if err := collector.RegisterMQTTPublication(pub); err != nil {
		t.Errorf("NoopCollector.RegisterMQTTPublication() should not return error, got: %v", err)
	}

	// Test RegisterMQTTSubscription
	sub := &MQTTSubscriptionInfo{
		OperationID: "test_sub",
		Topic:       "test/topic",
	}
	if err := collector.RegisterMQTTSubscription(sub); err != nil {
		t.Errorf("NoopCollector.RegisterMQTTSubscription() should not return error, got: %v", err)
	}

	// Test Generate
	if err := collector.Generate(); err != nil {
		t.Errorf("NoopCollector.Generate() should not return error, got: %v", err)
	}
}

func TestNoopCollectorImplementsInterfaces(t *testing.T) {
	t.Parallel()

	var _ RouteMetadataCollector = (*NoopCollector)(nil)
	var _ MQTTMetadataCollector = (*NoopCollector)(nil)
	var _ MetadataCollector = (*NoopCollector)(nil)
}

func TestTypeKindConstants(t *testing.T) {
	t.Parallel()

	expectedKinds := map[string]string{
		"TypeKindObject":     "object",
		"TypeKindStringEnum": "string_enum",
		"TypeKindNumberEnum": "number_enum",
		"TypeKindAlias":      "alias",
	}

	actualKinds := map[string]string{
		"TypeKindObject":     TypeKindObject,
		"TypeKindStringEnum": TypeKindStringEnum,
		"TypeKindNumberEnum": TypeKindNumberEnum,
		"TypeKindAlias":      TypeKindAlias,
	}

	for name, expected := range expectedKinds {
		if actual := actualKinds[name]; actual != expected {
			t.Errorf("%s = %q, want %q", name, actual, expected)
		}
	}
}

func TestFieldKindConstants(t *testing.T) {
	t.Parallel()

	expectedKinds := map[string]string{
		"FieldKindPrimitive": "primitive",
		"FieldKindArray":     "array",
		"FieldKindReference": "reference",
		"FieldKindEnum":      "enum",
		"FieldKindObject":    "object",
	}

	actualKinds := map[string]string{
		"FieldKindPrimitive": FieldKindPrimitive,
		"FieldKindArray":     FieldKindArray,
		"FieldKindReference": FieldKindReference,
		"FieldKindEnum":      FieldKindEnum,
		"FieldKindObject":    FieldKindObject,
	}

	for name, expected := range expectedKinds {
		if actual := actualKinds[name]; actual != expected {
			t.Errorf("%s = %q, want %q", name, actual, expected)
		}
	}
}

func TestNoopCollectorMultipleCalls(t *testing.T) {
	t.Parallel()

	collector := &NoopCollector{}

	// Multiple calls should all succeed
	for i := 0; i < 10; i++ {
		route := &RouteInfo{
			OperationID: "test",
			Method:      "GET",
			Path:        "/test",
		}
		if err := collector.RegisterRoute(route); err != nil {
			t.Errorf("Call %d: RegisterRoute() failed: %v", i, err)
		}
	}

	// Multiple Generate calls should succeed
	for i := 0; i < 5; i++ {
		if err := collector.Generate(); err != nil {
			t.Errorf("Call %d: Generate() failed: %v", i, err)
		}
	}
}

func TestNoopCollectorNilValues(t *testing.T) {
	t.Parallel()

	collector := &NoopCollector{}

	// Test with nil route (should not panic)
	if err := collector.RegisterRoute(nil); err != nil {
		t.Errorf("RegisterRoute(nil) should not return error, got: %v", err)
	}

	// Test with nil publication (should not panic)
	if err := collector.RegisterMQTTPublication(nil); err != nil {
		t.Errorf("RegisterMQTTPublication(nil) should not return error, got: %v", err)
	}

	// Test with nil subscription (should not panic)
	if err := collector.RegisterMQTTSubscription(nil); err != nil {
		t.Errorf("RegisterMQTTSubscription(nil) should not return error, got: %v", err)
	}
}
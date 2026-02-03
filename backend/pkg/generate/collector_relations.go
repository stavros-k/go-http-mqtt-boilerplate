package generate

// This file handles computing type relationships (References, ReferencedBy, UsedBy).

import (
	"fmt"
	"log/slog"
	"sort"
)

// computeTypeRelationships computes ReferencedBy and UsedBy for all types
// (References are already computed during type extraction).
func (g *OpenAPICollector) computeTypeRelationships() {
	g.l.Debug("Computing type relationships", slog.Int("typeCount", len(g.types)), slog.Int("operationCount", len(g.httpOps)))

	// Build ReferencedBy from References
	g.buildReferencedBy()
	g.l.Debug("Built ReferencedBy relationships")

	// Build UsedBy from routes
	g.buildUsedBy()
	g.l.Debug("Computed UsedBy relationships")
}

// buildReferencedBy builds the inverse of References for all types.
func (g *OpenAPICollector) buildReferencedBy() {
	// Track which types were modified so we only sort those
	modifiedTypes := make(map[string]struct{})

	for typeName, typeInfo := range g.types {
		for _, ref := range typeInfo.References {
			if refType, exists := g.types[ref]; exists {
				refType.ReferencedBy = append(refType.ReferencedBy, typeName)
				modifiedTypes[ref] = struct{}{}
			}
		}
	}

	// Sort only the types that received new references
	for typeName := range modifiedTypes {
		sort.Strings(g.types[typeName].ReferencedBy)
	}
}

// buildUsedBy tracks which operations use each type.
func (g *OpenAPICollector) buildUsedBy() {
	// Track HTTP operations
	for _, route := range g.httpOps {
		// Track request type
		if route.Request != nil {
			g.addUsage(route.Request.TypeName, route.OperationID, "request")
		}

		// Track response types
		for _, resp := range route.Responses {
			g.addUsage(resp.TypeName, route.OperationID, "response")
		}

		// Track parameter types
		for _, param := range route.Parameters {
			g.addUsage(param.TypeName, route.OperationID, "parameter")
		}
	}

	// Track MQTT publications
	for _, pub := range g.mqttPublications {
		g.addUsage(pub.TypeName, pub.OperationID, "mqtt_publication")
	}

	// Track MQTT subscriptions
	for _, sub := range g.mqttSubscriptions {
		g.addUsage(sub.TypeName, sub.OperationID, "mqtt_subscription")
	}

	// Deduplicate UsedBy entries
	for _, typ := range g.types {
		usages := make(map[string]struct{})
		dedupedUsages := []UsageInfo{}

		for _, usage := range typ.UsedBy {
			key := fmt.Sprintf("%s:%s", usage.OperationID, usage.Role)
			if _, used := usages[key]; used {
				continue
			}

			dedupedUsages = append(dedupedUsages, usage)
			usages[key] = struct{}{}
		}

		if len(dedupedUsages) == 0 {
			dedupedUsages = nil
		}

		typ.UsedBy = dedupedUsages
	}
}

// addUsage adds a UsageInfo entry to a type if it exists.
func (g *OpenAPICollector) addUsage(typeName, operationID, role string) {
	if typeName == "" {
		return
	}

	if typeInfo, exists := g.types[typeName]; exists {
		typeInfo.UsedBy = append(typeInfo.UsedBy, UsageInfo{
			OperationID: operationID,
			Role:        role,
		})
	}
}

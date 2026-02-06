import docsData from "../../../../api_docs.json";

// FieldType represents the structured type information for a field
export type FieldType = {
    kind: "primitive" | "array" | "reference" | "enum" | "object" | "map" | "unknown";
    type: string;
    format: string;
    required: boolean;
    nullable: boolean;
    itemsType?: FieldType;
    additionalProperties?: FieldType;
    mapKeyType?: FieldType;
};

// FieldInfo describes a field in a struct
export type FieldInfo = {
    name: string;
    displayType: string;
    typeInfo: FieldType;
    description: string;
    deprecated: string;
};

// EnumValue represents an enum constant with its documentation
export type EnumValue = {
    value: string | number;
    description: string;
    deprecated: string;
};

// UsageInfo tracks where a type is used in operations/routes
export type UsageInfo = {
    operationID: string;
    role: "request" | "response" | "parameter" | "mqtt_publication" | "mqtt_subscription" | "payload";
};

// Representations contains various format representations of a type
export type Representations = {
    json: string;
    jsonSchema: string;
    yamlSchema: string;
    go: string;
    ts: string;
};

// TypeInfo contains comprehensive metadata about a Go type
export type TypeInfo = {
    name: string;
    kind: "object" | "alias" | "enum";
    description: string;
    deprecated: string;
    fields?: FieldInfo[];
    enumValues?: EnumValue[];
    references: string[];
    referencedBy?: string[];
    usedBy?: UsageInfo[];
    representations: Representations;
    usedByHTTP: boolean;
    usedByMQTT: boolean;
    underlyingType?: FieldType;
};

// RequestInfo describes a request body
export type RequestInfo = {
    type: string;
    description: string;
    examples?: Record<string, string>;
};

// ParameterInfo describes a route parameter
export type ParameterInfo = {
    name: string;
    in: "path" | "query" | "header";
    type: string;
    description: string;
    required: boolean;
    deprecated: string;
};

// ResponseInfo describes a response
export type ResponseInfo = {
    statusCode: number;
    type: string;
    description: string;
    examples?: Record<string, string>;
};

// HTTP method union type
export type HTTPMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";

// RouteInfo contains metadata about a REST route
export type RouteInfo = {
    operationID: string;
    method: HTTPMethod;
    path: string;
    summary: string;
    description: string;
    group: string;
    deprecated: string;
    request?: RequestInfo;
    parameters?: ParameterInfo[];
    responses: Record<number, ResponseInfo>;
};

// MQTTTopicParameter describes a parameter in an MQTT topic pattern
export type MQTTTopicParameter = {
    name: string;
    description: string;
    type: string;
};

// MQTTPublicationInfo contains metadata about an MQTT publication
export type MQTTPublicationInfo = {
    operationID: string;
    topic: string;
    topicMQTT: string;
    topicParameters?: MQTTTopicParameter[];
    summary: string;
    description: string;
    group: string;
    deprecated: string;
    qos: number;
    retained: boolean;
    type: string;
    examples?: Record<string, string>;
};

// MQTTSubscriptionInfo contains metadata about an MQTT subscription
export type MQTTSubscriptionInfo = {
    operationID: string;
    topic: string;
    topicMQTT: string;
    topicParameters?: MQTTTopicParameter[];
    summary: string;
    description: string;
    group: string;
    deprecated: string;
    qos: number;
    type: string;
    examples?: Record<string, string>;
};

// ServerInfo contains server information
export type ServerInfo = {
    url: string;
    description: string;
};

// APIInfo contains API metadata
export type APIInfo = {
    title: string;
    version: string;
    description: string;
    servers: ServerInfo[];
};

// Database stats types
export type Column = {
    name: string;
    type: string;
    notNull: boolean;
    default?: string;
    primaryKey: boolean;
};

export type ForeignKey = {
    from: string;
    table: string;
    to: string;
};

export type Index = {
    name: string;
    unique: boolean;
    columns: string[];
};

export type Table = {
    name: string;
    columns: Column[];
    foreignKeys: ForeignKey[];
    indexes: Index[];
};

export type DatabaseStats = {
    tables: Table[];
};

export type Database = {
    dialect: string;
    schema: string;
    stats: DatabaseStats;
};

// APIDocumentation is the complete API documentation structure
export type APIDocumentation = {
    info: APIInfo;
    types: Record<string, TypeInfo>;
    httpOperations: Record<string, RouteInfo>;
    mqttPublications: Record<string, MQTTPublicationInfo>;
    mqttSubscriptions: Record<string, MQTTSubscriptionInfo>;
    database: Database;
    openapiSpec: string;
};

// Type aliases for convenience
export type TypeData = TypeInfo;
export type OperationData = RouteInfo;
export type MQTTPublicationData = MQTTPublicationInfo;
export type MQTTSubscriptionData = MQTTSubscriptionInfo;
export type FieldMetadata = FieldInfo;
export type Response = ResponseInfo;
export type UsedByItem = UsageInfo;

// Useful union types and keyof types
export type ItemType = "type" | "operation" | "mqtt-publication" | "mqtt-subscription";
export type TypeKeys = keyof APIDocumentation["types"];
export type OperationID = keyof APIDocumentation["httpOperations"];
export type MQTTPublicationID = keyof APIDocumentation["mqttPublications"];
export type MQTTSubscriptionID = keyof APIDocumentation["mqttSubscriptions"];

// Docs is an alias for APIDocumentation (for backwards compatibility)
export type Docs = APIDocumentation;

// Helper functions
export function getTypeJson(typeName: string | "null"): string | null {
    if (typeName === "null") return null;
    const type = docs.types[typeName];
    if (!type?.representations) return null;

    return type.representations.json;
}

export function getAllOperations(): OperationData[] {
    return Object.values(docs.httpOperations);
}

export function getAllMQTTPublications(): MQTTPublicationData[] {
    return Object.values(docs.mqttPublications);
}

export function getAllMQTTSubscriptions(): MQTTSubscriptionData[] {
    return Object.values(docs.mqttSubscriptions);
}

// Cast the imported JSON to the proper type
export const docs: APIDocumentation = docsData as unknown as APIDocumentation;

import docsData from "../../../../api_docs.json";

export type Docs = typeof docsData;

type HTTPOperations = Docs["httpOperations"];
type MQTTPublications = Docs["mqttPublications"];
type MQTTSubscriptions = Docs["mqttSubscriptions"];
type ApiDataTypes = Docs["types"];

export type OperationID = keyof HTTPOperations;
export type OperationData = HTTPOperations[OperationID];
export type Response = OperationData["responses"][keyof OperationData["responses"]];

export type MQTTPublicationID = keyof MQTTPublications;
export type MQTTPublicationData = MQTTPublications[MQTTPublicationID];

export type MQTTSubscriptionID = keyof MQTTSubscriptions;
export type MQTTSubscriptionData = MQTTSubscriptions[MQTTSubscriptionID];

// Get union of all HTTP methods across all operations
export type HTTPMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";

export type TypeKeys = keyof ApiDataTypes;
export type TypeData = ApiDataTypes[TypeKeys];

export type FieldMetadata = NonNullable<TypeData["fields"]>[number];
export type UsedByItem = NonNullable<TypeData["usedBy"]>[number];
export type EnumValue = NonNullable<TypeData["enumValues"]>[number];

export type ItemType = "type" | "operation" | "mqtt-publication" | "mqtt-subscription";

export function getTypeJson(typeName: TypeKeys | "null") {
    if (typeName === "null") return null;
    const type = docs.types[typeName];
    const representations = type?.representations;
    if (!representations) return null;

    // Use jsonSchema if json is empty, otherwise use json
    const jsonRep = type.representations.json;
    if (jsonRep && jsonRep.trim() !== "") {
        return jsonRep;
    }
    // Fallback to jsonSchema if available
    const jsonSchema = type.representations.jsonSchema;
    if (jsonSchema && jsonSchema.trim() !== "") {
        return jsonSchema;
    }
    return null;
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

export const docs: Docs = docsData;

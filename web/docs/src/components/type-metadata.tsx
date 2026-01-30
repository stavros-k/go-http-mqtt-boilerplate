import type { Route } from "next";
import Link from "next/link";
import { BiLinkExternal } from "react-icons/bi";
import type { EnumValue, FieldMetadata, TypeData, UsedByItem } from "@/data/api";
import { docs, getAllMQTTPublications, getAllMQTTSubscriptions, getAllOperations } from "@/data/api";
import { RoutePath } from "./route-path";
import { VerbBadge } from "./verb-badge";

interface TypeMetadataProps {
    data: TypeData;
    typeName: string;
}

// Helper function to check if a type name exists in the docs
const isTypeLink = (type: string) => {
    return type in docs.types;
};

function TypeDescription({ description }: { description?: string }) {
    if (!description) return null;

    return (
        <div>
            <h2 className='mb-3 font-semibold text-text-primary text-xl'>Description</h2>
            <p className='text-text-secondary'>{description}</p>
        </div>
    );
}

function UsedBySection({ usedBy }: { usedBy: UsedByItem[] | null }) {
    if (!usedBy || usedBy.length === 0) return null;

    const allOperations = getAllOperations();
    const allPublications = getAllMQTTPublications();
    const allSubscriptions = getAllMQTTSubscriptions();

    // Group by operationID and collect roles
    const grouped = usedBy.reduce(
        (acc, item) => {
            if (!acc[item.operationID]) {
                acc[item.operationID] = new Set();
            }
            acc[item.operationID].add(item.role);
            return acc;
        },
        {} as Record<string, Set<string>>
    );

    return (
        <div>
            <h2 className='mb-4 font-semibold text-text-primary text-xl'>Used By Operations</h2>
            <p className='mb-4 text-sm text-text-tertiary'>This type is used by the following operations:</p>
            <div className='grid grid-cols-1 gap-3'>
                {Object.entries(grouped).map(([operationID, roles]) => {
                    // Check roles to determine operation type
                    const isHTTP = Array.from(roles).some((role) =>
                        ["request", "response", "parameter"].includes(role)
                    );
                    const isMQTTPublication = roles.has("mqtt_publication");
                    const isMQTTSubscription = roles.has("mqtt_subscription");

                    // HTTP Operation
                    if (isHTTP) {
                        const operation = allOperations.find((op) => op.operationID === operationID);
                        return (
                            <Link
                                key={operationID}
                                href={`/api/operation/${operationID}` as Route}
                                className='block rounded-lg border-2 border-border-primary bg-bg-tertiary p-4 transition-all duration-200 hover:border-accent-green-light hover:shadow-md'>
                                <div className='flex flex-col gap-2'>
                                    <div className='flex items-center justify-between gap-3'>
                                        <div className='flex min-w-0 items-center gap-2'>
                                            {operation && (
                                                <>
                                                    <VerbBadge
                                                        verb={operation.method}
                                                        size='xs'
                                                    />
                                                    <span className='truncate font-mono font-semibold text-sm'>
                                                        <RoutePath path={operation.path} />
                                                    </span>
                                                </>
                                            )}
                                        </div>
                                        <div className='flex shrink-0 flex-wrap gap-1'>
                                            {Array.from(roles).map((role) => (
                                                <span
                                                    key={role}
                                                    className={`rounded px-2 py-0.5 font-medium text-xs ${
                                                        role === "request"
                                                            ? "border border-blue-500/30 bg-blue-500/20 text-blue-400"
                                                            : "border border-green-500/30 bg-green-500/20 text-green-400"
                                                    }`}>
                                                    {role}
                                                </span>
                                            ))}
                                        </div>
                                    </div>
                                    <code className='font-mono text-text-muted text-xs'>{operationID}</code>
                                </div>
                            </Link>
                        );
                    }

                    // MQTT Publication
                    if (isMQTTPublication) {
                        const publication = allPublications.find((pub) => pub.operationID === operationID);
                        return (
                            <Link
                                key={operationID}
                                href={`/api/mqtt/publication/${operationID}` as Route}
                                className='block rounded-lg border-2 border-border-primary bg-bg-tertiary p-4 transition-all duration-200 hover:border-accent-blue-light hover:shadow-md'>
                                <div className='flex flex-col gap-2'>
                                    <div className='flex items-center justify-between gap-3'>
                                        <div className='flex min-w-0 items-center gap-2'>
                                            {publication && (
                                                <>
                                                    <span className='rounded border border-accent-blue-border bg-accent-blue-bg px-2 py-0.5 font-bold text-accent-blue-text text-xs'>
                                                        PUB
                                                    </span>
                                                    <span className='truncate font-mono font-semibold text-sm'>
                                                        <RoutePath path={publication.topic} />
                                                    </span>
                                                </>
                                            )}
                                        </div>
                                        <div className='flex shrink-0 flex-wrap gap-1'>
                                            {Array.from(roles).map((role) => (
                                                <span
                                                    key={role}
                                                    className='rounded border border-blue-500/30 bg-blue-500/20 px-2 py-0.5 font-medium text-blue-400 text-xs'>
                                                    {role}
                                                </span>
                                            ))}
                                        </div>
                                    </div>
                                    <code className='font-mono text-text-muted text-xs'>{operationID}</code>
                                </div>
                            </Link>
                        );
                    }

                    // MQTT Subscription
                    if (isMQTTSubscription) {
                        const subscription = allSubscriptions.find((sub) => sub.operationID === operationID);
                        return (
                            <Link
                                key={operationID}
                                href={`/api/mqtt/subscription/${operationID}` as Route}
                                className='block rounded-lg border-2 border-border-primary bg-bg-tertiary p-4 transition-all duration-200 hover:border-accent-green-light hover:shadow-md'>
                                <div className='flex flex-col gap-2'>
                                    <div className='flex items-center justify-between gap-3'>
                                        <div className='flex min-w-0 items-center gap-2'>
                                            {subscription && (
                                                <>
                                                    <span className='rounded border border-accent-green-border bg-accent-green-bg px-2 py-0.5 font-bold text-accent-green-text text-xs'>
                                                        SUB
                                                    </span>
                                                    <span className='truncate font-mono font-semibold text-sm'>
                                                        <RoutePath path={subscription.topic} />
                                                    </span>
                                                </>
                                            )}
                                        </div>
                                        <div className='flex shrink-0 flex-wrap gap-1'>
                                            {Array.from(roles).map((role) => (
                                                <span
                                                    key={role}
                                                    className='rounded border border-green-500/30 bg-green-500/20 px-2 py-0.5 font-medium text-green-400 text-xs'>
                                                    {role}
                                                </span>
                                            ))}
                                        </div>
                                    </div>
                                    <code className='font-mono text-text-muted text-xs'>{operationID}</code>
                                </div>
                            </Link>
                        );
                    }

                    return null;
                })}
            </div>
        </div>
    );
}

function EnumValuesSection({ enumValues }: { enumValues: EnumValue[] | null }) {
    if (!enumValues || enumValues.length === 0) return null;

    return (
        <div>
            <h2 className='mb-4 font-semibold text-text-primary text-xl'>Possible Values</h2>
            <div className='flex flex-col gap-4'>
                {enumValues.map((enumValue) => (
                    <div
                        key={enumValue.value}
                        className='rounded-lg border-2 border-border-primary bg-bg-secondary p-4 transition-all duration-200 hover:border-accent-blue/50'>
                        <div className='mb-2 flex items-center gap-2'>
                            <code className='rounded-lg border-2 border-type-enum-border bg-type-enum-bg px-3 py-1.5 font-mono font-semibold text-sm text-type-enum'>
                                &quot;{enumValue.value}&quot;
                            </code>
                            {enumValue.deprecated && (
                                <span className='rounded border border-warning-border bg-warning-bg px-2 py-0.5 font-semibold text-warning-text text-xs'>
                                    DEPRECATED
                                </span>
                            )}
                        </div>
                        {enumValue.description && <p className='text-sm text-text-tertiary'>{enumValue.description}</p>}
                        {enumValue.deprecated && (
                            <p className='mt-2 text-sm text-warning-text italic'>{enumValue.deprecated}</p>
                        )}
                    </div>
                ))}
            </div>
        </div>
    );
}

function FieldItem({ field }: { field: FieldMetadata }) {
    const displayType = "displayType" in field ? field.displayType : "";
    // Use typeInfo.type for actual type reference, fallback to displayType
    const actualType = "typeInfo" in field && field.typeInfo ? field.typeInfo.type : displayType;
    const isClickableType = isTypeLink(actualType);

    // Check if this field has additionalProperties (map/dictionary type)
    const additionalProps = field.typeInfo?.additionalProperties;

    return (
        <div className='rounded-lg border-2 border-border-primary bg-bg-secondary p-4 transition-all duration-200 hover:border-accent-blue/50'>
            <div className='mb-2 flex items-start justify-between gap-4'>
                <div className='flex flex-wrap items-center gap-2'>
                    <code className='font-semibold text-lg text-text-primary'>{field.name}</code>
                    {"typeInfo" in field && field.typeInfo && field.typeInfo.required && (
                        <span className='rounded border border-red-500/30 bg-red-500/20 px-2 py-0.5 font-semibold text-red-400 text-xs'>
                            required
                        </span>
                    )}
                    {"typeInfo" in field && field.typeInfo && !field.typeInfo.required && (
                        <span className='rounded border border-yellow-500/30 bg-yellow-500/20 px-2 py-0.5 text-xs text-yellow-400'>
                            optional
                        </span>
                    )}
                    {"typeInfo" in field && field.typeInfo && field.typeInfo.nullable && (
                        <span className='rounded border border-blue-500/30 bg-blue-500/20 px-2 py-0.5 text-blue-400 text-xs'>
                            nullable
                        </span>
                    )}
                </div>
                <div className='shrink-0'>
                    {isClickableType ? (
                        <Link
                            href={`/api/type/${actualType}`}
                            className='inline-flex items-center gap-1 rounded-lg border-2 border-type-reference/30 bg-type-reference/10 px-3 py-1.5 font-mono font-semibold text-sm text-type-reference transition-all duration-200 hover:scale-105 hover:border-type-reference hover:bg-type-reference/20'>
                            {displayType}
                            <BiLinkExternal className='h-3.5 w-3.5' />
                        </Link>
                    ) : (
                        <code className='rounded-lg border-2 border-type-primitive/30 bg-type-primitive/10 px-3 py-1.5 font-mono font-semibold text-sm text-type-primitive'>
                            {displayType}
                        </code>
                    )}
                </div>
            </div>

            {"description" in field && field.description && (
                <p className='mt-2 text-sm text-text-tertiary'>{String(field.description)}</p>
            )}

            {additionalProps && (
                <div className='mt-3 rounded-lg border border-border-primary bg-bg-tertiary p-3'>
                    <div className='mb-1 flex items-center gap-2'>
                        <span className='font-semibold text-text-secondary text-xs uppercase tracking-wide'>
                            Map/Dictionary Type
                        </span>
                    </div>
                    <div className='flex items-center gap-2 text-sm'>
                        <span className='text-text-tertiary'>Values:</span>
                        {isTypeLink(additionalProps.type) ? (
                            <Link
                                href={`/api/type/${additionalProps.type}`}
                                className='inline-flex items-center gap-1 rounded border border-type-reference/30 bg-type-reference/10 px-2 py-0.5 font-mono font-semibold text-type-reference text-xs transition-all duration-200 hover:border-type-reference hover:bg-type-reference/20'>
                                {additionalProps.type}
                                <BiLinkExternal className='h-3 w-3' />
                            </Link>
                        ) : (
                            <code className='rounded border border-type-primitive/30 bg-type-primitive/10 px-2 py-0.5 font-mono font-semibold text-type-primitive text-xs'>
                                {additionalProps.type}
                            </code>
                        )}
                    </div>
                </div>
            )}
        </div>
    );
}

function FieldsSection({ fields }: { fields: FieldMetadata[] | undefined }) {
    if (!fields || fields.length === 0) return null;

    return (
        <div>
            <h2 className='mb-4 font-semibold text-text-primary text-xl'>Fields</h2>
            <div className='space-y-4'>
                {fields.map((field) => (
                    <FieldItem
                        key={field.name}
                        field={field}
                    />
                ))}
            </div>
        </div>
    );
}

function ReferencedBySection({ referencedBy, typeName }: { referencedBy: string[] | undefined; typeName: string }) {
    const hasReferencedBy = referencedBy && referencedBy.length > 0;

    return (
        <div>
            <h2 className='mb-4 font-semibold text-text-primary text-xl'>Referenced By</h2>
            {hasReferencedBy ? (
                <>
                    <p className='mb-3 text-sm text-text-tertiary'>Types that use {typeName}:</p>
                    <div className='flex flex-wrap gap-2'>
                        {referencedBy.map((ref: string) => (
                            <Link
                                key={ref}
                                href={`/api/type/${ref}`}
                                className='rounded-md border border-type-enum-border bg-type-enum-bg px-3 py-1.5 font-mono text-sm text-type-enum transition-colors hover:brightness-110'>
                                {ref}
                            </Link>
                        ))}
                    </div>
                </>
            ) : (
                <p className='text-sm text-text-tertiary'>No types reference this type.</p>
            )}
        </div>
    );
}

export function TypeMetadata({ data, typeName }: TypeMetadataProps) {
    const fields = "fields" in data ? data.fields : undefined;
    const isPrimitiveOrEnum = !fields || fields.length === 0;

    const enumValues =
        "enumValues" in data && Array.isArray(data.enumValues) && data.enumValues.length > 0 ? data.enumValues : null;
    const usedBy =
        "usedBy" in data && Array.isArray(data.usedBy) && data.usedBy.length > 0 ? (data.usedBy as UsedByItem[]) : null;
    const referencedBy = "referencedBy" in data ? data.referencedBy : undefined;

    return (
        <div className='space-y-6'>
            <TypeDescription description={data.description} />

            {isPrimitiveOrEnum && (
                <>
                    <div className='border-border-primary border-t' />
                    <EnumValuesSection enumValues={enumValues} />
                </>
            )}

            {fields && fields.length > 0 && (
                <>
                    <div className='border-border-primary border-t' />
                    <FieldsSection fields={fields} />
                </>
            )}

            {usedBy && usedBy.length > 0 && (
                <>
                    <div className='border-border-primary border-t' />
                    <UsedBySection usedBy={usedBy} />
                </>
            )}

            {referencedBy && referencedBy.length > 0 && (
                <>
                    <div className='border-border-primary border-t' />
                    <ReferencedBySection
                        referencedBy={referencedBy}
                        typeName={typeName}
                    />
                </>
            )}
        </div>
    );
}

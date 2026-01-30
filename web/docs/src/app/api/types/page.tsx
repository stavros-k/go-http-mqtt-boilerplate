"use client";

import type { Route } from "next";
import { useMemo, useState } from "react";
import { Breadcrumbs } from "@/components/breadcrumbs";
import { CollapsibleGroup } from "@/components/collapsible-group";
import { EmptyState } from "@/components/empty-state";
import { ItemCard } from "@/components/item-card";
import { PageHeader } from "@/components/page-header";
import { StatCard } from "@/components/stat-card";
import { type TypeFilterState, TypeFilters } from "@/components/type-filters";
import { getKindDisplayName, TypeKindBadge } from "@/components/type-kind-badge";
import { docs } from "@/data/api";

export default function TypesPage() {
    const allTypes = Object.entries(docs.types);
    const [filters, setFilters] = useState<TypeFilterState>({
        kind: "all",
        hideDeprecated: false,
        search: "",
    });

    // Extract unique kinds
    const kinds = useMemo(() => {
        const uniqueKinds = new Set(allTypes.map(([_, type]) => type.kind));
        return Array.from(uniqueKinds).sort();
    }, [allTypes]);

    // Filter types
    const filteredTypes = useMemo(() => {
        return allTypes.filter(([key, type]) => {
            // Kind filter
            if (filters.kind !== "all" && type.kind !== filters.kind) {
                return false;
            }

            // Deprecated filter
            if (filters.hideDeprecated && type.deprecated) {
                return false;
            }

            // Search filter
            if (filters.search) {
                const searchLower = filters.search.toLowerCase();
                const matchesName = key.toLowerCase().includes(searchLower);
                const matchesDescription = type.description?.toLowerCase().includes(searchLower);

                if (!matchesName && !matchesDescription) {
                    return false;
                }
            }

            return true;
        });
    }, [allTypes, filters]);

    // Group types by kind
    const groupedTypes = useMemo(() => {
        const grouped = filteredTypes.reduce(
            (acc, [key, type]) => {
                const kind = getKindDisplayName(type.kind);
                if (!acc[kind]) {
                    acc[kind] = [];
                }
                acc[kind].push([key, type] as const);
                return acc;
            },
            {} as Record<string, typeof filteredTypes>
        );
        return Object.entries(grouped).sort(([a], [b]) => a.localeCompare(b));
    }, [filteredTypes]);

    // Calculate statistics
    const deprecatedCount = useMemo(() => {
        return allTypes.filter(([_, type]) => type.deprecated).length;
    }, [allTypes]);

    const kindCounts = useMemo(() => {
        return allTypes.reduce(
            (acc, [_, type]) => {
                acc[type.kind] = (acc[type.kind] || 0) + 1;
                return acc;
            },
            {} as Record<string, number>
        );
    }, [allTypes]);

    return (
        <div className='flex-1 overflow-y-auto p-10'>
            <Breadcrumbs items={[{ label: "Types" }]} />

            <PageHeader
                title='Types'
                description='Browse all type definitions used in the API'
            />

            {/* Statistics */}
            <div className='mb-8 grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5'>
                <StatCard
                    label='Total'
                    value={allTypes.length}
                    color='blue'
                />
                {Object.entries(kindCounts)
                    .sort(([a], [b]) => a.localeCompare(b))
                    .map(([kind, count]) => (
                        <StatCard
                            key={kind}
                            label={getKindDisplayName(kind)}
                            value={count}
                            color='purple'
                        />
                    ))}
                {deprecatedCount > 0 && (
                    <StatCard
                        label='Deprecated'
                        value={deprecatedCount}
                        color='yellow'
                    />
                )}
            </div>

            <TypeFilters
                kinds={kinds}
                onFilterChange={setFilters}
            />

            {filteredTypes.length === 0 ? (
                <EmptyState
                    title='No types found'
                    description="Try adjusting your filters or search query to find what you're looking for."
                    icon='🔍'
                />
            ) : (
                <div>
                    {groupedTypes.map(([kindName, types]) => (
                        <CollapsibleGroup
                            key={kindName}
                            title={kindName}
                            defaultOpen={true}>
                            <div className='grid gap-5'>
                                {types.map(([key, type]) => {
                                    const fields = "fields" in type ? type.fields : undefined;
                                    const fieldCount = fields?.length || 0;
                                    const enumValues = "enumValues" in type ? type.enumValues : undefined;
                                    const enumCount = enumValues?.length || 0;
                                    const isDeprecated = !!type.deprecated;

                                    return (
                                        <ItemCard
                                            key={key}
                                            href={`/api/type/${key}` as Route}
                                            title={key}
                                            description={type.description}
                                            badges={
                                                <div className='flex max-w-50 flex-wrap items-start justify-end gap-2'>
                                                    <TypeKindBadge
                                                        kind={type.kind}
                                                        size='sm'
                                                    />
                                                    {fieldCount > 0 && (
                                                        <span className='whitespace-nowrap rounded-lg border border-blue-500/30 bg-blue-500/20 px-2 py-1 font-semibold text-blue-400 text-xs'>
                                                            {fieldCount} {fieldCount === 1 ? "field" : "fields"}
                                                        </span>
                                                    )}
                                                    {enumCount > 0 && (
                                                        <span className='whitespace-nowrap rounded-lg border border-purple-500/30 bg-purple-500/20 px-2 py-1 font-semibold text-purple-400 text-xs'>
                                                            {enumCount} {enumCount === 1 ? "value" : "values"}
                                                        </span>
                                                    )}
                                                </div>
                                            }
                                            deprecated={isDeprecated}
                                            hoverBorderColor='hover:border-accent-blue-light'
                                        />
                                    );
                                })}
                            </div>
                        </CollapsibleGroup>
                    ))}
                </div>
            )}
        </div>
    );
}

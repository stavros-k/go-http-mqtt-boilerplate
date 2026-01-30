"use client";

import type { Route } from "next";
import { useMemo, useState } from "react";
import { Breadcrumbs } from "@/components/breadcrumbs";
import { CollapsibleGroup } from "@/components/collapsible-group";
import { EmptyState } from "@/components/empty-state";
import { ItemCard } from "@/components/item-card";
import { type OperationFilterState, OperationFilters } from "@/components/operation-filters";
import { PageHeader } from "@/components/page-header";
import { RoutePath } from "@/components/route-path";
import { StatCard } from "@/components/stat-card";
import { VerbBadge } from "@/components/verb-badge";
import { getAllOperations } from "@/data/api";

export default function OperationsPage() {
    const allOperations = getAllOperations();
    const [filters, setFilters] = useState<OperationFilterState>({
        verb: "all",
        group: "all",
        hideDeprecated: false,
        search: "",
    });

    // Extract unique methods and groups
    const methods = useMemo(() => {
        const uniqueMethods = new Set(allOperations.map((op) => op.method.toUpperCase()));
        return Array.from(uniqueMethods).sort();
    }, [allOperations]);

    const groups = useMemo(() => {
        const uniqueGroups = new Set(allOperations.map((op) => op.group).filter(Boolean));
        return Array.from(uniqueGroups).sort() as string[];
    }, [allOperations]);

    // Filter operations
    const filteredOperations = useMemo(() => {
        return allOperations.filter((operation) => {
            // Method filter
            if (filters.verb !== "all" && operation.method.toUpperCase() !== filters.verb) {
                return false;
            }

            // Group filter
            if (filters.group !== "all" && operation.group !== filters.group) {
                return false;
            }

            // Deprecated filter
            if (filters.hideDeprecated && operation.deprecated) {
                return false;
            }

            // Search filter
            if (filters.search) {
                const searchLower = filters.search.toLowerCase();
                const matchesName = operation.operationID.toLowerCase().includes(searchLower);
                const matchesPath = operation.path.toLowerCase().includes(searchLower);
                const matchesSummary = operation.summary?.toLowerCase().includes(searchLower);
                const matchesDescription = operation.description?.toLowerCase().includes(searchLower);

                if (!matchesName && !matchesPath && !matchesSummary && !matchesDescription) {
                    return false;
                }
            }

            return true;
        });
    }, [allOperations, filters]);

    // Group operations by group
    const groupedOperations = useMemo(() => {
        const grouped = filteredOperations.reduce(
            (acc, operation) => {
                const group = operation.group || "Ungrouped";
                if (!acc[group]) {
                    acc[group] = [];
                }
                acc[group].push(operation);
                return acc;
            },
            {} as Record<string, typeof filteredOperations>
        );
        return Object.entries(grouped).sort(([a], [b]) => {
            // "Ungrouped" always last
            if (a === "Ungrouped") return 1;
            if (b === "Ungrouped") return -1;
            return a.localeCompare(b);
        });
    }, [filteredOperations]);

    // Calculate statistics
    const deprecatedCount = useMemo(() => {
        return allOperations.filter((op) => op.deprecated).length;
    }, [allOperations]);

    const methodCounts = useMemo(() => {
        return allOperations.reduce(
            (acc, op) => {
                const method = op.method.toUpperCase();
                acc[method] = (acc[method] || 0) + 1;
                return acc;
            },
            {} as Record<string, number>
        );
    }, [allOperations]);

    return (
        <div className='flex-1 overflow-y-auto p-10'>
            <Breadcrumbs items={[{ label: "Operations" }]} />

            <PageHeader
                title='Operations'
                description='Browse all API operations and endpoints'
            />

            {/* Statistics */}
            <div className='mb-8 grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6'>
                <StatCard
                    label='Total'
                    value={allOperations.length}
                    color='blue'
                />
                {Object.entries(methodCounts)
                    .sort(([a], [b]) => a.localeCompare(b))
                    .map(([method, count]) => {
                        const colorMap: Record<string, "blue" | "green" | "yellow" | "red" | "purple"> = {
                            GET: "blue",
                            POST: "green",
                            PUT: "yellow",
                            DELETE: "red",
                            PATCH: "purple",
                        };
                        return (
                            <StatCard
                                key={method}
                                label={method}
                                value={count}
                                color={colorMap[method] || "blue"}
                            />
                        );
                    })}
                {deprecatedCount > 0 && (
                    <StatCard
                        label='Deprecated'
                        value={deprecatedCount}
                        color='yellow'
                    />
                )}
            </div>

            <OperationFilters
                verbs={methods}
                groups={groups}
                onFilterChange={setFilters}
            />

            {filteredOperations.length === 0 ? (
                <EmptyState
                    title='No operations found'
                    description="Try adjusting your filters or search query to find what you're looking for."
                    icon='🔍'
                />
            ) : (
                <div>
                    {groupedOperations.map(([groupName, operations]) => (
                        <CollapsibleGroup
                            key={groupName}
                            title={groupName}
                            defaultOpen={true}>
                            <div className='grid gap-5'>
                                {operations.map((operation) => (
                                    <ItemCard
                                        key={operation.operationID}
                                        href={`/api/operation/${operation.operationID}` as Route}
                                        title={operation.operationID}
                                        subtitle={
                                            <div className='flex items-center gap-2'>
                                                <VerbBadge
                                                    verb={operation.method}
                                                    size='sm'
                                                />
                                                <RoutePath
                                                    path={operation.path}
                                                    className='font-mono'
                                                />
                                            </div>
                                        }
                                        description={operation.summary || operation.description}
                                        deprecated={!!operation.deprecated}
                                        hoverBorderColor='hover:border-accent-green-light'
                                    />
                                ))}
                            </div>
                        </CollapsibleGroup>
                    ))}
                </div>
            )}
        </div>
    );
}

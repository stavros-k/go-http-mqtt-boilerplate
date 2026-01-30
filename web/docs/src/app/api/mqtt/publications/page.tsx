"use client";

import type { Route } from "next";
import { useMemo, useState } from "react";
import { Breadcrumbs } from "@/components/breadcrumbs";
import { CollapsibleGroup } from "@/components/collapsible-group";
import { EmptyState } from "@/components/empty-state";
import { ItemCard } from "@/components/item-card";
import { type MQTTFilterState, MQTTFilters } from "@/components/mqtt-filters";
import { PageHeader } from "@/components/page-header";
import { PubBadge } from "@/components/pub-badge";
import { RoutePath } from "@/components/route-path";
import { StatCard } from "@/components/stat-card";
import { getAllMQTTPublications } from "@/data/api";

export default function MQTTPublicationsPage() {
    const allPublications = getAllMQTTPublications();
    const [filters, setFilters] = useState<MQTTFilterState>({
        group: "all",
        hideDeprecated: false,
        search: "",
    });

    // Extract unique groups
    const groups = useMemo(() => {
        const uniqueGroups = new Set(allPublications.map((pub) => pub.group).filter(Boolean));
        return Array.from(uniqueGroups).sort() as string[];
    }, [allPublications]);

    // Filter publications
    const filteredPublications = useMemo(() => {
        return allPublications.filter((publication) => {
            // Group filter
            if (filters.group !== "all" && publication.group !== filters.group) {
                return false;
            }

            // Deprecated filter
            if (filters.hideDeprecated && publication.deprecated) {
                return false;
            }

            // Search filter
            if (filters.search) {
                const searchLower = filters.search.toLowerCase();
                const matchesName = publication.operationID.toLowerCase().includes(searchLower);
                const matchesTopic = publication.topic.toLowerCase().includes(searchLower);
                const matchesSummary = publication.summary?.toLowerCase().includes(searchLower);
                const matchesDescription = publication.description?.toLowerCase().includes(searchLower);

                if (!matchesName && !matchesTopic && !matchesSummary && !matchesDescription) {
                    return false;
                }
            }

            return true;
        });
    }, [allPublications, filters]);

    // Group publications by group
    const groupedPublications = useMemo(() => {
        const grouped = filteredPublications.reduce(
            (acc, publication) => {
                const group = publication.group || "Ungrouped";
                if (!acc[group]) {
                    acc[group] = [];
                }
                acc[group].push(publication);
                return acc;
            },
            {} as Record<string, typeof filteredPublications>
        );
        return Object.entries(grouped).sort(([a], [b]) => {
            // "Ungrouped" always last
            if (a === "Ungrouped") return 1;
            if (b === "Ungrouped") return -1;
            return a.localeCompare(b);
        });
    }, [filteredPublications]);

    // Calculate statistics
    const deprecatedCount = useMemo(() => {
        return allPublications.filter((pub) => pub.deprecated).length;
    }, [allPublications]);

    const qosCounts = useMemo(() => {
        return allPublications.reduce(
            (acc, pub) => {
                acc[pub.qos] = (acc[pub.qos] || 0) + 1;
                return acc;
            },
            {} as Record<number, number>
        );
    }, [allPublications]);

    const retainedCount = useMemo(() => {
        return allPublications.filter((pub) => pub.retained).length;
    }, [allPublications]);

    return (
        <div className='flex-1 overflow-y-auto p-10'>
            <Breadcrumbs items={[{ label: "MQTT Publications" }]} />

            <PageHeader
                title='MQTT Publications'
                description='Browse all MQTT publication operations'
            />

            {/* Statistics */}
            <div className='mb-8 grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6'>
                <StatCard
                    label='Total'
                    value={allPublications.length}
                    color='purple'
                />
                {Object.entries(qosCounts)
                    .sort(([a], [b]) => Number(a) - Number(b))
                    .map(([qos, count]) => (
                        <StatCard
                            key={qos}
                            label={`QoS ${qos}`}
                            value={count}
                            color='blue'
                        />
                    ))}
                {retainedCount > 0 && (
                    <StatCard
                        label='Retained'
                        value={retainedCount}
                        color='green'
                    />
                )}
                {deprecatedCount > 0 && (
                    <StatCard
                        label='Deprecated'
                        value={deprecatedCount}
                        color='yellow'
                    />
                )}
            </div>

            <MQTTFilters
                groups={groups}
                onFilterChange={setFilters}
            />

            {filteredPublications.length === 0 ? (
                <EmptyState
                    title='No publications found'
                    description="Try adjusting your filters or search query to find what you're looking for."
                    icon='🔍'
                />
            ) : (
                <div>
                    {groupedPublications.map(([groupName, publications]) => (
                        <CollapsibleGroup
                            key={groupName}
                            title={groupName}
                            defaultOpen={true}>
                            <div className='grid gap-5'>
                                {publications.map((publication) => (
                                    <ItemCard
                                        key={publication.operationID}
                                        href={`/api/mqtt/publication/${publication.operationID}` as Route}
                                        title={publication.operationID}
                                        subtitle={
                                            <div className='flex items-center gap-2'>
                                                <PubBadge />
                                                <span className='font-mono text-sm'>
                                                    <RoutePath path={publication.topic} />
                                                </span>
                                            </div>
                                        }
                                        description={publication.summary || publication.description}
                                        deprecated={!!publication.deprecated}
                                        hoverBorderColor='hover:border-accent-blue-light'
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

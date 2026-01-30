"use client";

import type { Route } from "next";
import { useMemo, useState } from "react";
import { Breadcrumbs } from "@/components/breadcrumbs";
import { CollapsibleGroup } from "@/components/collapsible-group";
import { EmptyState } from "@/components/empty-state";
import { ItemCard } from "@/components/item-card";
import { type MQTTFilterState, MQTTFilters } from "@/components/mqtt-filters";
import { PageHeader } from "@/components/page-header";
import { RoutePath } from "@/components/route-path";
import { StatCard } from "@/components/stat-card";
import { SubBadge } from "@/components/sub-badge";
import { getAllMQTTSubscriptions } from "@/data/api";

export default function MQTTSubscriptionsPage() {
    const allSubscriptions = getAllMQTTSubscriptions();
    const [filters, setFilters] = useState<MQTTFilterState>({
        group: "all",
        hideDeprecated: false,
        search: "",
    });

    // Extract unique groups
    const groups = useMemo(() => {
        const uniqueGroups = new Set(allSubscriptions.map((sub) => sub.group).filter(Boolean));
        return Array.from(uniqueGroups).sort() as string[];
    }, [allSubscriptions]);

    // Filter subscriptions
    const filteredSubscriptions = useMemo(() => {
        return allSubscriptions.filter((subscription) => {
            // Group filter
            if (filters.group !== "all" && subscription.group !== filters.group) {
                return false;
            }

            // Deprecated filter
            if (filters.hideDeprecated && subscription.deprecated) {
                return false;
            }

            // Search filter
            if (filters.search) {
                const searchLower = filters.search.toLowerCase();
                const matchesName = subscription.operationID.toLowerCase().includes(searchLower);
                const matchesTopic = subscription.topic.toLowerCase().includes(searchLower);
                const matchesSummary = subscription.summary?.toLowerCase().includes(searchLower);
                const matchesDescription = subscription.description?.toLowerCase().includes(searchLower);

                if (!matchesName && !matchesTopic && !matchesSummary && !matchesDescription) {
                    return false;
                }
            }

            return true;
        });
    }, [allSubscriptions, filters]);

    // Group subscriptions by group
    const groupedSubscriptions = useMemo(() => {
        const grouped = filteredSubscriptions.reduce(
            (acc, subscription) => {
                const group = subscription.group || "Ungrouped";
                if (!acc[group]) {
                    acc[group] = [];
                }
                acc[group].push(subscription);
                return acc;
            },
            {} as Record<string, typeof filteredSubscriptions>
        );
        return Object.entries(grouped).sort(([a], [b]) => {
            // "Ungrouped" always last
            if (a === "Ungrouped") return 1;
            if (b === "Ungrouped") return -1;
            return a.localeCompare(b);
        });
    }, [filteredSubscriptions]);

    // Calculate statistics
    const deprecatedCount = useMemo(() => {
        return allSubscriptions.filter((sub) => sub.deprecated).length;
    }, [allSubscriptions]);

    const qosCounts = useMemo(() => {
        return allSubscriptions.reduce(
            (acc, sub) => {
                acc[sub.qos] = (acc[sub.qos] || 0) + 1;
                return acc;
            },
            {} as Record<number, number>
        );
    }, [allSubscriptions]);

    return (
        <div className='flex-1 overflow-y-auto p-10'>
            <Breadcrumbs items={[{ label: "MQTT Subscriptions" }]} />

            <PageHeader
                title='MQTT Subscriptions'
                description='Browse all MQTT subscription operations'
            />

            {/* Statistics */}
            <div className='mb-8 grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6'>
                <StatCard
                    label='Total'
                    value={allSubscriptions.length}
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

            {filteredSubscriptions.length === 0 ? (
                <EmptyState
                    title='No subscriptions found'
                    description="Try adjusting your filters or search query to find what you're looking for."
                    icon='🔍'
                />
            ) : (
                <div>
                    {groupedSubscriptions.map(([groupName, subscriptions]) => (
                        <CollapsibleGroup
                            key={groupName}
                            title={groupName}
                            defaultOpen={true}>
                            <div className='grid gap-5'>
                                {subscriptions.map((subscription) => (
                                    <ItemCard
                                        key={subscription.operationID}
                                        href={`/api/mqtt/subscription/${subscription.operationID}` as Route}
                                        title={subscription.operationID}
                                        subtitle={
                                            <div className='flex items-center gap-2'>
                                                <SubBadge />
                                                <span className='font-mono text-sm'>
                                                    <RoutePath path={subscription.topic} />
                                                </span>
                                            </div>
                                        }
                                        description={subscription.summary || subscription.description}
                                        deprecated={!!subscription.deprecated}
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

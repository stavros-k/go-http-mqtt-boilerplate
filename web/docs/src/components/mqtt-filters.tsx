"use client";

import { useState } from "react";

export type MQTTFilterState = {
    group: string;
    hideDeprecated: boolean;
    search: string;
};

type MQTTFiltersProps = {
    groups: string[];
    onFilterChange: (filter: MQTTFilterState) => void;
};

export function MQTTFilters({ groups, onFilterChange }: MQTTFiltersProps) {
    const [selectedGroup, setSelectedGroup] = useState<string>("all");
    const [hideDeprecated, setHideDeprecated] = useState(false);
    const [search, setSearch] = useState("");

    const handleFilterChange = (updates: Partial<MQTTFilterState>) => {
        const newFilter = {
            group: updates.group !== undefined ? updates.group : selectedGroup,
            hideDeprecated: updates.hideDeprecated !== undefined ? updates.hideDeprecated : hideDeprecated,
            search: updates.search !== undefined ? updates.search : search,
        };
        onFilterChange(newFilter);
    };

    const handleGroupChange = (group: string) => {
        setSelectedGroup(group);
        handleFilterChange({ group });
    };

    const handleDeprecatedToggle = () => {
        const newValue = !hideDeprecated;
        setHideDeprecated(newValue);
        handleFilterChange({ hideDeprecated: newValue });
    };

    const handleSearchChange = (newSearch: string) => {
        setSearch(newSearch);
        handleFilterChange({ search: newSearch });
    };

    return (
        <div className='mb-8 flex flex-col gap-4 rounded-xl border-2 border-border-primary bg-bg-secondary p-4'>
            <div className='flex flex-col flex-wrap gap-4 sm:flex-row sm:items-center'>
                {/* Group Filter */}
                {groups.length > 0 && (
                    <div className='flex items-center gap-3'>
                        <span className='font-bold text-sm text-text-primary'>Group:</span>
                        <select
                            value={selectedGroup}
                            onChange={(e) => handleGroupChange(e.target.value)}
                            className='cursor-pointer rounded-lg border-2 border-border-primary bg-bg-tertiary px-3 py-1.5 font-bold text-sm text-text-primary hover:border-accent-blue focus:border-accent-blue focus:ring-2 focus:ring-accent-blue'>
                            <option value='all'>All</option>
                            {groups.map((group) => (
                                <option
                                    key={group}
                                    value={group}>
                                    {group}
                                </option>
                            ))}
                        </select>
                    </div>
                )}

                {/* Hide Deprecated */}
                <label className='group flex cursor-pointer items-center gap-2.5'>
                    <input
                        type='checkbox'
                        checked={hideDeprecated}
                        onChange={handleDeprecatedToggle}
                        className='h-5 w-5 cursor-pointer rounded-md border-2 border-border-primary bg-bg-tertiary text-accent-blue focus:ring-2 focus:ring-accent-blue'
                    />
                    <span className='font-bold text-sm text-text-primary transition-colors group-hover:text-accent-blue'>
                        Hide deprecated
                    </span>
                </label>
            </div>

            {/* Search */}
            <div className='flex items-center gap-3'>
                <span className='font-bold text-sm text-text-primary'>Search:</span>
                <input
                    type='text'
                    value={search}
                    onChange={(e) => handleSearchChange(e.target.value)}
                    placeholder='Search by operation name, topic, or description...'
                    className='flex-1 rounded-lg border-2 border-border-primary bg-bg-tertiary px-3 py-2 text-sm text-text-primary placeholder-text-muted hover:border-accent-blue focus:border-accent-blue focus:ring-2 focus:ring-accent-blue'
                />
            </div>
        </div>
    );
}

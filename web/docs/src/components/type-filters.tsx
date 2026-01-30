"use client";

import { useState } from "react";
import { getKindDisplayName } from "./type-kind-badge";

export type TypeFilterState = {
    kind: string;
    hideDeprecated: boolean;
    search: string;
};

type TypeFiltersProps = {
    kinds: string[];
    onFilterChange: (filter: TypeFilterState) => void;
};

export function TypeFilters({ kinds, onFilterChange }: TypeFiltersProps) {
    const [selectedKind, setSelectedKind] = useState<string>("all");
    const [hideDeprecated, setHideDeprecated] = useState(false);
    const [search, setSearch] = useState("");

    const handleFilterChange = (updates: Partial<TypeFilterState>) => {
        const newFilter = {
            kind: updates.kind !== undefined ? updates.kind : selectedKind,
            hideDeprecated: updates.hideDeprecated !== undefined ? updates.hideDeprecated : hideDeprecated,
            search: updates.search !== undefined ? updates.search : search,
        };
        onFilterChange(newFilter);
    };

    const handleKindChange = (kind: string) => {
        setSelectedKind(kind);
        handleFilterChange({ kind });
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
                {/* Kind Filter */}
                <div className='flex items-center gap-3'>
                    <span className='font-bold text-sm text-text-primary'>Kind:</span>
                    <div className='flex flex-wrap gap-2'>
                        <button
                            type='button'
                            onClick={() => handleKindChange("all")}
                            className={`rounded-lg px-3 py-1.5 font-bold text-sm transition-all duration-200 ${
                                selectedKind === "all"
                                    ? "border-2 border-accent-blue-border bg-bg-tertiary text-text-primary shadow-md"
                                    : "border-2 border-border-primary bg-bg-secondary text-text-primary hover:border-border-secondary"
                            }`}>
                            All
                        </button>
                        {kinds.map((kind) => (
                            <button
                                key={kind}
                                type='button'
                                onClick={() => handleKindChange(kind)}
                                className={`rounded-lg px-3 py-1.5 font-bold text-sm transition-all duration-200 ${
                                    selectedKind === kind
                                        ? "border-2 border-accent-blue-border bg-bg-tertiary text-text-primary shadow-md"
                                        : "border-2 border-border-primary bg-bg-secondary text-text-primary hover:border-border-secondary"
                                }`}>
                                {getKindDisplayName(kind)}
                            </button>
                        ))}
                    </div>
                </div>

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
                    placeholder='Search by type name...'
                    className='flex-1 rounded-lg border-2 border-border-primary bg-bg-tertiary px-3 py-2 text-sm text-text-primary placeholder-text-muted hover:border-accent-blue focus:border-accent-blue focus:ring-2 focus:ring-accent-blue'
                />
            </div>
        </div>
    );
}

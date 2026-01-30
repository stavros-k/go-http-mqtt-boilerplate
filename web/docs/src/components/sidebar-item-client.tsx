"use client";
import Link from "next/link";
import { useParams } from "next/navigation";
import type { ItemType } from "@/data/api";
import { RoutePath } from "./route-path";
import type { getItemData } from "./sidebar";
import { TypeKindBadge } from "./type-kind-badge";
import { VerbBadge } from "./verb-badge";

type Props = {
    type: ItemType;
    item: ReturnType<typeof getItemData>;
};

export const SidebarItem = ({ type, item }: Props) => {
    const params = useParams();
    const currentName = type === "operation" ? params.operationId : params[type];
    const { urlPath, data, title } = item;
    const isDeprecated = !!data?.deprecated;
    const isActive = currentName === item.name;
    const method = "method" in item ? item.method : undefined;
    const path = "path" in item ? item.path : undefined;
    const topic = "topic" in item ? item.topic : undefined;
    const kind = "kind" in data ? data.kind : undefined;

    return (
        <Link
            key={urlPath}
            href={urlPath}
            className={`block ${
                isActive ? "bg-bg-tertiary text-text-primary shadow-md" : "bg-bg-secondary text-text-primary"
            } mb-1.5 rounded-lg border-2 px-2.5 py-2 text-sm no-underline transition-all duration-200 hover:shadow-sm ${
                isActive ? "border-accent-blue-border" : "border-border-primary hover:border-border-secondary"
            } ${isDeprecated ? "opacity-deprecated" : ""}`}>
            <div className='flex flex-col gap-1'>
                <div className='flex items-center justify-between gap-1.5'>
                    <span className='truncate font-semibold text-base'>{title || item.name}</span>
                    {kind && type === "type" && (
                        <TypeKindBadge
                            kind={kind}
                            size='xs'
                        />
                    )}
                </div>
                {method && path && (
                    <div className='flex items-center gap-1'>
                        <VerbBadge
                            verb={method}
                            size='xs'
                        />
                        <RoutePath
                            path={path}
                            className='truncate font-mono text-text-muted text-xs'
                        />
                    </div>
                )}
                {type === "mqtt-publication" && topic && (
                    <div className='flex items-center gap-1'>
                        <span className='rounded border border-accent-blue-border bg-accent-blue-bg px-1.5 py-0.5 font-bold text-[10px] text-accent-blue-text'>
                            PUB
                        </span>
                        <RoutePath
                            path={topic}
                            className='truncate font-mono text-accent-blue-light text-xs'
                        />
                    </div>
                )}
                {type === "mqtt-subscription" && topic && (
                    <div className='flex items-center gap-1'>
                        <span className='rounded border border-accent-green-border bg-accent-green-bg px-1.5 py-0.5 font-bold text-[10px] text-accent-green-text'>
                            SUB
                        </span>
                        <RoutePath
                            path={topic}
                            className='truncate font-mono text-accent-green-light text-xs'
                        />
                    </div>
                )}
            </div>
        </Link>
    );
};

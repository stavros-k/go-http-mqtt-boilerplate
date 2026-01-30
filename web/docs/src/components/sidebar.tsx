"use client";

import type { Route } from "next";
import Link from "next/link";
import { useState } from "react";
import { FaDatabase } from "react-icons/fa";
import { IoHome } from "react-icons/io5";
import { MdChevronLeft, MdChevronRight } from "react-icons/md";
import { TbApi, TbCode, TbFileDescription } from "react-icons/tb";
import {
    docs,
    getAllMQTTPublications,
    getAllMQTTSubscriptions,
    getAllOperations,
    type ItemType,
    type TypeKeys,
} from "@/data/api";
import { CodeThemeToggle } from "./code-theme-toggle";
import { ConnectionIndicator } from "./connection-indicator";
import { PubBadge } from "./pub-badge";
import { SidebarSection } from "./sidebar-section";
import { SubBadge } from "./sub-badge";

const SidebarLink = ({
    title,
    href,
    icon: Icon,
    iconColor = "text-text-secondary",
}: {
    title: string;
    href: Route;
    icon: React.ComponentType<{ className?: string }>;
    iconColor?: string;
}) => {
    return (
        <div className='mt-6 border-border-sidebar border-t-2 pt-6'>
            <Link
                href={href}
                className='flex items-center gap-2.5 rounded-lg px-3 py-2.5 font-medium transition-all duration-200 hover:bg-bg-hover'>
                <Icon className={`h-5 w-5 ${iconColor} shrink-0`} />
                {title}
            </Link>
        </div>
    );
};

export const Sidebar = () => {
    const [isExpanded, setIsExpanded] = useState(false);

    const toggleExpanded = () => {
        setIsExpanded((prev) => !prev);
    };

    return (
        <aside
            className={`sticky top-0 flex h-screen shrink-0 flex-col border-border-sidebar border-r-2 bg-bg-sidebar text-sm text-text-primary transition-[width,min-width,max-width] duration-250 ease-in-out ${isExpanded ? "w-fit min-w-80 max-w-md" : "w-16 min-w-16 max-w-16"}
            `}>
            {/* Expanded Header */}
            <div className={`border-border-sidebar border-b-2 p-6 pb-4 ${isExpanded ? "block" : "hidden"}`}>
                <div className='mb-4 flex items-center justify-between'>
                    <h1 className='font-bold text-xl'>
                        <Link
                            href='/'
                            className='group flex items-center gap-2.5 transition-colors'>
                            <IoHome className='h-8 w-8 shrink-0 text-text-primary transition-colors group-hover:text-accent-blue-text' />
                            <div className='flex flex-col'>
                                <span>{docs.info.title}</span>
                                <span className='font-normal text-text-muted text-xs'>v{docs.info.version}</span>
                            </div>
                        </Link>
                    </h1>
                    <div className='flex items-center gap-2'>
                        <ConnectionIndicator />
                        <button
                            type='button'
                            onClick={toggleExpanded}
                            className='rounded-lg border-2 border-border-primary bg-bg-tertiary p-2.5 shadow-sm transition-all duration-200 hover:border-accent-blue-border hover:bg-bg-hover hover:shadow-md active:scale-95'
                            aria-label='Collapse sidebar'
                            title='Collapse sidebar'>
                            <MdChevronLeft className='h-6 w-6 text-text-primary' />
                        </button>
                    </div>
                </div>
                <div className='space-y-3'>
                    <CodeThemeToggle />
                </div>
            </div>

            {/* Collapsed Icon Navigation */}
            <div className={`flex flex-1 flex-col gap-2 overflow-y-auto p-2 pt-4 ${isExpanded ? "hidden" : "flex"}`}>
                <button
                    type='button'
                    onClick={toggleExpanded}
                    className='flex items-center justify-center rounded-lg border-2 border-border-primary bg-bg-tertiary p-2.5 shadow-sm transition-all duration-200 hover:border-accent-blue-border hover:bg-bg-hover hover:shadow-md active:scale-95'
                    aria-label='Expand sidebar'
                    title='Expand sidebar'>
                    <MdChevronRight className='h-6 w-6 text-text-primary' />
                </button>

                <div className='my-2 border-border-sidebar border-t-2' />

                <Link
                    href='/'
                    className='group flex items-center justify-center rounded-lg p-2.5 transition-all duration-200 hover:bg-bg-hover'
                    title='Home'>
                    <IoHome className='h-6 w-6 text-text-primary group-hover:text-accent-blue-text' />
                </Link>

                <div className='my-2 border-border-sidebar border-t-2' />

                <Link
                    href='/api/operations'
                    className='group flex items-center justify-center rounded-lg p-2.5 transition-all duration-200 hover:bg-bg-hover'
                    title='HTTP Operations'>
                    <TbApi className='h-6 w-6 text-accent-blue group-hover:text-accent-blue-text' />
                </Link>

                <Link
                    href='/api/mqtt/publications'
                    className='flex items-center justify-center rounded-lg p-2.5 transition-all duration-200 hover:bg-bg-hover'
                    title='MQTT Publications'>
                    <PubBadge border={false} />
                </Link>

                <Link
                    href='/api/mqtt/subscriptions'
                    className='flex items-center justify-center rounded-lg p-2.5 transition-all duration-200 hover:bg-bg-hover'
                    title='MQTT Subscriptions'>
                    <SubBadge border={false} />
                </Link>

                <Link
                    href='/api/types'
                    className='group flex items-center justify-center rounded-lg p-2.5 transition-all duration-200 hover:bg-bg-hover'
                    title='Types'>
                    <TbCode className='h-6 w-6 text-success-text group-hover:text-accent-purple' />
                </Link>

                <div className='my-2 border-border-sidebar border-t-2' />

                <Link
                    href='/api/database/schema'
                    className='group flex items-center justify-center rounded-lg p-2.5 transition-all duration-200 hover:bg-bg-hover'
                    title='Database Schema'>
                    <FaDatabase className='h-5 w-5 text-info-text group-hover:text-info-border' />
                </Link>

                <Link
                    href='/api/openapi'
                    className='group flex items-center justify-center rounded-lg p-2.5 transition-all duration-200 hover:bg-bg-hover'
                    title='OpenAPI Specification'>
                    <TbFileDescription className='h-6 w-6 text-warning-text group-hover:text-warning-border' />
                </Link>
            </div>

            {/* Expanded Scrollable Content */}
            <div
                className={`flex-1 overflow-y-scroll p-6 pt-4 ${isExpanded ? "block" : "hidden"}`}
                style={{
                    scrollbarWidth: "auto",
                    scrollbarColor: "rgb(100 116 139) transparent",
                }}>
                <SidebarSection
                    title='HTTP Operations'
                    type='operation'
                    overviewHref='/api/operations'
                    icon={TbApi}
                    iconColor='text-accent-blue'
                />
                <SidebarSection
                    title='MQTT Publications'
                    type='mqtt-publication'
                    overviewHref='/api/mqtt/publications'
                    icon={() => <PubBadge border={false} />}
                />
                <SidebarSection
                    title='MQTT Subscriptions'
                    type='mqtt-subscription'
                    overviewHref='/api/mqtt/subscriptions'
                    icon={() => <SubBadge border={false} />}
                />
                <SidebarSection
                    title='Types'
                    type='type'
                    overviewHref='/api/types'
                    icon={TbCode}
                    iconColor='text-success-text'
                />

                <SidebarLink
                    title='Database Schema'
                    href='/api/database/schema'
                    icon={FaDatabase}
                    iconColor='text-info-text'
                />
                <SidebarLink
                    title='OpenAPI Specification'
                    href='/api/openapi'
                    icon={TbFileDescription}
                    iconColor='text-warning-text'
                />
            </div>
        </aside>
    );
};

type getItemProps = {
    type: ItemType;
    itemName: string;
};

export function getItemData({ type, itemName }: getItemProps) {
    if (type === "type") {
        const name = itemName as TypeKeys;
        if (!docs.types[name]) {
            throw new Error(`Type ${name} not found`);
        }
        return {
            type: type,
            name: name,
            urlPath: `/api/${type}/${name}`,
            data: docs.types[name],
            title: "",
            group: "",
        } as const;
    }
    if (type === "operation") {
        // itemName is operationID for operations
        const allOperations = getAllOperations();
        const operation = allOperations.find((op) => op.operationID === itemName);
        if (!operation) {
            throw new Error(`Operation ${itemName} not found`);
        }
        return {
            type: type,
            name: itemName,
            urlPath: `/api/operation/${itemName}`,
            data: operation,
            title: operation.operationID,
            method: operation.method,
            path: operation.path,
            group: operation.group || "",
        } as const;
    }
    if (type === "mqtt-publication") {
        const allPublications = getAllMQTTPublications();
        const publication = allPublications.find((pub) => pub.operationID === itemName);
        if (!publication) {
            throw new Error(`MQTT Publication ${itemName} not found`);
        }
        return {
            type: type,
            name: itemName,
            urlPath: `/api/mqtt/publication/${itemName}`,
            data: publication,
            title: publication.operationID,
            topic: publication.topic,
            group: publication.group || "",
        } as const;
    }
    if (type === "mqtt-subscription") {
        const allSubscriptions = getAllMQTTSubscriptions();
        const subscription = allSubscriptions.find((sub) => sub.operationID === itemName);
        if (!subscription) {
            throw new Error(`MQTT Subscription ${itemName} not found`);
        }
        return {
            type: type,
            name: itemName,
            urlPath: `/api/mqtt/subscription/${itemName}`,
            data: subscription,
            title: subscription.operationID,
            topic: subscription.topic,
            group: subscription.group || "",
        } as const;
    }
    throw new Error(`Invalid type: ${type}`);
}

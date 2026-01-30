"use client";
import { Fragment, useState } from "react";
import type { MQTTPublicationData, MQTTSubscriptionData } from "@/data/api";

type MQTTTopicHeaderProps = {
    topic: string;
    topicMQTT: string;
    topicParameters?: MQTTPublicationData["topicParameters"] | MQTTSubscriptionData["topicParameters"];
    type: "publication" | "subscription"; // publication or subscription
};

export function MQTTTopicHeader({ topic, topicMQTT, topicParameters, type }: MQTTTopicHeaderProps) {
    const [hoveredParam, setHoveredParam] = useState<string | null>(null);

    // Split the path into segments and identify parameters
    const segments = topic.split(/(\{[^}]+\})/g).filter(Boolean);

    const isPublication = type === "publication";
    const badgeColor = isPublication
        ? "bg-accent-blue-bg text-accent-blue-text border-accent-blue-border"
        : "bg-accent-green-bg text-accent-green-text border-accent-green-border";

    return (
        <div>
            {/* Header with badge and parameterized topic */}
            <div className='mb-4 flex items-center gap-3'>
                <span className={`rounded border-2 px-3 py-1.5 font-bold text-sm ${badgeColor}`}>
                    {isPublication ? "PUB" : "SUB"}
                </span>
                <h2 className='font-mono font-semibold text-xl'>
                    <span>
                        {segments.map((segment, idx) => {
                            const uniqueKey = `${segment}-${idx}`;
                            const isParam = segment.startsWith("{") && segment.endsWith("}");
                            const paramName = isParam ? segment.slice(1, -1) : null;
                            const isHovered = hoveredParam === paramName;
                            if (!isParam) {
                                return (
                                    <span
                                        key={uniqueKey}
                                        className='text-text-secondary'>
                                        {segment}
                                    </span>
                                );
                            }

                            return (
                                <Fragment key={uniqueKey}>
                                    <span className='text-type-enum'>{"{"}</span>
                                    <button
                                        type='button'
                                        className={`border-0 bg-transparent p-0 font-bold font-mono text-xl transition-colors ${
                                            isHovered ? "text-accent-blue-light" : "text-accent-blue"
                                        }`}
                                        onMouseEnter={() => setHoveredParam(paramName)}
                                        onMouseLeave={() => setHoveredParam(null)}
                                        aria-label={`Highlight ${paramName} parameter`}>
                                        {paramName}
                                    </button>
                                    <span className='text-type-enum'>{"}"}</span>
                                </Fragment>
                            );
                        })}
                    </span>
                </h2>
            </div>

            {/* MQTT Wildcard Format */}
            <div className='mb-4 text-sm text-text-muted'>
                <span className='font-semibold'>MQTT Format: </span>
                <code className='font-mono'>{topicMQTT}</code>
            </div>

            {/* Topic Parameters */}
            {topicParameters && topicParameters.length > 0 && (
                <div className='mb-8'>
                    <h3 className='mb-4 font-bold text-2xl text-text-primary'>Topic Parameters</h3>
                    <div className='space-y-3'>
                        {topicParameters.map((param) => {
                            const isHovered = hoveredParam === param.name;
                            return (
                                <button
                                    key={param.name}
                                    type='button'
                                    className={`w-full rounded-lg border-2 p-4 text-left transition-colors ${
                                        isHovered
                                            ? "border-accent-blue/50 bg-accent-blue/5"
                                            : "border-border-primary bg-bg-tertiary"
                                    }`}
                                    onMouseEnter={() => setHoveredParam(param.name)}
                                    onMouseLeave={() => setHoveredParam(null)}
                                    aria-label={`Highlight ${param.name} parameter in topic`}>
                                    <div className='mb-2 flex items-start justify-between gap-4'>
                                        <code className='font-semibold text-base text-text-primary'>{param.name}</code>
                                        {param.type && (
                                            <code className='rounded-lg border-2 border-type-primitive/30 bg-type-primitive/10 px-3 py-1.5 font-mono font-semibold text-sm text-type-primitive'>
                                                {param.type}
                                            </code>
                                        )}
                                    </div>
                                    {param.description && (
                                        <p className='text-sm text-text-tertiary'>{param.description}</p>
                                    )}
                                </button>
                            );
                        })}
                    </div>
                </div>
            )}
        </div>
    );
}

"use client";
import { Fragment, useState } from "react";
import type { OperationData } from "@/data/api";
import { VerbBadge } from "./verb-badge";

type OperationHeaderProps = {
    method: string;
    path: string;
    parameters?: OperationData["parameters"];
};

export function OperationHeader({ method, path, parameters }: OperationHeaderProps) {
    const [hoveredParam, setHoveredParam] = useState<string | null>(null);

    // Split the path into segments and identify parameters
    const segments = path.split(/(\{[^}]+\})/g).filter(Boolean);

    return (
        <div>
            <div className='mb-4 flex items-center gap-3'>
                <VerbBadge
                    verb={method}
                    size='lg'
                />
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

            {parameters && parameters.length > 0 && (
                <div className='mb-8'>
                    <h3 className='mb-4 font-bold text-2xl text-text-primary'>Parameters</h3>
                    <div className='space-y-3'>
                        {parameters.map((param) => {
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
                                    aria-label={`Highlight ${param.name} parameter in path`}>
                                    <div className='mb-2 flex items-start justify-between gap-4'>
                                        <div className='flex flex-wrap items-center gap-2'>
                                            <code className='font-semibold text-base text-text-primary'>
                                                {param.name}
                                            </code>
                                            {param.required && (
                                                <span className='rounded border border-red-500/30 bg-red-500/20 px-2 py-0.5 font-semibold text-red-400 text-xs'>
                                                    required
                                                </span>
                                            )}
                                            <span className='rounded border border-blue-500/30 bg-blue-500/20 px-2 py-0.5 text-blue-400 text-xs'>
                                                {param.in}
                                            </span>
                                        </div>
                                        <code className='rounded-lg border-2 border-type-primitive/30 bg-type-primitive/10 px-3 py-1.5 font-mono font-semibold text-sm text-type-primitive'>
                                            {param.type}
                                        </code>
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

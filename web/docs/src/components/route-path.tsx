import { Fragment } from "react";

type RoutePathProps = {
    path: string;
    className?: string;
};

export function RoutePath({ path, className = "" }: RoutePathProps) {
    // Split the path into segments and identify parameters (anything between {})
    const segments = path.split(/(\{[^}]+\})/g).filter(Boolean);

    // Create unique keys by tracking segment occurrences
    const segmentCounts = new Map<string, number>();

    return (
        <span className={className}>
            {segments.map((segment) => {
                const count = segmentCounts.get(segment) || 0;
                segmentCounts.set(segment, count + 1);
                const uniqueKey = count === 0 ? segment : `${segment}-${count}`;

                const isParam = segment.startsWith("{") && segment.endsWith("}");
                const paramName = isParam ? segment.slice(1, -1) : null;
                if (isParam) {
                    return (
                        <Fragment key={uniqueKey}>
                            <span className='text-type-enum'>{"{"}</span>
                            <span className='font-bold text-accent-blue'>{paramName}</span>
                            <span className='text-type-enum'>{"}"}</span>
                        </Fragment>
                    );
                }
                return (
                    <span
                        key={uniqueKey}
                        className='text-text-secondary'>
                        {segment}
                    </span>
                );
            })}
        </span>
    );
}

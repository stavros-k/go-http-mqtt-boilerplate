type VerbBadgeProps = {
    verb: string;
    size?: "xs" | "sm" | "md" | "lg";
};

export function VerbBadge({ verb, size = "md" }: VerbBadgeProps) {
    const upperVerb = verb.toUpperCase();

    const colorClasses =
        {
            GET: "bg-blue-500/20 text-blue-400 border-blue-500/30",
            POST: "bg-green-500/20 text-green-400 border-green-500/30",
            PUT: "bg-yellow-500/20 text-yellow-400 border-yellow-500/30",
            PATCH: "bg-purple-500/20 text-purple-400 border-purple-500/30",
            DELETE: "bg-red-500/20 text-red-400 border-red-500/30",
        }[upperVerb] || "bg-gray-500/20 text-gray-400 border-gray-500/30";

    const sizeClasses = {
        xs: "text-[9px] px-1.5 py-0.5",
        sm: "text-xs px-2 py-0.5",
        md: "text-sm px-3 py-1",
        lg: "text-base px-4 py-1.5",
    }[size];

    return (
        <span className={`${colorClasses} ${sizeClasses} inline-block rounded-lg border-2 font-bold font-mono`}>
            {upperVerb}
        </span>
    );
}

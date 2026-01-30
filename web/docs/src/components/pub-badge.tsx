type PubBadgeProps = {
    border?: boolean;
};
export function PubBadge({ border = true }: PubBadgeProps) {
    return (
        <span
            className={`rounded bg-accent-blue-bg px-1.5 py-0.5 font-bold text-[10px] ${border ? "border border-accent-blue-border" : ""} text-accent-blue-text`}>
            PUB
        </span>
    );
}

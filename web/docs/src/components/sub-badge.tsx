type SubBadgeProps = {
    border?: boolean;
};
export function SubBadge({ border = true }: SubBadgeProps) {
    return (
        <span
            className={`rounded bg-accent-green-bg px-1.5 py-0.5 font-bold text-[10px] ${border ? "border border-accent-green-border" : ""} text-accent-green-text`}>
            SUB
        </span>
    );
}

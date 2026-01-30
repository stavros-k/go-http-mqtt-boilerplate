type Props = {
    group: string;
    size?: "sm" | "md" | "lg";
};
export const Group = ({ group, size = "md" }: Props) => {
    if (!group) return null;

    const sizeClasses = {
        sm: "text-xs px-2 py-1",
        md: "text-sm px-3 py-1.5",
        lg: "text-base px-4 py-2",
    }[size];

    return (
        <span
            className={`${sizeClasses} inline-block rounded-lg border-2 border-tag-teal-border bg-tag-teal-bg font-bold text-tag-teal-text shadow-sm`}>
            {group}
        </span>
    );
};

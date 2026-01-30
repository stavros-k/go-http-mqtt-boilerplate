type TypeKindBadgeProps = {
    kind: string;
    size?: "xs" | "sm" | "md";
};

const kindDisplayNames: Record<string, string> = {
    string_enum: "String Enum",
    int_enum: "Int Enum",
    object: "Object",
    array: "Array",
    primitive: "Primitive",
    string: "String",
    number: "Number",
    boolean: "Boolean",
    integer: "Integer",
};

export function getKindDisplayName(kind: string): string {
    return kindDisplayNames[kind] || kind.charAt(0).toUpperCase() + kind.slice(1);
}

export function TypeKindBadge({ kind, size = "md" }: TypeKindBadgeProps) {
    const sizeClasses = {
        xs: "text-[9px] px-1.5 py-0.5",
        sm: "text-xs px-2 py-1",
        md: "text-sm px-3 py-1.5",
    }[size];

    return (
        <span
            className={`${sizeClasses} inline-block rounded-lg border-2 border-info-border bg-info-bg font-bold text-info-text`}>
            {getKindDisplayName(kind)}
        </span>
    );
}

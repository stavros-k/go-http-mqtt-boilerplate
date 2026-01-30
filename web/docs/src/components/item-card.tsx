import type { Route } from "next";
import Link from "next/link";
import type { ReactNode } from "react";

type ItemCardProps = {
    href: Route;
    title: string;
    subtitle?: ReactNode;
    subtitleColor?: string;
    description: string;
    badges?: ReactNode;
    tags?: ReactNode;
    deprecated?: boolean;
    hoverBorderColor?: string;
};

export const ItemCard = ({
    href,
    title,
    subtitle,
    subtitleColor = "text-accent-blue",
    description,
    badges,
    tags,
    deprecated,
    hoverBorderColor = "hover:border-accent-blue",
}: ItemCardProps) => {
    return (
        <Link
            href={href}
            className={`block rounded-xl border-2 border-border-primary bg-bg-secondary p-6 ${hoverBorderColor} relative transition-all duration-200 hover:shadow-lg ${
                deprecated ? "pb-12 opacity-deprecated hover:opacity-deprecated-hover" : ""
            }`}>
            <div className='flex items-start justify-between gap-4'>
                <div className='min-w-0 flex-1'>
                    <h3 className='mb-2 font-bold text-text-primary text-xl'>{title}</h3>
                    {subtitle && <div className={`mb-2 font-semibold text-base ${subtitleColor}`}>{subtitle}</div>}
                    <p className='text-sm text-text-secondary leading-relaxed'>{description}</p>
                </div>
                <div className='flex shrink-0 flex-col items-end gap-2'>
                    {tags && <div className='flex flex-wrap justify-end gap-2'>{tags}</div>}
                    {badges && <div className='flex flex-wrap justify-end gap-2'>{badges}</div>}
                </div>
            </div>
            {deprecated && (
                <div className='absolute right-4 bottom-4 rounded-lg border-2 border-warning-border bg-warning-bg px-3 py-1.5 font-bold text-warning-text text-xs'>
                    Deprecated
                </div>
            )}
        </Link>
    );
};

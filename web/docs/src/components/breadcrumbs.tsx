import type { Route } from "next";
import Link from "next/link";
import { IoChevronForward, IoHome } from "react-icons/io5";

type BreadcrumbItem = {
    label: string;
    href?: Route;
};

type BreadcrumbsProps = {
    items: BreadcrumbItem[];
};

export function Breadcrumbs({ items }: BreadcrumbsProps) {
    return (
        <nav className='mb-6 flex items-center gap-2 text-sm text-text-secondary'>
            <Link
                href='/'
                className='flex items-center gap-1 transition-colors hover:text-accent-blue'>
                <IoHome className='h-4 w-4' />
                <span>Home</span>
            </Link>
            {items.map((item, index) => {
                const isLast = index === items.length - 1;
                return (
                    <div
                        key={item.label}
                        className='flex items-center gap-2'>
                        <IoChevronForward className='h-3 w-3' />
                        {item.href && !isLast ? (
                            <Link
                                href={item.href}
                                className='transition-colors hover:text-accent-blue'>
                                {item.label}
                            </Link>
                        ) : (
                            <span className={isLast ? "font-semibold text-text-primary" : ""}>{item.label}</span>
                        )}
                    </div>
                );
            })}
        </nav>
    );
}

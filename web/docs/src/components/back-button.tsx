import type { Route } from "next";
import Link from "next/link";
import { IoArrowBack } from "react-icons/io5";

type BackButtonProps = {
    href: Route;
    label: string;
};

export function BackButton({ href, label }: BackButtonProps) {
    return (
        <Link
            href={href}
            className='group mb-6 inline-flex items-center gap-2 text-sm text-text-secondary transition-colors hover:text-accent-blue'>
            <IoArrowBack className='h-4 w-4 transition-transform group-hover:-translate-x-1' />
            <span>Back to {label}</span>
        </Link>
    );
}

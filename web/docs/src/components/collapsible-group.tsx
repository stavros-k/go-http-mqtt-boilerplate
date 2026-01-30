"use client";

import { useState } from "react";
import { MdExpandLess, MdExpandMore } from "react-icons/md";

type CollapsibleGroupProps = {
    title: string;
    children: React.ReactNode;
    defaultOpen?: boolean;
};

export function CollapsibleGroup({ title, children, defaultOpen = true }: CollapsibleGroupProps) {
    const [isOpen, setIsOpen] = useState(defaultOpen);

    return (
        <div className='mb-12'>
            <button
                type='button'
                onClick={() => setIsOpen((prev) => !prev)}
                className='group mb-6 flex w-full items-center justify-between border-border-primary border-b-2 pb-3 font-bold text-2xl text-text-primary transition-all duration-200 hover:border-accent-blue hover:text-accent-blue'>
                <h2 className='transition-transform group-hover:scale-105'>{title}</h2>
                <span className='text-text-muted transition-colors group-hover:text-accent-blue'>
                    {isOpen ? <MdExpandLess className='h-7 w-7' /> : <MdExpandMore className='h-7 w-7' />}
                </span>
            </button>
            {isOpen && <div className='grid gap-5'>{children}</div>}
        </div>
    );
}

type CollapsibleCardProps = {
    title: string;
    subtitle?: string;
    children: React.ReactNode;
    defaultOpen?: boolean;
};

export function CollapsibleCard({ title, subtitle, children, defaultOpen = false }: CollapsibleCardProps) {
    const [isOpen, setIsOpen] = useState(defaultOpen);

    return (
        <div className='mb-6 overflow-hidden rounded-xl border-2 border-border-secondary shadow-sm transition-shadow hover:shadow-md'>
            <button
                type='button'
                onClick={() => setIsOpen((prev) => !prev)}
                className='group flex w-full cursor-pointer items-center justify-between bg-bg-tertiary p-5 font-bold text-text-primary transition-colors duration-200 hover:bg-bg-hover'>
                <div className='text-left'>
                    <div className='transition-colors group-hover:text-accent-blue'>{title}</div>
                    {subtitle && <div className='mt-1.5 font-normal text-sm text-text-muted'>{subtitle}</div>}
                </div>
                <span className='text-text-muted transition-colors group-hover:text-accent-blue'>
                    {isOpen ? <MdExpandLess className='h-6 w-6' /> : <MdExpandMore className='h-6 w-6' />}
                </span>
            </button>
            {isOpen && <div className='bg-bg-secondary p-5'>{children}</div>}
        </div>
    );
}

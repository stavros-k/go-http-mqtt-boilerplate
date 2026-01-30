"use client";

import type { ReactNode } from "react";
import { useState } from "react";
import { MdExpandLess, MdExpandMore } from "react-icons/md";

type CollapsibleResponseProps = {
    statusCode: string;
    description: string;
    children: ReactNode;
};

export function CollapsibleResponse({ statusCode, description, children }: CollapsibleResponseProps) {
    const [isOpen, setIsOpen] = useState(false);

    return (
        <div className='overflow-hidden rounded-xl border-2 border-border-secondary shadow-sm transition-shadow hover:shadow-md'>
            <button
                type='button'
                onClick={() => setIsOpen((prev) => !prev)}
                className='group flex w-full cursor-pointer items-center justify-between bg-bg-tertiary p-5 font-bold text-text-primary transition-colors duration-200 hover:bg-bg-hover'>
                <div className='flex items-center gap-3'>
                    <span
                        className={`rounded-lg px-3 py-1.5 font-bold font-mono text-lg ${
                            statusCode.startsWith("2")
                                ? "border-2 border-green-500/30 bg-green-500/20 text-green-400"
                                : statusCode.startsWith("4")
                                  ? "border-2 border-yellow-500/30 bg-yellow-500/20 text-yellow-400"
                                  : "border-2 border-red-500/30 bg-red-500/20 text-red-400"
                        }`}>
                        {statusCode}
                    </span>
                    <span className='font-normal text-sm text-text-muted'>{description}</span>
                </div>
                <span className='text-text-muted transition-colors group-hover:text-accent-blue'>
                    {isOpen ? <MdExpandLess className='h-6 w-6' /> : <MdExpandMore className='h-6 w-6' />}
                </span>
            </button>
            {isOpen && <div className='bg-bg-secondary p-5'>{children}</div>}
        </div>
    );
}

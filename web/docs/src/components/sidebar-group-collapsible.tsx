"use client";

import { useState } from "react";
import { MdExpandLess, MdExpandMore } from "react-icons/md";

type SidebarGroupCollapsibleProps = {
    groupName: string;
    children: React.ReactNode;
};

export const SidebarGroupCollapsible = ({ groupName, children }: SidebarGroupCollapsibleProps) => {
    const [isOpen, setIsOpen] = useState(true);

    return (
        <div className='mb-3'>
            <button
                type='button'
                onClick={() => setIsOpen((prev) => !prev)}
                className='group flex w-full items-center justify-between rounded px-1 py-1 transition-colors hover:bg-bg-hover'>
                <h3 className='font-semibold text-text-dim text-xs uppercase transition-colors group-hover:text-text-primary'>
                    {groupName}
                </h3>
                <span className='text-text-dim transition-colors group-hover:text-text-primary'>
                    {isOpen ? <MdExpandLess className='h-4 w-4' /> : <MdExpandMore className='h-4 w-4' />}
                </span>
            </button>
            {isOpen && <div className='mt-1'>{children}</div>}
        </div>
    );
};

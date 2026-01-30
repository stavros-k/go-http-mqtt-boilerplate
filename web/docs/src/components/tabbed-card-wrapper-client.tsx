"use client";
import { useState } from "react";

type Tab = {
    // https://react-icons.github.io/react-icons/
    icon: React.ReactNode;
    title: string;
    code: React.ReactNode;
};

type Props = {
    tabs: Tab[];
};

export const TabbedCardWrapper = ({ tabs }: Props) => {
    const [activeTab, setActiveTab] = useState(0);

    if (!tabs || tabs.length === 0) return null;

    return (
        <div className='mb-6 rounded-lg border border-border-primary bg-bg-secondary p-6 shadow-sm'>
            <div>
                <div className='mb-4 flex overflow-x-auto border-border-secondary border-b'>
                    {tabs.map((tab, index) => (
                        <button
                            key={tab.title}
                            type='button'
                            onClick={() => setActiveTab(index)}
                            className={`flex cursor-pointer items-center gap-2 whitespace-nowrap border-b-2 px-4 py-2 transition-colors ${
                                activeTab === index
                                    ? "border-accent-blue-hover font-semibold text-accent-blue-hover"
                                    : "border-transparent text-text-muted hover:text-text-secondary"
                            }`}>
                            {tab.icon}
                            <span className='text-xs'>{tab.title}</span>
                        </button>
                    ))}
                </div>
                {tabs.map((tab, index) => (
                    <div
                        key={tab.title}
                        className={activeTab === index ? "block" : "hidden"}>
                        {tab.code}
                    </div>
                ))}
            </div>
        </div>
    );
};

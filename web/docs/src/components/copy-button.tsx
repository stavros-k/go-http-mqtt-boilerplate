"use client";

import { useEffect, useState } from "react";
import { FiCheck, FiCopy } from "react-icons/fi";

type Props = {
    code: string;
};

export const CopyButton = ({ code }: Props) => {
    const [copied, setCopied] = useState(false);

    useEffect(() => {
        if (!copied) return;

        const timeoutId = setTimeout(() => {
            setCopied(false);
        }, 2000);

        return () => {
            clearTimeout(timeoutId);
        };
    }, [copied]);

    const handleCopy = async () => {
        try {
            await navigator.clipboard.writeText(code);
            setCopied(true);
        } catch (err) {
            console.error("Failed to copy text:", err);
        }
    };

    return (
        <button
            onClick={handleCopy}
            className='group absolute top-3 right-3 rounded-lg border border-border-primary bg-bg-secondary p-2 transition-all duration-200 hover:bg-bg-tertiary'
            title={copied ? "Copied!" : "Copy code"}
            type='button'>
            {copied ? (
                <FiCheck className='h-4 w-4 text-green-500' />
            ) : (
                <FiCopy className='h-4 w-4 text-text-secondary transition-colors group-hover:text-text-primary' />
            )}
        </button>
    );
};

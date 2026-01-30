import type { Route } from "next";
import Link from "next/link";
import type { BundledLanguage } from "shiki";
import { codeToHtml } from "shiki";
import { getShikiOptions } from "@/utils/code-theme";
import { CopyButton } from "./copy-button";

type Props = {
    label: { text: string; href?: Route };
    lang: BundledLanguage;
} & ({ code: string; noCodeMessage?: never } | { code: string | null; noCodeMessage: string });

export const CodeWrapper = ({ label, code, lang, noCodeMessage }: Props) => {
    return (
        <div className='mb-6 last:mb-0'>
            <div className='mb-3 font-bold text-sm text-text-tertiary'>
                {label.href && code ? (
                    <Link
                        href={label.href}
                        className='text-accent-blue-hover transition-colors hover:text-accent-blue-light'>
                        {label.text}
                    </Link>
                ) : (
                    <span>{code ? label.text : noCodeMessage}</span>
                )}
            </div>
            {code && (
                <div className='relative overflow-hidden'>
                    <CopyButton code={code} />
                    <div className='overflow-x-auto'>
                        <CodeBlock
                            code={code}
                            lang={lang}
                        />
                    </div>
                </div>
            )}
        </div>
    );
};

type CodeBlockProps = {
    code: string;
    lang: BundledLanguage;
};
async function CodeBlock(props: CodeBlockProps) {
    const out = await codeToHtml(props.code, getShikiOptions(props.lang));

    return (
        // biome-ignore lint/security/noDangerouslySetInnerHtml: This is sanitized by shiki
        <div dangerouslySetInnerHTML={{ __html: out }} />
    );
}

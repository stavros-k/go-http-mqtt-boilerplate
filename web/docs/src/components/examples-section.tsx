import type { ReactNode } from "react";
import { CardBoxWrapper } from "./card-box-wrapper";
import { CodeWrapper } from "./code-wrapper";
import { CollapsibleCard } from "./collapsible-group";

type Props = {
    examples?: Record<string, string>;
    title?: string | false;
    noExamplesMessage?: ReactNode;
};

export function ExamplesSection({ examples, title = "Examples", noExamplesMessage }: Props) {
    if (!examples || Object.keys(examples).length === 0) {
        return noExamplesMessage ?? null;
    }

    const exampleEntries = Object.entries(examples);

    const content = (
        <div className='space-y-3'>
            {exampleEntries.map(([exampleKey, exampleValue]) => (
                <CollapsibleCard
                    key={exampleKey}
                    title={exampleKey}
                    defaultOpen={exampleEntries.length === 1}>
                    <CodeWrapper
                        label={{ text: "Example" }}
                        code={exampleValue}
                        lang='json'
                    />
                </CollapsibleCard>
            ))}
        </div>
    );

    return title !== false ? <CardBoxWrapper title={title}>{content}</CardBoxWrapper> : content;
}

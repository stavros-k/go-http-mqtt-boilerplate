import type { Route } from "next";
import { TbBrandGolang, TbBrandTypescript, TbFileCode, TbInfoCircle, TbJson } from "react-icons/tb";
import { BackButton } from "@/components/back-button";
import { Breadcrumbs } from "@/components/breadcrumbs";
import { CodeWrapper } from "@/components/code-wrapper";
import { Deprecation } from "@/components/deprecation";
import { TabbedCardWrapper } from "@/components/tabbed-card-wrapper-client";
import { TypeKindBadge } from "@/components/type-kind-badge";
import { TypeMetadata } from "@/components/type-metadata";
import { UsageBadges } from "@/components/usage-badges";
import { docs, type TypeKeys } from "@/data/api";

export function generateStaticParams() {
    return Object.keys(docs.types).map((type) => ({
        type: type,
    }));
}

export async function generateMetadata(props: PageProps<"/api/type/[type]">) {
    const params = await props.params;
    const { type } = params as { type: TypeKeys };
    return {
        title: `Type - [${type}]`,
    };
}

export default async function Type(props: PageProps<"/api/type/[type]">) {
    const params = await props.params;
    const { type } = params as { type: TypeKeys };
    const data = docs.types[type];

    return (
        <div className='flex-1 overflow-y-auto p-10'>
            <Breadcrumbs items={[{ label: "Types", href: "/api/types" as Route }, { label: type }]} />

            <BackButton
                href='/api/types'
                label='Types'
            />

            <div>
                <div className='mb-3 flex items-center justify-between gap-3'>
                    <h1 className='font-bold text-4xl text-text-primary'>{type}</h1>
                    <div className='flex items-center gap-3'>
                        <UsageBadges usedBy={data.usedBy || undefined} />
                        <TypeKindBadge kind={data.kind} />
                    </div>
                </div>

                <Deprecation
                    deprecated={data.deprecated}
                    itemType='type'
                />

                <div className='mb-8 border-border-primary border-b-2 pb-6 text-text-tertiary'>
                    <p>{data.description}</p>
                </div>
            </div>

            <TabbedCardWrapper
                tabs={[
                    {
                        title: "Overview",
                        icon: <TbInfoCircle className='h-8 w-8 text-lang-overview' />,
                        code: (
                            <TypeMetadata
                                data={data}
                                typeName={type}
                            />
                        ),
                    },
                    {
                        title: "TypeScript",
                        icon: <TbBrandTypescript className='h-8 w-8 text-lang-ts' />,
                        code: data.representations?.ts ? (
                            <CodeWrapper
                                code={data.representations.ts}
                                lang='typescript'
                                label={{ text: type }}
                            />
                        ) : (
                            <p className='p-4 text-sm text-text-tertiary'>
                                No TypeScript representation available for this type.
                            </p>
                        ),
                    },
                    {
                        title: "Go",
                        icon: <TbBrandGolang className='h-8 w-8 text-lang-go' />,
                        code: data.representations?.go ? (
                            <CodeWrapper
                                code={data.representations.go}
                                lang='go'
                                label={{ text: type }}
                            />
                        ) : (
                            <p className='p-4 text-sm text-text-tertiary'>
                                No Go representation available for this type.
                            </p>
                        ),
                    },
                    {
                        title: "JSON",
                        icon: <TbFileCode className='h-8 w-8 text-lang-json' />,
                        code: data.representations?.json ? (
                            <>
                                <CodeWrapper
                                    code={data.representations.json}
                                    lang='json'
                                    label={{ text: type }}
                                />
                                <div className='mb-4 rounded-lg border border-border-secondary bg-bg-tertiary p-3'>
                                    <p className='text-text-muted text-xs'>
                                        Example representation - actual values may vary
                                    </p>
                                </div>
                            </>
                        ) : (
                            <p className='p-4 text-sm text-text-tertiary'>No JSON available for this type.</p>
                        ),
                    },
                    {
                        title: "JSON Schema",
                        icon: <TbJson className='h-8 w-8 text-purple-400' />,
                        code: data.representations?.jsonSchema ? (
                            <CodeWrapper
                                code={data.representations.jsonSchema}
                                lang='json'
                                label={{ text: type }}
                            />
                        ) : (
                            <p className='p-4 text-sm text-text-tertiary'>No JSON schema available for this type.</p>
                        ),
                    },
                ]}
            />
        </div>
    );
}

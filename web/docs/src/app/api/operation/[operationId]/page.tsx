import type { Route } from "next";
import { BackButton } from "@/components/back-button";
import { Breadcrumbs } from "@/components/breadcrumbs";
import { CardBoxWrapper } from "@/components/card-box-wrapper";
import { CodeWrapper } from "@/components/code-wrapper";
import { CollapsibleCard } from "@/components/collapsible-group";
import { CollapsibleResponse } from "@/components/collapsible-response";
import { Deprecation } from "@/components/deprecation";
import { Group } from "@/components/group";
import { OperationHeader } from "@/components/operation-header";
import { getAllOperations, getTypeJson, type Response, type TypeKeys } from "@/data/api";

export function generateStaticParams() {
    const operations = getAllOperations();
    return operations.map((operation) => ({
        operationId: operation.operationID,
    }));
}

export async function generateMetadata(props: PageProps<"/api/operation/[operationId]">) {
    const params = await props.params;
    const { operationId } = params;
    return {
        title: `Operation - [${operationId}]`,
    };
}

export default async function OperationPage(props: PageProps<"/api/operation/[operationId]">) {
    const params = await props.params;
    const { operationId } = params;

    const allOperations = getAllOperations();
    const operation = allOperations.find((op) => op.operationID === operationId);

    if (!operation) {
        return <div>Operation not found</div>;
    }

    const requestJson = operation.request ? getTypeJson(operation.request.type as TypeKeys) : null;

    return (
        <div className='flex-1 overflow-y-auto p-10'>
            <Breadcrumbs
                items={[{ label: "Operations", href: "/api/operations" as Route }, { label: operation.operationID }]}
            />

            <BackButton
                href='/api/operations'
                label='Operations'
            />

            <div>
                <div className='mb-3 flex items-center justify-between gap-3'>
                    <h1 className='font-bold text-4xl text-text-primary'>{operation.operationID}</h1>
                    <Group
                        group={operation.group || ""}
                        size='md'
                    />
                </div>

                <Deprecation
                    deprecated={operation.deprecated}
                    itemType='operation'
                />

                <div className='mb-8 border-border-primary border-b-2 pb-6 text-text-tertiary'>
                    <p className='mb-2'>{operation.summary}</p>
                    {operation.description && operation.description !== operation.summary && (
                        <p className='text-sm'>{operation.description}</p>
                    )}
                </div>
            </div>

            <OperationHeader
                method={operation.method}
                path={operation.path}
                parameters={operation.parameters}
            />

            {operation.request && (
                <CardBoxWrapper title='Request Body'>
                    <CodeWrapper
                        label={{
                            text: operation.request.type,
                            href: `/api/type/${operation.request.type}` as Route,
                        }}
                        code={requestJson}
                        noCodeMessage='No request body'
                        lang='json'
                    />
                    {operation.request.description && (
                        <p className='mt-3 text-sm text-text-tertiary'>{operation.request.description}</p>
                    )}
                    {requestJson && (
                        <div className='mb-4 rounded-lg border border-border-secondary bg-bg-tertiary p-3'>
                            <p className='text-text-muted text-xs'>Example representation - actual values may vary</p>
                        </div>
                    )}
                </CardBoxWrapper>
            )}

            {operation.responses && Object.keys(operation.responses).length > 0 && (
                <CardBoxWrapper title='Responses'>
                    <div className='space-y-4'>
                        {Object.entries(operation.responses).map(([statusCode, response]) => {
                            const resp = response as Response;

                            return (
                                <CollapsibleResponse
                                    key={statusCode}
                                    statusCode={statusCode}
                                    description={resp.description}>
                                    <div className='mb-4'>
                                        <div className='mb-2 font-semibold text-sm text-text-tertiary'>
                                            Type:{" "}
                                            <a
                                                href={`/api/type/${resp.type}`}
                                                className='text-accent-blue-hover transition-colors hover:text-accent-blue-light'>
                                                {resp.type}
                                            </a>
                                        </div>
                                    </div>
                                    {resp.examples && Object.keys(resp.examples).length > 0 ? (
                                        <div className='space-y-3'>
                                            {Object.entries(resp.examples).map(([exampleKey, exampleValue]) => (
                                                <CollapsibleCard
                                                    key={exampleKey}
                                                    title={exampleKey}
                                                    defaultOpen={Object.keys(resp.examples).length === 1}>
                                                    <CodeWrapper
                                                        label={{ text: "Example" }}
                                                        code={exampleValue}
                                                        lang='json'
                                                    />
                                                </CollapsibleCard>
                                            ))}
                                        </div>
                                    ) : (
                                        <p className='rounded-lg border border-border-secondary bg-bg-tertiary p-4 text-sm text-text-tertiary'>
                                            No examples available for this response.
                                        </p>
                                    )}
                                </CollapsibleResponse>
                            );
                        })}
                    </div>
                </CardBoxWrapper>
            )}
        </div>
    );
}

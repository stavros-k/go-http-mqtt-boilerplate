import type { Route } from "next";
import { BackButton } from "@/components/back-button";
import { Breadcrumbs } from "@/components/breadcrumbs";
import { CardBoxWrapper } from "@/components/card-box-wrapper";
import { CodeWrapper } from "@/components/code-wrapper";
import { Deprecation } from "@/components/deprecation";
import { ExamplesSection } from "@/components/examples-section";
import { Group } from "@/components/group";
import { MQTTTopicHeader } from "@/components/mqtt-topic-header";
import { getAllMQTTPublications, getTypeJson, type TypeKeys } from "@/data/api";

export function generateStaticParams() {
    const publications = getAllMQTTPublications();
    if (publications.length === 0) {
        return [{ operationId: "__no-mqtt-publications__" }];
    }
    return publications.map((publication) => ({
        operationId: publication.operationID,
    }));
}

export async function generateMetadata(props: PageProps<"/api/mqtt/publication/[operationId]">) {
    const params = await props.params;
    const { operationId } = params;
    return {
        title: `MQTT Publication - [${operationId}]`,
    };
}

export default async function MQTTPublicationPage(props: PageProps<"/api/mqtt/publication/[operationId]">) {
    const params = await props.params;
    const { operationId } = params;

    const allPublications = getAllMQTTPublications();

    // Handle the case where there are no MQTT publications (e.g., cloud API)
    if (operationId === "__no-mqtt-publications__") {
        return (
            <div className='flex-1 overflow-y-auto p-10'>
                <div className='rounded-lg border border-border-secondary bg-bg-tertiary p-8 text-center'>
                    <h1 className='mb-2 font-bold text-2xl'>No MQTT Publications Available</h1>
                    <p className='text-text-secondary'>This API does not include any MQTT publication endpoints.</p>
                </div>
            </div>
        );
    }

    const publication = allPublications.find((pub) => pub.operationID === operationId);

    if (!publication) {
        throw new Error(`MQTT Publication "${operationId}" not found - generateStaticParams mismatch`);
    }

    const messageJson = publication.type ? getTypeJson(publication.type as TypeKeys) : null;

    return (
        <div className='flex-1 overflow-y-auto p-10'>
            <Breadcrumbs
                items={[
                    { label: "MQTT Publications", href: "/api/mqtt/publications" as Route },
                    { label: publication.operationID },
                ]}
            />

            <BackButton
                href='/api/mqtt/publications'
                label='MQTT Publications'
            />

            <div>
                <div className='mb-3 flex items-center justify-between gap-3'>
                    <h1 className='font-bold text-4xl text-text-primary'>{publication.operationID}</h1>
                    <Group
                        group={publication.group || ""}
                        size='md'
                    />
                </div>

                <Deprecation
                    deprecated={publication.deprecated}
                    itemType='mqtt publication'
                />

                <div className='mb-8 border-border-primary border-b-2 pb-6 text-text-tertiary'>
                    <p className='mb-2'>{publication.summary}</p>
                    {publication.description && publication.description !== publication.summary && (
                        <p className='text-sm'>{publication.description}</p>
                    )}
                </div>

                <div className='mb-6 rounded-lg border-2 border-accent-green-border bg-accent-green-bg p-4'>
                    <p className='text-accent-green-text text-sm'>
                        <strong>Note:</strong> The server publishes to this topic. Clients should subscribe to receive
                        messages from this topic.
                    </p>
                </div>
            </div>

            <MQTTTopicHeader
                topic={publication.topic}
                topicMQTT={publication.topicMQTT}
                topicParameters={publication.topicParameters}
                type='publication'
            />

            {/* MQTT Settings */}
            <CardBoxWrapper title='MQTT Settings'>
                <div className='grid grid-cols-2 gap-4'>
                    <div className='rounded-lg border border-border-secondary bg-bg-tertiary p-3'>
                        <div className='mb-1 text-text-muted text-xs'>QoS</div>
                        <div className='font-semibold text-sm text-text-primary'>{publication.qos}</div>
                    </div>
                    <div className='rounded-lg border border-border-secondary bg-bg-tertiary p-3'>
                        <div className='mb-1 text-text-muted text-xs'>Retained</div>
                        <div className='font-semibold text-sm text-text-primary'>
                            {publication.retained ? "Yes" : "No"}
                        </div>
                    </div>
                </div>
            </CardBoxWrapper>

            {/* Message Type */}
            <CardBoxWrapper title='Message Type'>
                <CodeWrapper
                    label={{
                        text: publication.type,
                        href: `/api/type/${publication.type}` as Route,
                    }}
                    code={messageJson}
                    noCodeMessage='No message type'
                    lang='json'
                />
                {messageJson && (
                    <div className='mb-4 rounded-lg border border-border-secondary bg-bg-tertiary p-3'>
                        <p className='text-text-muted text-xs'>Example representation - actual values may vary</p>
                    </div>
                )}
            </CardBoxWrapper>

            {/* Examples */}
            <ExamplesSection examples={publication.examples} />
        </div>
    );
}

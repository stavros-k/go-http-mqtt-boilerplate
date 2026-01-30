import type { Route } from "next";
import { BackButton } from "@/components/back-button";
import { Breadcrumbs } from "@/components/breadcrumbs";
import { CardBoxWrapper } from "@/components/card-box-wrapper";
import { CodeWrapper } from "@/components/code-wrapper";
import { CollapsibleCard } from "@/components/collapsible-group";
import { Deprecation } from "@/components/deprecation";
import { Group } from "@/components/group";
import { MQTTTopicHeader } from "@/components/mqtt-topic-header";
import { getAllMQTTSubscriptions, getTypeJson, type TypeKeys } from "@/data/api";

export function generateStaticParams() {
    const subscriptions = getAllMQTTSubscriptions();
    return subscriptions.map((subscription) => ({
        operationId: subscription.operationID,
    }));
}

export async function generateMetadata(props: PageProps<"/api/mqtt/subscription/[operationId]">) {
    const params = await props.params;
    const { operationId } = params;
    return {
        title: `MQTT Subscription - [${operationId}]`,
    };
}

export default async function MQTTSubscriptionPage(props: PageProps<"/api/mqtt/subscription/[operationId]">) {
    const params = await props.params;
    const { operationId } = params;

    const allSubscriptions = getAllMQTTSubscriptions();
    const subscription = allSubscriptions.find((sub) => sub.operationID === operationId);

    if (!subscription) {
        return <div>MQTT Subscription not found</div>;
    }

    const messageJson = subscription.type ? getTypeJson(subscription.type as TypeKeys) : null;

    return (
        <div className='flex-1 overflow-y-auto p-10'>
            <Breadcrumbs
                items={[
                    { label: "MQTT Subscriptions", href: "/api/mqtt/subscriptions" as Route },
                    { label: subscription.operationID },
                ]}
            />

            <BackButton
                href='/api/mqtt/subscriptions'
                label='MQTT Subscriptions'
            />

            <div>
                <div className='mb-3 flex items-center justify-between gap-3'>
                    <h1 className='font-bold text-4xl text-text-primary'>{subscription.operationID}</h1>
                    <Group
                        group={subscription.group || ""}
                        size='md'
                    />
                </div>

                <Deprecation
                    deprecated={subscription.deprecated}
                    itemType='mqtt subscription'
                />

                <div className='mb-8 border-border-primary border-b-2 pb-6 text-text-tertiary'>
                    <p className='mb-2'>{subscription.summary}</p>
                    {subscription.description && subscription.description !== subscription.summary && (
                        <p className='text-sm'>{subscription.description}</p>
                    )}
                </div>

                <div className='mb-6 rounded-lg border-2 border-accent-blue-border bg-accent-blue-bg p-4'>
                    <p className='text-accent-blue-text text-sm'>
                        <strong>Note:</strong> The server subscribes to this topic. Clients are expected to publish
                        (send) messages to this topic.
                    </p>
                </div>
            </div>

            <MQTTTopicHeader
                topic={subscription.topic}
                topicMQTT={subscription.topicMQTT}
                topicParameters={subscription.topicParameters}
                type='subscription'
            />

            {/* MQTT Settings */}
            <CardBoxWrapper title='MQTT Settings'>
                <div className='w-1/2 rounded-lg border border-border-secondary bg-bg-tertiary p-3'>
                    <div className='mb-1 text-text-muted text-xs'>QoS</div>
                    <div className='font-semibold text-sm text-text-primary'>{subscription.qos}</div>
                </div>
            </CardBoxWrapper>

            {/* Message Type */}
            <CardBoxWrapper title='Message Type'>
                <CodeWrapper
                    label={{
                        text: subscription.type,
                        href: `/api/type/${subscription.type}` as Route,
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
            {subscription.examples && Object.keys(subscription.examples).length > 0 && (
                <CardBoxWrapper title='Examples'>
                    <div className='space-y-3'>
                        {Object.entries(subscription.examples).map(([exampleKey, exampleValue]) => (
                            <CollapsibleCard
                                key={exampleKey}
                                title={exampleKey}
                                defaultOpen={Object.keys(subscription.examples).length === 1}>
                                <CodeWrapper
                                    label={{ text: "Example" }}
                                    code={exampleValue}
                                    lang='json'
                                />
                            </CollapsibleCard>
                        ))}
                    </div>
                </CardBoxWrapper>
            )}
        </div>
    );
}

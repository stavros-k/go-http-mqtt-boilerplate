import Link from "next/link";
import { TbApi, TbCode, TbDatabase, TbFileDescription } from "react-icons/tb";
import { PubBadge } from "@/components/pub-badge";
import { SubBadge } from "@/components/sub-badge";
import { VerbBadge } from "@/components/verb-badge";
import { docs, getAllMQTTPublications, getAllMQTTSubscriptions, getAllOperations } from "@/data/api";

// Calculate total number of operations
const operations = getAllOperations();
const operationCount = operations.length;

// Calculate MQTT operations
const mqttPublications = getAllMQTTPublications();
const mqttSubscriptions = getAllMQTTSubscriptions();
const mqttPublicationCount = mqttPublications.length;
const mqttSubscriptionCount = mqttSubscriptions.length;

// Count types used by HTTP operations only
const httpTypeCount = Object.values(docs.types).filter((type) => {
    if (!("usedBy" in type) || !type.usedBy) return false;
    return type.usedBy.some((usage) => ["request", "response", "parameter"].includes(usage.role));
}).length;

// Count types used by MQTT operations only
const mqttTypeCount = Object.values(docs.types).filter((type) => {
    if (!("usedBy" in type) || !type.usedBy) return false;
    return type.usedBy.some((usage) => ["mqtt_publication", "mqtt_subscription"].includes(usage.role));
}).length;

const tableCount = docs.database.tableCount || 0;

// Get unique paths for route count
const uniquePaths = new Set(operations.map((op) => op.path));
const routeCount = uniquePaths.size;

// Calculate HTTP method distribution
const httpMethods = operations.reduce(
    (acc, op) => {
        acc[op.method] = (acc[op.method] || 0) + 1;
        return acc;
    },
    {} as Record<string, number>
);

export default function Home() {
    return (
        <div className='min-h-screen flex-1 bg-bg-primary font-sans'>
            {/* Hero Section */}
            <div className='relative overflow-hidden bg-linear-to-br from-gradient-start via-gradient-mid to-gradient-end'>
                <div className='absolute inset-0 bg-black/30 dark:bg-black/40'></div>
                <div className='relative w-full px-6 py-16 sm:py-20 lg:px-8'>
                    <div className='text-center'>
                        <h1 className='mb-3 font-bold text-4xl text-white drop-shadow-lg sm:text-5xl'>
                            {docs.info.title}
                        </h1>
                        <div className='mb-4 inline-block rounded-full border-2 border-white/30 bg-white/90 px-4 py-2 shadow-lg backdrop-blur-sm dark:border-border-primary dark:bg-bg-secondary/80'>
                            <span className='font-bold text-gray-800 text-sm dark:text-text-secondary'>
                                Version {docs.info.version}
                            </span>
                        </div>
                        <p className='mx-auto max-w-2xl font-medium text-lg text-white/95 drop-shadow-md sm:text-xl'>
                            {docs.info.description}
                        </p>
                    </div>
                </div>
            </div>

            {/* HTTP Section */}
            <div className='w-full bg-bg-secondary/20 px-6 py-12 lg:px-8'>
                <h2 className='mb-10 text-center font-bold text-3xl text-text-primary'>HTTP API</h2>
                <div className='mx-auto mb-8 grid max-w-6xl grid-cols-1 gap-6 md:grid-cols-3'>
                    {/* HTTP Operations Card */}
                    <Link
                        href='/api/operations'
                        className='block'>
                        <div className='h-full cursor-pointer rounded-2xl border-2 border-border-primary bg-bg-secondary p-6 shadow-lg transition-all duration-300 hover:scale-105 hover:border-accent-blue hover:shadow-2xl'>
                            <div className='mb-4 flex items-center gap-3'>
                                <div className='flex h-12 w-12 items-center justify-center rounded-xl bg-accent-blue/20 shadow-md'>
                                    <TbApi className='h-8 w-8 text-accent-blue' />
                                </div>
                                <div className='font-bold text-lg text-text-primary'>Operations</div>
                            </div>
                            <div className='mb-2 font-bold text-4xl text-accent-blue'>{operationCount}</div>
                            <p className='text-text-secondary text-xs'>
                                Across {routeCount} route{routeCount !== 1 ? "s" : ""}
                            </p>
                        </div>
                    </Link>

                    {/* HTTP Types Card */}
                    <Link
                        href='/api/types'
                        className='block'>
                        <div className='h-full cursor-pointer rounded-2xl border-2 border-border-primary bg-bg-secondary p-6 shadow-lg transition-all duration-300 hover:scale-105 hover:border-success-border hover:shadow-2xl'>
                            <div className='mb-4 flex items-center gap-3'>
                                <div className='flex h-12 w-12 items-center justify-center rounded-xl bg-success-bg shadow-md'>
                                    <TbCode className='h-8 w-8 text-success-text' />
                                </div>
                                <div className='font-bold text-lg text-text-primary'>Types</div>
                            </div>
                            <div className='mb-2 font-bold text-4xl text-success-text'>{httpTypeCount}</div>
                            <p className='text-text-secondary text-xs'>HTTP types</p>
                        </div>
                    </Link>

                    {/* OpenAPI Specification Card */}
                    <Link
                        href='/api/openapi'
                        className='block'>
                        <div className='h-full cursor-pointer rounded-2xl border-2 border-border-primary bg-bg-secondary p-6 shadow-lg transition-all duration-300 hover:scale-105 hover:border-warning-border hover:shadow-2xl'>
                            <div className='mb-4 flex items-center gap-3'>
                                <div className='flex h-12 w-12 items-center justify-center rounded-xl bg-warning-bg shadow-md'>
                                    <TbFileDescription className='h-8 w-8 text-warning-text' />
                                </div>
                                <div className='font-bold text-lg text-text-primary'>OpenAPI</div>
                            </div>
                            <div className='mb-2 font-bold text-4xl text-warning-text'>3.0</div>
                            <p className='text-text-secondary text-xs'>Specification</p>
                        </div>
                    </Link>
                </div>

                {/* HTTP Methods */}
                <div className='mx-auto max-w-6xl'>
                    <h3 className='mb-6 text-center font-bold text-text-primary text-xl'>HTTP Methods</h3>
                    <div className='flex flex-wrap justify-center gap-4'>
                        {Object.entries(httpMethods)
                            .sort(([, a], [, b]) => b - a)
                            .map(([method, count]) => (
                                <div
                                    key={method}
                                    className='w-32 rounded-xl border-2 border-border-primary bg-bg-secondary p-4 shadow-lg transition-all duration-300 hover:scale-105 hover:shadow-xl'>
                                    <div className='text-center'>
                                        <VerbBadge
                                            verb={method}
                                            size='sm'
                                        />
                                        <div className='mt-3 font-bold text-2xl text-text-primary'>{count}</div>
                                    </div>
                                </div>
                            ))}
                    </div>
                </div>
            </div>

            {/* MQTT Section */}
            <div className='w-full px-6 py-12 lg:px-8'>
                <h2 className='mb-10 text-center font-bold text-3xl text-text-primary'>MQTT API</h2>
                <div className='mx-auto grid max-w-6xl grid-cols-1 gap-6 md:grid-cols-3'>
                    {/* MQTT Publications Card */}
                    <Link
                        href='/api/mqtt/publications'
                        className='block'>
                        <div className='h-full cursor-pointer rounded-2xl border-2 border-border-primary bg-bg-secondary p-6 shadow-lg transition-all duration-300 hover:scale-105 hover:border-accent-blue hover:shadow-2xl'>
                            <div className='mb-4 flex items-center gap-3'>
                                <div className='flex h-12 w-12 items-center justify-center rounded-xl bg-accent-blue/20 shadow-md'>
                                    <PubBadge border={false} />
                                </div>
                                <div className='font-bold text-lg text-text-primary'>Publications</div>
                            </div>
                            <div className='mb-2 font-bold text-4xl text-accent-blue'>{mqttPublicationCount}</div>
                            <p className='text-text-secondary text-xs'>Server publishes</p>
                        </div>
                    </Link>

                    {/* MQTT Subscriptions Card */}
                    <Link
                        href='/api/mqtt/subscriptions'
                        className='block'>
                        <div className='h-full cursor-pointer rounded-2xl border-2 border-border-primary bg-bg-secondary p-6 shadow-lg transition-all duration-300 hover:scale-105 hover:border-accent-green hover:shadow-2xl'>
                            <div className='mb-4 flex items-center gap-3'>
                                <div className='flex h-12 w-12 items-center justify-center rounded-xl bg-accent-green/20 shadow-md'>
                                    <SubBadge border={false} />
                                </div>
                                <div className='font-bold text-lg text-text-primary'>Subscriptions</div>
                            </div>
                            <div className='mb-2 font-bold text-4xl text-accent-green'>{mqttSubscriptionCount}</div>
                            <p className='text-text-secondary text-xs'>Server subscribes</p>
                        </div>
                    </Link>

                    {/* MQTT Types Card */}
                    <Link
                        href='/api/types'
                        className='block'>
                        <div className='h-full cursor-pointer rounded-2xl border-2 border-border-primary bg-bg-secondary p-6 shadow-lg transition-all duration-300 hover:scale-105 hover:border-accent-purple hover:shadow-2xl'>
                            <div className='mb-4 flex items-center gap-3'>
                                <div className='flex h-12 w-12 items-center justify-center rounded-xl bg-accent-purple/20 shadow-md'>
                                    <TbCode className='h-8 w-8 text-accent-purple' />
                                </div>
                                <div className='font-bold text-lg text-text-primary'>Types</div>
                            </div>
                            <div className='mb-2 font-bold text-4xl text-accent-purple'>{mqttTypeCount}</div>
                            <p className='text-text-secondary text-xs'>MQTT types</p>
                        </div>
                    </Link>
                </div>
            </div>

            {/* Other Resources Section */}
            <div className='w-full bg-bg-secondary/20 px-6 py-12 lg:px-8'>
                <h2 className='mb-10 text-center font-bold text-3xl text-text-primary'>Resources</h2>
                <div className='mx-auto grid max-w-4xl grid-cols-1 gap-6 md:grid-cols-2'>
                    {/* Types Card */}
                    <Link
                        href='/api/types'
                        className='block'>
                        <div className='h-full cursor-pointer rounded-2xl border-2 border-border-primary bg-bg-secondary p-6 shadow-lg transition-all duration-300 hover:scale-105 hover:border-success-border hover:shadow-2xl'>
                            <div className='mb-4 flex items-center gap-3'>
                                <div className='flex h-12 w-12 items-center justify-center rounded-xl bg-success-bg shadow-md'>
                                    <TbCode className='h-8 w-8 text-success-text' />
                                </div>
                                <div className='font-bold text-lg text-text-primary'>All Types</div>
                            </div>
                            <div className='mb-2 font-bold text-4xl text-success-text'>
                                {Object.keys(docs.types).length}
                            </div>
                            <p className='text-text-secondary text-xs'>Type definitions</p>
                        </div>
                    </Link>

                    {/* Database Schema Card */}
                    <Link
                        href='/api/database/schema'
                        className='block'>
                        <div className='h-full cursor-pointer rounded-2xl border-2 border-border-primary bg-bg-secondary p-6 shadow-lg transition-all duration-300 hover:scale-105 hover:border-info-border hover:shadow-2xl'>
                            <div className='mb-4 flex items-center gap-3'>
                                <div className='flex h-12 w-12 items-center justify-center rounded-xl bg-info-bg shadow-md'>
                                    <TbDatabase className='h-8 w-8 text-info-text' />
                                </div>
                                <div className='font-bold text-lg text-text-primary'>Database</div>
                            </div>
                            <div className='mb-2 font-bold text-4xl text-info-text'>{tableCount}</div>
                            <p className='text-text-secondary text-xs'>Table{tableCount !== 1 ? "s" : ""}</p>
                        </div>
                    </Link>
                </div>
            </div>
        </div>
    );
}

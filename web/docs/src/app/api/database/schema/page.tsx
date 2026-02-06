import { Breadcrumbs } from "@/components/breadcrumbs";
import { CardBoxWrapper } from "@/components/card-box-wrapper";
import { CodeWrapper } from "@/components/code-wrapper";
import { docs } from "@/data/api";
export async function generateMetadata() {
    return {
        title: "Database Schema",
    };
}

export default function DatabaseSchema() {
    const tables = docs.database.stats?.tables || [];

    return (
        <div className='flex-1 overflow-y-auto p-10'>
            <Breadcrumbs items={[{ label: "Database Schema" }]} />

            <div>
                <h1 className='mb-3 font-bold text-4xl text-text-primary'>Database Schema</h1>
                <h2 className='mb-4 text-text-primary text-xl'>Database structure definition</h2>

                <div className='mb-8 border-border-primary border-b-2 pb-6 text-text-tertiary'>
                    <p>This page displays the database schema used by the application.</p>
                </div>
            </div>

            {tables && (
                <div className='mb-8 space-y-8'>
                    <h2 className='font-bold text-2xl text-text-primary'>Tables</h2>
                    {tables.map((table) => (
                        <CardBoxWrapper
                            key={table.name}
                            title={table.name}>
                            <div className='space-y-6'>
                                {/* Columns Section */}
                                <div>
                                    <h3 className='mb-3 font-semibold text-lg text-text-secondary'>Columns</h3>
                                    <div className='overflow-x-auto'>
                                        <table className='w-full border-separate border-spacing-0'>
                                            <thead>
                                                <tr className='bg-bg-secondary'>
                                                    <th className='border-border-primary border-b-2 px-4 py-3 text-left font-semibold text-sm text-text-secondary'>
                                                        Name
                                                    </th>
                                                    <th className='border-border-primary border-b-2 px-4 py-3 text-left font-semibold text-sm text-text-secondary'>
                                                        Type
                                                    </th>
                                                    <th className='border-border-primary border-b-2 px-4 py-3 text-left font-semibold text-sm text-text-secondary'>
                                                        Nullable
                                                    </th>
                                                    <th className='border-border-primary border-b-2 px-4 py-3 text-left font-semibold text-sm text-text-secondary'>
                                                        Primary Key
                                                    </th>
                                                    <th className='border-border-primary border-b-2 px-4 py-3 text-left font-semibold text-sm text-text-secondary'>
                                                        Default
                                                    </th>
                                                </tr>
                                            </thead>
                                            <tbody>
                                                {table.columns?.map((column) => (
                                                    <tr
                                                        key={column.name}
                                                        className='hover:bg-bg-secondary/50'>
                                                        <td className='border-border-primary border-b px-4 py-3 font-mono text-sm text-text-primary'>
                                                            {column.name}
                                                        </td>
                                                        <td className='border-border-primary border-b px-4 py-3 font-mono text-sm text-text-secondary'>
                                                            {column.type}
                                                        </td>
                                                        <td className='border-border-primary border-b px-4 py-3 text-sm'>
                                                            {column.not_null ? (
                                                                <span className='rounded-full bg-red-500/20 px-2 py-1 font-medium text-red-400 text-xs'>
                                                                    NO
                                                                </span>
                                                            ) : (
                                                                <span className='rounded-full bg-green-500/20 px-2 py-1 font-medium text-green-400 text-xs'>
                                                                    YES
                                                                </span>
                                                            )}
                                                        </td>
                                                        <td className='border-border-primary border-b px-4 py-3 text-sm'>
                                                            {column.primary_key ? (
                                                                <span className='rounded-full bg-accent-blue/20 px-2 py-1 font-medium text-accent-blue text-xs'>
                                                                    YES
                                                                </span>
                                                            ) : (
                                                                <span className='text-text-tertiary'>-</span>
                                                            )}
                                                        </td>
                                                        <td className='border-border-primary border-b px-4 py-3 font-mono text-sm text-text-tertiary'>
                                                            {"default" in column && column.default
                                                                ? column.default
                                                                : "-"}
                                                        </td>
                                                    </tr>
                                                ))}
                                            </tbody>
                                        </table>
                                    </div>
                                </div>

                                {/* Foreign Keys Section */}
                                {table.foreign_keys && (
                                    <div>
                                        <h3 className='mb-3 font-semibold text-lg text-text-secondary'>Foreign Keys</h3>
                                        <div className='overflow-x-auto'>
                                            <table className='w-full border-separate border-spacing-0'>
                                                <thead>
                                                    <tr className='bg-bg-secondary'>
                                                        <th className='border-border-primary border-b-2 px-4 py-3 text-left font-semibold text-sm text-text-secondary'>
                                                            Column
                                                        </th>
                                                        <th className='border-border-primary border-b-2 px-4 py-3 text-left font-semibold text-sm text-text-secondary'>
                                                            References
                                                        </th>
                                                    </tr>
                                                </thead>
                                                <tbody>
                                                    {table.foreign_keys.map((fk, idx) => (
                                                        <tr
                                                            key={`${fk.table}.${fk.to}-${idx}`}
                                                            className='hover:bg-bg-secondary/50'>
                                                            <td className='border-border-primary border-b px-4 py-3 font-mono text-sm text-text-primary'>
                                                                {fk.from}
                                                            </td>
                                                            <td className='border-border-primary border-b px-4 py-3 font-mono text-sm text-text-secondary'>
                                                                {fk.table}.{fk.to}
                                                            </td>
                                                        </tr>
                                                    ))}
                                                </tbody>
                                            </table>
                                        </div>
                                    </div>
                                )}

                                {/* Indexes Section */}
                                {table.indexes && (
                                    <div>
                                        <h3 className='mb-3 font-semibold text-lg text-text-secondary'>Indexes</h3>
                                        <div className='overflow-x-auto'>
                                            <table className='w-full border-separate border-spacing-0'>
                                                <thead>
                                                    <tr className='bg-bg-secondary'>
                                                        <th className='border-border-primary border-b-2 px-4 py-3 text-left font-semibold text-sm text-text-secondary'>
                                                            Name
                                                        </th>
                                                        <th className='border-border-primary border-b-2 px-4 py-3 text-left font-semibold text-sm text-text-secondary'>
                                                            Columns
                                                        </th>
                                                        <th className='border-border-primary border-b-2 px-4 py-3 text-left font-semibold text-sm text-text-secondary'>
                                                            Unique
                                                        </th>
                                                    </tr>
                                                </thead>
                                                <tbody>
                                                    {table.indexes.map((index) => (
                                                        <tr
                                                            key={index.name}
                                                            className='hover:bg-bg-secondary/50'>
                                                            <td className='border-border-primary border-b px-4 py-3 font-mono text-sm text-text-primary'>
                                                                {index.name}
                                                            </td>
                                                            <td className='border-border-primary border-b px-4 py-3 font-mono text-sm text-text-secondary'>
                                                                {index.columns?.join(", ") || "-"}
                                                            </td>
                                                            <td className='border-border-primary border-b px-4 py-3 text-sm'>
                                                                {index.unique ? (
                                                                    <span className='rounded-full bg-accent-blue/20 px-2 py-1 font-medium text-accent-blue text-xs'>
                                                                        YES
                                                                    </span>
                                                                ) : (
                                                                    <span className='rounded-full bg-gray-500/20 px-2 py-1 font-medium text-gray-400 text-xs'>
                                                                        NO
                                                                    </span>
                                                                )}
                                                            </td>
                                                        </tr>
                                                    ))}
                                                </tbody>
                                            </table>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </CardBoxWrapper>
                    ))}
                </div>
            )}

            <div className='mt-8'>
                <CardBoxWrapper title='Raw Schema'>
                    <CodeWrapper
                        code={docs.database.schema}
                        label={{ text: "schema.sql" }}
                        lang='sql'
                    />
                </CardBoxWrapper>
            </div>
        </div>
    );
}

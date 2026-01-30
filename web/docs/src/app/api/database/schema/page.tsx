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
    const tableCount = docs.database.tableCount || 0;

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

            <div className='mb-8'>
                <div className='inline-block rounded-xl border-2 border-border-primary bg-bg-secondary p-6'>
                    <div className='flex items-center gap-4'>
                        <div className='text-center'>
                            <div className='mb-1 font-bold text-4xl text-accent-blue'>{tableCount}</div>
                            <div className='font-semibold text-sm text-text-secondary'>
                                {tableCount === 1 ? "Table" : "Tables"}
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <CardBoxWrapper title='Schema'>
                <CodeWrapper
                    code={docs.database.schema}
                    label={{ text: "schema.sql" }}
                    lang='sql'
                />
            </CardBoxWrapper>
        </div>
    );
}

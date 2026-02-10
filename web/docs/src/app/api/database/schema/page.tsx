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
    return (
        <div className='flex-1 overflow-y-auto p-10'>
            <Breadcrumbs items={[{ label: "Database Schema" }]} />

            <div>
                <h1 className='mb-3 font-bold text-4xl text-text-primary'>Database Schema</h1>
                <div className='mb-4 flex items-center gap-3'>
                    <h2 className='text-text-primary text-xl'>SQL Schema Definition</h2>
                    <span className='rounded-lg bg-info-bg px-3 py-1.5 font-mono font-semibold text-info-text text-sm'>
                        {docs.database.dialect}
                    </span>
                </div>

                <div className='mb-8 border-border-primary border-b-2 pb-6 text-text-tertiary'>
                    <p>This page displays the raw SQL schema used by the application.</p>
                </div>
            </div>

            <div className='mt-8'>
                <CardBoxWrapper title='Schema'>
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

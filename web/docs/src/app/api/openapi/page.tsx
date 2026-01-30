import { Breadcrumbs } from "@/components/breadcrumbs";
import { CardBoxWrapper } from "@/components/card-box-wrapper";
import { CodeWrapper } from "@/components/code-wrapper";
import { docs } from "@/data/api";

export async function generateMetadata() {
    return {
        title: "OpenAPI Specification",
    };
}

export default function OpenAPISpec() {
    const hasSpec = docs.openapiSpec && docs.openapiSpec.trim() !== "";

    return (
        <div className='flex-1 overflow-y-auto p-10'>
            <Breadcrumbs items={[{ label: "OpenAPI Specification" }]} />

            <div>
                <h1 className='mb-3 font-bold text-4xl text-text-primary'>OpenAPI Specification</h1>
                <h2 className='mb-4 text-text-primary text-xl'>Complete API specification in OpenAPI 3.0 format</h2>

                <div className='mb-8 border-border-primary border-b-2 pb-6 text-text-tertiary'>
                    <p>
                        This page displays the complete OpenAPI 3.0 specification for the API. You can use this
                        specification with tools like Swagger UI, Postman, or other OpenAPI-compatible clients.
                    </p>
                </div>
            </div>

            {hasSpec ? (
                <CardBoxWrapper title='OpenAPI 3.0 YAML'>
                    <CodeWrapper
                        code={docs.openapiSpec}
                        label={{ text: "openapi.yaml" }}
                        lang='yaml'
                    />
                </CardBoxWrapper>
            ) : (
                <div className='rounded-xl border-2 border-border-primary bg-bg-secondary p-6 text-text-secondary'>
                    No OpenAPI specification available.
                </div>
            )}
        </div>
    );
}

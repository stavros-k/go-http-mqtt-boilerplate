type Props = {
    deprecated: string;
    itemType: "type" | "operation" | "mqtt publication" | "mqtt subscription";
};

function getItemLabel(itemType: Props["itemType"]) {
    switch (itemType) {
        case "type":
            return "type";
        case "operation":
            return "operation";
        case "mqtt publication":
            return "MQTT publication";
        case "mqtt subscription":
            return "MQTT subscription";
    }
}
export const Deprecation = ({ deprecated, itemType = "type" }: Props) => {
    if (!deprecated) return null;

    return (
        <div className='mb-6 rounded-lg border-2 border-warning-border bg-warning-bg px-4 py-3 text-warning-text'>
            <div className='flex items-start gap-3'>
                <span className='text-xl'>⚠️</span>
                <div>
                    <p className='mb-1 font-bold'>This {getItemLabel(itemType)} is deprecated</p>
                    <p className='text-sm'>{deprecated}</p>
                </div>
            </div>
        </div>
    );
};

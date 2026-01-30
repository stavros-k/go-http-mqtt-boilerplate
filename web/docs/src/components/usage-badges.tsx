import { TbApi } from "react-icons/tb";
import type { UsedByItem } from "@/data/api";

interface UsageBadgesProps {
    usedBy: UsedByItem[] | undefined;
}

export function UsageBadges({ usedBy }: UsageBadgesProps) {
    if (!usedBy || usedBy.length === 0) {
        return null;
    }

    const hasHTTP = usedBy.some((item) => ["request", "response", "parameter"].includes(item.role));
    const hasMQTT = usedBy.some((item) => item.role === "mqtt_publication" || item.role === "mqtt_subscription");

    if (!hasHTTP && !hasMQTT) {
        return null;
    }

    return (
        <div className='flex items-center gap-2'>
            {hasHTTP && (
                <div className='flex items-center gap-1.5 rounded-lg border-2 border-accent-green-border bg-accent-green-bg px-3 py-1.5 font-semibold text-accent-green-text text-sm'>
                    <TbApi className='h-4 w-4' />
                    <span>HTTP</span>
                </div>
            )}
            {hasMQTT && (
                <div className='flex items-center gap-1.5 rounded-lg border-2 border-accent-purple-border bg-accent-purple-bg px-3 py-1.5 font-semibold text-accent-purple-text text-sm'>
                    <span>MQTT</span>
                </div>
            )}
        </div>
    );
}
